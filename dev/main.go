package main

import (
	"encoding/json"
	"flag"
	"fmt"
	"log"
	"os"
	"path/filepath"

	"github.com/siddhantadhav/excel_mapper"
)

type CLIFlags struct {
	JSONFilePath          string
	ClientFilePath        string
	OutputFilePath        string
	HTMLTemplatePath      string
	FinalHTMLTemplatePath string
	FinalPDFPath          string
}

func ParseJSON(jsonFilePath string) (*excel_mapper.MappingStorage, error) {
	jsonFileBytes, err := os.ReadFile(jsonFilePath)
	if err != nil {
		return nil, err
	}

	var mappingStorage excel_mapper.MappingStorage
	err = json.Unmarshal(jsonFileBytes, &mappingStorage)
	if err != nil {
		return nil, err
	}

	return &mappingStorage, nil
}

func GetColumnMapping(mappingStorage *excel_mapper.MappingStorage) ([]excel_mapper.ColumnMapping, error) {
	var columnMappings []excel_mapper.ColumnMapping
	for _, mapping := range mappingStorage.Mappings {
		columnMappings = append(columnMappings, mapping.ToColumnMapping())
	}
	return columnMappings, nil
}

func main() {
	jsonFilePath := flag.String("json", "", "Path to JSON mapping file")
	clientFilePath := flag.String("input", "", "Path to client Excel file")
	outputFilePath := flag.String("output", "", "Path to output Excel file")
	flag.Parse()

	// Validate required flags
	if *jsonFilePath == "" || *clientFilePath == "" || *outputFilePath == "" {
		log.Fatal("❌ Please provide --json, --input, and --output flags")
	}

	// Initialize flags structure
	flags := CLIFlags{
		JSONFilePath:          *jsonFilePath,
		ClientFilePath:        *clientFilePath,
		OutputFilePath:        *outputFilePath,
	}

	// --- Step 1: Load JSON mapping ---
	mappingStorage, err := ParseJSON(flags.JSONFilePath)
	if err != nil {
		log.Fatalf("❌ Failed to parse JSON mapping: %v", err)
	}

	columnMappings, err := GetColumnMapping(mappingStorage)
	if err != nil {
		log.Fatalf("❌ Failed to extract column mappings: %v", err)
	}

	// --- Step 2: Initialize input and output Excel files ---
	inputFile, err := excel_mapper.InitFile(filepath.Dir(flags.ClientFilePath), filepath.Base(flags.ClientFilePath))
	if err != nil {
		log.Fatalf("❌ Failed to load input Excel file: %v", err)
	}

	outputFile, err := excel_mapper.InitFile(filepath.Dir(flags.OutputFilePath), filepath.Base(flags.OutputFilePath))
	if err != nil {
		log.Fatalf("❌ Failed to create output Excel file: %v", err)
	}

	fmt.Printf("📥 Input headers: %v\n", inputFile.Col)
	fmt.Printf("🗺️ Mappings: %v\n", columnMappings)

	// --- Step 3: Fill output Excel file ---
	err = excel_mapper.FillFile(inputFile, outputFile, columnMappings)
	if err != nil {
		log.Fatalf("❌ Failed to fill output file: %v", err)
	}

	fmt.Println("✅ Output Excel file generated successfully!")
}
