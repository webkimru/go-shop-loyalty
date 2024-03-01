package store

import (
	"context"
	"github.com/webkimru/go-shop-loyalty/internal/gophermart/config"
	"github.com/webkimru/go-shop-loyalty/internal/gophermart/models"
)

type Repositories interface {
	Initialize(ctx context.Context, app config.AppConfig) error
	CreateUser(ctx context.Context, user models.User) (*models.User, error)
	GetIDUserByAuth(ctx context.Context, user models.User) (int64, error)

	CreateOrder(ctx context.Context, order models.Order) (number int64, userID int64, err error)
	GetOrders(ctx context.Context) ([]models.Order, error)
}
