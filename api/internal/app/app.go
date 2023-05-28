package app

import (
	"os"
	"os/signal"
	"syscall"

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

	err = telegram.NewBot(&telegram.Options{
		Config:   cfg,
		Logger:   logger,
		Storages: storages,
		Services: services,
	})
	if err != nil {
		logger.Error("app - Run - telegram.NewBot: " + err.Error())
	}

	interrupt := make(chan os.Signal, 1)
	signal.Notify(interrupt, os.Interrupt, syscall.SIGTERM)

	s := <-interrupt
	logger.Info("app - Run - signal: " + s.String())
}
