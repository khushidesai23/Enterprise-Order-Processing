package main

import (
	"log"

	"github.com/khushidesai23/Enterprise-Order-Processing/config"
	"github.com/khushidesai23/Enterprise-Order-Processing/internal/common"
	"github.com/khushidesai23/Enterprise-Order-Processing/internal/database"
	"github.com/khushidesai23/Enterprise-Order-Processing/pkg/logger"
)

func main() {

	if err := config.Load(); err != nil {
		log.Fatal(err)
	}

	if err := logger.Init(); err != nil {
		log.Fatal(err)
	}

	defer logger.Sync()

	if err := database.Connect(); err != nil {
		logger.L().Fatal(err.Error())
	}

	app := common.NewApplication()

	if err := app.Start(); err != nil {
		logger.L().Fatal(err.Error())
	}
}