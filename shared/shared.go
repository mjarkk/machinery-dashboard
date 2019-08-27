package shared

import (
	"time"

	"go.mongodb.org/mongo-driver/bson/primitive"
)

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
	Time    int64  `bson:"time"`
	Success bool   `bson:"success"` // True means no error
	Error   string `bson:"error"`   // If there is an error you can few it here
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
func (e *DBEntry) NewEntry(err error) {
	newPoint := DataPoint{
		Success: err == nil,
		Time:    time.Now().Unix(),
	}
	if err != nil {
		newPoint.Error = err.Error()
	}
	e.Points = append(e.Points, newPoint)
}
