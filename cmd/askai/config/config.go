package config

import (
	"fmt"
	"io"

	pkgcfg "github.com/pastdev/askai/pkg/config"
	cobracfg "github.com/pastdev/configloader/pkg/cobra"
	cfgldr "github.com/pastdev/configloader/pkg/config"
	"github.com/spf13/cobra"
)

const (
	DirectoryConfigDir = "./askai.d"
	SystemConfigDir    = "/etc/askai.d"
	UserConfigDir      = "~/.config/askai.d"
)

type Config struct {
	configSource cobracfg.ConfigLoader[pkgcfg.Config]
	endpoint     string
}

func (c *Config) Config() (*pkgcfg.Config, error) {
	cfg, err := c.configSource.Config()
	if err != nil {
		return nil, fmt.Errorf("config load: %w", err)
	}
	return cfg, nil
}

func (c *Config) EndpointConfig() (*pkgcfg.EndpointConfig, error) {
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

func (c *Config) AddConfigCommandTo(root *cobra.Command) {
	c.configSource.AddSubCommandTo(
		root,
		cobracfg.WithConfigCommandOutput(
			"endpoints",
			func(w io.Writer, cfg *pkgcfg.Config) error {
				for endpoint := range cfg.Endpoints {
					_, err := fmt.Fprintln(w, endpoint)
					if err != nil {
						return fmt.Errorf("print: %w", err)
					}
				}
				return nil
			}))
}

func AddConfig(root *cobra.Command) *Config {
	cfg := Config{
		configSource: cobracfg.ConfigLoader[pkgcfg.Config]{
			DefaultSources: cfgldr.Sources[pkgcfg.Config]{
				cfgldr.DirSource[pkgcfg.Config]{
					Path:      SystemConfigDir,
					Unmarshal: cfgldr.YamlValueTemplateUnmarshal[pkgcfg.Config](nil),
				},
				cfgldr.DirSource[pkgcfg.Config]{
					Path:      UserConfigDir,
					Unmarshal: cfgldr.YamlValueTemplateUnmarshal[pkgcfg.Config](nil),
				},
				cfgldr.DirSource[pkgcfg.Config]{
					Path:      DirectoryConfigDir,
					Unmarshal: cfgldr.YamlValueTemplateUnmarshal[pkgcfg.Config](nil),
				},
			},
		},
	}

	cfg.configSource.AddSubCommandTo(
		root,
		cobracfg.WithConfigCommandOutput(
			"endpoints",
			func(w io.Writer, cfg *pkgcfg.Config) error {
				for endpoint := range cfg.Endpoints {
					_, err := fmt.Fprintln(w, endpoint)
					if err != nil {
						return fmt.Errorf("print: %w", err)
					}
				}
				return nil
			}))

	cfg.configSource.PersistentFlags(root).FileSourceVarP(
		cfgldr.YamlUnmarshal[pkgcfg.Config](),
		"config",
		"c",
		"location of one or more config files")
	cfg.configSource.PersistentFlags(root).DirSourceVarP(
		cfgldr.YamlUnmarshal[pkgcfg.Config](),
		"config-dir",
		"d",
		"location of one or more config directories")
	root.PersistentFlags().StringVar(&cfg.endpoint, "endpoint", "", "the endpoint to use")

	return &cfg
}
