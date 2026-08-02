package mailer

import (
	"CertSend/internal/matcher"
	"CertSend/pkg/logger"
	"bytes"
	"errors"
	"fmt"
	"time"

	"github.com/wneessen/go-mail"
)

const DailyLimit = 500

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

		if i < len(datas)-1 {
			time.Sleep(delay)
		}
	}

	return errors.Join(errs...)
}
