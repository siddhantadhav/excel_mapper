package excel_mapper

import (
	"fmt"
	"os"
	"path/filepath"
	"strconv"
	"strings"

	"github.com/xuri/excelize/v2"
)

// File represents an Excel file
type File struct {
	Path string
	Name string
	Col  []string
}

// MappingFunc generates a value for a FileB column
type MappingFunc func(row []string, fA *File) interface{}

// ColumnMapping represents a mapping rule
type ColumnMapping struct {
	Target    string        // Column in FileB
	Source    []string      // Columns in FileA
	Transform MappingFunc   // Optional transform
	Default   interface{}   // Default value if missing
}

// FileIndex maps lowercase trimmed column names to indices
type FileIndex struct {
	ColMap map[string]int
	File   *File
}

// NewFileIndex creates a column index map
func NewFileIndex(f *File) *FileIndex {
	m := make(map[string]int)
	for i, col := range f.Col {
		m[strings.ToLower(strings.TrimSpace(col))] = i
	}
	return &FileIndex{ColMap: m, File: f}
}

// ColIndex returns the index of a column in a file
func ColIndex(name string, f *File) int {
	if f == nil { return -1 }
	idx, ok := NewFileIndex(f).ColMap[strings.ToLower(strings.TrimSpace(name))]
	if !ok { return -1 }
	return idx
}

// InitFile initializes a File, reading headers if it exists
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
	if err != nil { return nil, fmt.Errorf("open file: %v", err) }
	defer f.Close()

	sheets := f.GetSheetList()
	if len(sheets) == 0 { return &File{Path: path, Name: name}, nil }

	rows, err := f.GetRows(sheets[0])
	if err != nil || len(rows) == 0 { return &File{Path: path, Name: name}, nil }

	return &File{Path: path, Name: name, Col: rows[0]}, nil
}

// ParseFloatSafe parses a string into float64, returns 0 on error
func ParseFloatSafe(s string) float64 {
	v, err := strconv.ParseFloat(strings.TrimSpace(s), 64)
	if err != nil { return 0 }
	return v
}

// FillFile populates FileB from FileA using ColumnMapping rules
func FillFile(fileA, fileB *File, mappings []ColumnMapping, output string) error {
	if fileA == nil || fileB == nil {
		return fmt.Errorf("fileA or fileB is nil")
	}

	// Open FileA
	fA, err := excelize.OpenFile(fileA.Path)
	if err != nil { return fmt.Errorf("open FileA: %v", err) }
	defer fA.Close()

	rowsA, err := fA.GetRows(fA.GetSheetList()[0])
	if err != nil || len(rowsA) <= 1 { return fmt.Errorf("FileA has no data") }

	// Open or create FileB
	var fB *excelize.File
	if _, err := os.Stat(fileB.Path); os.IsNotExist(err) {
		fB = excelize.NewFile()
	} else {
		fB, err = excelize.OpenFile(fileB.Path)
		if err != nil { return fmt.Errorf("open FileB: %v", err) }
	}
	defer fB.Close()

	sheetB := ensureSheet(fB)

	// Determine headers
	headers := buildHeaders(fileB.Col, mappings)
	writeHeaders(fB, sheetB, headers)
	fileB.Col = headers

	colIdxB := NewFileIndex(&File{Col: headers})

	// Fill rows
	for i, row := range rowsA[1:] {
		vals := make([]interface{}, len(colIdxB.ColMap))
		for _, m := range mappings {
			idxB, ok := colIdxB.ColMap[strings.ToLower(strings.TrimSpace(m.Target))]
			if !ok { continue }

			var val interface{}
			if m.Transform != nil {
				val = m.Transform(row, fileA) // transform can use ColIndex directly
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
		_ = fB.SetSheetRow(sheetB, fmt.Sprintf("A%d", i+2), &vals)
	}

	if err := fB.SaveAs(output); err != nil {
		return fmt.Errorf("save FileB: %v", err)
	}
	return nil
}

// Helper: ensure at least one sheet exists
func ensureSheet(f *excelize.File) string {
	sheets := f.GetSheetList()
	if len(sheets) == 0 {
		name := "Sheet1"
		f.NewSheet(name)
		f.SetActiveSheet(0)
		return name
	}
	return sheets[0]
}

// Helper: build final headers
func buildHeaders(existing []string, mappings []ColumnMapping) []string {
	headerSet := make(map[string]struct{})
	for _, h := range existing { headerSet[strings.ToLower(strings.TrimSpace(h))] = struct{}{} }
	headers := append([]string{}, existing...)
	for _, m := range mappings {
		key := strings.TrimSpace(m.Target)
		if _, exists := headerSet[strings.ToLower(key)]; !exists {
			headers = append(headers, key)
		}
	}
	return headers
}

// Helper: write header row
func writeHeaders(f *excelize.File, sheet string, headers []string) {
	vals := make([]interface{}, len(headers))
	for i, h := range headers { vals[i] = h }
	_ = f.SetSheetRow(sheet, "A1", &vals)
}

// Transform helpers
func SumColumns(cols ...string) MappingFunc {
	return func(row []string, fA *File) interface{} {
		sum := 0.0
		for _, c := range cols {
			idx := ColIndex(c, fA)
			if idx != -1 && idx < len(row) {
				sum += ParseFloatSafe(row[idx])
			}
		}
		return sum
	}
}

func ConcatColumns(sep string, cols ...string) MappingFunc {
	return func(row []string, fA *File) interface{} {
		vals := []string{}
		for _, c := range cols {
			idx := ColIndex(c, fA)
			if idx != -1 && idx < len(row) && row[idx] != "" {
				vals = append(vals, row[idx])
			}
		}
		return strings.Join(vals, sep)
	}
}

func AverageColumns(cols ...string) MappingFunc {
	return func(row []string, fA *File) interface{} {
		sum := 0.0
		count := 0
		for _, c := range cols {
			idx := ColIndex(c, fA)
			if idx != -1 && idx < len(row) && row[idx] != "" {
				sum += ParseFloatSafe(row[idx])
				count++
			}
		}
		if count == 0 { return 0 }
		return sum / float64(count)
	}
}
