package main

import (
	"github.com/OksidGen/enrich_server/pkg"
	"github.com/rs/zerolog/log"

	"github.com/OksidGen/enrich_server/internal/app"
	"github.com/OksidGen/enrich_server/internal/config"
)

func main() {
	cfg, err := config.NewConfig()
	if err != nil {
		log.Fatal().Err(err)
	}

	pkg.SetupLogger()

	app.Run(cfg)
}
