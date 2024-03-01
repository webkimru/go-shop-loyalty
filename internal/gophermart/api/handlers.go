package api

import (
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"errors"
	"fmt"
	"github.com/webkimru/go-shop-loyalty/internal/gophermart/config"
	"github.com/webkimru/go-shop-loyalty/internal/gophermart/logger"
	"github.com/webkimru/go-shop-loyalty/internal/gophermart/models"
	"github.com/webkimru/go-shop-loyalty/internal/gophermart/repositories/store"
	"net/http"
	"strconv"
	"time"
)

// Repo - репозиторий используется хендлерами
var Repo *Repository
var app *config.AppConfig

var ErrDuplicate = errors.New("duplicate key value")

const bearerSchema = "Bearer "

// Repository описываем структуру репозитория для хендлеров
type Repository struct {
	Store store.Repositories
}

// NewRepo создаем новый репозиторий
func NewRepo(repository store.Repositories) *Repository {
	return &Repository{
		Store: repository,
	}
}

// NewHandlers устанавливаем репозиторий для хендлеров
func NewHandlers(r *Repository, a *config.AppConfig) {
	Repo = r
	app = a
}

func (m *Repository) Register(w http.ResponseWriter, r *http.Request) {
	//- `200` — пользователь успешно зарегистрирован и аутентифицирован;
	//- `400` — неверный формат запроса;
	//- `409` — логин уже занят;
	//- `500` — внутренняя ошибка сервера.
	var user models.User

	// разбираем POST данные на входе
	if err := json.NewDecoder(r.Body).Decode(&user); err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}
	if user.Login == "" || user.Password == "" {
		w.WriteHeader(http.StatusBadRequest)
	}

	// дописываем нужные значения в модель пользователя
	user.CreatedAt = time.Now().Format(time.RFC3339)
	hash := sha256.Sum256([]byte(user.Password))
	user.Password = hex.EncodeToString(hash[:])

	// пишем в базу
	resp, err := m.Store.CreateUser(r.Context(), user)
	if err != nil && !errors.Is(err, ErrDuplicate) {
		logger.Log.Errorln("failed CreateUser()= ", err)
		w.WriteHeader(http.StatusInternalServerError)
		return
	}
	if errors.Is(err, ErrDuplicate) {
		w.WriteHeader(http.StatusConflict)
		return
	}

	// выставляем токен для авторизации зарегистрированного пользователя
	token, err := BuildJWTString(resp.ID)
	if err != nil {
		logger.Log.Errorln("failed BuildJWTString()= ", err)
		w.WriteHeader(http.StatusInternalServerError)
		return
	}
	w.Header().Set("Authorization", fmt.Sprintf("Bearer %s", token))

	// отвечаем клиенту
	if err := m.WriteResponseJSON(w, *resp, http.StatusOK); err != nil {
		logger.Log.Errorln("failed WriteResponseJSON()=", err)
		w.WriteHeader(http.StatusInternalServerError)
	}
}

func (m *Repository) Login(w http.ResponseWriter, r *http.Request) {
	//- `200` — пользователь успешно аутентифицирован;
	//- `400` — неверный формат запроса;
	//- `401` — неверная пара логин/пароль;
	//- `500` — внутренняя ошибка сервера.
	var user models.User

	if err := json.NewDecoder(r.Body).Decode(&user); err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}
	if user.Login == "" || user.Password == "" {
		w.WriteHeader(http.StatusBadRequest)
		return
	}

	hash := sha256.Sum256([]byte(user.Password))
	user.Password = hex.EncodeToString(hash[:])

	// check auth
	id, err := m.Store.GetIDUserByAuth(r.Context(), user)
	if err != nil {
		w.WriteHeader(http.StatusUnauthorized)
		return
	}

	// set token
	token, err := BuildJWTString(id)
	if err != nil {
		logger.Log.Errorln("failed BuildJWTString()= ", err)
		w.WriteHeader(http.StatusInternalServerError)
		return
	}
	w.Header().Set("Authorization", fmt.Sprintf("Bearer %s", token))

	w.WriteHeader(http.StatusOK)
}

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
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	var order models.Order
	if !order.IsValid() {
		// `422` — неверный формат номера заказа;
		w.WriteHeader(http.StatusUnprocessableEntity)
		return
	}
	order.Number = orderNumber
	order.UserID = m.GetUserID(r)
	order.Status = models.OrderStateNew
	order.CreatedAt = time.Now().Format(time.RFC3339)
	orderNumberDB, userDB, err := m.Store.CreateOrder(r.Context(), order)
	if err != nil && !errors.Is(err, ErrDuplicate) {
		logger.Log.Errorln("failed CreateOrder()= ", err)
		w.WriteHeader(http.StatusInternalServerError)
		return
	}
	// `409` — номер заказа уже был загружен другим пользователем;
	if errors.Is(err, ErrDuplicate) && userDB != order.UserID {
		w.WriteHeader(http.StatusConflict)
		return
	}
	// `200` — номер заказа уже был загружен этим пользователем;
	if errors.Is(err, ErrDuplicate) && userDB == order.UserID {
		w.WriteHeader(http.StatusOK)
		return
	}

	// `202` — новый номер заказа принят в обработку;
	w.WriteHeader(http.StatusAccepted)
	_, err = w.Write([]byte(strconv.Itoa(int(orderNumberDB))))
	if err != nil {
		logger.Log.Errorln("failed Write()= ", err)
	}
}

func (m *Repository) GetOrders(w http.ResponseWriter, r *http.Request) {
	//- `200` — успешная обработка запроса.
	//- `204` — нет данных для ответа.
	//- `401` — пользователь не авторизован.
	//- `500` — внутренняя ошибка сервера.
	orders, err := m.Store.GetOrders(r.Context())
	if err != nil {
		logger.Log.Errorln("failed GetOrders()= ", err)
		w.WriteHeader(http.StatusInternalServerError)
		return
	}

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

func (m *Repository) GetBalance(w http.ResponseWriter, r *http.Request) {

}

func (m *Repository) PayWithdraw(w http.ResponseWriter, r *http.Request) {

}

func (m *Repository) GetWithdraw(w http.ResponseWriter, r *http.Request) {

}

func (m *Repository) WriteResponseJSON(w http.ResponseWriter, data interface{}, status int) error {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	if err := json.NewEncoder(w).Encode(data); err != nil {
		return err
	}

	return nil
}

func (m *Repository) GetUserID(r *http.Request) int64 {
	authorization := r.Header.Get("Authorization")
	token := authorization[len(bearerSchema):]
	userID := GetUserID(token)

	return userID
}
