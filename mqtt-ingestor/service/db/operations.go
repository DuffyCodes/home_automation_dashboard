package db

import (
	"context"
	"time"

	"go.mongodb.org/mongo-driver/bson"
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
