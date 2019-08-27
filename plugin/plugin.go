package plugin

import (
	"context"

	"github.com/RichardKnop/machinery/v1"
	"github.com/RichardKnop/machinery/v1/tasks"
	"github.com/globalsign/mgo/bson"
	"github.com/mjarkk/machinery-dashboard/shared"
	log "github.com/sirupsen/logrus"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
	"go.mongodb.org/mongo-driver/mongo/readpref"
)

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
func Init(worker *machinery.Worker, initOptions shared.Options) error {
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
		_, err := collection.InsertOne(c(), shared.NewEntry(worker.Queue))
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
	worker.SetErrorHandler(p.ErrorHandeler)
	worker.SetPostTaskHandler(p.PostTaskHandler)

	return nil
}

func (p *plugin) ErrorHandeler(err error) {
	logError(p.AddPoint(err))
}

func (p *plugin) PostTaskHandler(task *tasks.Signature) {
	logError(p.AddPoint(nil))
}

func logError(err error) {
	if err != nil {
		log.Infof("Failed update database entry: %s", err.Error())
	}
}

// AddPoint adds a data point to the database
func (p *plugin) AddPoint(point error) error {
	query := bson.M{"queue": p.worker.Queue}

	var data shared.DBEntry
	err := p.mongodb.collection.FindOne(c(), query).Decode(&data)
	if err != nil {
		return err
	}
	data.NewEntry(point)

	_, err = p.mongodb.collection.UpdateOne(c(), query, bson.M{"$set": bson.M{"points": data.Points}})
	return err
}
