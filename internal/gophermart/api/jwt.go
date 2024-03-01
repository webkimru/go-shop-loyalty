package api

import (
	"fmt"
	"github.com/golang-jwt/jwt/v4"
	"github.com/webkimru/go-shop-loyalty/internal/gophermart/logger"
	"time"
)

// Claims — структура утверждений, которая включает стандартные утверждения и
// одно пользовательское UserID
type Claims struct {
	jwt.RegisteredClaims
	UserID int64
}

// BuildJWTString создаёт токен и возвращает его в виде строки.
func BuildJWTString(userID int64) (string, error) {
	// создаём новый токен с алгоритмом подписи HS256 и утверждениями — Claims
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, Claims{
		RegisteredClaims: jwt.RegisteredClaims{
			// когда создан токен
			ExpiresAt: jwt.NewNumericDate(time.Now().Add(time.Hour * time.Duration(app.TokenExp))),
		},
		// собственное утверждение
		UserID: userID,
	})

	// создаём строку токена
	tokenString, err := token.SignedString([]byte(app.SecretKey))
	if err != nil {
		return "", err
	}

	// возвращаем строку токена
	return tokenString, nil
}

func GetUserID(tokenString string) int64 {
	// создаём экземпляр структуры с утверждениями
	claims := &Claims{}
	// парсим из строки токена tokenString в структуру claims
	token, err := jwt.ParseWithClaims(tokenString, claims, func(t *jwt.Token) (interface{}, error) {
		// проверка заголовка алгоритма токена
		if _, ok := t.Method.(*jwt.SigningMethodHMAC); !ok {
			return nil, fmt.Errorf("unexpected signing method: %v", t.Header["alg"])
		}
		return []byte(app.SecretKey), nil
	})
	if err != nil {
		logger.Log.Infoln(err)
		return -1
	}

	if !token.Valid {
		logger.Log.Infoln("Token is not valid")
		return -1
	}

	// возвращаем ID пользователя в читаемом виде
	return claims.UserID
}
