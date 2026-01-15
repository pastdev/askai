package config

import (
	"testing"

	"github.com/openai/openai-go/v3"
	"github.com/openai/openai-go/v3/packages/param"
	pkgcfg "github.com/pastdev/askai/pkg/config"
	cobracfg "github.com/pastdev/configloader/pkg/cobra"
	cfgldr "github.com/pastdev/configloader/pkg/config"
	"github.com/stretchr/testify/require"
)

func TestConfigLoader(t *testing.T) {
	tester := func(t *testing.T, cfgYaml string, expected *pkgcfg.Config) {
		cfgLoader := cobracfg.ConfigLoader[pkgcfg.Config]{
			DefaultSources: cfgldr.Sources[pkgcfg.Config]{
				cfgldr.RawSource[pkgcfg.Config]{
					Data:      []byte(cfgYaml),
					Unmarshal: YamlToJSONWithValueTemplateUnmarshal[pkgcfg.Config](),
				},
			},
		}
		actual, err := cfgLoader.Config()
		require.NoError(t, err)
		require.Equal(t, expected, actual)
	}

	t.Run("empty", func(t *testing.T) {
		tester(t,
			``,
			&pkgcfg.Config{
				DefaultEndpoint: "",
				Endpoints:       nil,
			})
	})

	t.Run("default endpoint", func(t *testing.T) {
		tester(t,
			`---
default_endpoint: foo
endpoints: {}
`,
			&pkgcfg.Config{
				DefaultEndpoint: "foo",
				Endpoints:       map[string]pkgcfg.EndpointConfig{},
			})
	})

	t.Run("simple endpoint no default", func(t *testing.T) {
		tester(t,
			`---
endpoints:
  grok:
    auth_token: api-key
`,
			&pkgcfg.Config{
				DefaultEndpoint: "",
				Endpoints: map[string]pkgcfg.EndpointConfig{
					"grok": {
						AuthToken: "api-key",
					},
				},
			})
	})

	t.Run("simple endpoint with default", func(t *testing.T) {
		tester(t,
			`---
default_endpoint: grok
endpoints:
  grok:
    auth_token: api-key
`,
			&pkgcfg.Config{
				DefaultEndpoint: "grok",
				Endpoints: map[string]pkgcfg.EndpointConfig{
					"grok": {
						AuthToken: "api-key",
					},
				},
			})
	})

	t.Run("chat completion defaults", func(t *testing.T) {
		tester(t,
			`---
default_endpoint: grok
endpoints:
  grok:
    chat_completion_defaults:
      messages:
      - content: keep it simple
        role: system
      model: grok-4-fast-non-reasoning
`,
			&pkgcfg.Config{
				DefaultEndpoint: "grok",
				Endpoints: map[string]pkgcfg.EndpointConfig{
					"grok": {
						ChatCompletionDefaults: &openai.ChatCompletionNewParams{
							Messages: []openai.ChatCompletionMessageParamUnion{
								{
									OfSystem: &openai.ChatCompletionSystemMessageParam{
										Role: "system",
										Content: openai.ChatCompletionSystemMessageParamContentUnion{
											OfString: param.NewOpt("keep it simple"),
										},
									},
								},
							},
							Model: "grok-4-fast-non-reasoning",
						},
					},
				},
			})
	})
}
