package pg

import (
	"context"
	"database/sql"
	"github.com/webkimru/go-shop-loyalty/internal/gophermart/api"
	"github.com/webkimru/go-shop-loyalty/internal/gophermart/config"
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

func (s *Store) CreateOrder(ctx context.Context, order models.Order) (int64, int64, error) {
	stmt, err := s.Conn.PrepareContext(ctx, `
		INSERT INTO gophermart.orders (number, user_id, status, created_at) VALUES($1, $2, $3, $4)
			ON CONFLICT (number) DO
			    UPDATE SET number = $1 RETURNING number, user_id, created_at
	`)

	if err != nil {
		return 0, 0, err
	}

	var number, userID int64
	var createdAt string
	err = stmt.QueryRowContext(ctx, order.Number, order.UserID, order.Status, order.CreatedAt).Scan(&number, &userID, &createdAt)
	if err != nil {
		return 0, 0, err
	}
	defer stmt.Close()

	// check duplicate
	if number == order.Number && createdAt != order.CreatedAt {
		return number, userID, api.ErrDuplicate
	}

	return number, userID, nil
}

func (s *Store) GetOrders(ctx context.Context) ([]models.Order, error) {
	stmt, err := s.Conn.PrepareContext(ctx, `
		SELECT number, accrual, status, created_at
			FROM gophermart.orders
				ORDER BY created_at DESC
	`)
	if err != nil {
		return nil, err
	}

	rows, err := stmt.QueryContext(ctx)
	if err != nil {
		return nil, err
	}

	defer rows.Close()

	var orders []models.Order
	for rows.Next() {
		var number int64
		var accrual sql.NullInt64
		var status, createdAt string
		err = rows.Scan(&number, &accrual, &status, &createdAt)
		if err != nil {
			return nil, err
		}
		orders = append(orders, models.Order{
			Number:    number,
			Accrual:   accrual.Int64,
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
