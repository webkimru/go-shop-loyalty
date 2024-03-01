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
	"time"
)

// Repo - репозиторий используется хендлерами
var Repo *Repository
var app *config.AppConfig

var ErrDuplicate = errors.New("duplicate key value")

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
	if err := m.WriteResponseJSON(w, *resp); err != nil {
		logger.Log.Errorln("failed WriteResponseJSON()=", err)
		w.WriteHeader(http.StatusInternalServerError)
	}
}

func (m *Repository) Login(w http.ResponseWriter, r *http.Request) {
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
	var order models.Order
	if err := json.NewDecoder(r.Body).Decode(&order); err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	if !order.IsValid() {
		w.WriteHeader(http.StatusUnprocessableEntity)
		return
	}

	w.WriteHeader(http.StatusOK)
}

func (m *Repository) GetOrders(w http.ResponseWriter, r *http.Request) {

}

func (m *Repository) GetBalance(w http.ResponseWriter, r *http.Request) {

}

func (m *Repository) PayWithdraw(w http.ResponseWriter, r *http.Request) {

}

func (m *Repository) GetWithdraw(w http.ResponseWriter, r *http.Request) {

}

func (m *Repository) WriteResponseJSON(w http.ResponseWriter, user models.User) error {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	if err := json.NewEncoder(w).Encode(user); err != nil {
		return err
	}

	return nil
}
