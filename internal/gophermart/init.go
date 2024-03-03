package gophermart

import (
	"context"
	"flag"
	"github.com/webkimru/go-shop-loyalty/internal/gophermart/api"
	"github.com/webkimru/go-shop-loyalty/internal/gophermart/config"
	"github.com/webkimru/go-shop-loyalty/internal/gophermart/logger"
	"github.com/webkimru/go-shop-loyalty/internal/gophermart/repositories/store"
	"github.com/webkimru/go-shop-loyalty/internal/gophermart/repositories/store/pg"
	"log"
	"os"
	"strconv"
)

var app config.AppConfig

func Setup(ctx context.Context) (*string, error) {
	// flags:
	serverAddress := flag.String("a", "localhost:8080", "gophermart server address")
	storeDriver := flag.String("s", "postgresql", "gophermart store driver")
	databaseURI := flag.String("d", "", "database uri")
	secretKey := flag.String("k", "", "secret key")
	tokenExp := flag.Int("t", 2, "token exp (hour)")
	flag.Parse()

	// envs:
	if envRunAddr := os.Getenv("RUN_ADDRESS"); envRunAddr != "" {
		serverAddress = &envRunAddr
	}
	if envStoreDriver := os.Getenv("STORE_DRIVER"); envStoreDriver != "" {
		storeDriver = &envStoreDriver
	}
	if envDatabaseURI := os.Getenv("DATABASE_URI"); envDatabaseURI != "" {
		databaseURI = &envDatabaseURI
	}
	if envSecretKey := os.Getenv("SECRET_KEY"); envSecretKey != "" {
		secretKey = &envSecretKey
	}
	if envTokenExp := os.Getenv("TOKEN_EXP"); envTokenExp != "" {
		te, err := strconv.Atoi(envTokenExp)
		if err != nil {
			log.Fatal(err)
		}
		tokenExp = &te
	}

	// init logger:
	if err := logger.Initialize("info"); err != nil {
		return nil, err
	}

	// config:
	a := config.AppConfig{
		ServerAddress:    *serverAddress,
		StoreDriver:      *storeDriver,
		StoreDatabaseURI: *databaseURI,
		SecretKey:        *secretKey,
		TokenExp:         *tokenExp,
	}
	app = a

	// print default config:
	logger.Log.Infoln(
		"Starting configuration:",
		"RUN_ADDRESS", app.ServerAddress,
		"STORE_DRIVER", app.StoreDriver,
		"DATABASE_URI", app.StoreDatabaseURI,
	)

	// init store:
	var db store.Repositories
	switch app.StoreDriver {
	case "postgresql", "postgres":
		db = &pg.Store{}
	default:
		logger.Log.Fatalf("Unknown storage app.StoreDriver=%s", app.StoreDriver)
	}
	if err := db.Initialize(ctx, app); err != nil {
		return nil, err
	}

	// init app:
	repo := api.NewRepo(db)
	api.NewHandlers(repo, &app)

	return serverAddress, nil
}
