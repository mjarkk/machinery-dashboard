package main

import (
	"fmt"
	"os"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/mjarkk/machinery-dashboard/shared"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
	"go.mongodb.org/mongo-driver/mongo/readpref"
)

func resErr(c *gin.Context, err error) bool {
	if err == nil {
		return false
	}

	c.JSON(400, gin.H{
		"status": false,
		"error":  err.Error(),
	})
	return true
}

func res(c *gin.Context, data interface{}) {
	c.JSON(200, gin.H{
		"status": true,
		"data":   data,
	})
}

func main() {
	config, err := getConfig()
	if err != nil {
		fmt.Println("Error while reading config:", err)
		os.Exit(1)
	}

	client, err := mongo.NewClient(options.Client().ApplyURI(config.Mongodb.ConnectionURI))
	if err != nil {
		fmt.Println("Error while connecting to database:", err)
		os.Exit(1)
	}

	err = client.Connect(c())
	if err != nil {
		fmt.Println("Error while connecting to database:", err)
		os.Exit(1)
	}

	err = client.Ping(c(), readpref.Primary())
	if err != nil {
		fmt.Println("Error while connecting to database:", err)
		os.Exit(1)
	}

	collection := client.Database(config.Mongodb.Database).Collection("machinery-stats")

	gin.SetMode(gin.ReleaseMode)

	r := gin.Default()

	r.Use(func(c *gin.Context) {
		c.Header("Access-Control-Allow-Origin", "*")
		c.Header("Access-Control-Allow-Methods", "POST, GET, OPTIONS, PUT, DELETE, UPDATE")
		c.Header("Access-Control-Allow-Headers", "Content-Type, Content-Length, Accept-Encoding, X-CSRF-Token, Authorization, X-Max")
		c.Header("Access-Control-Allow-Credentials", "true")
		c.Next()
	})

	r.Static("/build", "./frontend/build")
	r.StaticFile("/favicon.ico", "./frontend/build/favicon.ico")
	r.StaticFile("/", "./frontend/build/index.html")

	api := r.Group("/api")
	api.GET("", func(g *gin.Context) {
		// Get the current database data
		cur, err := collection.Find(c(), bson.M{})
		if resErr(g, err) {
			return
		}

		defer cur.Close(c())

		output := []shared.DBEntry{}
		for cur.Next(c()) {
			var newEntry shared.DBEntry
			err := cur.Decode(&newEntry)
			if resErr(g, err) {
				return
			}
			output = append(output, newEntry)
		}

		// Format the data
		toReturn := []apiOutput{}
		for _, entry := range output {
			timelineEntries := []timelineEntry{}
			for _, point := range entry.Points {
				pointTime := time.Unix(point.Time, 0)
				if len(timelineEntries) == 0 || pointTime.After(time.Unix(timelineEntries[len(timelineEntries)-1].From, 0).Add(time.Minute*30)) {
					timelineEntries = append(timelineEntries, timelineEntry{
						From:   point.Time,
						Points: []shared.DataPoint{},
					})
				}
				currentPoints := timelineEntries[len(timelineEntries)-1].Points
				currentPoints = append(currentPoints, point)
				timelineEntries[len(timelineEntries)-1].Points = currentPoints
			}
			toReturn = append(toReturn, apiOutput{
				ID:       entry.ID,
				Queue:    entry.Queue,
				Timeline: timelineEntries,
			})
		}

		res(g, toReturn)
	})

	port := 9090
	fmt.Printf("Running on port: %v\n", port)
	r.Run(fmt.Sprintf(":%v", port))
}
