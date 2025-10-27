package main

import (
	"encoding/json"
	"fmt"
	"log"
	"os"
	"time"

	"github.com/siddhantadhav/excel_mapper"
)

func main() {
	// ------------------------------
	// Read mapping JSON
	// ------------------------------
	jsonData, err := os.ReadFile("sample_data/mapping.json")
	if err != nil {
		log.Fatal("failed to read JSON:", err)
	}

	var input struct {
		UniqueId string                         `json:"unique_id"`
		Mappings []excel_mapper.DBColumnMapping `json:"mappings"`
	}
	if err := json.Unmarshal(jsonData, &input); err != nil {
		log.Fatal("invalid JSON:", err)
	}

	// ------------------------------
	// Connect to MongoDB
	// ------------------------------
	mongoCfg := excel_mapper.MongoConfig{
		URI:        "",
		Database:   "",
		Collection: "",
		Timeout:    10 * time.Second,
	}

	store, err := excel_mapper.ConnectMongo(mongoCfg)
	if err != nil {
		log.Fatal("connect mongo:", err)
	}
	defer store.Close()

	// ------------------------------
	// Save mappings to Mongo
	// ------------------------------
	if err := store.SaveMapping(input.UniqueId, input.Mappings); err != nil {
		log.Fatal("save mapping:", err)
	}
	fmt.Println("✅ Mappings saved to MongoDB")

	// ------------------------------
	// Load mappings back from Mongo
	// ------------------------------
	dbMappings, err := store.LoadMapping(input.UniqueId)
	if err != nil {
		log.Fatal("load mapping:", err)
	}
	fmt.Printf("✅ Loaded %d mappings from MongoDB\n", len(dbMappings))

	// ------------------------------
	// Convert DBColumnMapping to ColumnMapping
	// ------------------------------
	var colMappings []excel_mapper.ColumnMapping
	for _, m := range dbMappings {
		colMappings = append(colMappings, m.ToColumnMapping())
	}

	// ------------------------------
	// Initialize Excel files
	// ------------------------------
	fileA, err := excel_mapper.InitFile("sample_data", "fileA.xlsx")
	if err != nil {
		log.Fatal("init FileA:", err)
	}

	fileB, err := excel_mapper.InitFile("sample_data", "fileB.xlsx")
	if err != nil {
		log.Fatal("init FileB:", err)
	}

	// ------------------------------
	// Fill FileB from FileA
	// ------------------------------
	outputPath := "sample_data/fileB_filled.xlsx"
	if err := excel_mapper.FillFile(fileA, fileB, colMappings, outputPath); err != nil {
		log.Fatal("fill FileB:", err)
	}

	fmt.Println("✅ FileB has been filled successfully:", outputPath)
}
