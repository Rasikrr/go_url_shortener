package main

import (
	"go_url_chortener_api/internal/app"
	"go_url_chortener_api/internal/config"
	"go_url_chortener_api/internal/env"
)

func main() {
	cfg := config.MustLoad(env.EnvLocal)
	app.Run(cfg)
}
