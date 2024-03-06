package api

import (
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"errors"
	"fmt"
	"github.com/webkimru/go-shop-loyalty/internal/gophermart/logger"
	"github.com/webkimru/go-shop-loyalty/internal/gophermart/models"
	"net/http"
	"time"
)

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
