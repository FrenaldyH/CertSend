package certificate

import (
	"archive/zip"
	"os"
	"path/filepath"
	"testing"

	"github.com/FrenaldyH/CertSend/pkg/logger"
)

// TestMain initializes logger.Log once before any test in this
// package runs, since InputPDF/InputZip call logger.Log.Error/Debug
// internally and a nil *slog.Logger panics when a method is called on it.
func TestMain(m *testing.M) {
	dir, err := os.MkdirTemp("", "certificate-test-logs")
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

// writeTestZip is a test helper that creates a zip file at zipPath
// containing one entry per key/value pair in entries.
func writeTestZip(t *testing.T, zipPath string, entries map[string]string) {
	t.Helper()

	zf, err := os.Create(zipPath)
	if err != nil {
		t.Fatalf("failed to create zip file: %v", err)
	}
	defer zf.Close()

	w := zip.NewWriter(zf)
	defer w.Close()

	for name, content := range entries {
		fw, err := w.Create(name)
		if err != nil {
			t.Fatalf("failed to create zip entry %q: %v", name, err)
		}
		if _, err := fw.Write([]byte(content)); err != nil {
			t.Fatalf("failed to write zip entry %q: %v", name, err)
		}
	}
}

func TestInputPDF_Success(t *testing.T) {
	dir := t.TempDir()
	pdfPath := filepath.Join(dir, "budi.pdf")
	want := []byte("%PDF-1.4 fake certificate content")
	if err := os.WriteFile(pdfPath, want, 0644); err != nil {
		t.Fatalf("failed to set up test pdf: %v", err)
	}

	cert, err := InputPDF(pdfPath)
	if err != nil {
		t.Fatalf("InputPDF returned an unexpected error: %v", err)
	}

	if cert.FileName != "budi.pdf" {
		t.Errorf("got Name %q, want %q", cert.FileName, "budi.pdf")
	}
	if string(cert.File) != string(want) {
		t.Errorf("got File %q, want %q", cert.File, want)
	}
}

func TestInputPDF_FileNotFound(t *testing.T) {
	_, err := InputPDF(filepath.Join(t.TempDir(), "missing.pdf"))
	if err == nil {
		t.Fatal("expected an error for a missing file, got nil")
	}
}

func TestInputZip_Success(t *testing.T) {
	dir := t.TempDir()
	zipPath := filepath.Join(dir, "certificates.zip")

	entries := map[string]string{
		"budi.pdf": "%PDF-1.4 budi certificate",
		"ani.pdf":  "%PDF-1.4 ani certificate",
	}
	writeTestZip(t, zipPath, entries)

	certs, err := InputZip(zipPath)
	if err != nil {
		t.Fatalf("InputZip returned an unexpected error: %v", err)
	}
	if len(certs) != len(entries) {
		t.Fatalf("got %d certificates, want %d", len(certs), len(entries))
	}

	got := map[string]string{}
	for _, c := range certs {
		got[c.FileName] = string(c.File)
	}
	for name, content := range entries {
		if got[name] != content {
			t.Errorf("certificate %q: got content %q, want %q", name, got[name], content)
		}
	}
}

func TestInputZip_SkipsDirectoryEntries(t *testing.T) {
	dir := t.TempDir()
	zipPath := filepath.Join(dir, "with-folder.zip")

	zf, err := os.Create(zipPath)
	if err != nil {
		t.Fatalf("failed to create zip file: %v", err)
	}
	w := zip.NewWriter(zf)
	if _, err := w.Create("subfolder/"); err != nil {
		t.Fatalf("failed to write directory entry: %v", err)
	}
	fw, err := w.Create("subfolder/budi.pdf")
	if err != nil {
		t.Fatalf("failed to write file entry: %v", err)
	}
	if _, err := fw.Write([]byte("%PDF-1.4 budi certificate")); err != nil {
		t.Fatalf("failed to write file content: %v", err)
	}
	if err := w.Close(); err != nil {
		t.Fatalf("failed to close zip writer: %v", err)
	}
	if err := zf.Close(); err != nil {
		t.Fatalf("failed to close zip file: %v", err)
	}

	certs, err := InputZip(zipPath)
	if err != nil {
		t.Fatalf("InputZip returned an unexpected error: %v", err)
	}
	if len(certs) != 1 {
		t.Fatalf("got %d certificates, want 1 (the directory entry should be skipped)", len(certs))
	}
	if certs[0].FileName != "budi.pdf" {
		t.Errorf("got Name %q, want %q", certs[0].FileName, "budi.pdf")
	}
}

func TestInputZip_FileNotFound(t *testing.T) {
	_, err := InputZip(filepath.Join(t.TempDir(), "missing.zip"))
	if err == nil {
		t.Fatal("expected an error for a missing zip file, got nil")
	}
}
