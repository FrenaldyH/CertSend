package logger

import (
	"encoding/json"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

// TestInitLog_SetsUpLogger verifies that InitLog succeeds and leaves
// the package-level Log variable ready to use.
func TestInitLog_SetsUpLogger(t *testing.T) {
	dir := t.TempDir()

	if err := InitLog(dir); err != nil {
		t.Fatalf("InitLog returned an unexpected error: %v", err)
	}

	if Log == nil {
		t.Fatal("expected Log to be initialized, got nil")
	}
}

// TestInitLog_WritesJSONToFile verifies that log entries written
// through Log actually end up in today's log file, encoded as JSON,
// and include the source file/line thanks to AddSource.
func TestInitLog_WritesJSONToFile(t *testing.T) {
	dir := t.TempDir()

	if err := InitLog(dir); err != nil {
		t.Fatalf("InitLog returned an unexpected error: %v", err)
	}

	Log.Info("hello from test")

	entries, err := os.ReadDir(dir)
	if err != nil {
		t.Fatalf("failed to read log dir: %v", err)
	}
	if len(entries) != 1 {
		t.Fatalf("expected exactly 1 log file, found %d", len(entries))
	}

	content, err := os.ReadFile(filepath.Join(dir, entries[0].Name()))
	if err != nil {
		t.Fatalf("failed to read log file: %v", err)
	}

	line := strings.TrimSpace(string(content))
	var record map[string]any
	if err := json.Unmarshal([]byte(line), &record); err != nil {
		t.Fatalf("log line is not valid JSON: %v\nline: %s", err, line)
	}

	if record["msg"] != "hello from test" {
		t.Errorf("got msg %q, want %q", record["msg"], "hello from test")
	}
	if _, ok := record["source"]; !ok {
		t.Error("expected a \"source\" field in the log record (AddSource), but it was missing")
	}
}

// TestInitLog_PropagatesError verifies that InitLog returns an error
// (instead of panicking) when the underlying log file/folder can't
// be created, e.g. because the path is already occupied by a file.
func TestInitLog_PropagatesError(t *testing.T) {
	base := t.TempDir()
	blockedPath := filepath.Join(base, "logs")

	if err := os.WriteFile(blockedPath, []byte("i'm a file, not a folder"), 0644); err != nil {
		t.Fatalf("failed to set up test file: %v", err)
	}

	if err := InitLog(blockedPath); err == nil {
		t.Fatal("expected an error when path is occupied by a file, got nil")
	}
}
