package main

import (
	"log"

	"github.com/bambu-group-03/apigw/config"
	"github.com/bambu-group-03/apigw/internal/app"
)

func main() {
	// Configuration
	cfg, err := config.NewConfig()
	if err != nil {
		log.Fatalf("Config error: %s", err)
	}

	// Run
	app.Run(cfg)
}
