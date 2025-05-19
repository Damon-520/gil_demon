package dao

import (
	"gil_teacher/app/conf"

	"go.mongodb.org/mongo-driver/v2/mongo"
	"go.mongodb.org/mongo-driver/v2/mongo/options"
)

func NewMongoClient(c *conf.Mongo) (*mongo.Client, error) {
	return mongo.Connect(options.Client().ApplyURI(c.Url))
}
