package handler

import (
	"net/http"

	errorhandler "github.com/crcaniullan-commits/Tally/internal/error"
	"github.com/crcaniullan-commits/Tally/internal/service"
	"github.com/go-chi/chi/v5"
)

type Handler struct {
	Users interface {
		RegisterRoutes(chi.Router) chi.Router
		Create(http.ResponseWriter, *http.Request)
		Update(http.ResponseWriter, *http.Request)
		Delete(http.ResponseWriter, *http.Request)
		GetByID(http.ResponseWriter, *http.Request)
	}
	Incomes interface {
	}
	Goals interface {
	}
	Expenses interface {
	}
	Debtors interface {
	}
	Categories interface {
	}
	AccessKeys interface {
	}
}

func NewHandler(s service.Service, e errorhandler.ErrorsResponse) Handler {
	return Handler{
		Users:      NewUserHandler(s.ServiceUsers, e),
		Incomes:    NewIncomesHandler(s.ServiceIncomes, e),
		Goals:      NewGoalsHandler(s.ServiceGoals, e),
		Expenses:   NewExpensesHandler(s.ServiceExpenses, e),
		Debtors:    NewDebtorsHandler(s.ServiceDebtors, e),
		Categories: NewCategoriesHandler(s.ServiceCategories, e),
		AccessKeys: NewAccessKeysHandler(s.ServiceAccessKey, e),
	}
}
