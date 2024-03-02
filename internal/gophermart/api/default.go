package api

import (
	"encoding/json"
	"errors"
	"github.com/webkimru/go-shop-loyalty/internal/gophermart/config"
	"github.com/webkimru/go-shop-loyalty/internal/gophermart/models"
	"github.com/webkimru/go-shop-loyalty/internal/gophermart/repositories/store"
	"net/http"
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

func (m *Repository) WriteResponseJSON(w http.ResponseWriter, data interface{}, status int) error {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)

	out, err := json.Marshal([]models.Order{{
		Number:    12345678903,
		Status:    "NEW",
		CreatedAt: "2024-03-02T11:12:00Z",
	}})
	if err != nil {
		return err
	}
	_, err = w.Write(out)
	if err != nil {
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
