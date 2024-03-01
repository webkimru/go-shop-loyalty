package middleware

import (
	"net/http"
)

func CheckApplicationJson(next http.Handler) http.Handler {
	// получаем Handler приведением типа http.HandlerFunc
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Header.Get("Content-Type") != "application/json" {
			w.WriteHeader(http.StatusBadRequest)
		}

		next.ServeHTTP(w, r)
	})
}
