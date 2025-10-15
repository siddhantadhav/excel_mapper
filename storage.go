package excel_mapper

import (
	"context"
	"errors"
	"fmt"
	"time"

	"go.mongodb.org/mongo-driver/v2/mongo"
	"go.mongodb.org/mongo-driver/v2/mongo/options"
)

/*
=====================
Mongo Storage
=====================
*/

type DBColumnMapping struct {
	Target    string                 `bson:"target" json:"target"`
	Source    []string               `bson:"source" json:"source"`
	Transform string                 `bson:"transform" json:"transform"`                 // sum, concat, average, raw, none
	Formula   string                 `bson:"formula,omitempty" json:"formula,omitempty"` // raw calculations
	Params    map[string]interface{} `bson:"params,omitempty" json:"params,omitempty"`   // e.g. sep for concat
	Default   interface{}            `bson:"default,omitempty" json:"default,omitempty"`
}

type MappingStorage struct {
	UniqueId string            `bson:"unique_id" json:"unique_id"`
	Mappings []DBColumnMapping `bson:"mappings" json:"mappings"`
	Created  time.Time         `bson:"created_at" json:"created_at"`
	Updated  time.Time         `bson:"updated_at" json:"updated_at"`
}

type MongoStorage struct {
	client     *mongo.Client
	collection *mongo.Collection
}

type MongoConfig struct {
	URI        string
	Database   string
	Collection string
	Timeout    time.Duration
}

func ConnectMongo(cfg MongoConfig) (*MongoStorage, error) {
	if cfg.URI == "" || cfg.Database == "" || cfg.Collection == "" {
		return nil, errors.New("incomplete MongoDB config")
	}
	client, err := mongo.Connect(options.Client().ApplyURI(cfg.URI))
	if err != nil {
		return nil, err
	}
	ctx, cancel := context.WithTimeout(context.Background(), cfg.Timeout)
	defer cancel()
	if err := client.Ping(ctx, nil); err != nil {
		return nil, fmt.Errorf("ping failed: %v", err)
	}
	coll := client.Database(cfg.Database).Collection(cfg.Collection)
	return &MongoStorage{client: client, collection: coll}, nil
}

func (ms *MongoStorage) SaveMapping(uniqueId string, dbMappings []DBColumnMapping) error {
	if uniqueId == "" {
		return errors.New("uniqueId required")
	}
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	doc := MappingStorage{
		UniqueId: uniqueId,
		Mappings: dbMappings,
		Updated:  time.Now(),
	}

	// Only update these fields
	update := map[string]interface{}{
		"$set": map[string]interface{}{
			"mappings":   doc.Mappings,
			"updated_at": doc.Updated,
		},
		"$setOnInsert": map[string]interface{}{
			"unique_id":  doc.UniqueId,
			"created_at": time.Now(),
		},
	}

	opts := options.UpdateOne().SetUpsert(true)
	_, err := ms.collection.UpdateOne(ctx, map[string]interface{}{"unique_id": uniqueId}, update, opts)
	return err

}

func (ms *MongoStorage) LoadMapping(uniqueId string) ([]DBColumnMapping, error) {
	if uniqueId == "" {
		return nil, errors.New("uniqueId required")
	}
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()
	var stored MappingStorage
	err := ms.collection.FindOne(ctx, map[string]interface{}{"unique_id": uniqueId}).Decode(&stored)
	if err != nil {
		return nil, err
	}
	return stored.Mappings, nil
}

func (ms *MongoStorage) Close() error {
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	return ms.client.Disconnect(ctx)
}
