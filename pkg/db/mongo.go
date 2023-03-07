package db

import (
	"context"
	"fmt"
	"strings"

	"github.com/CalebQ42/stupid-backend/pkg/crash"
	"github.com/google/uuid"
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
	if res.Err() == mongo.ErrNoDocuments {
		return ErrNotFound
	} else if res.Err() != nil {
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

func (m MongoTable) Find(values map[string]any, v any) (err error) {
	res := m.c.FindOne(context.TODO(), values)
	if res.Err() == mongo.ErrNoDocuments {
		return ErrNotFound
	} else if res.Err() != nil {
		return res.Err()
	}
	return res.Decode(v)
}

func (m MongoTable) FindMany(values map[string]any, v any) (err error) {
	res, err := m.c.Find(context.TODO(), values)
	if err == mongo.ErrNoDocuments {
		return ErrNotFound
	}
	return res.All(context.TODO(), v)
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

func (m MongoTable) Contains(values map[string]any) (bool, error) {
	res := m.c.FindOne(context.TODO(), values)
	if res.Err() == mongo.ErrNoDocuments {
		return false, nil
	} else if res.Err() != nil {
		return false, res.Err()
	}
	return true, nil
}

func (m MongoTable) Delete(key string) error {
	_, err := m.c.DeleteOne(context.TODO(), bson.M{"_id": key})
	return err
}

func (m MongoTable) AddCrash(c crash.Individual) error {
	first, _, _ := strings.Cut(c.Stack, "\n")
	res := m.c.FindOne(context.TODO(), bson.M{"crashes._id": c.ID})
	if res.Err() != nil && res.Err() != mongo.ErrNoDocuments {
		return res.Err()
	}
	res = m.c.FindOneAndUpdate(context.TODO(), bson.M{"error": c.Error, "first": first}, bson.D{{Key: "$addToSet", Value: bson.M{"crashes": c}}})
	if res.Err() == mongo.ErrNoDocuments {
		newGroup := crash.Group{
			ID:        uuid.NewString(),
			Error:     c.Error,
			FirstLine: first,
			Crashes:   []crash.Individual{c},
		}
		_, err := m.c.InsertOne(context.TODO(), newGroup)
		return err
	}
	return res.Err()
}

func (m MongoTable) IncrementFailed(id string) error {
	res := m.c.FindOneAndUpdate(context.TODO(), bson.M{"_id": id}, bson.M{"$inc": bson.M{"failed": 1}})
	if res.Err() == mongo.ErrNoDocuments {
		return ErrNotFound
	}
	return res.Err()
}

func (m MongoTable) IncrementAndUpdateLastTimeout(id string, t int64) error {
	res := m.c.FindOneAndUpdate(context.TODO(), bson.M{"_id": id}, bson.M{"$inc": bson.M{"failed": 1}, "$set": bson.M{"lastTimeout": t}})
	if res.Err() == mongo.ErrNoDocuments {
		return ErrNotFound
	}
	return res.Err()
}
