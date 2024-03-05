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
			ON CONFLICT (login) DO NOTHING
				RETURNING id, login, password, created_at
	`)
	if err != nil {
		return nil, err
	}
	defer stmt.Close()

	var id int64
	var login, password, createdAt string
	err = stmt.QueryRowContext(ctx, user.Login, user.Password, user.CreatedAt).Scan(&id, &login, &password, &createdAt)
	switch {
	case err == sql.ErrNoRows:
		return nil, api.ErrDuplicate
	case err != nil:
		return nil, err
	default:
		user.ID = id
		user.Password = password
		user.CreatedAt = createdAt
		return &user, nil
	}
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
		WITH cte AS (
			INSERT INTO gophermart.orders (number, user_id, status, created_at) VALUES($1, $2, $3, $4)
				ON CONFLICT (number)
					DO NOTHING RETURNING number, user_id, created_at
		)
		SELECT * FROM cte
		UNION
			SELECT number, user_id, created_at
				FROM gophermart.orders
					WHERE number = $1 and user_id != $2
	`)
	if err != nil {
		return "", 0, err
	}
	defer stmt.Close()

	var userDB int64
	var numberDB, createdAtDB string
	err = stmt.QueryRowContext(ctx, order.Number, order.UserID, order.Status, order.CreatedAt).Scan(&numberDB, &userDB, &createdAtDB)
	switch {
	case err == sql.ErrNoRows: // owner duplicate
		return "", 0, api.ErrDuplicate
	case err != nil:
		return "", 0, err
	default: // first and duplicate another
		return numberDB, userDB, nil
	}
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
		money := models.Money(accrual.Int64)
		orders = append(orders, models.Order{
			Number:    number,
			Accrual:   models.Money(money.Get()),
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

func (s *Store) GetBalance(ctx context.Context, userID int64) (*models.Balance, error) {
	stmt, err := s.Conn.PrepareContext(ctx, `
		SELECT user_id, current, withdrawn FROM gophermart.balance
			WHERE user_id = $1
	`)
	if err != nil {
		return nil, err
	}
	defer stmt.Close()

	var userDB int64
	var current, withdrawn models.Money
	err = stmt.QueryRowContext(ctx, userID).Scan(&userDB, &current, &withdrawn)
	balance := models.Balance{
		UserID:    userDB,
		Current:   models.Money(current.Get()),
		Withdrawn: models.Money(withdrawn.Get()),
	}
	switch {
	case err == sql.ErrNoRows:
		return &balance, nil
	case err != nil:
		return nil, err
	default:
		return &balance, nil
	}
}

func (s *Store) SetBalance(ctx context.Context, balance models.Balance, userID int64) error {
	row := s.Conn.QueryRowContext(ctx, `
		INSERT INTO gophermart.balance (user_id, current, withdrawn) VALUES($1, $2, $3)
			ON CONFLICT (user_id) DO
				UPDATE SET current = gophermart.balance.current + $2
	`, userID, balance.Current, balance.Withdrawn)

	if err := row.Err(); err != nil {
		return err
	}

	return nil
}

func (s *Store) UpdateOrder(ctx context.Context, order models.Order) error {
	row := s.Conn.QueryRowContext(ctx, `
		UPDATE gophermart.orders SET accrual = $1, status = $2
			WHERE number = $3
	`, order.Accrual, order.Status, order.Number)

	if err := row.Err(); err != nil {
		return err
	}
	return nil
}
