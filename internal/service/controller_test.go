package service

import (
	"archive/zip"
	"net"
	"os"
	"path/filepath"
	"strconv"
	"testing"
	"time"

	"CertSend/pkg/logger"
)

// TestMain initializes logger.Log once before any test in this
// package runs, since SendCertificates/loadCertificates call
// logger.Log.Error/Warn internally and a nil *slog.Logger panics when
// a method is called on it.
func TestMain(m *testing.M) {
	dir, err := os.MkdirTemp("", "service-test-logs")
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

// writeTestCSV writes a synthetic CSV file (never real participant
// data) to a temp file and returns its path.
func writeTestCSV(t *testing.T, content string) string {
	t.Helper()
	path := filepath.Join(t.TempDir(), "data.csv")
	if err := os.WriteFile(path, []byte(content), 0644); err != nil {
		t.Fatalf("failed to set up test csv: %v", err)
	}
	return path
}

// unreachableSMTPConfig returns an SMTPConfig pointed at a local TCP
// port that nothing is listening on, so any real send attempt fails
// fast and deterministically without needing a real SMTP server.
func unreachableSMTPConfig(t *testing.T) SMTPConfig {
	t.Helper()

	l, err := net.Listen("tcp", "127.0.0.1:0")
	if err != nil {
		t.Fatalf("failed to reserve a local port: %v", err)
	}
	_, portStr, _ := net.SplitHostPort(l.Addr().String())
	port, _ := strconv.Atoi(portStr)
	l.Close()

	return SMTPConfig{Host: "127.0.0.1", Port: port, Username: "test", Password: "test"}
}

func TestLoadCertificates_DispatchesPDF(t *testing.T) {
	path := filepath.Join(t.TempDir(), "Budi Santoso_Seminar.pdf")
	if err := os.WriteFile(path, []byte("%PDF-1.4 fake content"), 0644); err != nil {
		t.Fatalf("failed to set up test pdf: %v", err)
	}

	certs, err := loadCertificates(path)
	if err != nil {
		t.Fatalf("loadCertificates returned an unexpected error: %v", err)
	}
	if len(certs) != 1 {
		t.Fatalf("got %d certificates, want 1", len(certs))
	}
	if certs[0].FileName != "Budi Santoso_Seminar.pdf" {
		t.Errorf("got FileName %q, want %q", certs[0].FileName, "Budi Santoso_Seminar.pdf")
	}
}

func TestLoadCertificates_DispatchesZip(t *testing.T) {
	zipPath := filepath.Join(t.TempDir(), "certificates.zip")
	zf, err := os.Create(zipPath)
	if err != nil {
		t.Fatalf("failed to create test zip: %v", err)
	}
	w := zip.NewWriter(zf)
	for _, name := range []string{"Budi Santoso_Seminar.pdf", "Ani Wijaya_Seminar.pdf"} {
		fw, err := w.Create(name)
		if err != nil {
			t.Fatalf("failed to write zip entry %q: %v", name, err)
		}
		if _, err := fw.Write([]byte("%PDF-1.4 fake content")); err != nil {
			t.Fatalf("failed to write zip entry content: %v", err)
		}
	}
	if err := w.Close(); err != nil {
		t.Fatalf("failed to close zip writer: %v", err)
	}
	if err := zf.Close(); err != nil {
		t.Fatalf("failed to close zip file: %v", err)
	}

	certs, err := loadCertificates(zipPath)
	if err != nil {
		t.Fatalf("loadCertificates returned an unexpected error: %v", err)
	}
	if len(certs) != 2 {
		t.Fatalf("got %d certificates, want 2", len(certs))
	}
}

func TestLoadCertificates_UnsupportedExtension(t *testing.T) {
	path := filepath.Join(t.TempDir(), "notes.docx")
	if err := os.WriteFile(path, []byte("not a certificate"), 0644); err != nil {
		t.Fatalf("failed to set up test file: %v", err)
	}

	_, err := loadCertificates(path)
	if err == nil {
		t.Fatal("expected an error for an unsupported file extension, got nil")
	}
}

// TestSendCertificates_NoMatchesReturnsError verifies the pipeline
// stops right after matching (before ever touching SMTP) when nothing
// matched, using a garbage SMTPConfig to prove it's never dialed.
func TestSendCertificates_NoMatchesReturnsError(t *testing.T) {
	certPath := filepath.Join(t.TempDir(), "Someone Unrelated_Seminar.pdf")
	if err := os.WriteFile(certPath, []byte("%PDF-1.4 fake content"), 0644); err != nil {
		t.Fatalf("failed to set up test pdf: %v", err)
	}
	csvPath := writeTestCSV(t, "name,nrp,email\nBudi Santoso,123,budi@example.com\n")

	garbageSMTP := SMTPConfig{Host: "", Port: 0}
	err := SendCertificates(certPath, csvPath, garbageSMTP, time.Millisecond)
	if err == nil {
		t.Fatal("expected an error when no certificate matches any recipient, got nil")
	}
}

// TestSendCertificates_PropagatesMailerError verifies that a full,
// successful match still surfaces an error when the SMTP server is
// unreachable, proving the pipeline actually reaches the send step.
func TestSendCertificates_PropagatesMailerError(t *testing.T) {
	certPath := filepath.Join(t.TempDir(), "Budi Santoso_Seminar.pdf")
	if err := os.WriteFile(certPath, []byte("%PDF-1.4 fake content"), 0644); err != nil {
		t.Fatalf("failed to set up test pdf: %v", err)
	}
	csvPath := writeTestCSV(t, "name,nrp,email\nBudi Santoso,123,budi@example.com\n")

	err := SendCertificates(certPath, csvPath, unreachableSMTPConfig(t), time.Millisecond)
	if err == nil {
		t.Fatal("expected an error since the SMTP server is unreachable, got nil")
	}
}
