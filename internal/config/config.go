package config

import (
	"time"
)

type Config struct {
	MqttConfig MqttConfig    `embed:"" prefix:"mqtt-"`
	PromConfig PromConfig    `embed:"" prefix:"prometheus-"`
	Interval   time.Duration `help:"Scrapping interval" default:"1m"`
	Topics     Topic         `kong:"-"`
}

type MqttConfig struct {
	Host     string `help:"MQTT broker to send messages to" required:""`
	Username string `help:"Username to authenticate with"`
	Password string `help:"Password to authenticate with"`
}

type PromConfig struct {
	Host string `help:"Host running the Prometheus collector that shall be scrapped" required:""`
	Path string `help:"Path to use when requesting metrics from host" default:"/"`
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
