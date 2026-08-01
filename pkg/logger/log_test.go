package logger

import (
	"os"
	"path/filepath"
	"testing"
	"time"
)

// TestMakeFolderLogs_CreatesFolder verifies that calling makeFolderLogs
// actually creates the target folder on disk.
func TestMakeFolderLogs_CreatesFolder(t *testing.T) {
	base := t.TempDir()
	logsDir := filepath.Join(base, "logs")

	if err := makeFolderLogs(logsDir); err != nil {
		t.Fatalf("makeFolderLogs returned an unexpected error: %v", err)
	}

	info, err := os.Stat(logsDir)
	if err != nil {
		t.Fatalf("expected folder to exist after makeFolderLogs, but Stat failed: %v", err)
	}
	if !info.IsDir() {
		t.Fatalf("expected %q to be a directory, but it is not", logsDir)
	}
}

// TestMakeFolderLogs_Idempotent verifies that calling makeFolderLogs
// multiple times on the same path never errors, even though the
// folder already exists after the first call.
func TestMakeFolderLogs_Idempotent(t *testing.T) {
	logsDir := filepath.Join(t.TempDir(), "logs")

	if err := makeFolderLogs(logsDir); err != nil {
		t.Fatalf("first call failed: %v", err)
	}
	if err := makeFolderLogs(logsDir); err != nil {
		t.Fatalf("second call on an already-existing folder failed: %v", err)
	}
}

// TestMakeFolderLogs_PathIsAFile verifies that makeFolderLogs returns
// an error (instead of panicking) when the target path is already
// occupied by a regular file.
func TestMakeFolderLogs_PathIsAFile(t *testing.T) {
	base := t.TempDir()
	blockedPath := filepath.Join(base, "logs")

	if err := os.WriteFile(blockedPath, []byte("i'm a file, not a folder"), 0644); err != nil {
		t.Fatalf("failed to set up test file: %v", err)
	}

	if err := makeFolderLogs(blockedPath); err == nil {
		t.Fatal("expected an error when path is occupied by a file, got nil")
	}
}

// TestMakeFileLogs_CreatesFileWithCorrectName verifies that the log
// file created by MakeFileLogs is named after today's date in
// Asia/Jakarta time, using the YYYY-MM-DD.log format.
func TestMakeFileLogs_CreatesFileWithCorrectName(t *testing.T) {
	dir := t.TempDir()

	file, err := makeFileLogs(dir)
	if err != nil {
		t.Fatalf("MakeFileLogs returned an unexpected error: %v", err)
	}
	defer file.Close()

	loc, err := time.LoadLocation("Asia/Jakarta")
	if err != nil {
		t.Fatalf("failed to load Asia/Jakarta location: %v", err)
	}
	wantName := time.Now().In(loc).Format("2006-01-02") + ".log"
	wantPath := filepath.Join(dir, wantName)

	if file.Name() != wantPath {
		t.Errorf("got log file path %q, want %q", file.Name(), wantPath)
	}
}

// TestMakeFileLogs_AppendsAcrossCalls is a regression test for a bug
// where the log file was opened with O_TRUNC, wiping out previous
// entries every time MakeFileLogs was called. It verifies that
// writing, closing, then calling MakeFileLogs again preserves the
// earlier content instead of erasing it.
func TestMakeFileLogs_AppendsAcrossCalls(t *testing.T) {
	dir := t.TempDir()

	file1, err := makeFileLogs(dir)
	if err != nil {
		t.Fatalf("first MakeFileLogs call failed: %v", err)
	}
	if _, err := file1.WriteString("first line\n"); err != nil {
		t.Fatalf("failed to write first line: %v", err)
	}
	if err := file1.Close(); err != nil {
		t.Fatalf("failed to close file after first write: %v", err)
	}

	file2, err := makeFileLogs(dir)
	if err != nil {
		t.Fatalf("second MakeFileLogs call failed: %v", err)
	}
	if _, err := file2.WriteString("second line\n"); err != nil {
		t.Fatalf("failed to write second line: %v", err)
	}
	if err := file2.Close(); err != nil {
		t.Fatalf("failed to close file after second write: %v", err)
	}

	content, err := os.ReadFile(file2.Name())
	if err != nil {
		t.Fatalf("failed to read back log file: %v", err)
	}

	got := string(content)
	want := "first line\nsecond line\n"
	if got != want {
		t.Errorf("log file content = %q, want %q (previous entries were likely overwritten)", got, want)
	}
}
