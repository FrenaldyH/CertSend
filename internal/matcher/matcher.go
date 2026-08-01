package matcher

import (
	"CertSend/internal/certificate"
	"CertSend/internal/csvmap"
	"CertSend/pkg/logger"
	"strings"
)

// Data pairs one certificate file with its recipient, ready to be
// handed off to the mailer.
type Data struct {
	Email      string
	PersonName string
	FileName   string
	FileData   []byte
}

// Matcher pairs each certificate with its recipient from entries by
// comparing names. A certificate's file name is expected to follow
// the "{Person Name}_{anything}.ext" convention (e.g.
// "Budi Santoso_Peserta Seminar Dosen Vol 2.pdf"); the part before
// the first underscore is compared, case- and whitespace-insensitively,
// against each Entry.Name.
//
// Certificates whose file name doesn't contain an underscore, or
// whose derived name has no matching Entry, are logged and skipped
// rather than failing the whole batch — the returned slice only
// contains successful pairs.
func Matcher(entries []csvmap.Entry, certificates []certificate.Certificate) ([]Data, error) {
	lookup := make(map[string]csvmap.Entry)
	for _, e := range entries {
		keyName := strings.ToLower(
			strings.TrimSpace(e.Name),
		)
		lookup[keyName] = e
	}

	var datas []Data
	for _, c := range certificates {
		cPersonName, _, found := strings.Cut(c.FileName, "_")
		if !found {
			msgWarn := "unexpected certificate format name"
			logger.Log.Warn(
				msgWarn,
				"certificateName", c.FileName,
			)
			continue
		}

		certKey := strings.ToLower(
			strings.TrimSpace(cPersonName),
		)

		person, found := lookup[certKey]
		if !found {
			msgWarn := "person not found"
			logger.Log.Warn(
				msgWarn,
				"name", cPersonName,
				"fileName", c.FileName,
			)
			continue
		}

		datas = append(
			datas,
			Data{
				Email:      person.Email,
				PersonName: person.Name,
				FileName:   c.FileName,
				FileData:   c.File,
			},
		)
	}

	return datas, nil
}
