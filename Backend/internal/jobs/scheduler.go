package jobs

import (
	"context"
	"time"

	"github.com/robfig/cron/v3"
	"go.uber.org/zap"
)

const (
	DefaultCronRecordatorioDiario = "0 20 * * *" // Todos los días a las 20:00
	DefaultCronCobrosVencidos     = "0 9 * * *"  // Todos los días a las 09:00
)

type IServicioJobs interface {
	RecordatorioDiario(ctx context.Context) error
	CobrosVencidos(ctx context.Context) error
	EjecutarTodos(ctx context.Context) error
}

type PlanificadorJobs struct {
	cron    *cron.Cron
	service IServicioJobs
	logger  *zap.SugaredLogger
}

func NuevoPlanificadorJobs(
	service IServicioJobs,
	loc *time.Location,
	cronDiario, cronCobros string,
	logger *zap.SugaredLogger,
) (*PlanificadorJobs, error) {
	if loc == nil {
		var err error
		loc, err = time.LoadLocation("America/Santiago")
		if err != nil {
			loc = time.UTC
		}
	}

	if cronDiario == "" {
		cronDiario = DefaultCronRecordatorioDiario
	}
	if cronCobros == "" {
		cronCobros = DefaultCronCobrosVencidos
	}

	c := cron.New(
		cron.WithLocation(loc),
		cron.WithChain(cron.Recover(cron.DefaultLogger)),
	)

	_, err := c.AddFunc(cronDiario, func() {
		ctx, cancel := context.WithTimeout(context.Background(), 5*time.Minute)
		defer cancel()
		if err := service.RecordatorioDiario(ctx); err != nil && logger != nil {
			logger.Errorw("cron: fallo en recordatorio diario", "error", err)
		}
	})
	if err != nil {
		return nil, err
	}

	_, err = c.AddFunc(cronCobros, func() {
		ctx, cancel := context.WithTimeout(context.Background(), 5*time.Minute)
		defer cancel()
		if err := service.CobrosVencidos(ctx); err != nil && logger != nil {
			logger.Errorw("cron: fallo en cobros vencidos", "error", err)
		}
	})
	if err != nil {
		return nil, err
	}

	return &PlanificadorJobs{
		cron:    c,
		service: service,
		logger:  logger,
	}, nil
}

func (p *PlanificadorJobs) Start() {
	if p.cron != nil {
		p.cron.Start()
		if p.logger != nil {
			p.logger.Infow("planificador de jobs iniciado")
		}
	}
}

func (p *PlanificadorJobs) Stop() context.Context {
	if p.cron != nil {
		return p.cron.Stop()
	}
	ctx, cancel := context.WithCancel(context.Background())
	cancel()
	return ctx
}
