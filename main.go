package main

import (
	"bytes"
	"context"
	"fmt"
	"io"
	"os"

	"github.com/alecthomas/kong"
	kongyaml "github.com/alecthomas/kong-yaml"
	"github.com/rs/zerolog"
	"github.com/rs/zerolog/log"
	"gopkg.in/yaml.v3"

	"github.com/florianloch/prom2mqtt/internal/config"
	"github.com/florianloch/prom2mqtt/internal/scrape"
)

const envKeyConfig = "PROM2MQTT_CONFIG_PATH"

func main() {
	log.Logger = log.Output(zerolog.ConsoleWriter{Out: os.Stderr})

	cli := &struct {
		Config config.Config `embed:""`
	}{}

	kong.ConfigureHelp(kong.HelpOptions{Compact: false, Summary: true})

	kong.Name("prom2mqtt")
	// TODO: Update!
	kong.Description("Small daemon reading temperature and humidity from a dnt RoomLogg Pro base station via USB and pushing it into an InfluxDB.")

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

	log.Info().Interface("cfg", cfg).Msg("Loaded config")

	scraper := scrape.New("http://localhost:8080/internal/metrics")

	metrics, err := scraper.ScrapeURL(context.Background())
	if err != nil {
		log.Fatal().Err(err).Msg("Failed to scrape metrics")
	}

	vals, err := metrics.ExtractValues(config.PromSeriesSpecifier{
		Name: "my_metric",
		Labels: []config.PromLabelSpecifier{{
			Name:  "my_label_name",
			Value: "my_label_value",
		}},
	})
	if err != nil {
		log.Fatal().Err(err).Msg("Failed to extract series values")
	}

	log.Info().Interface("values", vals).Msg("Extracted series")
}
