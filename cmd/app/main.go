// @title           XBank API
// @version         1.0
// @description     XBank Banking API - DDD architecture with Go and Fiber
// @host            localhost:3000
// @BasePath        /api/v1
// @securityDefinitions.apikey BearerAuth
// @in header
// @name Authorization
// @description Enter "Bearer {token}" to authorize
package main

import (
	"github.com/BakhodiribnYashinibnMansur/XBank/internal/app"
	"github.com/BakhodiribnYashinibnMansur/XBank/internal/kernel/infrastructure/config"
	"github.com/BakhodiribnYashinibnMansur/XBank/pkg/logger"
)

func main() {
	logger.Init(true)
	defer logger.Sync()

	cfg := config.Load("config.yml")

	app.Run(cfg)
}
