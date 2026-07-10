package main

import (
	"github.com/khushidesai23/Enterprise-Order-Processing/config"
	"github.com/khushidesai23/Enterprise-Order-Processing/internal/common"
	"github.com/khushidesai23/Enterprise-Order-Processing/internal/database"
	"github.com/khushidesai23/Enterprise-Order-Processing/pkg/logger"
)

func main() {

	config.LoadConfig()

	logger.Init()

	defer logger.Sync()

	if err := database.Connect(); err != nil {
		panic(err)
	}

	common.Start()
}