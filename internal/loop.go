package internal

import (
	"context"
	"github.com/florianloch/prom2mqtt/internal/config"
	"github.com/florianloch/prom2mqtt/internal/scrape"
	"github.com/rs/zerolog/log"
	"strconv"
	"time"
)

type MQTTer interface {
	Publish(ctx context.Context, topic, payload string) error
}

type Scraper interface {
	ScrapeURL(ctx context.Context, targetURL string) (*scrape.Metrics, error)
}

func Loop(ctx context.Context, scraper Scraper, mqttClient MQTTer, cfg *config.Config) {
	ticker := eagerTicker(ctx, cfg.Interval)

	for {
		select {
		case <-ticker:
			scrapeAndPublish(ctx, scraper, mqttClient, cfg)
		case <-ctx.Done():
			return
		}
	}
}

func scrapeAndPublish(ctx context.Context, scraper Scraper, mqttClient MQTTer, cfg *config.Config) {
	for _, group := range cfg.Groups {
		logger := log.With().Str("target", group.Prometheus.URL).Logger()

		metrics, err := scraper.ScrapeURL(ctx, group.Prometheus.URL)
		if err != nil {
			logger.Error().Err(err).Str("target", group.Prometheus.URL).Msg("Failed to scrape metrics")

			continue
		}

		for topic, seriesSpecifier := range group.Topics {
			logger = logger.With().Str("topic", topic).Interface("specifier", seriesSpecifier).Logger()

			values, err := metrics.ExtractValues(seriesSpecifier)
			if err != nil {
				logger.Error().Err(err).Msg("Failed to extract series values")

				continue
			}

			if len(values) > 1 {
				logger.Error().Msg("Extracted more than one value; check existing label specifiers or add additional ones ensuring only one series gets matched")

				continue
			}

			if len(values) == 0 {
				logger.Warn().Msg("Extracted no value; possibly label specifiers are too restrictive")

				continue
			}

			value := values[0]

			stringifiedValue := strconv.FormatFloat(value, 'G', -1, 64)

			logger.Debug().Float64("value", value).Msg("Publishing message...")

			func() {
				publishCtx, cancelFn := context.WithTimeout(ctx, 5*time.Second)
				defer cancelFn()

				if err := mqttClient.Publish(publishCtx, topic, stringifiedValue); err != nil {
					logger.Error().
						Err(err).
						Msg("Failed to publish series value via MQTT")

					return
				}
			}()

			logger.Debug().Msg("Successfully published message")
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
