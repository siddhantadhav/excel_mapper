package excel_mapper

import (
	"strconv"
	"strings"

	"github.com/xuri/excelize/v2"
)

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
	for _, h := range existing {
		headerSet[strings.ToLower(strings.TrimSpace(h))] = struct{}{}
	}
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
	for i, h := range headers {
		vals[i] = h
	}
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
		if count == 0 {
			return 0
		}
		return sum / float64(count)
	}
}

// ParseFloatSafe parses a string into float64, returns 0 on error
func ParseFloatSafe(s string) float64 {
	v, err := strconv.ParseFloat(strings.TrimSpace(s), 64)
	if err != nil {
		return 0
	}
	return v
}
