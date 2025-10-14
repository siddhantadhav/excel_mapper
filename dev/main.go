package main

import (
	"fmt"
	"path/filepath"

	"github.com/siddhantadhav/excel_mapper"
)

func main() {
	dir := "./sample_data"

	fileA, err := excel_mapper.InitFile(dir, "fileA.xlsx")
	if err != nil {
		fmt.Println("Error initializing fileA:", err)
		return
	}

	fileB, err := excel_mapper.InitFile(dir, "fileB.xlsx")
	if err != nil {
		fmt.Println("Error initializing fileB:", err)
		return
	}

	mappings := []excel_mapper.ColumnMapping{
		{Target: "First Name", Source: []string{"First Name"}},
		{Target: "Last Name", Source: []string{"Last Name"}},
		{Target: "Full Name", Transform: excel_mapper.ConcatColumns(" ", "First Name", "Last Name")},
		{Target: "Score", Transform: excel_mapper.SumColumns("Math", "Science")},
		{Target: "Age", Source: []string{"Age"}},
		{Target: "Average", Transform: excel_mapper.AverageColumns("Math", "Science")},
	}

	output := filepath.Join(dir, "fileB_filled.xlsx")
	err = excel_mapper.FillFile(fileA, fileB, mappings, output)
	if err != nil {
		fmt.Println("Error filling file:", err)
		return
	}

	fmt.Println("✅ Mapping completed successfully. Output:", output)
}
