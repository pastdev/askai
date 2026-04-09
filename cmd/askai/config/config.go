package config

import (
	"encoding/json"
	"fmt"
	"io"

	yaml2json "github.com/invopop/yaml"
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
				cfgldr.DirSource[pkgcfg.Config]{
					Path:      SystemConfigDir,
					Unmarshal: YamlToJSONWithValueTemplateUnmarshal[pkgcfg.Config](),
				},
				cfgldr.DirSource[pkgcfg.Config]{
					Path:      UserConfigDir,
					Unmarshal: YamlToJSONWithValueTemplateUnmarshal[pkgcfg.Config](),
				},
				cfgldr.DirSource[pkgcfg.Config]{
					Path:      DirectoryConfigDir,
					Unmarshal: YamlToJSONWithValueTemplateUnmarshal[pkgcfg.Config](),
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

// YamlToJSONWithValueTemplateUnmarshal is an Unmarshal function that converts
// yaml to json, then unmarshals using the json unmarshal so that json tags are
// respected, then processes each _value_ individually through the go template
// engine then reserializes the result to json before unmarshaling into T.
func YamlToJSONWithValueTemplateUnmarshal[T any]() func(b []byte, cfg *T) error {
	return func(b []byte, cfg *T) error {
		var valueMap any
		err := yaml2json.Unmarshal(b, &valueMap)
		if err != nil {
			return fmt.Errorf("yamlunmarshal to valueMap: %w", err)
		}

		// walk the map and template each value
		err = cfgldr.Walk(cfgldr.NewTemplate(cfgldr.DefaultFuncMap()), valueMap)
		if err != nil {
			return fmt.Errorf("yamlunmarshal walk valueMap: %w", err)
		}

		data, err := json.Marshal(valueMap)
		if err != nil {
			return fmt.Errorf("yamlunmarshal from valueMap: %w", err)
		}

		err = json.Unmarshal(data, cfg)
		if err != nil {
			return fmt.Errorf("yamlunmarshal to type: %w", err)
		}
		return nil
	}
}
