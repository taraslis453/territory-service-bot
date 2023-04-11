package service

import (
	"github.com/taraslis453/territory-service-bot/config"
	"github.com/taraslis453/territory-service-bot/pkg/logging"
)

// Services stores all service layer interfaces
type Services struct {
}

// Options provides options for creating a new service instance via New.
type Options struct {
	Cfg      *config.Config
	Logger   logging.Logger
	Storages Storages
}
