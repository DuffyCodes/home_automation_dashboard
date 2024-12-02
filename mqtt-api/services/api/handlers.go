package api

import (
	"context"
	"net/http"
	"time"

	"github.com/gin-gonic/gin"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
)

type Handler struct {
	client     *mongo.Client
	database   string
	collection string
}

// NewHandler creates a new API handler
func NewHandler(client *mongo.Client, database string) *Handler {
	return &Handler{
		client:     client,
		database:   database,
		collection: "mqtt_events",
	}
}

// GetAverageTemperature retrieves the average temperature
func (h *Handler) GetAverageTemperature(c *gin.Context) {
	h.aggregateTemperature(c, "$avg")
}

// GetMaxTemperature retrieves the maximum temperature
func (h *Handler) GetMaxTemperature(c *gin.Context) {
	h.aggregateTemperature(c, "$max")
}

// GetMinTemperature retrieves the minimum temperature
func (h *Handler) GetMinTemperature(c *gin.Context) {
	h.aggregateTemperature(c, "$min")
}

// aggregateTemperature is a helper for temperature aggregation
func (h *Handler) aggregateTemperature(c *gin.Context, aggregation string) {
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	collection := h.client.Database(h.database).Collection(h.collection)

	topic := c.Query("topic")
	if topic == "" {
		topic = "home/kitchen_temperature/state"
	}

	startDate := c.Query("start")
	endDate := c.Query("end")

	layout := "2006-01-02"

	startTime, err := time.Parse(layout, startDate)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid start date format. Expected 'YYYY-MM-DD'"})
		return
	}

	// Increment the end date by one day to include the entire day in the range
	endTime, err := time.Parse(layout, endDate)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid end date format. Expected 'YYYY-MM-DD'"})
		return
	}
	endTime = endTime.Add(24 * time.Hour)
	// Define aggregation pipeline
	pipeline := bson.A{
		bson.D{{Key: "$match", Value: bson.D{
			{Key: "topic", Value: topic},
			{Key: "time", Value: bson.D{
				{Key: "$gte", Value: startTime},
				{Key: "$lt", Value: endTime},
			}},
		}}},
		bson.D{{Key: "$addFields", Value: bson.D{
			{Key: "numeric_payload", Value: bson.D{{Key: "$toDouble", Value: "$payload"}}},
		}}},
		bson.D{{Key: "$group", Value: bson.D{
			{Key: "_id", Value: nil},
			{Key: "result", Value: bson.D{{Key: aggregation, Value: "$numeric_payload"}}},
		}}},
	}

	// Execute aggregation
	cursor, err := collection.Aggregate(ctx, pipeline)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	defer cursor.Close(ctx)

	// Parse results
	var result []bson.M
	if err = cursor.All(ctx, &result); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	if len(result) == 0 {
		c.JSON(http.StatusNotFound, gin.H{"message": "No data found"})
		return
	}

	// Return the aggregated value
	c.JSON(http.StatusOK, gin.H{"value": result[0]["result"]})
}
