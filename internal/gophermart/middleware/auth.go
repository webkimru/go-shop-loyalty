package middleware

import (
	"github.com/webkimru/go-shop-loyalty/internal/gophermart/api"
	"net/http"
	"strings"
)

const bearerSchema = "Bearer "

func CheckAuth(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		authorization := r.Header.Get("Authorization")
		if authorization == "" || !strings.HasPrefix(authorization, "Bearer ") {
			w.WriteHeader(http.StatusUnauthorized)
			return
		}
		token := authorization[len(bearerSchema):]
		userID := api.GetUserID(token)
		if userID < 0 {
			w.WriteHeader(http.StatusUnauthorized)
			return
		}

		next.ServeHTTP(w, r)
	})
}
