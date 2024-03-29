package main

import (
	"bytes"
	"context"
	"fmt"
	"github.com/alecthomas/kong"
	kongyaml "github.com/alecthomas/kong-yaml"
	"github.com/florianloch/prom2mqtt/internal"
	"github.com/florianloch/prom2mqtt/internal/scrape"
	"github.com/rs/zerolog"
	"github.com/rs/zerolog/log"
	"gopkg.in/yaml.v3"
	"io"
	"os"
	"os/signal"

	"github.com/florianloch/prom2mqtt/internal/config"
	"github.com/florianloch/prom2mqtt/internal/mqtt"
)

const envKeyConfig = "PROM2MQTT_CONFIG_PATH"

func main() {
	log.Logger = log.Output(zerolog.ConsoleWriter{Out: os.Stderr})

	zerolog.SetGlobalLevel(zerolog.InfoLevel)

	cli := &struct {
		Config config.Config `embed:""`
	}{}

	kong.ConfigureHelp(kong.HelpOptions{Compact: false, Summary: true})

	kong.Name("prom2mqtt")
	// TODO: Update!
	kong.Description("Small daemon scraping a service exporting Prometheus metrics and sending these to an MQTT server.")

	configPath := os.Getenv(envKeyConfig)

	if configPath == "" {
		configPath = "./prom2mqtt.config.yaml"
	}

	loader := kong.ConfigurationLoader(func(r io.Reader) (kong.Resolver, error) {
		input, err := io.ReadAll(r)
		if err != nil {
			return nil, fmt.Errorf("reading config: %w", err)
		}

		if err := yaml.Unmarshal(input, &cli.Config); err != nil {
			return nil, fmt.Errorf("parsing topic mapping: %w", err)
		}

		return kongyaml.Loader(bytes.NewReader(input))
	})

	ktx := kong.Parse(cli, kong.Configuration(loader, configPath))
	if ktx.Error != nil {
		log.Fatal().Err(ktx.Error).Msg("Failed to parse input parameters/commands")
	}

	cfg := &cli.Config

	if cfg.Verbose {
		zerolog.SetGlobalLevel(zerolog.DebugLevel)
	}

	log.Debug().Interface("cfg", cfg).Msg("Loaded config")

	ctx, cancelFn := signal.NotifyContext(context.Background(), os.Interrupt)
	defer cancelFn()

	mqttClient, err := mqtt.New(ctx, cfg.Mqtt)
	if err != nil {
		log.Fatal().Err(err).Msg("Failed to init MQTT client")
	}

	internal.Loop(ctx, scrape.New(), mqttClient, cfg)
}
