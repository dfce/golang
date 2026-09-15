package main

import (
	"log"
	"mini-srs/internal/bootstrap"
	"mini-srs/internal/config"
)

func main() {
	cfg := config.DefaultConfig()
	app := bootstrap.NewApp(cfg)

	if err := app.Start(); err != nil {
		log.Fatalf("Server exit with error: %v", err)
	}
}
