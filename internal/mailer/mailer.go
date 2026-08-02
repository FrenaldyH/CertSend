package mailer

import (
	"bytes"
	"errors"
	"fmt"
	"time"

	"github.com/FrenaldyH/CertSend/pkg/logger"

	"github.com/FrenaldyH/CertSend/internal/matcher"

	"github.com/wneessen/go-mail"
)

// DailyLimit is Gmail's documented daily sending cap for standard
// personal accounts (500 recipients/emails per rolling 24 hours).
// See https://support.google.com/mail/answer/22839.
const DailyLimit = 500

// SendCertificate builds and sends a single certificate email to
// d.Email, with d.FileData attached under the name d.FileName.
func SendCertificate(d matcher.Data, client *mail.Client) error {
	msg := mail.NewMsg()

	if err := msg.From("frenaldyh@gmail.com"); err != nil {
		return err
	}

	if err := msg.To(d.Email); err != nil {
		return err
	}
	msg.Subject("Sertifikat Seminar - " + d.PersonName)
	msg.SetBodyString(mail.TypeTextPlain, "Halo "+d.PersonName+", berikut sertifikatmu.")

	reader := bytes.NewReader(d.FileData)
	if err := msg.AttachReader(d.FileName, reader); err != nil {
		return err
	}

	return client.DialAndSend(msg)
}

// SendBatch sends a certificate email to every entry in datas, pausing
// for delay between each send to avoid tripping the sending provider's
// rate-limit/spam heuristics. If len(datas) exceeds DailyLimit, the
// batch is rejected outright rather than risking a partial send that
// could get the sending account throttled or flagged.
//
// A failure sending to one recipient does not stop the rest of the
// batch; every error encountered is collected and returned together
// via errors.Join (nil if every send succeeded).
func SendBatch(datas []matcher.Data, client *mail.Client, delay time.Duration) error {
	if len(datas) > DailyLimit {
		warnErr := "recipient count exceeds the typical daily sending limit"
		logger.Log.Error(
			warnErr,
			"count", len(datas),
			"limit", DailyLimit,
		)
		return fmt.Errorf("recipient count exceeds the typical daily sending limit")
	}

	var errs []error
	for i, d := range datas {
		if i > 0 {
			time.Sleep(delay)
		}

		if err := SendCertificate(d, client); err != nil {
			msgErr := "failed to send certificate"
			logger.Log.Error(
				msgErr,
				"email", d.Email,
				"error", err,
			)
			errs = append(errs, err)
			continue
		}
	}

	return errors.Join(errs...)
}
