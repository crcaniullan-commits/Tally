package util

import "github.com/crcaniullan-commits/Tally/internal/model"

// GoalPeriod vive en model; se re-exporta acá para no cambiar los call sites.
type GoalPeriod = model.GoalPeriod

const (
	GoalPeriodDia    = model.GoalPeriodDia
	GoalPeriodSemana = model.GoalPeriodSemana
	GoalPeriodMes    = model.GoalPeriodMes
)
