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
