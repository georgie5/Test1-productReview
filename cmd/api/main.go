package main

import (
	"context"
	"database/sql"
	"flag"
	"log/slog"
	"os"
	"time"

	"github.com/georgie5/productReview/internal/data"
	_ "github.com/lib/pq" // PostgreSQL driver
)

// Define the application version
const appVersion = "1.0.0"

// Define server configuration structure
type serverConfig struct {
	port        int
	environment string
	db          struct {
		dsn string
	}

	limiter struct {
		rps     float64 // requests per second
		burst   int     // initial requests possible
		enabled bool    // enable or disable rate limiter
	}
}

// Define application dependencies structure
type applicationDependencies struct {
	config       serverConfig
	logger       *slog.Logger
	productModel data.ProductModel // ProductModel for managing products
	reviewModel  data.ReviewModel  // ReviewModel for managing reviews
}

func main() {
	// Initialize server configuration with default values
	var settings serverConfig

	flag.IntVar(&settings.port, "port", 4000, "API server port")
	flag.StringVar(&settings.environment, "env", "development", "Environment (development|staging|production)")

	flag.StringVar(&settings.db.dsn, "db-dsn", "postgres://productreview:product@localhost/productreview?sslmode=disable", "PostgreSQL DSN")

	flag.Float64Var(&settings.limiter.rps, "limiter-rps", 2, "Rate Limiter maximum requests per second")

	flag.IntVar(&settings.limiter.burst, "limiter-burst", 5, "Rate Limiter maximum burst")

	flag.BoolVar(&settings.limiter.enabled, "limiter-enabled", true, "Enable rate limiter")

	flag.Parse()

	// Initialize the logger
	logger := slog.New(slog.NewTextHandler(os.Stdout, nil))

	// Set up the database connection (optional for now since we’re not using it)
	db, err := openDB(settings)
	if err != nil {
		logger.Error(err.Error())
		os.Exit(1)
	}
	defer db.Close()
	logger.Info("Database connection pool established")

	// Initialize application dependencies
	appInstance := &applicationDependencies{
		config:       settings,
		logger:       logger,
		productModel: data.ProductModel{DB: db},
		reviewModel:  data.ReviewModel{DB: db},
	}

	err = appInstance.serve()
	if err != nil {
		logger.Error(err.Error())
		os.Exit(1)
	}
}

// openDB sets up a connection pool to the database
func openDB(settings serverConfig) (*sql.DB, error) {
	db, err := sql.Open("postgres", settings.db.dsn)
	if err != nil {
		return nil, err
	}

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	err = db.PingContext(ctx)
	if err != nil {
		db.Close()
		return nil, err
	}

	return db, nil
}
