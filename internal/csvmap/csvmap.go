package csvmap

import (
	"CertSend/pkg/logger"
	"encoding/csv"
	"fmt"
	"os"
	"strings"
)

// Entry represents one row of the CSV mapping: a participant's name,
// NRP, and email address.
type Entry struct {
	Name  string
	NRP   string
	Email string
}

// ParseCSV reads the CSV file at pathCSV and converts each data row
// into an Entry. The header row must contain "name", "nrp", and
// "email" columns (case-insensitive, in any order); a UTF-8 BOM on
// the first header cell (common in Excel-exported CSVs) is stripped
// automatically.
func ParseCSV(pathCSV string) ([]Entry, error) {
	fileCSV, err := os.Open(pathCSV)
	if err != nil {
		errMsg := "failed to open csv file"
		logger.Log.Error(
			errMsg,
			"csvPath", pathCSV,
			"error", err,
		)
		return nil, fmt.Errorf("%s : %w", errMsg, err)
	}
	defer fileCSV.Close()

	reader := csv.NewReader(fileCSV)
	records, err := reader.ReadAll()
	if err != nil {
		errMsg := "failed to read csv file"
		logger.Log.Error(
			errMsg,
			"csvPath", pathCSV,
			"error", err,
		)
		return nil, fmt.Errorf("%s : %w", errMsg, err)
	}
	if len(records) == 0 {
		errMsg := "csv file is empty"
		logger.Log.Warn(
			errMsg,
			"csvPath", pathCSV,
		)
		return nil, fmt.Errorf("%s : %s", errMsg, pathCSV)
	}

	header := records[0]
	header[0] = strings.TrimPrefix(header[0], "\ufeff")

	// Pengecekan letak index masing2 kolom header
	idxName, idxNRP, idxEmail := -1, -1, -1
	for i, col := range header {
		switch strings.ToLower(strings.TrimSpace(col)) {
		case "name":
			idxName = i
		case "nrp":
			idxNRP = i
		case "email":
			idxEmail = i
		}
	}
	if idxName == -1 || idxNRP == -1 || idxEmail == -1 {
		errMsg := "csv header not found (name/nrp/email)"
		logger.Log.Warn(
			errMsg,
			"csvPath", pathCSV,
		)
		return nil, fmt.Errorf("%s : %s", errMsg, pathCSV)
	}

	// Ekstrak masing2 row ke dalam array entries
	var entries []Entry
	for _, row := range records[1:] {
		entries = append(entries, Entry{
			Name:  strings.TrimSpace(row[idxName]),
			NRP:   strings.TrimSpace(row[idxNRP]),
			Email: strings.TrimSpace(row[idxEmail]),
		})
	}

	return entries, nil
}
