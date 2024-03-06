package api

import (
	"encoding/json"
	"errors"
	"github.com/webkimru/go-shop-loyalty/internal/gophermart/logger"
	"github.com/webkimru/go-shop-loyalty/internal/gophermart/models"
	"net/http"
)

func (m *Repository) PostWithdrawal(w http.ResponseWriter, r *http.Request) {
	//- `200` — успешная обработка запроса;
	//- `401` — пользователь не авторизован;
	//- `402` — на счету недостаточно средств;
	//- `422` — неверный номер заказа;
	//- `500` — внутренняя ошибка сервера.
	var withdrawal models.Withdrawal
	if err := json.NewDecoder(r.Body).Decode(&withdrawal); err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	logger.Log.Infoln(
		"Withdrawal:",
		"withdrawal.Order", withdrawal.Order,
		"withdrawal.Sum", withdrawal.Sum,
	)

	var order models.Order
	order.Number = withdrawal.Order
	if !order.IsValid() {
		// `422` — неверный формат номера заказа;
		w.WriteHeader(http.StatusUnprocessableEntity)
		return
	}

	authUserID := m.GetUserID(r)
	withdrawal.UserID = authUserID
	withdrawal.Sum = models.Money(withdrawal.Sum.Set())
	err := m.Store.SetWithdrawal(r.Context(), withdrawal)
	if err != nil && !errors.Is(err, ErrNotEnoughMoney) {
		logger.Log.Errorln("failed SetWithdrawal()= ", err)
		// `500` — внутренняя ошибка сервера.
		w.WriteHeader(http.StatusInternalServerError)
		return
	}
	// `402` — на счету недостаточно средств;
	if errors.Is(err, ErrNotEnoughMoney) {
		w.WriteHeader(http.StatusPaymentRequired)
		return
	}

	// `200` — успешная обработка запроса;
	w.WriteHeader(http.StatusOK)
}

func (m *Repository) GetWithdrawals(w http.ResponseWriter, r *http.Request) {
	//- `200` — успешная обработка запроса.
	//- `204` — нет ни одного списания.
	//- `401` — пользователь не авторизован.
	//- `500` — внутренняя ошибка сервера.
	authUserID := m.GetUserID(r)
	withdrawals, err := m.Store.GetWithdrawals(r.Context(), authUserID)
	if err != nil {
		logger.Log.Errorln("failed GetWithdrawals()= ", err)
		w.WriteHeader(http.StatusInternalServerError)
		return
	}

	logger.Log.Infoln("Requested by authUserID", authUserID, "|", "withdrawals", withdrawals)

	// 204` — нет данных для ответа.
	if len(withdrawals) == 0 {
		w.WriteHeader(http.StatusNoContent)
		return
	}

	if err := m.WriteResponseJSON(w, withdrawals, http.StatusOK); err != nil {
		logger.Log.Errorln("failed WriteResponseJSON()=", err)
		w.WriteHeader(http.StatusInternalServerError)
	}
}
