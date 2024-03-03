package api

import (
	"encoding/json"
	"errors"
	"github.com/webkimru/go-shop-loyalty/internal/gophermart/logger"
	"github.com/webkimru/go-shop-loyalty/internal/gophermart/models"
	"net/http"
	"strconv"
	"time"
)

func (m *Repository) CreateOrder(w http.ResponseWriter, r *http.Request) {
	//- `200` — номер заказа уже был загружен этим пользователем;
	//- `202` — новый номер заказа принят в обработку;
	//- `400` — неверный формат запроса;
	//- `401` — пользователь не аутентифицирован;
	//- `409` — номер заказа уже был загружен другим пользователем;
	//- `422` — неверный формат номера заказа;
	//- `500` — внутренняя ошибка сервера.
	var orderNumber int64
	if err := json.NewDecoder(r.Body).Decode(&orderNumber); err != nil {
		logger.Log.Errorln("orderNumber", orderNumber, err)
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	var order models.Order
	order.Number = strconv.Itoa(int(orderNumber))
	if !order.IsValid() {
		// `422` — неверный формат номера заказа;
		w.WriteHeader(http.StatusUnprocessableEntity)
		return
	}
	order.UserID = m.GetUserID(r)
	order.Status = models.OrderStateNew
	order.CreatedAt = time.Now().Format(time.RFC3339)
	orderNumberDB, userDB, err := m.Store.CreateOrder(r.Context(), order)
	if err != nil && !errors.Is(err, ErrDuplicate) {
		logger.Log.Errorln("failed CreateOrder()= ", err)
		w.WriteHeader(http.StatusInternalServerError)
		return
	}

	// `200` — номер заказа уже был загружен этим пользователем;
	if errors.Is(err, ErrDuplicate) {
		w.WriteHeader(http.StatusOK)
		return
	}

	// `409` — номер заказа уже был загружен другим пользователем;
	// строго после проверки errors.Is(err, ErrDuplicate), чтобы не проверять на userDB == 0
	if userDB != order.UserID {
		w.WriteHeader(http.StatusConflict)
		return
	}

	// `202` — новый номер заказа принят в обработку;
	w.WriteHeader(http.StatusAccepted)
	_, err = w.Write([]byte(orderNumberDB))
	if err != nil {
		logger.Log.Errorln("failed Write()= ", err)
	}
}

func (m *Repository) GetOrders(w http.ResponseWriter, r *http.Request) {
	//- `200` — успешная обработка запроса.
	//- `204` — нет данных для ответа.
	//- `401` — пользователь не авторизован.
	//- `500` — внутренняя ошибка сервера.
	authUserID := m.GetUserID(r)
	orders, err := m.Store.GetOrders(r.Context(), authUserID)
	if err != nil {
		logger.Log.Errorln("failed GetOrders()= ", err)
		w.WriteHeader(http.StatusInternalServerError)
		return
	}

	logger.Log.Infoln("Requested by authUserID", authUserID, "|", "orders", orders)

	// 204` — нет данных для ответа.
	if len(orders) == 0 {
		w.WriteHeader(http.StatusNoContent)
		return
	}

	if err := m.WriteResponseJSON(w, orders, http.StatusOK); err != nil {
		logger.Log.Errorln("failed WriteResponseJSON()=", err)
		w.WriteHeader(http.StatusInternalServerError)
	}
}
