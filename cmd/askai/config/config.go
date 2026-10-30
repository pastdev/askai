package config

import (
	"encoding/json"
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
		return nil, fmt.Errorf("endpointconfig named config: %w", err)
	}

	return endpoint, nil
}

func (c *Config) McpConfig(name string) (*pkgcfg.McpConfig, error) {
	cfg, err := c.Config()
	if err != nil {
		return nil, fmt.Errorf("mcpconfig load config: %w", err)
	}

	mcp, err := cfg.McpConfig(name)
	if err != nil {
		return nil, fmt.Errorf("mcpconfig named config: %w", err)
	}

	return mcp, nil
}

func AddConfig(root *cobra.Command) *Config {
	cfg := Config{
		configSource: cobracfg.ConfigLoader[pkgcfg.Config]{
			DefaultSources: cfgldr.Sources[pkgcfg.Config]{
				Sources: []cfgldr.SourceLoader{
					cfgldr.DirSource{
						Path:      SystemConfigDir,
						Unmarshal: yamlValueTemplateUnmarshal(),
					},
					cfgldr.DirSource{
						Path:      UserConfigDir,
						Unmarshal: yamlValueTemplateUnmarshal(),
					},
					cfgldr.DirSource{
						Path:      DirectoryConfigDir,
						Unmarshal: yamlValueTemplateUnmarshal(),
					},
				},
				Convert: jsonConvert,
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
		cfgldr.YamlUnmarshal(),
		"config",
		"c",
		"location of one or more config files")
	cfg.configSource.PersistentFlags(root).DirSourceVarP(
		cfgldr.YamlUnmarshal(),
		"config-dir",
		"d",
		"location of one or more config directories")
	root.PersistentFlags().StringVar(&cfg.endpoint, "endpoint", "", "the endpoint to use")

	return &cfg
}

// jsonConvert converts the merged document into cfg using the json
// marshaler so that json tags are respected (required by the openai types).
func jsonConvert(merged any, cfg *pkgcfg.Config) error {
	data, err := json.Marshal(merged)
	if err != nil {
		return fmt.Errorf("json marshal: %w", err)
	}
	err = json.Unmarshal(data, cfg)
	if err != nil {
		return fmt.Errorf("json unmarshal: %w", err)
	}
	return nil
}

// yamlValueTemplateUnmarshal unmarshals yaml then processes each _value_
// individually through the go template engine.
func yamlValueTemplateUnmarshal() cfgldr.UnmarshalFunc {
	return cfgldr.YamlValueTemplateUnmarshal(cfgldr.NewTemplate(cfgldr.DefaultFuncMap()))
}
