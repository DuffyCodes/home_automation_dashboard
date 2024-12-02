package db

import (
	"context"
	"log"
	"time"

	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
)

// InsertDocument inserts a document into a specified collection.
func (db *MongoDB) InsertDocument(collectionName string, document interface{}) error {
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	collection := db.Database.Collection(collectionName)
	_, err := collection.InsertOne(ctx, document)
	return err
}

// FindDocuments queries documents from a specified collection using a filter.
func (db *MongoDB) FindDocuments(collectionName string, filter interface{}) ([]bson.M, error) {
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	collection := db.Database.Collection(collectionName)
	cursor, err := collection.Find(ctx, filter)
	if err != nil {
		return nil, err
	}
	defer cursor.Close(ctx)

	var results []bson.M
	for cursor.Next(ctx) {
		var doc bson.M
		if err := cursor.Decode(&doc); err != nil {
			return nil, err
		}
		results = append(results, doc)
	}

	return results, nil
}

// UpdateDocument updates a document in a specified collection.
func (db *MongoDB) UpdateDocument(collectionName string, filter, update interface{}) error {
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	collection := db.Database.Collection(collectionName)
	_, err := collection.UpdateOne(ctx, filter, update)
	return err
}

// DeleteDocument deletes a document from a specified collection.
func (db *MongoDB) DeleteDocument(collectionName string, filter interface{}) error {
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	collection := db.Database.Collection(collectionName)
	_, err := collection.DeleteOne(ctx, filter)
	return err
}

func FetchAverageTemperature(ctx context.Context, client *mongo.Client, dbName, collectionName string, startTime, endTime time.Time) (float64, error) {
	collection := client.Database(dbName).Collection(collectionName)

	pipeline := bson.D{
		{Key: "$group", Value: bson.D{
			{Key: "_id", Value: "$topic"},
			{Key: "average", Value: bson.D{
				{Key: "$avg", Value: "$value"},
			}},
		}},
	}

	cursor, err := collection.Aggregate(ctx, pipeline)
	if err != nil {
		log.Printf("Error in aggregation: %v", err)
		return 0, err
	}
	defer cursor.Close(ctx)

	var result []bson.M
	if err = cursor.All(ctx, &result); err != nil {
		log.Printf("Error reading cursor: %v", err)
		return 0, err
	}

	if len(result) > 0 {
		return result[0]["averageTemperature"].(float64), nil
	}
	return 0, nil
}

func GetMaxMinTemperature(collection *mongo.Collection, topic string, startTime, endTime time.Time) (max float64, min float64, err error) {
	pipeline := mongo.Pipeline{
		bson.D{{Key: "$match", Value: bson.D{
			{Key: "topic", Value: topic},
			{Key: "timestamp", Value: bson.D{
				{Key: "$gte", Value: startTime},
				{Key: "$lt", Value: endTime},
			}},
		}}},
		bson.D{{Key: "$group", Value: bson.D{
			{Key: "_id", Value: "$topic"},
			{Key: "maxTemperature", Value: bson.D{{Key: "$max", Value: bson.D{{Key: "$toDouble", Value: "$payload"}}}}},
			{Key: "minTemperature", Value: bson.D{{Key: "$min", Value: bson.D{{Key: "$toDouble", Value: "$payload"}}}}},
		}}},
	}

	cursor, err := collection.Aggregate(context.TODO(), pipeline)
	if err != nil {
		return 0, 0, err
	}
	defer cursor.Close(context.TODO())

	if cursor.Next(context.TODO()) {
		var result struct {
			MaxTemperature float64 `bson:"maxTemperature"`
			MinTemperature float64 `bson:"minTemperature"`
		}
		if err := cursor.Decode(&result); err != nil {
			return 0, 0, err
		}
		return result.MaxTemperature, result.MinTemperature, nil
	}
	return 0, 0, nil
}

// CountOnOffEvents calculates the number of ON/OFF events for a switch topic.
func CountOnOffEvents(collection *mongo.Collection, topic string, startTime, endTime time.Time) (onCount, offCount int64, err error) {
	pipeline := mongo.Pipeline{
		bson.D{{Key: "$match", Value: bson.D{
			{Key: "topic", Value: topic},
			{Key: "timestamp", Value: bson.D{
				{Key: "$gte", Value: startTime},
				{Key: "$lt", Value: endTime},
			}},
		}}},
		bson.D{{Key: "$group", Value: bson.D{
			{Key: "_id", Value: "$payload"},
			{Key: "count", Value: bson.D{{Key: "$sum", Value: 1}}},
		}}},
	}

	cursor, err := collection.Aggregate(context.TODO(), pipeline)
	if err != nil {
		return 0, 0, err
	}
	defer cursor.Close(context.TODO())

	for cursor.Next(context.TODO()) {
		var result struct {
			ID    string `bson:"_id"`
			Count int64  `bson:"count"`
		}
		if err := cursor.Decode(&result); err != nil {
			return 0, 0, err
		}
		if result.ID == "on" {
			onCount = result.Count
		} else if result.ID == "off" {
			offCount = result.Count
		}
	}
	return onCount, offCount, nil
}
