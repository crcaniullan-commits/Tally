package email

import (
	"context"
	"fmt"
	"html/template"
	"strings"
	"time"

	"github.com/wneessen/go-mail"
)

const (
	gmailSMTPHost = "smtp.gmail.com"
	gmailSMTPPort = 587
)

type GmailClient struct {
	username string
	password string
	from     string
}

func NewGmailClient(username, password, from string) (*GmailClient, error) {
	username = strings.TrimSpace(username)
	from = strings.TrimSpace(from)

	if username == "" {
		return nil, fmt.Errorf("gmail: falta el usuario SMTP")
	}
	if password == "" {
		return nil, fmt.Errorf("gmail: falta la contraseña de aplicación")
	}
	if from == "" {
		return nil, fmt.Errorf("gmail: falta el remitente")
	}

	return &GmailClient{
		username: username,
		password: password,
		from:     from,
	}, nil
}

func (c *GmailClient) SendActivationCode(ctx context.Context, to, code string, expiresAt time.Time) error {
	if strings.TrimSpace(to) == "" {
		return fmt.Errorf("gmail: el destinatario está vacio")
	}
	if code == "" {
		return fmt.Errorf("gmail: el código de activación está vacio")
	}

	message := mail.NewMsg()

	if err := message.From(c.from); err != nil {
		return fmt.Errorf("gmail: configurar remitente: %w", err)
	}
	if err := message.To(to); err != nil {
		return fmt.Errorf("gmail: configurar destinatario: %w", err)
	}

	message.Subject("Tu código de activación - Tally")

	const body = `
		<h2>Código de activación</h2>
		<p>Tu código de acceso es:</p>
		<p style="font-size: 24px; font-weight: bold; letter-spacing: 2px;">
			{{.Code}}
		</p>
		<p>Vence el {{.ExpiresAt}}.</p>
	`

	tmpl, err := template.New("activation").Parse(body)
	if err != nil {
		return fmt.Errorf("gmail: crear plantilla: %w", err)
	}

	data := struct {
		Code      string
		ExpiresAt string
	}{
		Code:      code,
		ExpiresAt: expiresAt.Format("02/01/2006 15:04"),
	}

	var html strings.Builder

	if err := tmpl.Execute(&html, data); err != nil {
		return fmt.Errorf("gmail: generar contenido: %w", err)
	}

	message.SetBodyString(mail.TypeTextHTML, html.String())

	client, err := mail.NewClient(
		gmailSMTPHost,
		mail.WithPort(gmailSMTPPort),
		mail.WithSMTPAuth(mail.SMTPAuthPlain),
		mail.WithUsername(c.username),
		mail.WithPassword(c.password),
		mail.WithTLSPolicy(mail.TLSMandatory),
	)

	if err != nil {
		return fmt.Errorf("gmail: configurar cliente SMTP: %w", err)
	}

	if err := client.DialAndSendWithContext(ctx, message); err != nil {
		return fmt.Errorf("gmail: enviar código de activación: %w", err)
	}

	return nil
}
