package correo

import (
	"context"
	"fmt"
	"time"

	"github.com/resend/resend-go/v4"
)

/**
* Resend necesita de un dominio para funcionar
* una alternativa parece posible usando gmail,
* su implementación queda pendiente, se dejara
* el codigo de resend en caso de
 */
type Client struct {
	resend *resend.Client
	from   string
}

func NuevoCliente(apiKey, from string) *Client {
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

// SendPasswordResetCode manda el código de recuperación de contraseña.
func (c *Client) SendPasswordResetCode(ctx context.Context, to, code string, expiresAt time.Time) error {
	html := fmt.Sprintf(`
		<h2>Recuperación de contraseña</h2>
		<p>Tu código para recuperar la contraseña es:</p>
		<p style="font-size: 24px; font-weight: bold; letter-spacing: 2px;">%s</p>
		<p>Vence el %s.</p>
		<p>Si no solicitaste el cambio, ignora este mensaje.</p>
	`, code, expiresAt.Format("02/01/2006 15:04"))

	params := &resend.SendEmailRequest{
		From:    c.from,
		To:      []string{to},
		Subject: "Tu código de recuperación - Tally",
		Html:    html,
	}

	_, err := c.resend.Emails.SendWithContext(ctx, params)
	return err
}
