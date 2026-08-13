package mailer

import (
	"io"
	"mime"
	"mime/multipart"
	"mime/quotedprintable"
	"net/mail"
	"strings"
	"testing"

	notificationmodel "github.com/accelolabs/avito-tamagochi/backend/internal/notifications/model"
)

func TestBuildMessageCreatesUTF8MultipartAlternative(t *testing.T) {
	message := notificationmodel.Message{
		Recipient: "registered@example.com",
		Subject:   "Питомец ждёт вас",
		TextBody:  "Я по тебе соскучился... 🥺\nНавестишь меня?",
		HTMLBody:  "<p><strong>Я по тебе соскучился... 🥺</strong><br>Навестишь меня?</p>",
	}
	raw, err := buildMessage("no-reply@tamagochi.local", message.Recipient, message)
	if err != nil {
		t.Fatalf("build message: %v", err)
	}
	if strings.Contains(string(raw), "\n") && strings.Contains(strings.ReplaceAll(string(raw), "\r\n", ""), "\n") {
		t.Fatal("message contains a bare LF")
	}

	parsed, err := mail.ReadMessage(strings.NewReader(string(raw)))
	if err != nil {
		t.Fatalf("parse message: %v", err)
	}
	if got := parsed.Header.Get("To"); got != message.Recipient {
		t.Fatalf("To = %q, want %q", got, message.Recipient)
	}
	decodedSubject, err := new(mime.WordDecoder).DecodeHeader(parsed.Header.Get("Subject"))
	if err != nil {
		t.Fatalf("decode subject: %v", err)
	}
	if decodedSubject != message.Subject {
		t.Fatalf("Subject = %q, want %q", decodedSubject, message.Subject)
	}

	mediaType, parameters, err := mime.ParseMediaType(parsed.Header.Get("Content-Type"))
	if err != nil {
		t.Fatalf("parse content type: %v", err)
	}
	if mediaType != "multipart/alternative" {
		t.Fatalf("media type = %q, want multipart/alternative", mediaType)
	}
	reader := multipart.NewReader(parsed.Body, parameters["boundary"])
	want := []struct {
		contentType string
		body        string
	}{{"text/plain; charset=UTF-8", message.TextBody}, {"text/html; charset=UTF-8", message.HTMLBody}}
	for index, expected := range want {
		part, err := reader.NextPart()
		if err != nil {
			t.Fatalf("read MIME part %d: %v", index, err)
		}
		if got := part.Header.Get("Content-Type"); got != expected.contentType {
			t.Fatalf("part %d Content-Type = %q, want %q", index, got, expected.contentType)
		}
		decoded, err := io.ReadAll(quotedprintable.NewReader(part))
		if err != nil {
			t.Fatalf("decode MIME part %d: %v", index, err)
		}
		got := strings.ReplaceAll(string(decoded), "\r\n", "\n")
		if got != expected.body {
			t.Fatalf("part %d body = %q, want %q", index, got, expected.body)
		}
	}
	if _, err := reader.NextPart(); err != io.EOF {
		t.Fatalf("extra MIME part or unexpected error: %v", err)
	}
}

func TestBuildMessageRejectsHeaderInjection(t *testing.T) {
	tests := []struct {
		name      string
		from      string
		recipient string
		subject   string
	}{
		{"sender", "sender@example.com\r\nBcc: attacker@example.com", "user@example.com", "Subject"},
		{"recipient", "sender@example.com", "user@example.com\nBcc: attacker@example.com", "Subject"},
		{"subject", "sender@example.com", "user@example.com", "Subject\r\nBcc: attacker@example.com"},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			from, fromErr := mailboxAddress(test.from)
			recipient, recipientErr := mailboxAddress(test.recipient)
			if fromErr != nil || recipientErr != nil {
				return
			}
			_, err := buildMessage(from, recipient, notificationmodel.Message{Subject: test.subject})
			if err == nil {
				t.Fatal("header injection was accepted")
			}
		})
	}
}
