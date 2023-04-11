package main

import (
	"github.com/taraslis453/territory-service-bot/config"
	"github.com/taraslis453/territory-service-bot/internal/app"

	"github.com/taraslis453/territory-service-bot/pkg/logging"
)

func main() {
	logger := logging.NewZap("main")

	cfg := config.Get()
	logger.Info("read config", "config", cfg)

	app.Run(cfg)
}
