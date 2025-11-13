package excel_mapper

import (
	"context"
	"fmt"
	"time"

	"go.mongodb.org/mongo-driver/v2/bson"
	"go.mongodb.org/mongo-driver/v2/mongo"
	"go.mongodb.org/mongo-driver/v2/mongo/options"
)

/*
=====================
Mongo Storage
=====================
*/

type DBColumnMapping struct {
	Target    string         `bson:"target" json:"target"`
	Source    []string       `bson:"source" json:"source"`
	Transform string         `bson:"transform" json:"transform"`
	Formula   string         `bson:"formula,omitempty" json:"formula,omitempty"`
	Params    map[string]any `bson:"params,omitempty" json:"params,omitempty"`
	Default   any            `bson:"default,omitempty" json:"default,omitempty"`
	Unique    bool           `bson:"unique,omitempty" json:"unique,omitempty"`
}

type MappingStorage struct {
	UniqueId string            `bson:"unique_id" json:"unique_id"`
	Mappings []DBColumnMapping `bson:"mappings" json:"mappings"`
}

func CreateIndex(ctx context.Context, collection *mongo.Collection) error {
	c, cancel := context.WithTimeout(ctx, 2*time.Second)
	defer cancel()

	indexModel := mongo.IndexModel{
		Keys: bson.D{
			{Key: "unique_id", Value: 1},
		},
		Options: options.Index().SetUnique(true).SetName("unique_id"),
	}

	_, err := collection.Indexes().CreateOne(c, indexModel)
	if err != nil {
		return err
	}

	return nil
}

func SaveMapping(ctx context.Context, collection *mongo.Collection, document *MappingStorage) (*mongo.InsertOneResult, error) {
	err := CreateIndex(ctx, collection)
	if err != nil {
		return nil, err
	}

	c, cancel := context.WithTimeout(ctx, 2*time.Second)
	defer cancel()

	result, err := collection.InsertOne(c, document)

	if err != nil {
		if writeException, ok := err.(mongo.WriteException); ok {
			for _, we := range writeException.WriteErrors {
				if we.Code == 11000 {
					return nil, fmt.Errorf("document with unique_id %s already exists", document.UniqueId)
				}
			}
		}
		return nil, err
	}

	return result, nil
}

func LoadMapping(ctx context.Context, collection *mongo.Collection, uniqueId string) (*MappingStorage, error) {
	c, cancel := context.WithTimeout(ctx, 2*time.Second)
	defer cancel()

	var record MappingStorage

	err := collection.FindOne(c, bson.M{
		"unique_id": uniqueId,
	}).Decode(&record)

	if err != nil {
		return nil, err
	}

	return &record, nil
}
