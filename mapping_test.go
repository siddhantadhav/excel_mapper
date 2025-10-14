package excel_mapper

import (
	"fmt"
	"path/filepath"
	"testing"

	"github.com/xuri/excelize/v2"
)

func createSampleXLSX(t *testing.T, dir, name string) string {
	t.Helper()
	f := excelize.NewFile()
	sheet := f.GetSheetName(0)
	_ = f.SetSheetRow(sheet, "A1", &[]string{"Name", "Age"})
	_ = f.SetSheetRow(sheet, "A2", &[]string{"Alice", "25"})

	fullPath := filepath.Join(dir, name)
	if err := f.SaveAs(fullPath); err != nil {
		t.Fatalf("failed to create sample xlsx: %v", err)
	}
	return fullPath
}

func TestInitFile(t *testing.T) {
	tmpDir := t.TempDir()
	fileName := "sample.xlsx"
	fullPath := createSampleXLSX(t, tmpDir, fileName)

	file, err := InitFile(tmpDir, fileName)
	if err != nil {
		t.Fatalf("InitFile failed: %v", err)
	}

	if file == nil {
		t.Fatalf("expected non-nil *File")
	}

	if len(file.col) != 2 {
		t.Errorf("expected 2 columns, got %d", len(file.col))
	}

	if file.path != fullPath {
		t.Errorf("expected path %q, got %q", fullPath, file.path)
	}

	if file.name != fileName {
		t.Errorf("expected name %q, got %q", fileName, file.name)
	}

	fmt.Println(file.name)
	fmt.Println(file.path)
	fmt.Println(file.col)
}

func TestInitFile_InvalidExtension(t *testing.T) {
	tmpDir := t.TempDir()
	_, err := InitFile(tmpDir, "test.txt")
	if err == nil {
		t.Fatal("expected error for non-xlsx file, got nil")
	}
}


