package main

import (
	"agones-minecraft/config"
	"agones-minecraft/db"
	"agones-minecraft/log"
	"agones-minecraft/routers"
	"agones-minecraft/services/auth/jwt"
	"agones-minecraft/services/auth/sessions"
	"agones-minecraft/services/auth/twitch"
	"agones-minecraft/services/k8s"
	"agones-minecraft/services/k8s/agones"
	"agones-minecraft/services/validator"
)

func main() {
	config.LoadConfig()
	log.SetLog()

	k8s.InitConfig()
	agones.Init()

	sessions.NewStore()
	db.Init()

	twitch.NewODICProvider()
	jwt.New()
	validator.InitV1()

	r := routers.NewRouter()

	port := config.GetPort()
	r.Run(":" + port)
}
