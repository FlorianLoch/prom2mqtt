package mqtt

import (
	"context"
	"fmt"
	"net/url"

	"github.com/eclipse/paho.golang/autopaho"
	"github.com/eclipse/paho.golang/paho"

	"github.com/florianloch/prom2mqtt/internal/config"
)

type Client struct {
	conn *autopaho.ConnectionManager
}

const qosExactlyOnce = 2

func New(ctx context.Context, cfg config.MqttConfig) (*Client, error) {
	u, err := url.Parse(cfg.URL)
	if err != nil {
		return nil, fmt.Errorf("parsing broker URL '%s': %w", cfg.URL, err)
	}

	if u.Scheme != "mqtt" && u.Scheme != "tls" {
		return nil, fmt.Errorf("URL scheme '%s' not supported, should be 'mqtt' or 'tls'", u.Scheme)
	}

	mqttConfig := autopaho.ClientConfig{
		BrokerUrls: []*url.URL{u},
		TlsCfg:     nil,
		KeepAlive:  5 * 60,
	}

	mqttConfig.SetUsernamePassword(cfg.Username, []byte(cfg.Password))

	conn, err := autopaho.NewConnection(ctx, mqttConfig)

	return &Client{
		conn: conn,
	}, nil
}

func (c *Client) Publish(ctx context.Context, topic, payload string) error {
	resp, err := c.conn.Publish(ctx, &paho.Publish{
		QoS:     qosExactlyOnce,
		Topic:   topic,
		Retain:  false,
		Payload: []byte(payload),
	})
	if err != nil {
		return fmt.Errorf("publishing message: %w", err)
	} else if resp.ReasonCode != 0 && resp.ReasonCode != 16 { // 16 = Server received message but there are no subscribers
		return fmt.Errorf("received unexpected response reason code %d when publishing message", resp.ReasonCode)
	}

	return nil
}
