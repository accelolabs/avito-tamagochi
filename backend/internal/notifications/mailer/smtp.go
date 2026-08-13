package mailer

import (
	"bytes"
	"context"
	"fmt"
	"mime"
	"mime/multipart"
	"mime/quotedprintable"
	"net"
	"net/mail"
	"net/smtp"
	"net/textproto"
	"strings"
	"time"

	notificationmodel "github.com/accelolabs/avito-tamagochi/backend/internal/notifications/model"
)

type SMTP struct {
	address string
	from    string
	timeout time.Duration
}

func NewSMTP(address, from string, timeout time.Duration) *SMTP {
	return &SMTP{address: address, from: from, timeout: timeout}
}

func (s *SMTP) Send(ctx context.Context, message notificationmodel.Message) error {
	from, err := mailboxAddress(s.from)
	if err != nil {
		return fmt.Errorf("invalid sender: %w", err)
	}
	recipient, err := mailboxAddress(message.Recipient)
	if err != nil {
		return fmt.Errorf("invalid recipient: %w", err)
	}
	payload, err := buildMessage(from, recipient, message)
	if err != nil {
		return err
	}

	dialer := net.Dialer{Timeout: s.timeout}
	connection, err := dialer.DialContext(ctx, "tcp", s.address)
	if err != nil {
		return fmt.Errorf("connect to SMTP server: %w", err)
	}
	defer connection.Close()

	deadline := time.Now().Add(s.timeout)
	if contextDeadline, ok := ctx.Deadline(); ok && contextDeadline.Before(deadline) {
		deadline = contextDeadline
	}
	if err := connection.SetDeadline(deadline); err != nil {
		return fmt.Errorf("set SMTP deadline: %w", err)
	}

	host, _, err := net.SplitHostPort(s.address)
	if err != nil {
		return fmt.Errorf("invalid SMTP address: %w", err)
	}
	client, err := smtp.NewClient(connection, host)
	if err != nil {
		return fmt.Errorf("start SMTP client: %w", err)
	}
	defer client.Close()

	if err := client.Mail(from); err != nil {
		return fmt.Errorf("set SMTP sender: %w", err)
	}
	if err := client.Rcpt(recipient); err != nil {
		return fmt.Errorf("set SMTP recipient: %w", err)
	}
	data, err := client.Data()
	if err != nil {
		return fmt.Errorf("open SMTP message: %w", err)
	}
	if _, err := data.Write(payload); err != nil {
		_ = data.Close()
		return fmt.Errorf("write SMTP message: %w", err)
	}
	if err := data.Close(); err != nil {
		return fmt.Errorf("close SMTP message: %w", err)
	}
	if err := client.Quit(); err != nil {
		return fmt.Errorf("quit SMTP session: %w", err)
	}
	return nil
}

func buildMessage(from, recipient string, message notificationmodel.Message) ([]byte, error) {
	if err := validateHeaderValue(message.Subject); err != nil {
		return nil, fmt.Errorf("invalid subject: %w", err)
	}

	var body bytes.Buffer
	multipartWriter := multipart.NewWriter(&body)

	textHeader := make(textproto.MIMEHeader)
	textHeader.Set("Content-Type", "text/plain; charset=UTF-8")
	textHeader.Set("Content-Transfer-Encoding", "quoted-printable")
	textPart, err := multipartWriter.CreatePart(textHeader)
	if err != nil {
		return nil, fmt.Errorf("create text MIME part: %w", err)
	}
	if err := writeQuotedPrintable(textPart, message.TextBody); err != nil {
		return nil, fmt.Errorf("write text MIME part: %w", err)
	}

	htmlHeader := make(textproto.MIMEHeader)
	htmlHeader.Set("Content-Type", "text/html; charset=UTF-8")
	htmlHeader.Set("Content-Transfer-Encoding", "quoted-printable")
	htmlPart, err := multipartWriter.CreatePart(htmlHeader)
	if err != nil {
		return nil, fmt.Errorf("create HTML MIME part: %w", err)
	}
	if err := writeQuotedPrintable(htmlPart, message.HTMLBody); err != nil {
		return nil, fmt.Errorf("write HTML MIME part: %w", err)
	}
	if err := multipartWriter.Close(); err != nil {
		return nil, fmt.Errorf("close MIME message: %w", err)
	}

	var result bytes.Buffer
	fmt.Fprintf(&result, "From: %s\r\n", from)
	fmt.Fprintf(&result, "To: %s\r\n", recipient)
	fmt.Fprintf(&result, "Subject: %s\r\n", mime.BEncoding.Encode("UTF-8", message.Subject))
	fmt.Fprintf(&result, "MIME-Version: 1.0\r\n")
	fmt.Fprintf(&result, "Content-Type: multipart/alternative; boundary=%q\r\n", multipartWriter.Boundary())
	fmt.Fprintf(&result, "\r\n")
	result.Write(body.Bytes())
	return result.Bytes(), nil
}

func writeQuotedPrintable(destination interface{ Write([]byte) (int, error) }, value string) error {
	writer := quotedprintable.NewWriter(destination)
	canonical := strings.ReplaceAll(strings.ReplaceAll(value, "\r\n", "\n"), "\n", "\r\n")
	if _, err := writer.Write([]byte(canonical)); err != nil {
		return err
	}
	return writer.Close()
}

func mailboxAddress(value string) (string, error) {
	if err := validateHeaderValue(value); err != nil {
		return "", err
	}
	parsed, err := mail.ParseAddress(value)
	if err != nil {
		return "", err
	}
	return parsed.Address, nil
}

func validateHeaderValue(value string) error {
	if value == "" || strings.ContainsAny(value, "\r\n") {
		return fmt.Errorf("header value is empty or contains a newline")
	}
	return nil
}
