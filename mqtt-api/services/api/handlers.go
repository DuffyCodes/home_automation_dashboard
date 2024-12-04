package api

import (
	"context"
	"log"
	"net/http"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/prometheus/client_golang/prometheus"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
)

// Handler struct for API
type Handler struct {
	client     *mongo.Client
	database   string
	collection string
}

// Prometheus metrics
var (
	switchOnDuration = prometheus.NewGaugeVec(
		prometheus.GaugeOpts{
			Name: "switch_on_duration_seconds",
			Help: "Current ON duration for switches, in seconds",
		},
		[]string{"topic"},
	)
	kitchenTemperature = prometheus.NewGaugeVec(
		prometheus.GaugeOpts{
			Name: "kitchen_temperature",
			Help: "Current kitchen temperature",
		},
		[]string{"unit"},
	)
	temperatureHistogram = prometheus.NewHistogramVec(
		prometheus.HistogramOpts{
			Name:    "temperature_values",
			Help:    "Histogram of temperature values",
			Buckets: prometheus.LinearBuckets(10, 5, 10), // Buckets start at 10, step by 5, up to 10 buckets
		},
		[]string{"topic", "unit"},
	)
	switchTotalOnDuration = prometheus.NewGaugeVec(
		prometheus.GaugeOpts{
			Name: "switch_total_on_duration_seconds",
			Help: "Total ON duration for switches, in seconds",
		},
		[]string{"topic"},
	)
)

func init() {
	prometheus.MustRegister(switchOnDuration)
	prometheus.MustRegister(switchTotalOnDuration)
	prometheus.MustRegister(kitchenTemperature)
	prometheus.MustRegister(temperatureHistogram)
}

// NewHandler initializes a new Handler instance
func NewHandler(client *mongo.Client, database string) *Handler {
	return &Handler{
		client:     client,
		database:   database,
		collection: "mqtt_events",
	}
}

// GetAverageTemperature handles requests to calculate the average temperature
func (h *Handler) GetAverageTemperature(c *gin.Context) {
	h.aggregateTemperature(c, "$avg")
}

// GetMaxTemperature handles requests to calculate the maximum temperature
func (h *Handler) GetMaxTemperature(c *gin.Context) {
	h.aggregateTemperature(c, "$max")
}

// GetMinTemperature handles requests to calculate the minimum temperature
func (h *Handler) GetMinTemperature(c *gin.Context) {
	h.aggregateTemperature(c, "$min")
}

// aggregateTemperature performs aggregation on temperature data
func (h *Handler) aggregateTemperature(c *gin.Context, aggregation string) {
	topic, startTime, endTime, err := h.parseQueryParams(c)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	pipeline := h.buildAggregationPipeline(topic, startTime, endTime, aggregation)

	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	collection := h.client.Database(h.database).Collection(h.collection)
	cursor, err := collection.Aggregate(ctx, pipeline)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	defer cursor.Close(ctx)

	var result []bson.M
	if err = cursor.All(ctx, &result); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	if len(result) == 0 {
		c.JSON(http.StatusNotFound, gin.H{"message": "No data found"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"value": result[0]["result"]})
}

// parseQueryParams extracts and validates query parameters
func (h *Handler) parseQueryParams(c *gin.Context) (string, time.Time, time.Time, error) {
	topic := c.Query("topic")
	if topic == "" {
		topic = "home/kitchen_temperature/state"
	}

	layout := "2006-01-02"
	startDate := c.Query("start")
	endDate := c.Query("end")

	startTime, err := time.Parse(layout, startDate)
	if err != nil {
		return "", time.Time{}, time.Time{}, err
	}

	endTime, err := time.Parse(layout, endDate)
	if err != nil {
		return "", time.Time{}, time.Time{}, err
	}

	endTime = endTime.Add(24 * time.Hour) // Include the entire day
	return topic, startTime, endTime, nil
}

// buildAggregationPipeline constructs the MongoDB aggregation pipeline
func (h *Handler) buildAggregationPipeline(topic string, startTime, endTime time.Time, aggregation string) bson.A {
	return bson.A{
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
}

func (h *Handler) UpdateSwitchMetrics() {
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	collection := h.client.Database(h.database).Collection(h.collection)

	// Aggregation pipeline to get the latest state for each switch
	pipeline := bson.A{
		bson.D{{Key: "$match", Value: bson.D{
			{Key: "topic", Value: bson.D{{Key: "$in", Value: []string{
				"home/ezras_room_heater/state",
				"home/nikos_room_heater/state",
				"home/bulb_b/state",
				"home/bulb_d/state",
				"home/doorbell_motion/state",
				"home/aquarium_power_monitor/state",
			}}}},
		}}}, // Filter documents with matching topics
		bson.D{{Key: "$sort", Value: bson.D{{Key: "timestamp", Value: -1}}}}, // Sort by latest timestamp first
		bson.D{{Key: "$group", Value: bson.D{
			{Key: "_id", Value: "$topic"},
			{Key: "lastState", Value: bson.D{{Key: "$first", Value: "$value"}}},
			{Key: "lastTimestamp", Value: bson.D{{Key: "$first", Value: "$timestamp"}}},
		}}},
	}

	cursor, err := collection.Aggregate(ctx, pipeline)
	if err != nil {
		log.Printf("Failed to aggregate switch metrics: %v", err)
		return
	}
	defer cursor.Close(ctx)

	currentTime := time.Now()

	for cursor.Next(ctx) {
		var result struct {
			Topic         string    `bson:"_id"`
			LastState     string    `bson:"lastState"`
			LastTimestamp time.Time `bson:"lastTimestamp"`
		}

		if err := cursor.Decode(&result); err != nil {
			log.Printf("Failed to decode aggregation result: %v", err)
			continue
		}

		if result.LastState == "on" {
			// Calculate the duration since the last ON state
			onDuration := currentTime.Sub(result.LastTimestamp).Seconds()

			// Update Prometheus metric
			switchOnDuration.WithLabelValues(result.Topic).Set(onDuration)
			log.Printf("Topic: %s, Running Duration: %.2f seconds", result.Topic, onDuration)
		} else {
			// If the switch is OFF, set the duration to 0
			switchOnDuration.WithLabelValues(result.Topic).Set(0)
			log.Printf("Topic: %s, Currently OFF", result.Topic)
		}
	}

	if err := cursor.Err(); err != nil {
		log.Printf("Cursor error: %v", err)
	}
}

func (h *Handler) UpdateTotalSwitchOnDuration() {
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	collection := h.client.Database(h.database).Collection(h.collection)

	// Aggregation pipeline to calculate total ON duration
	pipeline := bson.A{
		bson.D{{Key: "$match", Value: bson.D{
			{Key: "topic", Value: bson.D{{Key: "$in", Value: []string{
				"home/ezras_room_heater/state",
				"home/nikos_room_heater/state",
				"home/bulb_b/state",
				"home/bulb_d/state",
				"home/doorbell_motion/state",
				"home/aquarium_power_monitor/state",
			}}}},
		}}}, // Filter documents with matching topics
		bson.D{{Key: "$sort", Value: bson.D{
			{Key: "topic", Value: 1},     // Sort by topic
			{Key: "timestamp", Value: 1}, // Then sort by timestamp (ascending)
		}}},
		bson.D{{Key: "$group", Value: bson.D{
			{Key: "_id", Value: "$topic"},
			{Key: "events", Value: bson.D{{Key: "$push", Value: bson.D{
				{Key: "value", Value: "$value"},
				{Key: "timestamp", Value: "$timestamp"},
			}}}},
		}}},
	}

	cursor, err := collection.Aggregate(ctx, pipeline)
	if err != nil {
		log.Printf("Failed to aggregate switch metrics: %v", err)
		return
	}
	defer cursor.Close(ctx)

	for cursor.Next(ctx) {
		var result struct {
			Topic  string `bson:"_id"`
			Events []struct {
				Value     string    `bson:"value"`
				Timestamp time.Time `bson:"timestamp"`
			} `bson:"events"`
		}

		if err := cursor.Decode(&result); err != nil {
			log.Printf("Failed to decode aggregation result: %v", err)
			continue
		}

		totalOnDuration := calculateTotalOnDuration(result.Events)

		// Update Prometheus metric for total ON duration
		switchTotalOnDuration.WithLabelValues(result.Topic).Set(totalOnDuration)
		log.Printf("Topic: %s, Total ON Duration: %.2f seconds", result.Topic, totalOnDuration)
	}

	if err := cursor.Err(); err != nil {
		log.Printf("Cursor error: %v", err)
	}
}

func calculateTotalOnDuration(events []struct {
	Value     string    `bson:"value"`
	Timestamp time.Time `bson:"timestamp"`
}) float64 {
	var totalDuration float64
	var lastOnTime *time.Time
	inOnState := false

	for _, event := range events {
		log.Printf("Value: %s, Time: %s", event.Value, event.Timestamp)
		if event.Value == "on" {
			if !inOnState {
				lastOnTime = &event.Timestamp
				inOnState = true
			}
			// Ignore extra 'on' events
		} else if event.Value == "off" {
			if inOnState && lastOnTime != nil {
				duration := event.Timestamp.Sub(*lastOnTime).Seconds()
				totalDuration += duration
				log.Printf("Adding duration: %.2f seconds", duration)
				lastOnTime = nil
				inOnState = false
			}
		}
	}

	// Handle the case where the last event is 'on' without a corresponding 'off'
	if inOnState && lastOnTime != nil {
		currentTime := time.Now()
		duration := currentTime.Sub(*lastOnTime).Seconds()
		totalDuration += duration
		log.Printf("Adding duration for unpaired 'on': %.2f seconds", duration)
	}

	log.Printf("Total ON Duration: %.2f seconds", totalDuration)
	return totalDuration
}
