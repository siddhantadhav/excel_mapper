package main

import (
	"fmt"
	"log"
	"strconv"

	"github.com/siddhantadhav/excel_mapper"
)

func main() {
	fileA, err := excel_mapper.InitFile("./sample_data", "fileA.xlsx")
	if err != nil {
		log.Fatalf("Failed to load fileA: %v", err)
	}

	fileB, err := excel_mapper.InitFile("./sample_data", "fileB.xlsx")
	if err != nil {
		log.Fatalf("Failed to load fileB: %v", err)
	}

	mapping := map[string]interface{}{
		"Full Name": excel_mapper.MappingFunc(func(row []string, fA *excel_mapper.File) interface{} {
			firstIdx := excel_mapper.ColIndex("First Name", fA)
			lastIdx := excel_mapper.ColIndex("Last Name", fA)
			if firstIdx < 0 || lastIdx < 0 {
				return ""
			}
			return fmt.Sprintf("%s %s", row[firstIdx], row[lastIdx])
		}),
		"Age": "Age",
		"Score": excel_mapper.MappingFunc(func(row []string, fA *excel_mapper.File) interface{} {
			mathIdx := excel_mapper.ColIndex("Math", fA)
			sciIdx := excel_mapper.ColIndex("Science", fA)
			if mathIdx < 0 || sciIdx < 0 {
				return nil
			}
			mathVal, _ := strconv.ParseFloat(row[mathIdx], 64)
			sciVal, _ := strconv.ParseFloat(row[sciIdx], 64)
			return mathVal + sciVal
		}),
		"Average": excel_mapper.MappingFunc(func(row []string, fA *excel_mapper.File) interface{} {
			mathIdx := excel_mapper.ColIndex("Math", fA)
			sciIdx := excel_mapper.ColIndex("Science", fA)
			if mathIdx < 0 || sciIdx < 0 {
				return nil
			}
			mathVal, _ := strconv.ParseFloat(row[mathIdx], 64)
			sciVal, _ := strconv.ParseFloat(row[sciIdx], 64)
			return ( mathVal + sciVal )/2
		}),
	}

	colMapping := &excel_mapper.ColMapping{
		FileA:   fileA,
		FileB:   fileB,
		Mapping: mapping,
	}

	outputPath := "./sample_data/fileB_filled.xlsx"
	if err := excel_mapper.FillFile(colMapping, outputPath); err != nil {
		log.Fatalf("❌ Failed to fill FileB: %v", err)
	}

	fmt.Printf("✅ FileB successfully filled and saved at: %s\n", outputPath)
}
