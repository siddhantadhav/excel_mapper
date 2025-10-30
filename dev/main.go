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
	JSONFilePath   string
	ClientFilePath string
	OutputFilePath string
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
	jsonFilePath := flag.String("json", "", "path to json file")
	clientFilePath := flag.String("input", "", "path to client excel file")
	outputFilePath := flag.String("output", "", "path to output excel file")
	flag.Parse()

	if *jsonFilePath == "" || *clientFilePath == "" || *outputFilePath == "" {
		log.Fatal("Please provide --json, --input, and --output flags")
	}

	flags := CLIFlags{
		JSONFilePath:   *jsonFilePath,
		ClientFilePath: *clientFilePath,
		OutputFilePath: *outputFilePath,
	}

	// Load JSON mapping
	mappingStorage, err := ParseJSON(flags.JSONFilePath)
	if err != nil {
		log.Fatal(err)
	}

	columnMappings, err := GetColumnMapping(mappingStorage)
	if err != nil {
		log.Fatal(err)
	}

	// Init input and output files
	inputFile, err := excel_mapper.InitFile(filepath.Dir(flags.ClientFilePath), filepath.Base(flags.ClientFilePath))
	if err != nil {
		log.Fatal(err)
	}

	outputFile, err := excel_mapper.InitFile(filepath.Dir(flags.OutputFilePath), filepath.Base(flags.OutputFilePath))
	if err != nil {
		log.Fatal(err)
	}

	fmt.Printf("Input headers: %v\n", inputFile.Col)
	fmt.Printf("Mappings: %v\n", columnMappings)

	// Fill output file
	err = excel_mapper.FillFile(inputFile, outputFile, columnMappings)
	if err != nil {
		log.Fatal(err)
	}

	fmt.Println("Output file generated successfully!")
}
