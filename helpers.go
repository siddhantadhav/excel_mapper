package excel_mapper

import (
	"fmt"
	"strconv"
	"strings"

	"github.com/Knetic/govaluate"
)

/*
=====================
Transform Helpers
=====================
*/

func ParseFloatSafe(v string) float64 {
	f, _ := strconv.ParseFloat(v, 64)
	return f
}

func SumColumns(cols ...string) MappingFunc {
	return func(row []string, f *File) any {
		sum := 0.0
		for _, c := range cols {
			idx := ColIndex(c, f)
			if idx >= 0 && idx < len(row) {
				sum += ParseFloatSafe(row[idx])
			}
		}
		return sum
	}
}

func AverageColumns(cols ...string) MappingFunc {
	return func(row []string, f *File) any {
		sum := 0.0
		count := 0
		for _, c := range cols {
			idx := ColIndex(c, f)
			if idx >= 0 && idx < len(row) {
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

func ConcatColumns(sep string, cols ...string) MappingFunc {
	return func(row []string, f *File) any {
		var parts []string
		for _, c := range cols {
			idx := ColIndex(c, f)
			if idx >= 0 && idx < len(row) {
				parts = append(parts, row[idx])
			}
		}
		return strings.Join(parts, sep)
	}
}

func Unique(sourceCol string) MappingFunc {
	seen := map[string]bool{}

	return func(row []string, f *File) any {
		idx := ColIndex(sourceCol, f)
		if idx < 0 || idx >= len(row) {
			return ""
		}

		val := row[idx]

		if seen[val] {
			return ""
		}

		seen[val] = true
		return val
	}
}

func Count(sources ...string) MappingFunc {
	counts := map[string]int{}
	return func(row []string, f *File) any {
		parts := make([]string, len(sources))
		for i, col := range sources {
			idx := ColIndex(col, f)
			if idx >= 0 && idx < len(row) {
				parts[i] = row[idx]
			} else {
				parts[i] = ""
			}
		}
		key := strings.Join(parts, "|")
		counts[key]++
		return fmt.Sprintf("%s=%d", key, counts[key])
	}
}

func EvalFormula(formula string, sources []string, row []string, f *File) any {
	expr, err := govaluate.NewEvaluableExpression(formula)
	if err != nil {
		return fmt.Sprintf("invalid formula: %s", formula)
	}
	params := make(map[string]any)
	for _, c := range f.Col {
		idx := ColIndex(c, f)
		if idx >= 0 && idx < len(row) {
			params[c] = ParseFloatSafe(row[idx])
		}
	}
	result, err := expr.Evaluate(params)
	if err != nil {
		return fmt.Sprintf("error: %v", err)
	}
	return result
}

/*
=====================
Conversions
=====================
*/

func (dbMap *DBColumnMapping) ToColumnMapping() ColumnMapping {
	var transform MappingFunc
	switch dbMap.Transform {
	case "sum":
		transform = SumColumns(dbMap.Source...)
	case "concat":
		sep := " "
		if s, ok := dbMap.Params["sep"].(string); ok {
			sep = s
		}
		transform = ConcatColumns(sep, dbMap.Source...)
	case "average":
		transform = AverageColumns(dbMap.Source...)
	case "unique":
		transform = Unique(dbMap.Source[0])
	case "count":
		transform = Count(dbMap.Source...)
	case "raw":
		formula := dbMap.Formula
		transform = func(row []string, f *File) any {
			return EvalFormula(formula, dbMap.Source, row, f)
		}
	default:
		transform = nil
	}
	return ColumnMapping{
		Target:    dbMap.Target,
		Source:    dbMap.Source,
		Transform: transform,
		Formula:   dbMap.Formula,
		Params:    dbMap.Params,
		Default:   dbMap.Default,
	}
}

func (m *ColumnMapping) ToDBMapping() DBColumnMapping {
	var transformType string
	switch {
	case m.Transform == nil && m.Formula == "":
		transformType = "none"
	case m.Formula != "":
		transformType = "raw"
	default:
		transformType = "custom"
	}
	return DBColumnMapping{
		Target:    m.Target,
		Source:    m.Source,
		Transform: transformType,
		Formula:   m.Formula,
		Params:    m.Params,
		Default:   m.Default,
	}
}
