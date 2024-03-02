package pg

import (
	"context"
	"database/sql"
	"github.com/webkimru/go-shop-loyalty/internal/gophermart/api"
	"github.com/webkimru/go-shop-loyalty/internal/gophermart/config"
	"github.com/webkimru/go-shop-loyalty/internal/gophermart/logger"
	"github.com/webkimru/go-shop-loyalty/internal/gophermart/models"
)

type Store struct {
	Conn *sql.DB
}

func (s *Store) Initialize(ctx context.Context, app config.AppConfig) error {
	var err error
	if s.Conn, err = ConnectToDB(app.StoreDatabaseURI); err != nil {
		return err
	}

	if err = Bootstrap(ctx, s.Conn); err != nil {
		return err
	}

	return nil
}

func (s *Store) CreateUser(ctx context.Context, user models.User) (*models.User, error) {
	stmt, err := s.Conn.PrepareContext(ctx, `
		INSERT INTO gophermart.users (login, password, created_at) VALUES($1, $2, $3)
			ON CONFLICT (login) DO
			    UPDATE SET login = $1 RETURNING id, login, password, created_at
	`)

	if err != nil {
		return nil, err
	}

	var id int64
	var login, password, createdAt string
	err = stmt.QueryRowContext(ctx, user.Login, user.Password, user.CreatedAt).Scan(&id, &login, &password, &createdAt)
	if err != nil {
		return nil, err
	}
	defer stmt.Close()

	// check duplicate login
	if login == user.Login && createdAt != user.CreatedAt {
		return nil, api.ErrDuplicate
	}

	user.ID = id
	user.Password = password
	user.CreatedAt = createdAt

	return &user, nil
}

func (s *Store) GetIDUserByAuth(ctx context.Context, user models.User) (int64, error) {
	stmt, err := s.Conn.PrepareContext(ctx, `
		SELECT id FROM gophermart.users
			WHERE login = $1 AND password = $2
	`)
	if err != nil {
		return 0, err
	}

	var res int64
	err = stmt.QueryRowContext(ctx, user.Login, user.Password).Scan(&res)
	if err != nil {
		return 0, err
	}
	defer stmt.Close()

	return res, nil
}

func (s *Store) CreateOrder(ctx context.Context, order models.Order) (string, int64, error) {
	stmt, err := s.Conn.PrepareContext(ctx, `
		INSERT INTO gophermart.orders (number, user_id, status, created_at) VALUES($1, $2, $3, $4)
			ON CONFLICT (number) DO
			    UPDATE SET number = $1 RETURNING number, user_id, created_at
	`)

	if err != nil {
		return "", 0, err
	}

	var userDB int64
	var numberDB, createdAtDB string
	err = stmt.QueryRowContext(ctx, order.Number, order.UserID, order.Status, order.CreatedAt).Scan(&numberDB, &userDB, &createdAtDB)
	if err != nil {
		return "", 0, err
	}
	defer stmt.Close()

	logger.Log.Infoln(
		"Added order numberFromDB", numberDB, "|",
		"userAuth", order.UserID, "|",
		"userDB", userDB, "|",
		"createdAtDB", createdAtDB,
	)

	// check duplicate
	if numberDB == order.Number && createdAtDB != order.CreatedAt {
		return numberDB, userDB, api.ErrDuplicate
	}

	return numberDB, userDB, nil
}

func (s *Store) GetOrders(ctx context.Context, userID int64) ([]models.Order, error) {
	stmt, err := s.Conn.PrepareContext(ctx, `
		SELECT number, accrual, status, created_at
			FROM gophermart.orders
				WHERE user_id = $1
				ORDER BY created_at DESC
	`)
	if err != nil {
		return nil, err
	}

	rows, err := stmt.QueryContext(ctx, userID)
	if err != nil {
		return nil, err
	}

	defer rows.Close()

	var orders []models.Order
	for rows.Next() {
		var accrual sql.NullInt64
		var number, status, createdAt string
		err = rows.Scan(&number, &accrual, &status, &createdAt)
		if err != nil {
			return nil, err
		}
		orders = append(orders, models.Order{
			Number:    number,
			Accrual:   float32(accrual.Int64) / 100,
			Status:    models.OrderState(status),
			CreatedAt: createdAt,
		})
	}

	// необходимо проверить ошибки уровня курсора
	if err := rows.Err(); err != nil {
		return nil, err
	}

	return orders, nil
}
