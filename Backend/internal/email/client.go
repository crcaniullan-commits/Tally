package email

import (
	"context"
	"fmt"
	"time"

	"github.com/resend/resend-go/v4"
)

/**
* Resend necesita de un dominio para funcionar
* una alternativa parece posbible usando gmail,
* su implementación queda pendiente, se dejara
* el codigo de resend en caso de
 */
type Client struct {
	resend *resend.Client
	from   string
}

func Newclient(apiKey, from string) *Client {
	return &Client{
		resend: resend.NewClient(apiKey),
		from:   from,
	}
}

func (c *Client) SendActivationCode(ctx context.Context, to, code string, expiresAt time.Time) error {
	html := fmt.Sprintf(`
		<h2>Código de activación</h2>
		<p>Tu código de acceso es:</p>
		<p style="font-size: 24px; font-weight: bold; letter-spacing: 2px;">%s</p>
		<p>Vence el %s.</p>
	`, code, expiresAt.Format("02/01/2006 15:04"))

	params := &resend.SendEmailRequest{
		From:    c.from,
		To:      []string{to},
		Subject: "Tu código de activación - Tally",
		Html:    html,
	}

	_, err := c.resend.Emails.SendWithContext(ctx, params)
	return err
}
