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

	r.Group(func(r chi.Router) {
		r.Use(middleware.CheckApplicationJSON)

		r.Post("/api/user/register", api.Repo.Register)
		r.Post("/api/user/login", api.Repo.Login)
	})

	r.Group(func(r chi.Router) {
		r.Use(middleware.CheckAuth)

		r.Post("/api/user/orders", api.Repo.CreateOrder)
		r.With(middleware.CheckApplicationJSON).Get("/api/user/orders", api.Repo.GetOrders)
		r.With(middleware.CheckApplicationJSON).Get("/api/user/balance", api.Repo.GetBalance)
		r.With(middleware.CheckApplicationJSON).Post("/api/user/balance/withdraw", api.Repo.PayWithdraw)
		r.With(middleware.CheckApplicationJSON).Get("/api/user/withdrawals", api.Repo.GetWithdraw)
	})

	return r
}
