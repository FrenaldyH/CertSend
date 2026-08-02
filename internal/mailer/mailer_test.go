package mailer

import (
	"net"
	"os"
	"strconv"
	"testing"
	"time"

	"github.com/FrenaldyH/CertSend/pkg/logger"

	"github.com/FrenaldyH/CertSend/internal/matcher"

	"github.com/wneessen/go-mail"
)

// TestMain initializes logger.Log once before any test in this
// package runs, since SendBatch calls logger.Log.Error internally and
// a nil *slog.Logger panics when a method is called on it.
func TestMain(m *testing.M) {
	dir, err := os.MkdirTemp("", "mailer-test-logs")
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

// newUnreachableClient returns a *mail.Client pointed at a local TCP
// port that nothing is listening on, so every DialAndSend attempt
// fails fast and deterministically ("connection refused") without
// needing a real SMTP server, credentials, or network access.
func newUnreachableClient(t *testing.T) *mail.Client {
	t.Helper()

	l, err := net.Listen("tcp", "127.0.0.1:0")
	if err != nil {
		t.Fatalf("failed to reserve a local port: %v", err)
	}
	_, portStr, _ := net.SplitHostPort(l.Addr().String())
	port, err := strconv.Atoi(portStr)
	if err != nil {
		t.Fatalf("failed to parse reserved port: %v", err)
	}
	l.Close() // free the port immediately so nothing is listening on it

	client, err := mail.NewClient(
		"127.0.0.1",
		mail.WithPort(port),
		mail.WithTLSPolicy(mail.NoTLS),
		mail.WithTimeout(2*time.Second),
	)
	if err != nil {
		t.Fatalf("failed to create test mail client: %v", err)
	}
	return client
}

func TestSendBatch_RejectsOverDailyLimit(t *testing.T) {
	datas := make([]matcher.Data, DailyLimit+1)
	for i := range datas {
		datas[i] = matcher.Data{Email: "someone@example.com"}
	}

	err := SendBatch(datas, newUnreachableClient(t), 0)
	if err == nil {
		t.Fatal("expected an error when recipient count exceeds DailyLimit, got nil")
	}
}

// TestSendBatch_ContinuesAfterFailureAndAggregatesErrors verifies that
// a failure sending to one recipient doesn't stop the rest of the
// batch, and that every failure (not just the first) is reflected in
// the returned error.
func TestSendBatch_ContinuesAfterFailureAndAggregatesErrors(t *testing.T) {
	datas := []matcher.Data{
		{Email: "budi@example.com", PersonName: "Budi", FileName: "budi.pdf", FileData: []byte("content")},
		{Email: "ani@example.com", PersonName: "Ani", FileName: "ani.pdf", FileData: []byte("content")},
	}

	err := SendBatch(datas, newUnreachableClient(t), time.Millisecond)
	if err == nil {
		t.Fatal("expected an aggregated error since the SMTP server is unreachable, got nil")
	}

	unwrapper, ok := err.(interface{ Unwrap() []error })
	if !ok {
		t.Fatalf("expected the error from errors.Join to support Unwrap() []error, got %T", err)
	}
	if got := len(unwrapper.Unwrap()); got != len(datas) {
		t.Errorf("got %d joined errors, want %d (every recipient's failure should be recorded, not just the first)", got, len(datas))
	}
}
