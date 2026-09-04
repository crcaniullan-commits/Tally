package handler

import (
	"net/http"

	"github.com/crcaniullan-commits/Tally/internal/service"
	"github.com/go-chi/chi/v5"
)

type Handler struct {
	Users interface {
		RegisterRoutes(chi.Router)
		Create(http.ResponseWriter, *http.Request)
		Update(http.ResponseWriter, *http.Request)
		Delete(http.ResponseWriter, *http.Request)
		GetByID(http.ResponseWriter, *http.Request)
	}
	PaymentMethods interface {
		RegisterRoutes(chi.Router)
		Create(http.ResponseWriter, *http.Request)
		Update(http.ResponseWriter, *http.Request)
		Delete(http.ResponseWriter, *http.Request)
		GetByID(http.ResponseWriter, *http.Request)
	}
	Incomes interface {
		RegisterRoutes(chi.Router)
		Create(http.ResponseWriter, *http.Request)
		Update(http.ResponseWriter, *http.Request)
		Delete(http.ResponseWriter, *http.Request)
		GetByID(http.ResponseWriter, *http.Request)
	}
	Goals interface {
		RegisterRoutes(chi.Router)
		Create(http.ResponseWriter, *http.Request)
		Update(http.ResponseWriter, *http.Request)
		Delete(http.ResponseWriter, *http.Request)
		GetByID(http.ResponseWriter, *http.Request)
	}
	Expenses interface {
		RegisterRoutes(chi.Router)
		Create(http.ResponseWriter, *http.Request)
		Update(http.ResponseWriter, *http.Request)
		Delete(http.ResponseWriter, *http.Request)
		GetByID(http.ResponseWriter, *http.Request)
	}
	Debtors interface {
		RegisterRoutes(chi.Router)
		Create(http.ResponseWriter, *http.Request)
		Update(http.ResponseWriter, *http.Request)
		Delete(http.ResponseWriter, *http.Request)
		GetByID(http.ResponseWriter, *http.Request)
	}
	Categories interface {
		RegisterRoutes(chi.Router)
		Create(http.ResponseWriter, *http.Request)
		Update(http.ResponseWriter, *http.Request)
		Delete(http.ResponseWriter, *http.Request)
		GetByID(http.ResponseWriter, *http.Request)
	}
	AccessKeys interface {
		RegisterRoutes(chi.Router)
		Create(http.ResponseWriter, *http.Request)
		Update(http.ResponseWriter, *http.Request)
		Delete(http.ResponseWriter, *http.Request)
		GetByID(http.ResponseWriter, *http.Request)
	}
}

func NewHandler(s service.Service) Handler {
	return Handler{
		Users:          NewUserHandler(s.ServiceUsers),
		PaymentMethods: NewPaymentMethodsHandler(s.ServicePaymentMethods),
		Incomes:        NewIncomesHandler(s.ServiceIncomes),
		Goals:          NewGoalsHandler(s.ServiceGoals),
		Expenses:       NewExpensesHandler(s.ServiceExpenses),
		Debtors:        NewDebtorsHandler(s.ServiceDebtors),
		Categories:     NewCategoriesHandler(s.ServiceCategories),
		AccessKeys:     NewAccessKeysHandler(s.ServiceAccessKey),
	}
}
