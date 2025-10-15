package main

import (
	"fmt"
	"path/filepath"
	"time"

	"github.com/siddhantadhav/excel_mapper"
)

func main() {
	dir := "./sample_data"

	// Initialize FileA (source)
	fileA, err := excel_mapper.InitFile(dir, "fileA.xlsx")
	if err != nil {
		fmt.Println("Error initializing fileA:", err)
		return
	}

	// Initialize FileB (destination)
	fileB, err := excel_mapper.InitFile(dir, "fileB.xlsx")
	if err != nil {
		fmt.Println("Error initializing fileB:", err)
		return
	}

	// Define mappings
	mappings := []excel_mapper.ColumnMapping{
		{Target: "First Name", Source: []string{"First Name"}},
		{Target: "Last Name", Source: []string{"Last Name"}},
		{Target: "Age", Source: []string{"Age"}},
		{Target: "Full Name", Transform: excel_mapper.ConcatColumns(" ", "First Name", "Last Name")},
		{Target: "Score", Transform: excel_mapper.SumColumns("Math", "Science")},
		{Target: "Average", Transform: excel_mapper.AverageColumns("Math", "Science")},
		{Target: "Adjusted Score", Transform: excel_mapper.MappingFunc(func(row []string, f *excel_mapper.File) interface{} {
			mathIdx := excel_mapper.ColIndex("Math", f)
			sciIdx := excel_mapper.ColIndex("Science", f)
			if mathIdx == -1 || sciIdx == -1 {
				return 0
			}
			math := excel_mapper.ParseFloatSafe(row[mathIdx])
			science := excel_mapper.ParseFloatSafe(row[sciIdx])
			return (math*0.6 + science*0.4)
		})},
		{Target: "Percentage", Transform: excel_mapper.MappingFunc(func(row []string, f *excel_mapper.File) interface{} {
			scoreIdx := excel_mapper.ColIndex("Score", f)
			if scoreIdx == -1 {
				return 0
			}
			score := excel_mapper.ParseFloatSafe(row[scoreIdx])
			return (score / 200) * 100
		})},
	}

	// Output Excel file
	output := filepath.Join(dir, "fileB_filled.xlsx")

	// Fill FileB using mappings
	err = excel_mapper.FillFile(fileA, fileB, mappings, output)
	if err != nil {
		fmt.Println("Error filling file:", err)
		return
	}
	fmt.Println("✅ Excel mapping completed successfully. Output:", output)

	// ----------------------
	// MongoDB Storage Example
	// ----------------------

	cfg := excel_mapper.MongoConfig{
		URI:        "mongodb://localhost:27017",
		Database:   "excel_mapper_db",
		Collection: "mappings",
		Timeout:    5 * time.Second,
	}

	store, err := excel_mapper.ConnectMongo(cfg)
	if err != nil {
		fmt.Println("Error connecting to MongoDB:", err)
		return
	}
	defer store.Close()

	uniqueId := "mapping_001"

	// Save mapping to MongoDB
	err = store.SaveMapping(uniqueId, mappings)
	if err != nil {
		fmt.Println("Error saving mapping to MongoDB:", err)
		return
	}
	fmt.Println("✅ Mapping saved to MongoDB with uniqueId:", uniqueId)

	// Load mapping back from MongoDB
	loadedMappings, err := store.LoadMapping(uniqueId)
	if err != nil {
		fmt.Println("Error loading mapping from MongoDB:", err)
		return
	}
	fmt.Printf("✅ Loaded %d mappings from MongoDB for uniqueId %s\n", len(loadedMappings), uniqueId)
}
