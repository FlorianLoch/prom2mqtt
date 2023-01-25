package mqtt

import "github.com/florianloch/prom2mqtt/internal/config"

type Client struct {
	host     string
	username string
	password string
}

func New(cfg config.MqttConfig) *Client {
	return &Client{
		host:     cfg.Host,
		username: cfg.Username,
		password: cfg.Password,
	}
}

func (c *Client) PublishTopic(topic string, values []float64) error {
	// TODO
	panic("Implement me!")
}
