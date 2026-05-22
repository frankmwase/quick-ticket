package delivery

import (
	"context"
	"fmt"
	"net/smtp"
	"strings"

	"quick-ticket/domain"
)

// SMTPConfig holds connection parameters for the SMTP server.
type SMTPConfig struct {
	Host     string
	Port     int
	Username string
	Password string
	From     string
	FromName string
}

// SMTPNotifier implements the domain.NotificationEngine interface
// using standard SMTP with MIME multipart for PDF attachments.
type SMTPNotifier struct {
	config SMTPConfig
}

func NewSMTPNotifier(config SMTPConfig) domain.NotificationEngine {
	return &SMTPNotifier{config: config}
}

func (n *SMTPNotifier) SendEmailWithAttachment(ctx context.Context, toEmail string, subject string, body string, pdfBytes []byte) error {
	boundary := "==QUICK_TICKET_BOUNDARY=="

	var msg strings.Builder
	msg.WriteString(fmt.Sprintf("From: %s <%s>\r\n", n.config.FromName, n.config.From))
	msg.WriteString(fmt.Sprintf("To: %s\r\n", toEmail))
	msg.WriteString(fmt.Sprintf("Subject: %s\r\n", subject))
	msg.WriteString("MIME-Version: 1.0\r\n")
	msg.WriteString(fmt.Sprintf("Content-Type: multipart/mixed; boundary=\"%s\"\r\n", boundary))
	msg.WriteString("\r\n")

	// Text body
	msg.WriteString(fmt.Sprintf("--%s\r\n", boundary))
	msg.WriteString("Content-Type: text/plain; charset=\"utf-8\"\r\n")
	msg.WriteString("\r\n")
	msg.WriteString(body)
	msg.WriteString("\r\n")

	// PDF attachment
	if len(pdfBytes) > 0 {
		msg.WriteString(fmt.Sprintf("--%s\r\n", boundary))
		msg.WriteString("Content-Type: application/pdf\r\n")
		msg.WriteString("Content-Transfer-Encoding: base64\r\n")
		msg.WriteString("Content-Disposition: attachment; filename=\"ticket.pdf\"\r\n")
		msg.WriteString("\r\n")

		encoded := encodeBase64(pdfBytes)
		msg.WriteString(encoded)
		msg.WriteString("\r\n")
	}

	msg.WriteString(fmt.Sprintf("--%s--\r\n", boundary))

	addr := fmt.Sprintf("%s:%d", n.config.Host, n.config.Port)
	auth := smtp.PlainAuth("", n.config.Username, n.config.Password, n.config.Host)

	return smtp.SendMail(addr, auth, n.config.From, []string{toEmail}, []byte(msg.String()))
}

func encodeBase64(data []byte) string {
	const lineLen = 76
	encoded := make([]byte, 0, len(data)*2)

	// Standard base64 encoding
	const base64Chars = "ABCDEFGHIJKLMNOPQRSTUVWXYZabcdefghijklmnopqrstuvwxyz0123456789+/"

	for i := 0; i < len(data); i += 3 {
		var b0, b1, b2 byte
		b0 = data[i]
		if i+1 < len(data) {
			b1 = data[i+1]
		}
		if i+2 < len(data) {
			b2 = data[i+2]
		}

		encoded = append(encoded, base64Chars[(b0>>2)&0x3F])
		encoded = append(encoded, base64Chars[((b0<<4)|(b1>>4))&0x3F])

		if i+1 < len(data) {
			encoded = append(encoded, base64Chars[((b1<<2)|(b2>>6))&0x3F])
		} else {
			encoded = append(encoded, '=')
		}

		if i+2 < len(data) {
			encoded = append(encoded, base64Chars[b2&0x3F])
		} else {
			encoded = append(encoded, '=')
		}
	}

	// Insert line breaks every 76 characters
	var result strings.Builder
	for i := 0; i < len(encoded); i += lineLen {
		end := i + lineLen
		if end > len(encoded) {
			end = len(encoded)
		}
		result.Write(encoded[i:end])
		result.WriteString("\r\n")
	}

	return result.String()
}
