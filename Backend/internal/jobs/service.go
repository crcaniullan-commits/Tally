package jobs

import (
	"context"
	"fmt"
	"strconv"
	"time"

	"github.com/crcaniullan-commits/Tally/internal/model"
	"github.com/google/uuid"
	"go.uber.org/zap"
)

type RepositorioJobs interface {
	UsuariosSinMovimientos(ctx context.Context, start, end time.Time) ([]uuid.UUID, error)
	CobrosVencidosPorUsuario(ctx context.Context, day string) ([]ResumenCobrosVencidos, error)
}

type Notificador interface {
	Send(ctx context.Context, userID uuid.UUID, n model.Notification) error
	Claim(ctx context.Context, userID uuid.UUID, kind string, day string) (bool, error)
	Unclaim(ctx context.Context, userID uuid.UUID, kind string, day string) error
}

type ServicioJobs struct {
	store    RepositorioJobs
	notifier Notificador
	loc      *time.Location
	logger   *zap.SugaredLogger
}

func NuevoServicioJobs(
	store RepositorioJobs,
	notifier Notificador,
	loc *time.Location,
	logger *zap.SugaredLogger,
) *ServicioJobs {
	if loc == nil {
		var err error
		loc, err = time.LoadLocation("America/Santiago")
		if err != nil {
			loc = time.UTC
		}
	}
	return &ServicioJobs{
		store:    store,
		notifier: notifier,
		loc:      loc,
		logger:   logger,
	}
}

// FormatearCLP formatea montos enteros a pesos chilenos con separador de miles: 45000 -> "$45.000"
func FormatearCLP(n int64) string {
	s := strconv.FormatInt(n, 10)
	for i := len(s) - 3; i > 0; i -= 3 {
		s = s[:i] + "." + s[i:]
	}
	return "$" + s
}

func (s *ServicioJobs) RecordatorioDiario(ctx context.Context) error {
	now := time.Now().In(s.loc)
	start := time.Date(now.Year(), now.Month(), now.Day(), 0, 0, 0, 0, s.loc)
	end := start.AddDate(0, 0, 1)
	day := start.Format("2006-01-02")

	users, err := s.store.UsuariosSinMovimientos(ctx, start, end)
	if err != nil {
		if s.logger != nil {
			s.logger.Errorw("daily reminder: error en consulta de usuarios sin movimientos", "err", err)
		}
		return err
	}

	for _, id := range users {
		ok, err := s.notifier.Claim(ctx, id, "daily", day)
		if err != nil || !ok {
			continue
		}

		err = s.notifier.Send(ctx, id, model.Notification{
			Title: "¿Cómo va tu día?",
			Body:  "Aún no registras movimientos hoy. Anótalos para no olvidarlos.",
			Data: map[string]string{
				"screen": "incomes",
			},
		})
		if err != nil {
			_ = s.notifier.Unclaim(ctx, id, "daily", day)
			if s.logger != nil {
				s.logger.Errorw("daily reminder: error en envío", "user_id", id, "err", err)
			}
		}
	}

	return nil
}

func (s *ServicioJobs) CobrosVencidos(ctx context.Context) error {
	now := time.Now().In(s.loc)
	day := now.Format("2006-01-02")

	resumenes, err := s.store.CobrosVencidosPorUsuario(ctx, day)
	if err != nil {
		if s.logger != nil {
			s.logger.Errorw("overdue debts: error en consulta de cobros vencidos", "err", err)
		}
		return err
	}

	for _, r := range resumenes {
		ok, err := s.notifier.Claim(ctx, r.UserID, "overdue", day)
		if err != nil || !ok {
			continue
		}

		cuerpo := fmt.Sprintf("Tienes %d cobros vencidos por %s.", r.Cantidad, FormatearCLP(r.Total))
		if r.Cantidad == 1 {
			cuerpo = fmt.Sprintf("Tienes 1 cobro vencido por %s.", FormatearCLP(r.Total))
		}

		err = s.notifier.Send(ctx, r.UserID, model.Notification{
			Title: "Cobros vencidos",
			Body:  cuerpo,
			Data: map[string]string{
				"screen": "debtors",
			},
		})
		if err != nil {
			_ = s.notifier.Unclaim(ctx, r.UserID, "overdue", day)
			if s.logger != nil {
				s.logger.Errorw("overdue debts: error en envío", "user_id", r.UserID, "err", err)
			}
		}
	}

	return nil
}

func (s *ServicioJobs) EjecutarTodos(ctx context.Context) error {
	if err := s.RecordatorioDiario(ctx); err != nil {
		return err
	}
	return s.CobrosVencidos(ctx)
}
