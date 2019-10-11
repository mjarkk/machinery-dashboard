package plugin

import (
	"context"
	"errors"
	"time"

	"github.com/RichardKnop/machinery/v1"
	"github.com/RichardKnop/machinery/v1/tasks"
	"github.com/globalsign/mgo/bson"
	"github.com/jasonlvhit/gocron"
	log "github.com/sirupsen/logrus"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
	"go.mongodb.org/mongo-driver/mongo/readpref"
)

const (
	taskErr     = "Machinery_dashboard_Error"
	taskSuccess = "Machinery_dashboard_Success"
)

// TaskStatus contains the status of a job
type TaskStatus uint8

const (
	StatusOke TaskStatus = iota + 1
	StatusRetry
	StatusError
)

// IsOke tells if the task status is oke
func (s TaskStatus) IsOke() bool {
	return s == StatusOke
}

// IsRetry tells if the task status is retry
func (s TaskStatus) IsRetry() bool {
	return s == StatusRetry
}

// IsError tells if the task status is error
func (s TaskStatus) IsError() bool {
	return s == StatusError
}

// Options are the options for the Init function
// And contains all settings for the frontend
// Currently we only support mongodb as database
type Options struct {
	Mongodb MongoDBConnectOptions `json:"mongodb"`
}

// MongoDBConnectOptions contains the settings for connecting a mongodb server
type MongoDBConnectOptions struct {
	ConnectionURI string `json:"connectionURL"`
	Database      string `json:"database"`
}

// DBEntry is one entry in the
type DBEntry struct {
	ID     primitive.ObjectID `bson:"_id"`
	Queue  string             `bson:"queue"`
	Points []DataPoint        `bson:"points"`
}

// DataPoint tells if an entry is successfull or not
type DataPoint struct {
	From      int64    `bson:"from" json:"from"`
	Successes int      `bson:"successes" json:"successes"`
	Retries   int      `bson:"retries" json:"retries"`
	Errors    []string `bson:"errors" json:"errors"` // Maybe for later
}

// NewEntry creates a new DBEntry
func NewEntry(queue string) DBEntry {
	return DBEntry{
		ID:     primitive.NewObjectID(),
		Queue:  queue,
		Points: []DataPoint{},
	}
}

// NewEntry ads an entry to the (*DBEntry).Points
func (e *DBEntry) NewEntry(status TaskStatus, ifErrorMsg error) {
	now := time.Now()

	for i, point := range e.Points {
		pointTime := time.Unix(point.From, 0)
		if !pointTime.Add(time.Minute * 30).After(now) {
			continue
		}
		switch status {
		case StatusOke:
			point.Successes++
		case StatusError:
			point.Errors = append(point.Errors, ifErrorMsg.Error())
		case StatusRetry:
			point.Retries++
		}
		e.Points[i] = point
		return
	}

	newPoint := DataPoint{
		From:   now.Unix(),
		Errors: []string{},
	}
	switch status {
	case StatusOke:
		newPoint.Successes++
	case StatusError:
		newPoint.Errors = []string{ifErrorMsg.Error()}
	case StatusRetry:
		newPoint.Retries++
	}

	e.Points = append(e.Points, newPoint)
}

func c() context.Context {
	return context.Background()
}

type plugin struct {
	worker  *machinery.Worker
	mongodb pluginMongo
}

type pluginMongo struct {
	collection *mongo.Collection
}

// Init adds the event listeners to the machinery worker
func Init(worker *machinery.Worker, initOptions Options) error {
	client, err := mongo.NewClient(options.Client().ApplyURI(initOptions.Mongodb.ConnectionURI))
	if err != nil {
		return err
	}

	err = client.Connect(c())
	if err != nil {
		return err
	}

	err = client.Ping(c(), readpref.Primary())
	if err != nil {
		return err
	}

	collection := client.Database(initOptions.Mongodb.Database).Collection("machinery-stats")

	// Check if this worker is already in the database, and if not so add it
	if collection.FindOne(c(), bson.M{"queue": worker.Queue}).Err() != nil {
		_, err := collection.InsertOne(c(), NewEntry(worker.Queue))
		if err != nil {
			return err
		}
	}

	p := &plugin{
		worker: worker,
		mongodb: pluginMongo{
			collection: collection,
		},
	}

	worker.SetPreTaskHandler(p.PreTaskHandler)

	gocron.Every(1).Hour().Do(p.Cleanup)
	go func() {
		time.Sleep(time.Minute * 2)
		p.Cleanup()
	}()
	p.listenForJobs()

	return nil
}

func (p *plugin) PreTaskHandler(task *tasks.Signature) {
	if task.Name == taskSuccess {
		return
	}
	if task.Name == taskErr {
		return
	}

	for _, tasks := range [][]*tasks.Signature{task.OnSuccess, task.OnError} {
		if tasks != nil && len(tasks) > 0 {
			for _, task := range tasks {
				if task != nil && (task.Name == taskSuccess || task.Name == taskErr) {
					// This task is already registered thisone can be seen as fails
					p.AddPoint(StatusRetry, nil)
					return
				}
			}
		}
	}

	if task.OnError == nil {
		task.OnError = []*tasks.Signature{}
	}
	if task.OnSuccess == nil {
		task.OnSuccess = []*tasks.Signature{}
	}

	args := []tasks.Arg{{
		Type:  "string",
		Value: task.Name,
	}}

	errTask, _ := tasks.NewSignature(taskErr, args)
	errTask.RoutingKey = p.worker.Queue
	task.OnError = append(task.OnError, errTask)

	successTask, _ := tasks.NewSignature(taskSuccess, args)
	successTask.RoutingKey = p.worker.Queue
	task.OnSuccess = append(task.OnSuccess, successTask)
}

func (p *plugin) listenForJobs() {
	server := p.worker.GetServer()
	server.RegisterTasks(map[string]interface{}{
		taskErr: func(err, jobName string) error {
			p.AddPoint(StatusError, errors.New(err))
			return nil
		},
		taskSuccess: func(jobName string) error {
			p.AddPoint(StatusOke, nil)
			return nil
		},
	})
}

func logError(err error) {
	if err != nil {
		log.Infof("Failed update database entry: %s", err.Error())
	}
}

// AddPoint adds a data point to the database
func (p *plugin) AddPoint(status TaskStatus, ifErrMsg error) error {
	query := bson.M{"queue": p.worker.Queue}

	var data DBEntry
	err := p.mongodb.collection.FindOne(c(), query).Decode(&data)
	if err != nil {
		return err
	}
	data.NewEntry(status, ifErrMsg)

	_, err = p.mongodb.collection.UpdateOne(c(), query, bson.M{"$set": bson.M{"points": data.Points}})
	return err
}

// Cleanup cleans up the database
func (p *plugin) Cleanup() {
	query := bson.M{"queue": p.worker.Queue}

	var data DBEntry
	err := p.mongodb.collection.FindOne(c(), query).Decode(&data)
	if err != nil {
		return
	}

	removeLaterThen := time.Now().Add(-(time.Hour * 24 * 3))

	removedSomething := false
	for _, point := range data.Points {
		pointTime := time.Unix(point.From, 0)
		if pointTime.After(removeLaterThen) {
			break
		}
		removedSomething = true
		data.Points = data.Points[1:]
	}

	if !removedSomething {
		return
	}

	p.mongodb.collection.UpdateOne(c(), query, bson.M{"$set": bson.M{"points": data.Points}})
}
