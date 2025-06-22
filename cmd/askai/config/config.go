package config

import (
	"fmt"
	"os"

	"github.com/pastdev/askai/pkg/config"
	"github.com/pastdev/askai/pkg/log"
	"github.com/spf13/cobra"
	"gopkg.in/yaml.v3"
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

func (c *Config) Config() (*config.Config, error) {
	if c.config == nil {
		log.Trace().Interface("ConfigSource", c.configSource).Msg("load configuration")
		cfg, err := c.configSource.Load()
		if err != nil {
			return nil, fmt.Errorf("newclient load: %w", err)
		}
		c.config = cfg
	}

	return c.config, nil
}

func (c *Config) EndpointConfig() (*config.EndpointConfig, error) {
	cfg, err := c.Config()
	if err != nil {
		return nil, fmt.Errorf("endpointconfig load config: %w", err)
	}

	endpoint, err := cfg.EndpointConfig(c.endpoint)
	if err != nil {
		return nil, fmt.Errorf("endpointconfig client config: %w", err)
	}

	return endpoint, nil
}

func New(c *Config) *cobra.Command {
	var output string

	cmd := cobra.Command{
		Use:   "config",
		Short: `Print out configuration information`,
		//nolint: revive // required to match upstream signature
		RunE: func(cmd *cobra.Command, args []string) error {
			cfg, err := c.Config()
			if err != nil {
				return fmt.Errorf("new load config: %w", err)
			}

			switch output {
			case "yaml":
				err := yaml.NewEncoder(os.Stdout).Encode(cfg)
				if err != nil {
					return fmt.Errorf("new yaml encode: %w", err)
				}
			case "endpoints":
				for endpoint := range cfg.Endpoints {
					fmt.Println(endpoint)
				}
			default:
				return fmt.Errorf("new unsupported output format: %s", output)
			}

			return nil
		},
	}

	cmd.Flags().StringVar(
		&output,
		"output",
		"content",
		"Format of output, one of: yaml, endpoints")

	return &cmd
}
