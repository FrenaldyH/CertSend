package service

import (
	"errors"
	"fmt"
	"path/filepath"
	"strings"
	"time"

	"github.com/FrenaldyH/CertSend/pkg/logger"

	"github.com/FrenaldyH/CertSend/internal/matcher"

	"github.com/FrenaldyH/CertSend/internal/mailer"

	"github.com/FrenaldyH/CertSend/internal/csvmap"

	"github.com/FrenaldyH/CertSend/internal/certificate"

	"github.com/wneessen/go-mail"
)

// SMTPConfig holds the credentials needed to connect to an SMTP
// server for sending emails.
type SMTPConfig struct {
	Host     string
	Port     int
	Username string
	Password string
}

// loadCertificates reads certPath and returns every certificate found
// in it, dispatching to certificate.InputPDF or certificate.InputZip
// based on the file extension.
func loadCertificates(certPath string) ([]certificate.Certificate, error) {
	switch strings.ToLower(filepath.Ext(certPath)) {
	case ".pdf":
		cert, err := certificate.InputPDF(certPath)
		if err != nil {
			return nil, err
		}
		return []certificate.Certificate{cert}, nil

	case ".zip":
		return certificate.InputZip(certPath)

	default:
		msgErr := "unsupported certificate file type"
		logger.Log.Error(
			msgErr,
			"certPath", certPath,
		)
		return nil, fmt.Errorf("%s: %s (expected .pdf or .zip)", msgErr, certPath)
	}
}

// SendCertificates runs the full pipeline: it loads certificate
// file(s) from certPath (a single .pdf or a .zip full of them), reads
// the recipient mapping from csvPath, matches each certificate to its
// recipient, then sends one email per match through the SMTP server
// described by smtp, pausing delay between each send.
//
// This is the single entry point meant to be called from app.go; it
// owns the order in which the certificate/csvmap/matcher/mailer
// packages are invoked so that callers don't need to know about it.
func SendCertificates(certPath string, csvPath string, smtp SMTPConfig, delay time.Duration) error {
	certs, err := loadCertificates(certPath)
	if err != nil {
		return err
	}

	entries, err := csvmap.ParseCSV(csvPath)
	if err != nil {
		return err
	}

	matched, err := matcher.Matcher(entries, certs)
	if err != nil {
		return err
	}
	if len(matched) == 0 {
		msg := "no certificates matched any recipient, nothing to send"
		logger.Log.Warn(
			msg,
			"certPath", certPath,
			"csvPath", csvPath,
		)
		return errors.New(msg)
	}

	client, err := mail.NewClient(
		smtp.Host,
		mail.WithPort(smtp.Port),
		mail.WithUsername(smtp.Username),
		mail.WithPassword(smtp.Password),
		mail.WithTLSPolicy(mail.TLSMandatory),
	)
	if err != nil {
		msg := "failed to create SMTP client"
		logger.Log.Error(
			msg,
			"host", smtp.Host,
			"port", smtp.Port,
			"error", err,
		)
		return fmt.Errorf("%s: %w", msg, err)
	}

	return mailer.SendBatch(matched, client, delay)
}
