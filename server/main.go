package main

import (
	"probig/config"
	"probig/database"
	"probig/embed"
	"probig/router"
	"probig/services"
)

func main() {
	config.Init()
	database.Init()
	services.LoadConfigs()

	r := router.SetupRouter()
	embed.ServeFrontend(r)
	r.Run(":" + config.ServerPort)
}
