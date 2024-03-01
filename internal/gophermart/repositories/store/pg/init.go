package pg

import (
	"context"
	"database/sql"
	_ "github.com/jackc/pgx/v5/stdlib"
	"github.com/webkimru/go-shop-loyalty/internal/gophermart/logger"
	"time"
)

// Bootstrap подготавливает БД к работе, создавая необходимые таблицы и индексы
func Bootstrap(ctx context.Context, conn *sql.DB) error {
	// запускаем транзакцию
	tx, err := conn.BeginTx(ctx, nil)
	if err != nil {
		return err
	}

	// в случае неуспешного коммита все изменения транзакции будут отменены
	defer tx.Rollback()

	// создаем схему
	tx.ExecContext(ctx, `CREATE SCHEMA IF NOT EXISTS gophermart`)
	tx.ExecContext(ctx, `SET search_path TO gophermart`)

	// создаем таблицы и индексы
	tx.ExecContext(ctx, `
		CREATE TABLE IF NOT EXISTS users (
			id BIGSERIAL PRIMARY KEY,
			login VARCHAR(50) NOT NULL,
		    password VARCHAR(100) NOT NULL,
			created_at TIMESTAMP WITH TIME ZONE DEFAULT NOW()
		)
	`)
	tx.ExecContext(ctx, `CREATE UNIQUE INDEX IF NOT EXISTS user_idx ON users (login)`)
	// orders:
	tx.ExecContext(ctx, `
		CREATE TABLE IF NOT EXISTS orders (
		    number BIGINT NOT NULL,
			user_id BIGINT NOT NULL,
			accrual BIGINT NOT NULL,
			status VARCHAR(25) NOT NULL,
			created_at TIMESTAMP WITH TIME ZONE DEFAULT NOW(),
			updated_at TIMESTAMP WITH TIME ZONE DEFAULT NOW()
		)
	`)
	tx.ExecContext(ctx, `CREATE UNIQUE INDEX IF NOT EXISTS order_idx ON orders (number)`)
	// balance:
	tx.ExecContext(ctx, `
		CREATE TABLE IF NOT EXISTS balance (
			user_id BIGINT NOT NULL,
			current BIGINT NOT NULL,
			withdrawn BIGINT NOT NULL,
			created_at TIMESTAMP WITH TIME ZONE DEFAULT NOW(),
			updated_at TIMESTAMP WITH TIME ZONE DEFAULT NOW()
		)
	`)
	tx.ExecContext(ctx, `CREATE UNIQUE INDEX IF NOT EXISTS order_idx ON orders (user_id)`)

	// триггер для поля updated_at
	tx.ExecContext(ctx, `
		CREATE OR REPLACE FUNCTION updated_at()
		RETURNS TRIGGER AS $$
		BEGIN
			NEW.updated_at = now();
			RETURN NEW;
		END;
		$$ language 'plpgsql';
	`)
	tx.ExecContext(ctx, `
		DO
		$$BEGIN
			CREATE TRIGGER orders_updated_at
				BEFORE UPDATE
				ON
					gophermart.orders
				FOR EACH ROW
			EXECUTE PROCEDURE updated_at();
		EXCEPTION
		   WHEN duplicate_object THEN
			  NULL;
		END;$$;
	`)
	tx.ExecContext(ctx, `
		DO
		$$BEGIN
			CREATE TRIGGER balance_updated_at
				BEFORE UPDATE
				ON
					gophermart.balance
				FOR EACH ROW
			EXECUTE PROCEDURE updated_at();
		EXCEPTION
		   WHEN duplicate_object THEN
			  NULL;
		END;$$;
	`)

	// коммитим транзакцию
	return tx.Commit()
}

func OpenDB(dsn string) (*sql.DB, error) {
	db, err := sql.Open("pgx", dsn)
	if err != nil {
		return nil, err
	}

	err = db.Ping()
	if err != nil {
		return nil, err
	}

	return db, nil
}

func ConnectToDB(dsn string) (*sql.DB, error) {
	// ретраи для переподключения к базе при старте
	// 1s, 3s, 5s
	backoff := [3]int{1, 3, 5}
	var cnt = 0

	for {
		connection, err := OpenDB(dsn)
		if err != nil {
			logger.Log.Infoln("Postgres not yet ready...")
			cnt++
		} else {
			logger.Log.Infoln("Connected to Postgres")
			return connection, nil
		}

		if cnt > 3 {
			logger.Log.Errorf("Shutdown after 3 attempts to connect to the database: %v", err)
			return nil, err
		}

		logger.Log.Infof("Backing off for %d seconds...", backoff[cnt-1])
		time.Sleep(time.Duration(backoff[cnt-1]) * time.Second)

		continue
	}
}
