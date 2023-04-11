package app

import (
	"os"
	"os/signal"
	"syscall"

	"github.com/taraslis453/territory-service-bot/config"
	"github.com/taraslis453/territory-service-bot/internal/controller/telegram"
	"github.com/taraslis453/territory-service-bot/internal/service"
	"github.com/taraslis453/territory-service-bot/pkg/logging"
)

func Run(cfg *config.Config) {
	logger := logging.NewZap(cfg.Log.Level)

	storages := service.Storages{}

	serviceOptions := &service.Options{
		Cfg:      cfg,
		Logger:   logger,
		Storages: storages,
	}

	services := service.Services{
		Bot: service.NewBotService(serviceOptions),
	}

	err := telegram.NewBot(&telegram.Options{
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
