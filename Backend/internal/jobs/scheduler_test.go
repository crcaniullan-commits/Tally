package jobs

import (
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"go.uber.org/zap"
)

func TestPlanificadorJobs(t *testing.T) {
	svc := new(IServicioJobsMock)
	logger := zap.NewNop().Sugar()
	loc, _ := time.LoadLocation("America/Santiago")

	t.Run("inicializa correctamente con valores por defecto", func(t *testing.T) {
		p, err := NuevoPlanificadorJobs(svc, loc, "", "", logger)
		require.NoError(t, err)
		require.NotNil(t, p)

		p.Start()
		doneCtx := p.Stop()
		assert.NotNil(t, doneCtx)
	})

	t.Run("falla si la expresión cron diaria es inválida", func(t *testing.T) {
		p, err := NuevoPlanificadorJobs(svc, loc, "expresion_invalida", "", logger)
		assert.Error(t, err)
		assert.Nil(t, p)
	})

	t.Run("falla si la expresión cron de cobros es inválida", func(t *testing.T) {
		p, err := NuevoPlanificadorJobs(svc, loc, "0 20 * * *", "invalida", logger)
		assert.Error(t, err)
		assert.Nil(t, p)
	})
}
