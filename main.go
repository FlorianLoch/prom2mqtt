package main

import (
	"bytes"
	"context"
	"fmt"
	"io"
	"os"
	"os/signal"
	"strconv"
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

	zerolog.SetGlobalLevel(zerolog.InfoLevel)

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

	loop(ctx, mqttClient, cfg)
}

func loop(ctx context.Context, mqttClient *mqtt.Client, cfg *config.Config) {
	scraper := scrape.New(cfg.Prometheus.URL)

	ticker := eagerTicker(ctx, cfg.Interval)

	for {
		select {
		case <-ticker:
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

					for i := range values {
						value := strconv.FormatFloat(values[i], 'G', -1, 64)

						go func() {
							log.Debug().Msg("Publishing message...")

							publishCtx, cancelFn := context.WithTimeout(ctx, 10*time.Second)
							defer cancelFn()

							if err := mqttClient.Publish(publishCtx, topic, value); err != nil {
								log.Error().
									Str("topic", topic).
									Interface("specifier", seriesSpecifiers).
									Err(err).
									Msg("Failed to publish series values via MQTT")

								return
							}

							log.Debug().Msg("Successfully published message")
						}()
					}
				}
			}
		case <-ctx.Done():
			return
		}
	}
}

func eagerTicker(ctx context.Context, interval time.Duration) <-chan time.Time {
	ch := make(chan time.Time, 1)
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

	ch <- time.Now()

	return ch
}
