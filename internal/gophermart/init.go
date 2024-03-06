package gophermart

import (
	"context"
	"flag"
	"fmt"
	"github.com/webkimru/go-shop-loyalty/internal/gophermart/api"
	"github.com/webkimru/go-shop-loyalty/internal/gophermart/config"
	"github.com/webkimru/go-shop-loyalty/internal/gophermart/logger"
	"github.com/webkimru/go-shop-loyalty/internal/gophermart/repositories/store"
	"github.com/webkimru/go-shop-loyalty/internal/gophermart/repositories/store/pg"
	"log"
	"os"
	"strconv"
	"strings"
)

var app config.AppConfig

func Setup(ctx context.Context) (*string, error) {
	// flags:
	serverAddress := flag.String("a", "localhost:8080", "gophermart server address")
	storeDriver := flag.String("s", "postgresql", "gophermart store driver")
	databaseURI := flag.String("d", "", "database uri")
	secretKey := flag.String("k", "", "secret key")
	tokenExp := flag.Int("t", 2, "token exp (hour)")
	accrualSystemAddress := flag.String("r", "localhost:8181", "accrual system address")
	accrualPollInterval := flag.Int("i", 1, "accrual poll interval (sec)")

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
	if envAccrualSystemAddress := os.Getenv("ACCRUAL_SYSTEM_ADDRESS"); envAccrualSystemAddress != "" {
		accrualSystemAddress = &envAccrualSystemAddress
	}
	if envAccrualPollInterval := os.Getenv("ACCRUAL_POLL_INTERVAL"); envAccrualPollInterval != "" {
		pi, err := strconv.Atoi(envAccrualPollInterval)
		if err != nil {
			log.Fatal(err)
		}
		accrualPollInterval = &pi
	}

	// init logger:
	if err := logger.Initialize("info"); err != nil {
		return nil, err
	}

	// config:
	a := config.AppConfig{
		ServerAddress:        *serverAddress,
		StoreDriver:          *storeDriver,
		StoreDatabaseURI:     *databaseURI,
		SecretKey:            *secretKey,
		TokenExp:             *tokenExp,
		AccrualSystemAddress: URL(*accrualSystemAddress),
		AccrualPollInterval:  *accrualPollInterval,
	}
	app = a

	// print default config:
	logger.Log.Infoln(
		"Starting configuration:",
		"RUN_ADDRESS", app.ServerAddress,
		"STORE_DRIVER", app.StoreDriver,
		"DATABASE_URI", app.StoreDatabaseURI,
		"SECRET_KEY", app.SecretKey,
		"TOKEN_EXP", app.TokenExp,
		"ACCRUAL_SYSTEM_ADDRESS", app.AccrualSystemAddress,
		"ACCRUAL_POLL_INTERVAL", app.AccrualPollInterval,
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

func URL(rawUrl string) string {
	if !strings.HasPrefix(rawUrl, "http") {
		return fmt.Sprintf("http://%s", rawUrl)
	}

	return rawUrl
}
