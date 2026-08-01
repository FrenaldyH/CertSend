package certificate

import (
	"CertSend/pkg/logger"
	"archive/zip"
	"fmt"
	"io"
	"os"
	"path/filepath"
)

// Certificate holds a single certificate file loaded into memory,
// identified by its original file name (including extension), ready
// to be matched against a recipient and attached to an email.
type Certificate struct {
	Name string
	File []byte
}

// InputPDF reads a single PDF certificate file from disk into memory.
func InputPDF(path string) (Certificate, error) {
	file, err := os.ReadFile(path)
	if err != nil {
		msgErr := "failed to read pdf file"
		logger.Log.Error(
			msgErr,
			"pdfPath", path,
			"error", err,
		)
		return Certificate{}, fmt.Errorf("%s : %w", msgErr, err)
	}

	name := filepath.Base(path)

	certificate := Certificate{
		Name: name,
		File: file,
	}

	return certificate, nil
}

// InputZip opens a ZIP archive and reads every certificate file inside
// it into memory. Directory entries are skipped. If an individual
// entry fails to open or read, it is logged and skipped rather than
// failing the whole batch.
func InputZip(path string) ([]Certificate, error) {
	folder, err := zip.OpenReader(path)
	if err != nil {
		msgErr := "failed to read zip file"
		logger.Log.Error(
			msgErr,
			"zipPath", path,
			"error", err,
		)
		return nil, fmt.Errorf("%s : %w", msgErr, err)
	}
	defer folder.Close()

	var certificates []Certificate
	for idxRow, f := range folder.File {
		if f.FileInfo().IsDir() {
			dbgMsg := fmt.Sprintf("row %d is a folder not a file", idxRow)
			logger.Log.Debug(
				dbgMsg,
			)
			continue
		}

		folderChunk, err := f.Open()
		if err != nil {
			msgErr := "failed to open file on the zip folder"
			logger.Log.Error(
				msgErr,
				"index", idxRow,
				"error", err,
			)
			continue
		}

		content, err := io.ReadAll(folderChunk)
		if err != nil {
			msgErr := "failed to read file on the zip folder"
			logger.Log.Error(
				msgErr,
				"index", idxRow,
				"error", err,
			)
			continue
		}
		folderChunk.Close()

		name := filepath.Base(f.Name)

		certificates = append(
			certificates,
			Certificate{
				Name: name,
				File: content,
			},
		)
	}
	return certificates, nil
}
