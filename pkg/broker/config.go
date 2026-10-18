// Copyright 2026 Sloth. All rights reserved.
// License can be found in the LICENSE file.

package broker

import (
	"github.com/spf13/viper"
)

// Config holds broker configuration.
type Config struct {
	Provider string
	NATS     NATSConfig
}

// NATSConfig holds NATS connection settings.
type NATSConfig struct {
	URL   string
	Name  string
	Queue string
}

// GetConfig loads broker settings from app.broker config.
func GetConfig() Config {
	return Config{
		Provider: viper.GetString("app.broker.provider"),
		NATS: NATSConfig{
			URL:   viper.GetString("app.broker.nats.url"),
			Name:  viper.GetString("app.broker.nats.name"),
			Queue: viper.GetString("app.broker.nats.queue"),
		},
	}
}
