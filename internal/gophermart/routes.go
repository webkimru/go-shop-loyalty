package gophermart

import (
	"github.com/go-chi/chi/v5"
	"github.com/webkimru/go-shop-loyalty/internal/gophermart/api"
	"github.com/webkimru/go-shop-loyalty/internal/gophermart/middleware"
	"net/http"
)

func Routes() http.Handler {
	r := chi.NewRouter()

	r.Use(middleware.WithLogging)
	r.Use(middleware.Gzip)
	r.Use(middleware.CheckApplicationJson)

	r.Group(func(r chi.Router) {
		r.Post("/api/user/register", api.Repo.Register)
		r.Post("/api/user/login", api.Repo.Login)

		r.Post("/api/user/orders", api.Repo.CreateOrder)
		r.Get("/api/user/orders", api.Repo.GetOrders)

		r.Get("/api/user/balance", api.Repo.GetBalance)
		r.Post("/api/user/balance/withdraw", api.Repo.PayWithdraw)
		r.Get("/api/user/withdrawals", api.Repo.GetWithdraw)
	})

	return r
}
