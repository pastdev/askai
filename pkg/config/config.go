package config

import (
	"context"
	"crypto/tls"
	"crypto/x509"
	"errors"
	"fmt"
	"net/http"
	"net/http/httputil"
	"os"

	"github.com/openai/openai-go/v3"
	"github.com/openai/openai-go/v3/option"
	"github.com/pastdev/askai/pkg/log"
	"github.com/pastdev/askai/pkg/mcp"
)

var ExampleConfig = `---
default_endpoint: mixtral_22b
endpoints:
  mixtral_22b:
    api_type: OPEN_AI
    base_url: http://172.22.144.1:11434/v1
  codestral:
    api_type: OPEN_AI
    base_url: http://172.22.144.2:11434/v1
mcp_servers:
  jira:
    # if true, skip all confirmation and immediatly invoke tools
    auto_confirm: false
    # if true, the default value of confirmation will be "Y", but user will
	# still be prompted for confirmation for each tool invocation
	confirm_default_yes: true
	# unique name among mcp servers, should match key, but not required to. this
	# value will be used to prefix all functions provided by this server to
	# prevent function name collisions when using multiple mcp servers
    name: jira
	# the mcp server url. currently only http servers are supported
	url: "https://mcp.atlassian.com/v1/forge/mcp"
`

type Config struct {
	Endpoints       map[string]EndpointConfig `json:"endpoints" yaml:"endpoints"`
	DefaultEndpoint string                    `json:"default_endpoint" yaml:"default_endpoint"`
	McpServers      map[string]McpConfig      `json:"mcp_servers" yaml:"mcp_servers"`
}

// EndpointConfig is a configuration of a client.
type EndpointConfig struct {
	APIVersion             string                          `json:"api_version" yaml:"api_version"`
	AuthToken              string                          `json:"auth_token" yaml:"auth_token"`
	BaseURL                string                          `json:"base_url" yaml:"base_url"`
	ChatCompletionDefaults *openai.ChatCompletionNewParams `json:"chat_completion_defaults" yaml:"chat_completion_defaults"`
	CACerts                string                          `json:"cacerts" yaml:"cacerts"`
	EmptyMessagesLimit     uint                            `json:"empty_messages_limit" yaml:"empty_messages_limit"`
	ImageDefaults          *openai.ImageGenerateParams     `json:"image_defaults" yaml:"image_defaults"`
	InsecureSkipTLS        bool                            `json:"insecure_skip_tls" yaml:"insecure_skip_tls"`
	OrgID                  string                          `json:"org_id" yaml:"org_id"`
}

type McpConfig struct {
	// If true, tool call confirmation will be skipped, instead logging the
	// call at info level.
	AutoConfirm bool `json:"auto_confirm" yaml:"auto_confirm"`
	// If true, the default value for tool confirmation will be yes. This is the
	// value that will be used if the user confirms with an empty response, but
	// confirmation is still required.
	ConfirmDefaultYes bool `json:"confirm_default_yes" yaml:"confirm_default_yes"`
	// A name to prefix each tool name with in order to prevent name collisions
	// when multiple mcp's are in use.
	Name string `json:"name" yaml:"name"`
	// The url of the mcp server.
	URL string `json:"url" yaml:"url"`
}

type loggingTransport struct {
	wrapped http.RoundTripper
}

func (c *Config) EndpointConfig(endpoint string) (*EndpointConfig, error) {
	if endpoint == "" {
		if c.DefaultEndpoint == "" {
			return nil, errors.New("no explicit endpoint and default not configured")
		}
		endpoint = c.DefaultEndpoint
	}

	clientCfg, ok := c.Endpoints[endpoint]
	if !ok {
		return nil, fmt.Errorf("endpoint %s not configured", endpoint)
	}

	return &clientCfg, nil
}

func (c *Config) McpConfig(mcp string) (*McpConfig, error) {
	cfg, ok := c.McpServers[mcp]
	if !ok {
		return nil, fmt.Errorf("mcp server %s not configured", mcp)
	}

	return &cfg, nil
}

func (c *EndpointConfig) NewHTTPClient() *http.Client {
	tlsConfig := &tls.Config{
		//nolint: gosec // allow _explicit_ user configured skipping
		InsecureSkipVerify: c.InsecureSkipTLS,
		// linter wants a min version, but defaulttransport doesn't
		// set one, so we may need to reconsider having this or
		// maybe adding a way to tune it.
		MinVersion: tls.VersionTLS12,
	}

	if c.CACerts != "" {
		log.Trace().Str("cacerts", c.CACerts).Msg("adding configured certificate")
		rootCAs, _ := x509.SystemCertPool()
		if rootCAs == nil {
			rootCAs = x509.NewCertPool()
		}

		ok := rootCAs.AppendCertsFromPEM([]byte(c.CACerts))
		if !ok {
			log.Warn().Msg("failed to add configured endpoint cacerts")
		}

		tlsConfig.RootCAs = rootCAs
	}

	var transport http.RoundTripper
	transport = &http.Transport{TLSClientConfig: tlsConfig}

	if log.Trace().Enabled() {
		transport = &loggingTransport{wrapped: transport}
	}

	return &http.Client{Transport: transport}
}

func (c *EndpointConfig) NewClient() openai.Client {
	opts := []option.RequestOption{
		option.WithHTTPClient(c.NewHTTPClient()),
	}
	if c.BaseURL != "" {
		opts = append(opts, option.WithBaseURL(c.BaseURL))
	}
	if c.AuthToken != "" {
		opts = append(opts, option.WithAPIKey(c.AuthToken))
	}
	if c.OrgID != "" {
		opts = append(opts, option.WithOrganization(c.OrgID))
	}
	return openai.NewClient(opts...)
}

func (c *McpConfig) NewMcpClient(ctx context.Context) (*mcp.Client, error) {
	mcp, err := mcp.NewClient(
		ctx,
		c.Name,
		c.URL,
		mcp.WithAutoConfirm(c.AutoConfirm),
		mcp.WithConfirmDefaultYes(c.ConfirmDefaultYes))
	if err != nil {
		return nil, fmt.Errorf("new mcp client: %w", err)
	}

	return mcp, nil
}

func (s *loggingTransport) RoundTrip(r *http.Request) (*http.Response, error) {
	dumpBody := false
	if _, ok := os.LookupEnv("HTTP_CLIENT_DUMP_BODY"); ok {
		dumpBody = true
	} else {
		log.Trace().Msg("to dump request/response body content, set env HTTP_CLIENT_DUMP_BODY=1")
	}

	req, _ := httputil.DumpRequestOut(r, dumpBody)
	log.Trace().Bytes("request", req).Msg("request")

	transport := s.wrapped
	if transport == nil {
		transport = http.DefaultTransport
	}

	resp, err := transport.RoundTrip(r)
	if err != nil {
		// err is returned after dumping the response
		err = fmt.Errorf("inspected response: %w", err)
	}

	if resp == nil {
		log.Trace().Err(err).Bytes("response", []byte{}).Msg("response")
	} else {
		res, _ := httputil.DumpResponse(resp, dumpBody)
		log.Trace().Err(err).Bytes("response", res).Msg("response")
	}

	return resp, err
}
