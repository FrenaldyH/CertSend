package csvmap

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/FrenaldyH/CertSend/pkg/logger"
)

// TestMain initializes logger.Log once before any test in this
// package runs, since ParseCSV calls logger.Log.Error/Warn internally
// and a nil *slog.Logger panics when a method is called on it.
func TestMain(m *testing.M) {
	dir, err := os.MkdirTemp("", "csvmap-test-logs")
	if err != nil {
		panic(err)
	}
	if err := logger.InitLog(dir); err != nil {
		panic(err)
	}

	code := m.Run()
	os.RemoveAll(dir)
	os.Exit(code)
}

// writeTestCSV is a test helper that writes content to a temp CSV
// file and returns its path.
func writeTestCSV(t *testing.T, content string) string {
	t.Helper()

	dir := t.TempDir()
	path := filepath.Join(dir, "data.csv")
	if err := os.WriteFile(path, []byte(content), 0644); err != nil {
		t.Fatalf("failed to set up test csv: %v", err)
	}
	return path
}

func TestParseCSV_Success(t *testing.T) {
	path := writeTestCSV(t, "name,nrp,email\nBudi Santoso,123,budi@mail.com\nAni Wijaya,456,ani@mail.com\n")

	entries, err := ParseCSV(path)
	if err != nil {
		t.Fatalf("ParseCSV returned an unexpected error: %v", err)
	}
	if len(entries) != 2 {
		t.Fatalf("got %d entries, want 2", len(entries))
	}

	want := []Entry{
		{Name: "Budi Santoso", NRP: "123", Email: "budi@mail.com"},
		{Name: "Ani Wijaya", NRP: "456", Email: "ani@mail.com"},
	}
	for i, w := range want {
		if entries[i] != w {
			t.Errorf("entry %d = %+v, want %+v", i, entries[i], w)
		}
	}
}

func TestParseCSV_TrimsWhitespace(t *testing.T) {
	path := writeTestCSV(t, "name,nrp,email\n  Budi Santoso  , 123 , budi@mail.com \n")

	entries, err := ParseCSV(path)
	if err != nil {
		t.Fatalf("ParseCSV returned an unexpected error: %v", err)
	}
	want := Entry{Name: "Budi Santoso", NRP: "123", Email: "budi@mail.com"}
	if entries[0] != want {
		t.Errorf("got %+v, want %+v (whitespace should be trimmed)", entries[0], want)
	}
}

func TestParseCSV_ColumnOrderIndependent(t *testing.T) {
	path := writeTestCSV(t, "email,name,nrp\nbudi@mail.com,Budi Santoso,123\n")

	entries, err := ParseCSV(path)
	if err != nil {
		t.Fatalf("ParseCSV returned an unexpected error: %v", err)
	}
	want := Entry{Name: "Budi Santoso", NRP: "123", Email: "budi@mail.com"}
	if entries[0] != want {
		t.Errorf("got %+v, want %+v (column order shouldn't matter)", entries[0], want)
	}
}

func TestParseCSV_StripsBOM(t *testing.T) {
	path := writeTestCSV(t, "\ufeffname,nrp,email\nBudi Santoso,123,budi@mail.com\n")

	entries, err := ParseCSV(path)
	if err != nil {
		t.Fatalf("ParseCSV returned an unexpected error: %v", err)
	}
	if len(entries) != 1 {
		t.Fatalf("got %d entries, want 1 (a BOM on the header shouldn't break parsing)", len(entries))
	}
	want := Entry{Name: "Budi Santoso", NRP: "123", Email: "budi@mail.com"}
	if entries[0] != want {
		t.Errorf("got %+v, want %+v", entries[0], want)
	}
}

func TestParseCSV_FileNotFound(t *testing.T) {
	_, err := ParseCSV(filepath.Join(t.TempDir(), "missing.csv"))
	if err == nil {
		t.Fatal("expected an error for a missing file, got nil")
	}
}

func TestParseCSV_EmptyFile(t *testing.T) {
	path := writeTestCSV(t, "")

	_, err := ParseCSV(path)
	if err == nil {
		t.Fatal("expected an error for an empty csv file, got nil")
	}
}

func TestParseCSV_MissingColumns(t *testing.T) {
	path := writeTestCSV(t, "name,email\nBudi Santoso,budi@mail.com\n")

	_, err := ParseCSV(path)
	if err == nil {
		t.Fatal("expected an error when the nrp column is missing, got nil")
	}
}
