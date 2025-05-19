package dao

import (
	"context"
	"fmt"
	"testing"
	"time"

	"gil_teacher/app/conf"

	"go.mongodb.org/mongo-driver/v2/bson"
)

func TestNewMongoClient(t *testing.T) {
	c := &conf.Mongo{
		Url: "mongodb://127.0.0.1:27017/?directConnection=true&serverSelectionTimeoutMS=2000&appName=mongosh+2.3.9",
	}

	mongoClient, err := NewMongoClient(c)
	if err != nil {
		fmt.Printf("new mongo client failed.")
		return
	}

	defer func() { _ = mongoClient.Disconnect(context.Background()) }()

	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	collection := mongoClient.Database("db_test").Collection("mcol_example")
	cur, err := collection.Find(ctx, bson.D{{Key: "status", Value: 2}})
	if err != nil {
		fmt.Printf("collection find failed. err:%+v\r\n", err)
		return
	}

	defer func() { _ = cur.Close(ctx) }()

	for cur.Next(ctx) {
		var result bson.D
		if err := cur.Decode(&result); err != nil {
			fmt.Printf("cursor decode failed. err:%+v\r\n", err)
			return
		}

		fmt.Printf("success result:%+v\r\n", result.String())
	}

}
