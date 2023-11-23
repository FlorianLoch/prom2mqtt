//go:generate mockgen -destination=./mocks/mocks.go -package=mocks -source=loop.go -typed

package internal

import (
	"context"
	"github.com/florianloch/prom2mqtt/internal/config"
	"github.com/florianloch/prom2mqtt/internal/mocks"
	"github.com/florianloch/prom2mqtt/internal/scrape"
	prom "github.com/prometheus/client_model/go"
	"go.uber.org/mock/gomock"
	"testing"
)

func Test_scrapeAndPublish(t *testing.T) {
	mockCtrl := gomock.NewController(t)
	defer mockCtrl.Finish()

	scraperMock := mocks.NewMockScraper(mockCtrl)
	mqttClientMock := mocks.NewMockMQTTer(mockCtrl)

	fakeMetrics := map[string]*prom.MetricFamily{
		"my_fake_metric": {
			Name: ptr("my_fake_metric"),
			Help: ptr("unhelpful help"),
			Type: ptr(prom.MetricType_COUNTER),
			Metric: []*prom.Metric{{
				Counter: &prom.Counter{Value: ptr(42.0)},
				Label: []*prom.LabelPair{{
					Name:  ptr("my_label"),
					Value: ptr("my_value"),
				}},
			}},
		},
	}

	scraperMock.EXPECT().ScrapeURL(gomock.Any(), "http://localhost:9090/metrics").
		Return(scrape.WrapRawMetrics(fakeMetrics), nil)

	mqttClientMock.EXPECT().Publish(gomock.Any(), "my_topic", "42").Return(nil)

	scrapeAndPublish(context.Background(), scraperMock, mqttClientMock, &config.Config{
		Groups: []config.Group{{
			Prometheus: config.PromConfig{
				URL: "http://localhost:9090/metrics",
			},
			Topics: map[string]config.PromSeriesSpecifier{
				"my_topic": {
					Name: "my_fake_metric",
					Labels: []config.PromLabelSpecifier{{
						Name:  "my_label",
						Value: "my_value",
					}},
				},
			},
		}},
	})
}

func ptr[T any](val T) *T {
	return &val
}
