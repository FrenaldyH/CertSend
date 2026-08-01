package certificate

import (
	"CertSend/pkg/logger"
	"archive/zip"
	"fmt"
	"io"
	"os"
	"path/filepath"
)

type Certificate struct {
	Name string
	File []byte
}

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
