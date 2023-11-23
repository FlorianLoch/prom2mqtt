package scrape

import (
	"context"
	"errors"
	"fmt"
	"io"
	"net/http"

	prom "github.com/prometheus/client_model/go"
	"github.com/prometheus/common/expfmt"

	"github.com/florianloch/prom2mqtt/internal/config"
)

type PromScraper struct {
	httpClient *http.Client
}

func New() *PromScraper {
	return &PromScraper{
		httpClient: &http.Client{},
	}
}

func (s *PromScraper) ScrapeURL(ctx context.Context, targetURL string) (*Metrics, error) {
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, targetURL, nil)
	if err != nil {
		return nil, fmt.Errorf("creating request to %q: %w", targetURL, err)
	}

	resp, err := s.httpClient.Do(req)
	defer resp.Body.Close()
	if err != nil {
		return nil, fmt.Errorf("performing request to %q: %w", targetURL, err)
	}

	return s.Scrape(resp.Body)
}

func (s *PromScraper) Scrape(r io.Reader) (*Metrics, error) {
	var parser expfmt.TextParser

	families, err := parser.TextToMetricFamilies(r)
	if err != nil {
		return nil, err
	}

	return WrapRawMetrics(families), nil
}

var (
	noMetricWithNameErr      = errors.New("no metric with given name")
	unsupportedMetricTypeErr = errors.New("unsupported metric type: only COUNTER and GAUGE are supported")
)

type Metrics struct {
	mF map[string]*prom.MetricFamily
}

func WrapRawMetrics(metricFamilies map[string]*prom.MetricFamily) *Metrics {
	return &Metrics{
		mF: metricFamilies,
	}
}

func (m *Metrics) ExtractValues(specifier config.PromSeriesSpecifier) ([]float64, error) {
	metric := m.mF[specifier.Name]
	if metric == nil {
		return nil, fmt.Errorf("%w: %s", noMetricWithNameErr, specifier.Name)
	}

	if metric.Type == nil || (*metric.Type != prom.MetricType_COUNTER && *metric.Type != prom.MetricType_GAUGE) {
		return nil, unsupportedMetricTypeErr
	}

	var values []float64

	for _, series := range metric.Metric {
		if !matchLabels(series, specifier.Labels) {
			continue
		}

		switch *metric.Type {
		case prom.MetricType_COUNTER:
			values = append(values, series.GetCounter().GetValue())
		case prom.MetricType_GAUGE:
			values = append(values, series.GetGauge().GetValue())
		}
	}

	return values, nil
}

func matchLabels(series *prom.Metric, labels []config.PromLabelSpecifier) bool {
	labelLUT := make(map[string]string, len(series.Label))

	for _, label := range series.Label {
		labelLUT[label.GetName()] = label.GetValue()
	}

	for _, label := range labels {
		value, ok := labelLUT[label.Name]
		if !ok {
			return false
		}

		if value != label.Value {
			return false
		}
	}

	return true
}
