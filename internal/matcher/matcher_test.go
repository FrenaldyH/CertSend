package matcher

import (
	"os"
	"testing"

	"github.com/FrenaldyH/CertSend/pkg/logger"

	"github.com/FrenaldyH/CertSend/internal/csvmap"

	"github.com/FrenaldyH/CertSend/internal/certificate"
)

// TestMain initializes logger.Log once before any test in this
// package runs, since Matcher calls logger.Log.Warn internally and a
// nil *slog.Logger panics when a method is called on it.
func TestMain(m *testing.M) {
	dir, err := os.MkdirTemp("", "matcher-test-logs")
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

// TestMatcher_Success verifies the happy path: a certificate whose
// file name follows the "{Name}_{anything}.ext" convention gets
// paired with the matching CSV entry, and every Data field is filled
// in correctly.
func TestMatcher_Success(t *testing.T) {
	entries := []csvmap.Entry{
		{Name: "Budi Santoso", NRP: "123", Email: "budi@mail.com"},
	}
	certs := []certificate.Certificate{
		{FileName: "Budi Santoso_Peserta Seminar Dosen Vol 2.pdf", File: []byte("%PDF-1.4 budi")},
	}

	datas, err := Matcher(entries, certs)
	if err != nil {
		t.Fatalf("Matcher returned an unexpected error: %v", err)
	}
	if len(datas) != 1 {
		t.Fatalf("got %d results, want 1", len(datas))
	}

	want := Data{
		Email:      "budi@mail.com",
		PersonName: "Budi Santoso",
		FileName:   "Budi Santoso_Peserta Seminar Dosen Vol 2.pdf",
		FileData:   []byte("%PDF-1.4 budi"),
	}
	got := datas[0]
	if got.Email != want.Email || got.PersonName != want.PersonName || got.FileName != want.FileName || string(got.FileData) != string(want.FileData) {
		t.Errorf("got %+v, want %+v", got, want)
	}
}

// TestMatcher_CaseAndWhitespaceInsensitive is a regression test for
// the case-sensitivity and stray-whitespace bugs found during review:
// a CSV name with extra surrounding spaces must still match a
// certificate file name with different letter casing.
func TestMatcher_CaseAndWhitespaceInsensitive(t *testing.T) {
	entries := []csvmap.Entry{
		{Name: "  Budi Santoso  ", Email: "budi@mail.com"},
	}
	certs := []certificate.Certificate{
		{FileName: "BUDI SANTOSO_Peserta Seminar Dosen Vol 2.pdf", File: []byte("content")},
	}

	datas, err := Matcher(entries, certs)
	if err != nil {
		t.Fatalf("Matcher returned an unexpected error: %v", err)
	}
	if len(datas) != 1 {
		t.Fatalf("got %d results, want 1 (matching should ignore case and surrounding whitespace)", len(datas))
	}
}

// TestMatcher_SkipsCertificateWithoutUnderscore verifies that a
// certificate file name with no underscore (an unexpected format) is
// logged and skipped instead of being guessed at or causing an error.
func TestMatcher_SkipsCertificateWithoutUnderscore(t *testing.T) {
	entries := []csvmap.Entry{
		{Name: "Budi Santoso", Email: "budi@mail.com"},
	}
	certs := []certificate.Certificate{
		{FileName: "Budi Santoso.pdf", File: []byte("content")}, // no underscore
	}

	datas, err := Matcher(entries, certs)
	if err != nil {
		t.Fatalf("Matcher returned an unexpected error: %v", err)
	}
	if len(datas) != 0 {
		t.Fatalf("got %d results, want 0 (file name without an underscore should be skipped)", len(datas))
	}
}

// TestMatcher_SkipsCertificateWithNoMatchingEntry verifies that a
// certificate with a well-formed name but no corresponding CSV entry
// is skipped rather than causing Matcher to fail entirely.
func TestMatcher_SkipsCertificateWithNoMatchingEntry(t *testing.T) {
	entries := []csvmap.Entry{
		{Name: "Budi Santoso", Email: "budi@mail.com"},
	}
	certs := []certificate.Certificate{
		{FileName: "Someone Else_Peserta Seminar Dosen Vol 2.pdf", File: []byte("content")},
	}

	datas, err := Matcher(entries, certs)
	if err != nil {
		t.Fatalf("Matcher returned an unexpected error: %v", err)
	}
	if len(datas) != 0 {
		t.Fatalf("got %d results, want 0 (a certificate with no matching entry should be skipped)", len(datas))
	}
}

// TestMatcher_MultipleCertificatesMixedResults runs a batch with a
// mix of matchable, unmatchable, and malformed certificates, and
// verifies that only the genuinely matched ones make it into the
// result while the rest are skipped without affecting each other.
func TestMatcher_MultipleCertificatesMixedResults(t *testing.T) {
	entries := []csvmap.Entry{
		{Name: "Budi Santoso", Email: "budi@mail.com"},
		{Name: "Ani Wijaya", Email: "ani@mail.com"},
	}
	certs := []certificate.Certificate{
		{FileName: "Budi Santoso_Peserta Seminar Dosen Vol 2.pdf", File: []byte("budi content")},
		{FileName: "Ani Wijaya_Peserta Seminar Dosen Vol 2.pdf", File: []byte("ani content")},
		{FileName: "Unknown Person_Peserta Seminar Dosen Vol 2.pdf", File: []byte("unknown content")},
		{FileName: "no-underscore-here.pdf", File: []byte("bad format")},
	}

	datas, err := Matcher(entries, certs)
	if err != nil {
		t.Fatalf("Matcher returned an unexpected error: %v", err)
	}
	if len(datas) != 2 {
		t.Fatalf("got %d results, want 2 (only Budi and Ani should match)", len(datas))
	}

	gotEmails := map[string]bool{}
	for _, d := range datas {
		gotEmails[d.Email] = true
	}
	if !gotEmails["budi@mail.com"] || !gotEmails["ani@mail.com"] {
		t.Errorf("expected both budi@mail.com and ani@mail.com in results, got %+v", datas)
	}
}
