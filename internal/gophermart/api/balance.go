package api

import (
	"github.com/webkimru/go-shop-loyalty/internal/gophermart/logger"
	"net/http"
)

func (m *Repository) GetBalance(w http.ResponseWriter, r *http.Request) {
	//- `200` — успешная обработка запроса.
	//- `401` — пользователь не авторизован.
	//- `500` — внутренняя ошибка сервера.
	authUserID := m.GetUserID(r)
	balance, err := m.Store.GetBalance(r.Context(), authUserID)
	if err != nil {
		logger.Log.Errorln("failed GetBalance()= ", err)
		w.WriteHeader(http.StatusInternalServerError)
		return
	}

	if err := m.WriteResponseJSON(w, balance, http.StatusOK); err != nil {
		logger.Log.Errorln("failed WriteResponseJSON()=", err)
		w.WriteHeader(http.StatusInternalServerError)
	}
}

func (m *Repository) PayWithdraw(w http.ResponseWriter, r *http.Request) {

}

func (m *Repository) GetWithdraw(w http.ResponseWriter, r *http.Request) {

}
