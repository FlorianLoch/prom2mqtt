package config

import (
	"time"
)

type Config struct {
	Mqtt       MqttConfig    `embed:"" prefix:"mqtt-"`
	Prometheus PromConfig    `embed:"" prefix:"prometheus-"`
	Interval   time.Duration `help:"Scrapping interval" default:"1m"`
	Topics     Topic         `kong:"-"`
	Verbose    bool          `name:"verbose" short:"v"`
}

type MqttConfig struct {
	URL      string `help:"MQTT broker to send messages to" required:""`
	Username string `help:"Username to authenticate with"`
	Password string `help:"Password to authenticate with"`
}

type PromConfig struct {
	URL string `help:"URL from where to scrape metrics" required:""`
	// TODO: Support Basic Auth
}

type Topic map[string][]PromSeriesSpecifier

type PromSeriesSpecifier struct {
	Name   string
	Labels []PromLabelSpecifier // TODO: Add support for partial/RegExp matching in addition to label equality?
}

type PromLabelSpecifier struct {
	Name  string
	Value string
}
