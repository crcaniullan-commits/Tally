package notificaciones

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"time"

	"github.com/crcaniullan-commits/Tally/internal/model"
)

const (
	ExpoPushEndpointDefault = "https://exp.host/--/api/v2/push/send"
	DefaultTimeoutExpo      = 10 * time.Second
)

type ClienteExpoPush struct {
	endpoint   string
	httpClient *http.Client
}

func NuevoClienteExpoPush(httpClient *http.Client) *ClienteExpoPush {
	if httpClient == nil {
		httpClient = &http.Client{Timeout: DefaultTimeoutExpo}
	}
	return &ClienteExpoPush{
		endpoint:   ExpoPushEndpointDefault,
		httpClient: httpClient,
	}
}

func (c *ClienteExpoPush) ConEndpoint(endpoint string) *ClienteExpoPush {
	c.endpoint = endpoint
	return c
}

type expoMessage struct {
	To    string            `json:"to"`
	Title string            `json:"title"`
	Body  string            `json:"body"`
	Data  map[string]string `json:"data,omitempty"`
	Sound string            `json:"sound,omitempty"`
}

type expoTicketDetails struct {
	Error string `json:"error,omitempty"`
}

type expoTicket struct {
	Status  string            `json:"status"`
	Message string            `json:"message,omitempty"`
	Details expoTicketDetails `json:"details,omitempty"`
}

type expoResponse struct {
	Data   []expoTicket `json:"data"`
	Errors []struct {
		Message string `json:"message"`
		Code    string `json:"code"`
	} `json:"errors,omitempty"`
}

func (c *ClienteExpoPush) EnviarPush(ctx context.Context, tokens []string, n model.Notification) ([]string, error) {
	if len(tokens) == 0 {
		return nil, nil
	}

	messages := make([]expoMessage, 0, len(tokens))
	for _, t := range tokens {
		messages = append(messages, expoMessage{
			To:    t,
			Title: n.Title,
			Body:  n.Body,
			Data:  n.Data,
			Sound: "default",
		})
	}

	payloadBytes, err := json.Marshal(messages)
	if err != nil {
		return nil, fmt.Errorf("error serializando mensajes expo: %w", err)
	}

	req, err := http.NewRequestWithContext(ctx, http.MethodPost, c.endpoint, bytes.NewReader(payloadBytes))
	if err != nil {
		return nil, fmt.Errorf("error creando request expo: %w", err)
	}
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Accept", "application/json")

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("error realizando request a expo: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		return nil, fmt.Errorf("expo push api respondió con status %d", resp.StatusCode)
	}

	var expoRes expoResponse
	if err := json.NewDecoder(resp.Body).Decode(&expoRes); err != nil {
		return nil, fmt.Errorf("error decodificando respuesta expo: %w", err)
	}

	var tokensInvalidos []string
	for i, ticket := range expoRes.Data {
		if ticket.Status == "error" && ticket.Details.Error == "DeviceNotRegistered" {
			if i < len(tokens) {
				tokensInvalidos = append(tokensInvalidos, tokens[i])
			}
		}
	}

	return tokensInvalidos, nil
}
