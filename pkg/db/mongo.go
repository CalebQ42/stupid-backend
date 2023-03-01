package db

import (
	"context"
	"fmt"

	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
)

// Satisfies both db.Table and db.CrashTable.
type MongoTable struct {
	c *mongo.Collection
}

func NewMongoTable(c *mongo.Collection) *MongoTable {
	return &MongoTable{
		c: c,
	}
}

func (m MongoTable) Get(key string, v any) error {
	res := m.c.FindOne(context.TODO(), bson.D{{Key: "_id", Value: key}})
	if res.Err() != nil {
		return res.Err()
	}
	return res.Decode(v)
}

func (m MongoTable) Add(v any) (string, error) {
	res, err := m.c.InsertOne(context.TODO(), v)
	if err != nil {
		return "", err
	}
	return fmt.Sprint(res.InsertedID), nil
}

func (m MongoTable) Update(key string, v any) error {
	res := m.c.FindOneAndReplace(context.TODO(), bson.M{"_id": key}, v)
	return res.Err()
}

func (m MongoTable) Has(key string) (bool, error) {
	res := m.c.FindOne(context.TODO(), bson.M{"_id": key})
	if res.Err() == mongo.ErrNoDocuments {
		return false, nil
	} else if res.Err() != nil {
		return false, res.Err()
	}
	return true, nil
}

func (m MongoTable) ContainsIndividualCrash(id string) (bool, error) {

}
