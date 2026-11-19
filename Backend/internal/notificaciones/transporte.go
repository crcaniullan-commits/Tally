package notificaciones

import (
	"context"

	"github.com/crcaniullan-commits/Tally/internal/model"
	"go.uber.org/zap"
)

type EnviadorPush interface {
	EnviarPush(ctx context.Context, tokens []string, n model.Notification) ([]string, error)
}

// LogNotifier es la implementación de prueba/desarrollo que registra en logs zap.
type LogNotifier struct {
	logger *zap.SugaredLogger
}

func NuevoLogNotifier(logger *zap.SugaredLogger) *LogNotifier {
	return &LogNotifier{logger: logger}
}

func (l *LogNotifier) EnviarPush(ctx context.Context, tokens []string, n model.Notification) ([]string, error) {
	if l.logger != nil {
		l.logger.Infow("push (mock)",
			"tokens_count", len(tokens),
			"title", n.Title,
			"body", n.Body,
			"data", n.Data,
		)
	}
	return nil, nil
}
