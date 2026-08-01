package logger

import (
	"os"
	"path/filepath"
	"time"
)

// makeFolderLogs creates the log folder at the given path if it doesn't already exist.
func makeFolderLogs(path string) error {
	// Buat folder logs
	err := os.MkdirAll(path, 0755)
	if err != nil {
		return err
	}
	return nil
}

// MakeFileLogs opens (or creates) today's log file inside the given
// folder path, in append mode, so entries accumulate across the day
// instead of overwriting previous ones. The file is named after the
// current date in Asia/Jakarta time, formatted as YYYY-MM-DD.log.
// The caller is responsible for closing the returned file.
func makeFileLogs(path string) (*os.File, error) {

	// Track waktu berdasarkan lokasi
	loc, err := time.LoadLocation("Asia/Jakarta")
	if err != nil {
		return nil, err
	}
	now := time.Now().In(loc)

	// Open file log
	filename := now.Format("2006-01-02") + ".log"
	file, err := os.OpenFile(
		filepath.Join(path, filename),
		os.O_APPEND|os.O_WRONLY|os.O_CREATE,
		0666,
	)
	if err != nil {
		return nil, err
	}

	return file, err
}
