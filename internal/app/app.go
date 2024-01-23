package app

import (
	"context"
	"errors"
	"fmt"
	"github.com/OksidGen/enrich_server/internal/config"
	"github.com/golang-migrate/migrate/v4"
	"github.com/golang-migrate/migrate/v4/database/pgx/v5"
	_ "github.com/golang-migrate/migrate/v4/source/file"
	"github.com/labstack/echo/v4/middleware"
	"github.com/rs/zerolog"
	"github.com/rs/zerolog/log"
	"os"
	"os/signal"
	"time"

	"github.com/OksidGen/enrich_server/internal/delivery"
	"github.com/OksidGen/enrich_server/internal/repository"
	"github.com/OksidGen/enrich_server/internal/usecase"
	_ "github.com/jackc/pgx/v5/stdlib"
	"github.com/jmoiron/sqlx"
	"github.com/labstack/echo/v4"
)

func Run(cfg *config.Config) {

	log.Debug().Msg("Connecting to database...")
	db, err := sqlx.Connect("pgx", fmt.Sprintf(
		"postgres://%s:%s@%s:%d/%s",
		cfg.PG.USER,
		cfg.PG.PASSWORD,
		cfg.PG.HOST,
		cfg.PG.PORT,
		cfg.PG.DATABASE,
	))
	if err != nil {
		log.Fatal().Err(err)
	}
	defer func(db *sqlx.DB) {
		log.Debug().Msg("Closing database...")
		err := db.Close()
		if err != nil {
			log.Fatal().Err(err)
		}
		log.Debug().Msg("Database closed")
	}(db)

	log.Debug().Msg("Pinging database...")
	err = db.Ping()
	if err != nil {
		log.Fatal().Err(err)
	}

	log.Debug().Msg("Running migrations...")
	driver, err := pgx.WithInstance(db.DB, &pgx.Config{})
	if err != nil {
		log.Fatal().Err(err)
	}
	m, err := migrate.NewWithDatabaseInstance(
		"file://internal/repository/migrations",
		"verceldb",
		driver,
	)
	if err := m.Up(); err != nil && !errors.Is(err, migrate.ErrNoChange) {
		log.Fatal().Err(err)
	}

	log.Debug().Msg("Initializing repository...")
	repo := repository.NewPostgresRepository(db)

	log.Debug().Msg("Initializing usecase...")
	uc := usecase.NewUsecase(repo)

	log.Debug().Msg("Initializing server...")
	e := echo.New()

	logger := zerolog.New(os.Stdout).With().Timestamp().Logger()
	e.Use(middleware.RequestLoggerWithConfig(middleware.RequestLoggerConfig{
		LogURI:    true,
		LogStatus: true,
		LogValuesFunc: func(c echo.Context, v middleware.RequestLoggerValues) error {
			logger.Info().
				Str("uri", c.Request().RequestURI).
				Str("method", c.Request().Method).
				Int("status", v.Status).
				Msg("Request")

			return nil
		},
	}))

	log.Debug().Msg("Registering routes...")
	deliveryHandler := delivery.NewDelivery(uc)
	deliveryHandler.RegisterRoutes(e)

	log.Info().Msg("Starting server...")
	go func() {
		if err := e.Start(":8080"); err != nil {
			e.Logger.Fatal(err)
		}
	}()

	quit := make(chan os.Signal, 1)
	signal.Notify(quit, os.Interrupt)
	<-quit

	log.Debug().Msg("Shutting down server...")
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	if err := e.Shutdown(ctx); err != nil {
		e.Logger.Fatal(err)
	}

	log.Info().Msg("Server gracefully shutdown")
}
