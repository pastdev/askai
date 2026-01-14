package config

import (
	"crypto/tls"
	"crypto/x509"
	"errors"
	"fmt"
	"net/http"
	"net/http/httputil"
	"os"

	oai "github.com/openai/openai-go/v3"
	oaiopt "github.com/openai/openai-go/v3/option"
	"github.com/pastdev/askai/pkg/log"
	"github.com/sashabaranov/go-openai"
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
`

type Config struct {
	Endpoints       map[string]EndpointConfig `json:"endpoints" yaml:"endpoints"`
	DefaultEndpoint string                    `json:"default_endpoint" yaml:"default_endpoint"`
}

// EndpointConfig is a configuration of a client.
type EndpointConfig struct {
	APIType                openai.APIType               `json:"api_type" yaml:"api_type"`
	APIVersion             string                       `json:"api_version" yaml:"api_version"`
	AuthToken              string                       `json:"auth_token" yaml:"auth_token"`
	BaseURL                string                       `json:"base_url" yaml:"base_url"`
	ChatCompletionDefaults *oai.ChatCompletionNewParams `json:"chat_completion_defaults" yaml:"chat_completion_defaults"`
	CACerts                string                       `json:"cacerts" yaml:"cacerts"`
	EmptyMessagesLimit     uint                         `json:"empty_messages_limit" yaml:"empty_messages_limit"`
	ImageDefaults          *openai.ImageRequest         `json:"image_defaults" yaml:"image_defaults"`
	InsecureSkipTLS        bool                         `json:"insecure_skip_tls" yaml:"insecure_skip_tls"`
	OrgID                  string                       `json:"org_id" yaml:"org_id"`
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

func (c *EndpointConfig) NewClientLegacy() *openai.Client {
	cfg := openai.DefaultConfig(c.AuthToken)

	if c.BaseURL != "" {
		cfg.BaseURL = c.BaseURL
	}

	if c.OrgID != "" {
		cfg.OrgID = c.OrgID
	}

	if c.APIType != "" {
		cfg.APIType = c.APIType
	}
	if c.APIVersion != "" {
		cfg.APIVersion = c.APIVersion
	}

	if c.EmptyMessagesLimit > 0 {
		cfg.EmptyMessagesLimit = c.EmptyMessagesLimit
	}

	cfg.HTTPClient = c.NewHTTPClient()

	return openai.NewClientWithConfig(cfg)
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

func (c *EndpointConfig) NewClient() oai.Client {
	opts := []oaiopt.RequestOption{
		oaiopt.WithHTTPClient(c.NewHTTPClient()),
	}
	if c.BaseURL != "" {
		opts = append(opts, oaiopt.WithBaseURL(c.BaseURL))
	}
	if c.AuthToken != "" {
		opts = append(opts, oaiopt.WithAPIKey(c.AuthToken))
	}
	if c.OrgID != "" {
		opts = append(opts, oaiopt.WithOrganization(c.OrgID))
	}
	return oai.NewClient(opts...)
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
