package main

import (
	"context"

	"github.com/mjarkk/machinery-dashboard/plugin"

	"go.mongodb.org/mongo-driver/bson/primitive"
)

func c() context.Context {
	return context.Background()
}

type apiOutput struct {
	ID       primitive.ObjectID `json:"_id"`
	Queue    string             `json:"queue"`
	Timeline []plugin.DataPoint `json:"timeline"`
}
