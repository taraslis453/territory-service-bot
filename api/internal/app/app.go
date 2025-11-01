package app

import (
	"context"
	"fmt"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/taraslis453/territory-service-bot/config"
	"github.com/taraslis453/territory-service-bot/internal/controller/telegram"
	"github.com/taraslis453/territory-service-bot/internal/entity"
	"github.com/taraslis453/territory-service-bot/internal/service"
	"github.com/taraslis453/territory-service-bot/internal/storage"
	"github.com/taraslis453/territory-service-bot/pkg/database"
	"github.com/taraslis453/territory-service-bot/pkg/logging"
)

func Run(cfg *config.Config) {
	logger := logging.NewZap(cfg.Log.Level)

	sql, err := database.NewPostgreSQL(&database.PostgreSQLConfig{
		User:     cfg.PostgreSQL.User,
		Password: cfg.PostgreSQL.Password,
		Host:     cfg.PostgreSQL.Host,
		Database: cfg.PostgreSQL.Database,
	})
	if err != nil {
		logger.Fatal("failed to init postgresql", "err", err)
	}

	err = sql.DB.Exec(`CREATE EXTENSION IF NOT EXISTS "uuid-ossp"`).Error
	if err != nil {
		logger.Fatal("failed to create extension uuid-ossp", "err", err)
	}

	err = sql.DB.AutoMigrate(
		&entity.User{},
		&entity.Congregation{},
		&entity.CongregationTerritory{},
		&entity.CongregationTerritoryNote{},
		&entity.CongregationTerritoryGroup{},
		&entity.RequestActionState{},
	)
	if err != nil {
		logger.Fatal("automigration failed", "err", err)
	}

	storages := service.Storages{
		User:         storage.NewUserStorage(sql),
		Congregation: storage.NewCongregationStorage(sql),
		Chat:         storage.NewChatStorage(sql),
	}

	serviceOptions := &service.Options{
		Cfg:      cfg,
		Logger:   logger,
		Storages: storages,
	}

	services := service.Services{
		Bot: service.NewBotService(serviceOptions),
	}

	// Start health check HTTP server for Cloud Run
	// Cloud Run requires containers to listen on PORT for health checks
	port := os.Getenv("PORT")
	if port == "" {
		port = "8080"
	}

	// Create HTTP server with health check endpoint
	mux := http.NewServeMux()
	mux.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		fmt.Fprintf(w, "OK - Territory Service Bot is running")
	})
	mux.HandleFunc("/health", func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		fmt.Fprintf(w, "healthy")
	})

	httpServer := &http.Server{
		Addr:    ":" + port,
		Handler: mux,
	}

	// Start HTTP server in goroutine
	go func() {
		logger.Info("starting health check server", "port", port)
		if err := httpServer.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			logger.Error("health check server error", "err", err)
		}
	}()

	// Start Telegram bot
	err = telegram.NewBot(&telegram.Options{
		Config:   cfg,
		Logger:   logger,
		Storages: storages,
		Services: services,
	})
	if err != nil {
		logger.Error("app - Run - telegram.NewBot: " + err.Error())
	}

	// Wait for interrupt signal
	interrupt := make(chan os.Signal, 1)
	signal.Notify(interrupt, os.Interrupt, syscall.SIGTERM)

	s := <-interrupt
	logger.Info("app - Run - signal: " + s.String())

	// Graceful shutdown of HTTP server
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	if err := httpServer.Shutdown(ctx); err != nil {
		logger.Error("failed to shutdown http server", "err", err)
	}
}
