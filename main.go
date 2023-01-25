package main

import (
	"bytes"
	"context"
	"fmt"
	"io"
	"os"
	"os/signal"
	"time"

	"github.com/alecthomas/kong"
	kongyaml "github.com/alecthomas/kong-yaml"
	"github.com/rs/zerolog"
	"github.com/rs/zerolog/log"
	"gopkg.in/yaml.v3"

	"github.com/florianloch/prom2mqtt/internal/config"
	"github.com/florianloch/prom2mqtt/internal/mqtt"
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

	log.Debug().Interface("cfg", cfg).Msg("Loaded config")

	mqttClient := mqtt.New(cfg.Mqtt)

	ctx, cancelFn := signal.NotifyContext(context.Background(), os.Interrupt)
	defer cancelFn()

	loop(ctx, mqttClient, cfg)
}

func loop(ctx context.Context, mqttClient *mqtt.Client, cfg *config.Config) {
	scraper := scrape.New(cfg.Prometheus.URL)

	ticker := eagerTicker(ctx, cfg.Interval)

	for range ticker {
		if ctx.Err() != nil {
			return
		}

		metrics, err := scraper.ScrapeURL(ctx)
		if err != nil {
			log.Error().Err(err).Msg("Failed to scrape metrics")

			continue
		}

		for topic, seriesSpecifiers := range cfg.Topics {
			for _, seriesSpecifier := range seriesSpecifiers {
				values, err := metrics.ExtractValues(seriesSpecifier)
				if err != nil {
					log.Fatal().Err(err).Msg("Failed to extract series values")
				}

				log.Debug().
					Str("topic", topic).
					Interface("specifier", seriesSpecifiers).
					Interface("values", values).
					Msg("Extracted series")

				if err := mqttClient.PublishTopic(topic, values); err != nil {
					log.Error().
						Str("topic", topic).
						Interface("specifier", seriesSpecifiers).
						Err(err).
						Msg("Failed to publish series values via MQTT")
				}
			}
		}
	}
}

func eagerTicker(ctx context.Context, interval time.Duration) <-chan time.Time {
	ch := make(chan time.Time)
	ticker := time.NewTicker(interval)

	go func() {
		for {
			select {
			case <-ctx.Done():
				ticker.Stop()

				return
			case t := <-ticker.C:
				ch <- t
			}
		}
	}()

	return ch
}
