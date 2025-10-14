package excel_mapper

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/xuri/excelize/v2"
)

type File struct {
	Col  []string
	Path string
	Name string
}

type MappingFunc func(rowA []string, fileA *File) interface{}

type ColMapping struct {
	FileA   *File
	FileB   *File
	Mapping map[string]interface{}
}

func InitFile(filePath, fileName string) (*File, error) {
	fullPath := filepath.Join(filePath, fileName)
	if _, err := os.Stat(fullPath); os.IsNotExist(err) {
		f := excelize.NewFile()
		if err := f.SaveAs(fullPath); err != nil {
			return nil, fmt.Errorf("failed to create file: %v", err)
		}
		return &File{Col: []string{}, Path: fullPath, Name: fileName}, nil
	}

	f, err := excelize.OpenFile(fullPath)
	if err != nil {
		return nil, fmt.Errorf("failed to open file: %v", err)
	}
	defer f.Close()

	sheets := f.GetSheetList()
	if len(sheets) == 0 {
		return &File{Col: []string{}, Path: fullPath, Name: fileName}, nil
	}
	sheetName := sheets[0]

	rows, err := f.GetRows(sheetName)
	if err != nil {
		return nil, fmt.Errorf("failed to read rows: %v", err)
	}

	var headers []string
	if len(rows) > 0 {
		headers = rows[0]
	}

	return &File{
		Col:  headers,
		Path: fullPath,
		Name: fileName,
	}, nil
}

func ColIndex(name string, f *File) int {
	for i, col := range f.Col {
		if strings.EqualFold(strings.TrimSpace(col), strings.TrimSpace(name)) {
			return i
		}
	}
	return -1
}

func FillFile(colMapping *ColMapping, outputPath string) error {
	var fB *excelize.File
	if _, err := os.Stat(colMapping.FileB.Path); os.IsNotExist(err) {
		fB = excelize.NewFile()
	} else {
		tmp, err := excelize.OpenFile(colMapping.FileB.Path)
		if err != nil {
			return fmt.Errorf("failed to open fileB: %v", err)
		}
		fB = tmp
	}

	sheets := fB.GetSheetList()
	var sheetB string
	if len(sheets) == 0 {
		sheetB = "Sheet1"
		fB.NewSheet(sheetB)
		fB.SetActiveSheet(0)
	} else {
		sheetB = sheets[0]
	}

	fA, err := excelize.OpenFile(colMapping.FileA.Path)
	if err != nil {
		return fmt.Errorf("failed to open fileA: %v", err)
	}
	defer fA.Close()

	sheetA := fA.GetSheetList()[0]
	rowsA, err := fA.GetRows(sheetA)
	if err != nil {
		return fmt.Errorf("failed to read rows from fileA: %v", err)
	}
	if len(rowsA) <= 1 {
		return fmt.Errorf("no data rows found in fileA")
	}

	headersB := append([]string{}, colMapping.FileB.Col...)
	headerChanged := false
	existingHeaderMap := make(map[string]bool)
	for _, h := range headersB {
		existingHeaderMap[strings.TrimSpace(h)] = true
	}

	for key := range colMapping.Mapping {
		if !existingHeaderMap[strings.TrimSpace(key)] {
			headersB = append(headersB, key)
			existingHeaderMap[strings.TrimSpace(key)] = true
			headerChanged = true
		}
	}

	if len(colMapping.FileB.Col) == 0 || headerChanged {
		for i, h := range headersB {
			cell, _ := excelize.CoordinatesToCellName(i+1, 1)
			fB.SetCellValue(sheetB, cell, h)
		}
		colMapping.FileB.Col = headersB
	}

	colIndexB := make(map[string]int)
	for i, col := range colMapping.FileB.Col {
		colIndexB[strings.TrimSpace(col)] = i
	}

	startRow := 2
	for r := 1; r < len(rowsA); r++ {
		rowA := rowsA[r]
		targetRow := startRow + (r - 1)

		for colB, mapping := range colMapping.Mapping {
			idxB, ok := colIndexB[strings.TrimSpace(colB)]
			if !ok {
				continue
			}

			var value interface{}
			switch m := mapping.(type) {
			case string:
				idxA := ColIndex(m, colMapping.FileA)
				if idxA >= 0 && idxA < len(rowA) {
					value = rowA[idxA]
				}
			case MappingFunc:
				value = m(rowA, colMapping.FileA)
			}

			cellB, _ := excelize.CoordinatesToCellName(idxB+1, targetRow)
			if err := fB.SetCellValue(sheetB, cellB, value); err != nil {
				return fmt.Errorf("failed to set cell %s: %v", cellB, err)
			}
		}
	}

	if err := fB.SaveAs(outputPath); err != nil {
		return fmt.Errorf("failed to save fileB: %v", err)
	}

	return nil
}
