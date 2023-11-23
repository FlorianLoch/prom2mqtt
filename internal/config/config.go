package config

import (
	"time"
)

type Config struct {
	Mqtt     MqttConfig    `embed:"" prefix:"mqtt-"`
	Interval time.Duration `help:"Scrapping interval" default:"1m"`
	Groups   []Group       `kong:"-"`
	Verbose  bool          `name:"verbose" short:"v"`
}

type MqttConfig struct {
	URL      string `help:"MQTT broker to send messages to" required:""`
	Username string `help:"Username to authenticate with"`
	Password string `help:"Password to authenticate with"`
}

type Group struct {
	Prometheus PromConfig
	Topics     map[string]PromSeriesSpecifier // topic -> series specifiers
}

type PromConfig struct {
	URL string
	// TODO: Support Basic Auth
}

type PromSeriesSpecifier struct {
	Name   string
	Labels []PromLabelSpecifier // TODO: Add support for partial/RegExp matching in addition to label equality?
}

type PromLabelSpecifier struct {
	Name  string
	Value string
}
