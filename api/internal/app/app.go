package app

import (
	"github.com/taraslis453/territory-service-bot/config"
	"github.com/taraslis453/territory-service-bot/internal/service"
	"github.com/taraslis453/territory-service-bot/pkg/logging"
)

func Run(cfg *config.Config) {
	logger := logging.NewZap(cfg.Log.Level)

	storages := service.Storages{}

	_ = &service.Options{
		Cfg:      cfg,
		Logger:   logger,
		Storages: storages,
	}

	_ = service.Services{}
}
