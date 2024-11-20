package config

import (
	"fmt"

	"github.com/pastdev/askai/pkg/config"
	"github.com/pastdev/askai/pkg/log"
	"github.com/spf13/cobra"
)

const (
	DirectoryConfigDir = "./askai.d"
	SystemConfigDir    = "/etc/askai.d"
	UserConfigDir      = "~/.config/askai.d"
)

var (
	defaultConfigDirs = []string{
		SystemConfigDir,
		UserConfigDir,
		DirectoryConfigDir,
	}
)

type Config struct {
	configSource config.Source
	config       *config.Config
	endpoint     string
}

func (c *Config) AddFlags(cmd *cobra.Command) {
	cmd.PersistentFlags().StringArrayVarP(
		&c.configSource.Files,
		"config",
		"c",
		nil,
		"location of one or more config files")
	cmd.PersistentFlags().StringArrayVarP(
		&c.configSource.Dirs,
		"config-dir",
		"d",
		defaultConfigDirs,
		"location of one or more config directories")
	cmd.PersistentFlags().StringVar(&c.endpoint, "endpoint", "", "the enpoint to use")
}

func (c *Config) EndpointConfig() (*config.EndpointConfig, error) {
	if c.config == nil {
		log.Trace().Interface("ConfigSource", c.configSource).Msg("load configuration")
		cfg, err := c.configSource.Load()
		if err != nil {
			return nil, fmt.Errorf("newclient load: %w", err)
		}
		c.config = cfg
	}

	endpoint, err := c.config.EndpointConfig(c.endpoint)
	if err != nil {
		return nil, fmt.Errorf("client config: %w", err)
	}

	return endpoint, nil
}
