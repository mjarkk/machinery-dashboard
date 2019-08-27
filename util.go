package main

import (
	"context"

	"github.com/mjarkk/machinery-dashboard/shared"
	"go.mongodb.org/mongo-driver/bson/primitive"
)

func c() context.Context {
	return context.Background()
}

type apiOutput struct {
	ID       primitive.ObjectID `json:"_id"`
	Queue    string             `json:"queue"`
	Timeline []timelineEntry    `json:"timeline"`
}

type timelineEntry struct {
	Points []shared.DataPoint `json:"timelineEntry"`
	From   int64              `json:"from"`
}
