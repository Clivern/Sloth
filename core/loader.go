// Copyright 2026 Cognit. All rights reserved.
// License can be found in the LICENSE file.

package core

import (
	"bytes"
	"fmt"
	"os"

	"github.com/drone/envsubst"
	"github.com/spf13/viper"
)

// Load reads and parses the configuration file
func Load(configPath string) error {
	configUnparsed, err := os.ReadFile(configPath)
	if err != nil {
		return fmt.Errorf("error while reading config file [%s]: %w", configPath, err)
	}

	configParsed, err := envsubst.EvalEnv(string(configUnparsed))
	if err != nil {
		return fmt.Errorf("error while parsing config file [%s]: %w", configPath, err)
	}

	viper.SetConfigType("yaml")
	err = viper.ReadConfig(bytes.NewBuffer([]byte(configParsed)))
	if err != nil {
		return fmt.Errorf("error while loading configs [%s]: %w", configPath, err)
	}

	viper.SetDefault("config", configPath)

	return nil
}
