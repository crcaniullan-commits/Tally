package utilidades

import "github.com/crcaniullan-commits/Tally/internal/model"

// PeriodoMeta vive en model; se re-exporta acá para no cambiar los call sites.
type PeriodoMeta = model.GoalPeriod

const (
	PeriodoMetaDia    = model.GoalPeriodDia
	PeriodoMetaSemana = model.GoalPeriodSemana
	PeriodoMetaMes    = model.GoalPeriodMes
)
