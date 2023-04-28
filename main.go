package main

import (
	"os"
	"time"

	"github.com/hum2/ycl/cmd"
	"github.com/mattn/go-colorable"
	"github.com/rs/zerolog"
	"github.com/rs/zerolog/log"
)

func main() {
	setLogger()
	c := cmd.New()
	if err := c.Execute(); err != nil {
		log.Fatal().Err(err)
		os.Exit(1)
	}
}

func setLogger() {
	zerolog.SetGlobalLevel(zerolog.InfoLevel)
	log.Logger = log.Output(zerolog.ConsoleWriter{
		Out:        colorable.NewColorableStdout(),
		TimeFormat: time.DateTime,
		NoColor:    false,
	})
}
