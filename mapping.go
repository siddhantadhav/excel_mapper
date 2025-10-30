package excel_mapper

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/xuri/excelize/v2"
)

/*
=====================
Excel File Structures
=====================
*/

type File struct {
	Path string
	Name string
	Col  []string
}

// MappingFunc generates a value for a FileB column
type MappingFunc func(row []string, fA *File) any

// ColumnMapping represents a mapping rule
type ColumnMapping struct {
	Target    string
	Source    []string
	Transform MappingFunc
	Formula   string
	Params    map[string]any
	Default   any
}

/*
=====================
File Indexing
=====================
*/

type FileIndex struct {
	ColMap map[string]int
	File   *File
}

func NewFileIndex(f *File) *FileIndex {
	m := make(map[string]int)
	for i, col := range f.Col {
		m[strings.ToLower(strings.TrimSpace(col))] = i
	}
	return &FileIndex{ColMap: m, File: f}
}

func ColIndex(name string, f *File) int {
	if f == nil {
		return -1
	}
	idx, ok := NewFileIndex(f).ColMap[strings.ToLower(strings.TrimSpace(name))]
	if !ok {
		return -1
	}
	return idx
}

/*
=====================
File Initialization
=====================
*/

func InitFile(dir, name string) (*File, error) {
	path := filepath.Join(dir, name)
	if _, err := os.Stat(path); os.IsNotExist(err) {
		f := excelize.NewFile()
		if err := f.SaveAs(path); err != nil {
			return nil, fmt.Errorf("create file: %v", err)
		}
		return &File{Path: path, Name: name}, nil
	}

	f, err := excelize.OpenFile(path)
	if err != nil {
		return nil, fmt.Errorf("open file: %v", err)
	}
	defer f.Close()

	sheets := f.GetSheetList()
	if len(sheets) == 0 {
		return &File{Path: path, Name: name}, nil
	}

	rows, err := f.GetRows(sheets[0])
	if err != nil || len(rows) == 0 {
		return &File{Path: path, Name: name}, nil
	}

	return &File{Path: path, Name: name, Col: rows[0]}, nil
}

/*
=====================
File Filling
=====================
*/

func FillFile(fileA, fileB *File, mappings []ColumnMapping) error {
	if fileA == nil || fileB == nil {
		return fmt.Errorf("fileA or fileB is nil")
	}

	// Open FileA
	fA, err := excelize.OpenFile(fileA.Path)
	if err != nil {
		return fmt.Errorf("open FileA: %v", err)
	}
	defer fA.Close()

	rowsA, err := fA.GetRows(fA.GetSheetList()[0])
	if err != nil || len(rowsA) <= 1 {
		return fmt.Errorf("FileA has no data")
	}

	// Open or create FileB
	var fB *excelize.File
	if _, err := os.Stat(fileB.Path); os.IsNotExist(err) {
		fB = excelize.NewFile()
	} else {
		fB, err = excelize.OpenFile(fileB.Path)
		if err != nil {
			return fmt.Errorf("open FileB: %v", err)
		}
	}
	defer fB.Close()

	sheetB := ensureSheet(fB)

	// Build headers and update FileB
	headers := buildHeaders(fileB.Col, mappings)
	writeHeaders(fB, sheetB, headers)
	fileB.Col = headers
	colIdxB := NewFileIndex(fileB) // use updated headers

	// Fill rows
	for i, row := range rowsA[1:] {
		vals := make([]any, len(fileB.Col)) // slice matches headers length
		for _, m := range mappings {
			idxB, ok := colIdxB.ColMap[strings.ToLower(strings.TrimSpace(m.Target))]
			if !ok {
				continue
			}

			var val any
			if m.Transform != nil {
				val = m.Transform(row, fileA)
			} else if m.Formula != "" {
				val = EvalFormula(m.Formula, m.Source, row, fileA)
			} else if len(m.Source) > 0 {
				srcIdx := ColIndex(m.Source[0], fileA)
				if srcIdx != -1 && srcIdx < len(row) {
					val = row[srcIdx]
				} else {
					val = m.Default
				}
			} else {
				val = m.Default
			}
			vals[idxB] = val
		}

		// Write row to sheet
		err := fB.SetSheetRow(sheetB, fmt.Sprintf("A%d", i+2), &vals)
		if err != nil {
			return fmt.Errorf("writing row %d: %v", i+2, err)
		}
	}

	// Save FileB
	if err := fB.SaveAs(fileB.Path); err != nil {
		return fmt.Errorf("save FileB: %v", err)
	}
	return nil
}

// ensureSheet returns the first sheet or creates one if none exist
func ensureSheet(f *excelize.File) string {
	sheets := f.GetSheetList()
	if len(sheets) > 0 {
		return sheets[0] // existing sheet name
	}
	// Create new sheet
	sheetID, err := f.NewSheet("Sheet1")
	if err != nil {
		// fallback: use default first sheet "Sheet1"
		return "Sheet1"
	}
	return f.GetSheetName(sheetID)
}

// buildHeaders combines existing headers with new mappings
func buildHeaders(existing []string, mappings []ColumnMapping) []string {
	headers := make([]string, len(existing))
	copy(headers, existing)

	headerMap := make(map[string]bool)
	for _, h := range existing {
		headerMap[strings.ToLower(h)] = true
	}

	for _, m := range mappings {
		if _, ok := headerMap[strings.ToLower(m.Target)]; !ok {
			headers = append(headers, m.Target)
			headerMap[strings.ToLower(m.Target)] = true
		}
	}

	return headers
}

// writeHeaders writes the headers to the first row
func writeHeaders(f *excelize.File, sheet string, headers []string) {
	_ = f.SetSheetRow(sheet, "A1", &headers)
}
