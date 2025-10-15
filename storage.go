package excel_mapper

import (
	"context"
	"errors"
	"fmt"
	"time"

	"github.com/Knetic/govaluate"
	"go.mongodb.org/mongo-driver/v2/mongo"
	"go.mongodb.org/mongo-driver/v2/mongo/options"
)

/*
=====================
Database Structures
=====================
*/

type DBColumnMapping struct {
	Target    string                 `bson:"target" json:"target"`
	Source    []string               `bson:"source" json:"source"`
	Transform string                 `bson:"transform" json:"transform"`                 // "sum", "concat", "average", "raw", "none"
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

/*
=====================
MongoDB Storage Layer
=====================
*/
type MongoStorage struct {
	client     *mongo.Client
	collection *mongo.Collection
}

// MongoConfig represents user-supplied connection settings
type MongoConfig struct {
	URI        string
	Database   string
	Collection string
	Timeout    time.Duration
}

// ConnectMongo initializes a MongoDB connection
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

	// Ping to verify connection
	if err := client.Ping(ctx, nil); err != nil {
		return nil, fmt.Errorf("ping failed: %v", err)
	}

	coll := client.Database(cfg.Database).Collection(cfg.Collection)
	return &MongoStorage{client: client, collection: coll}, nil
}

// SaveMapping stores or updates mappings by UniqueId
func (ms *MongoStorage) SaveMapping(uniqueId string, mappings []ColumnMapping) error {
	if uniqueId == "" {
		return errors.New("uniqueId required")
	}

	var dbMappings []DBColumnMapping
	for _, m := range mappings {
		dbMappings = append(dbMappings, m.ToDBMapping())
	}

	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	doc := MappingStorage{
		UniqueId: uniqueId,
		Mappings: dbMappings,
		Updated:  time.Now(),
	}

	filter := map[string]interface{}{"unique_id": uniqueId}
	update := map[string]interface{}{
		"$set": doc,
		"$setOnInsert": map[string]interface{}{
			"created_at": time.Now(),
		},
	}

	opts := options.UpdateOne().SetUpsert(true)

	_, err := ms.collection.UpdateOne(ctx, filter, update, opts)
	return err
}

// LoadMapping retrieves mappings by UniqueId
func (ms *MongoStorage) LoadMapping(uniqueId string) ([]ColumnMapping, error) {
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

	var result []ColumnMapping
	for _, dbMap := range stored.Mappings {
		result = append(result, dbMap.ToColumnMapping())
	}
	return result, nil
}

/*
=====================
Conversion Functions
=====================
*/

func (dbMap *DBColumnMapping) ToColumnMapping() ColumnMapping {
	var transform MappingFunc

	switch dbMap.Transform {
	case "sum":
		transform = SumColumns(dbMap.Source...)
	case "concat":
		sep := " "
		if s, ok := dbMap.Params["sep"].(string); ok {
			sep = s
		}
		transform = ConcatColumns(sep, dbMap.Source...)
	case "average":
		transform = AverageColumns(dbMap.Source...)
	case "raw":
		formula := dbMap.Formula
		transform = func(row []string, f *File) interface{} {
			expr, err := govaluate.NewEvaluableExpression(formula)
			if err != nil {
				return fmt.Sprintf("invalid formula: %s", formula)
			}

			params := make(map[string]interface{})
			for _, col := range f.Col {
				idx := ColIndex(col, f)
				if idx >= 0 && idx < len(row) {
					params[col] = ParseFloatSafe(row[idx])
				}
			}

			result, err := expr.Evaluate(params)
			if err != nil {
				return fmt.Sprintf("error: %v", err)
			}
			return result
		}
	default:
		transform = nil
	}

	return ColumnMapping{
		Target:    dbMap.Target,
		Source:    dbMap.Source,
		Transform: transform,
		Default:   dbMap.Default,
	}
}

func (m *ColumnMapping) ToDBMapping() DBColumnMapping {
	transformType := "none"

	switch {
	case m.Transform == nil:
		transformType = "none"
	default:
		transformType = "raw"
	}

	return DBColumnMapping{
		Target:    m.Target,
		Source:    m.Source,
		Transform: transformType,
		Default:   m.Default,
	}
}

/*
=====================
	Utility
=====================
*/

func (ms *MongoStorage) Close() error {
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	return ms.client.Disconnect(ctx)
}
