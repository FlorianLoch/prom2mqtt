package scrape

import (
	"context"
	"testing"

	"github.com/h2non/gock"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/florianloch/prom2mqtt/internal/config"
)

func TestScraping(t *testing.T) {
	defer gock.Off()

	gock.New("host").
		Get("/metrics").
		Reply(200).
		BodyString(textFormatResponse)

	scraper := New("http://host/metrics")

	metrics, err := scraper.ScrapeURL(context.Background())
	require.NoError(t, err)

	values, err := metrics.ExtractValues(config.PromSeriesSpecifier{
		Name:   "go_memstats_stack_sys_bytes",
		Labels: []config.PromLabelSpecifier{},
	})
	require.NoError(t, err)
	assert.EqualValues(t, []float64{688128}, values)

	values, err = metrics.ExtractValues(config.PromSeriesSpecifier{
		Name: "promhttp_metric_handler_requests_total",
		Labels: []config.PromLabelSpecifier{{
			Name:  "code",
			Value: "200",
		}},
	})
	require.NoError(t, err)
	assert.EqualValues(t, []float64{12909}, values)

	values, err = metrics.ExtractValues(config.PromSeriesSpecifier{
		Name: "promhttp_metric_handler_requests_total",
		Labels: []config.PromLabelSpecifier{{
			Name:  "code",
			Value: "200",
		}, {
			// This label does not exist
			Name:  "success",
			Value: "false",
		}},
	})
	require.NoError(t, err)
	assert.Len(t, values, 0)

	values, err = metrics.ExtractValues(config.PromSeriesSpecifier{
		Name: "promhttp_metric_handler_requests_total",
		Labels: []config.PromLabelSpecifier{{
			Name:  "code",
			Value: "500",
		}},
	})
	require.NoError(t, err)
	assert.EqualValues(t, []float64{0}, values)

	assert.False(t, gock.IsPending())
	assert.True(t, gock.IsDone())
}

const textFormatResponse = `
# HELP go_memstats_stack_sys_bytes Number of bytes obtained from system for stack allocator.
# TYPE go_memstats_stack_sys_bytes gauge
go_memstats_stack_sys_bytes 688128
# HELP go_memstats_sys_bytes Number of bytes obtained from system.
# TYPE go_memstats_sys_bytes gauge
go_memstats_sys_bytes 1.535168e+07
# HELP go_threads Number of OS threads created.
# TYPE go_threads gauge
# HELP promhttp_metric_handler_requests_total Total number of scrapes by HTTP status code.
# TYPE promhttp_metric_handler_requests_total counter
promhttp_metric_handler_requests_total{code="200"} 12909
promhttp_metric_handler_requests_total{code="500"} 0
promhttp_metric_handler_requests_total{code="503"} 0
`
