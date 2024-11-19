package askai

import (
	"errors"
	"fmt"
	"net/http"
	"net/http/httputil"
	"os"
	"path/filepath"

	"dario.cat/mergo"
	"github.com/pastdev/askai/pkg/log"
	"github.com/sashabaranov/go-openai"
	"gopkg.in/yaml.v3"
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
	Endpoints       map[string]ClientConfig `json:"endpoints" yaml:"endpoints"`
	DefaultEndpoint string                  `json:"default_endpoint" yaml:"default_endpoint"`
}

// ClientConfig is a configuration of a client.
type ClientConfig struct {
	AuthToken          string         `json:"auth_token" yaml:"auth_token"`
	BaseURL            string         `json:"base_url" yaml:"base_url"`
	OrgID              string         `json:"org_id" yaml:"org_id"`
	APIType            openai.APIType `json:"api_type" yaml:"api_type"`
	APIVersion         string         `json:"api_version" yaml:"api_version"`
	EmptyMessagesLimit uint           `json:"empty_messages_limit" yaml:"empty_messages_limit"`
}

// ConfigSource contains sources of yaml/json to load (unmarshal) into a Config.
type ConfigSource struct {
	// Dirs is a list of directories containing config files to load. The
	// directories will be processed in order loading the files within each
	// directory in order sorted by filename with later values overriding
	// existing values.
	Dirs []string
	// Files are config files to load. They will be loaded after files found in
	// Dirs and, like dirs, later values override former.
	Files []string
	// Memory are strings containing config yaml to load. They will be loaded
	// after files found in Filess and, like dirs, later values override former.
	Memory []string
}

type loggingTransport struct {
	wrapped http.RoundTripper
}

func (c *Config) NewClient(endpoint string) (*openai.Client, error) {
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

	return clientCfg.NewClient(), nil
}

func (c *ClientConfig) NewClient() *openai.Client {
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

	if log.Trace().Enabled() {
		cfg.HTTPClient = &http.Client{
			Transport: &loggingTransport{wrapped: cfg.HTTPClient.Transport},
		}
	}

	return openai.NewClientWithConfig(cfg)
}

func (c *Config) LoadBytes(b []byte) error {
	var overrides *Config
	err := yaml.Unmarshal(b, &overrides)
	if err != nil {
		return fmt.Errorf("unmarshal config yml: %w", err)
	}

	err = mergo.Merge(c, overrides, mergo.WithOverride)
	if err != nil {
		return fmt.Errorf("merge config overrides: %w", err)
	}

	return nil
}

func (c *Config) LoadFile(f string) error {
	b, err := os.ReadFile(f)
	if err != nil {
		log.Debug().Str("file", f).Msg("config not found")
		return nil
	}

	return c.LoadBytes(b)
}

func (c *Config) LoadString(s string) error {
	return c.LoadBytes([]byte(s))
}

func (s ConfigSource) Load() (*Config, error) {
	c := &Config{}
	for _, dir := range s.Dirs {
		dir = s.normalizePath(dir)
		listing, err := os.ReadDir(dir)
		if err != nil {
			log.Debug().Str("dir", dir).Msg("no configs found")
			continue
		}

		for _, entry := range listing {
			name := entry.Name()
			if !entry.Type().IsRegular() {
				if entry.IsDir() {
					log.Debug().
						Str("dir", dir).
						Str("subdir", entry.Name()).
						Msg("skipping subdir")
					continue
				}

				path, err := filepath.EvalSymlinks(filepath.Join(dir, entry.Name()))
				if err != nil {
					return c, fmt.Errorf("eval symlink: %w", err)
				}
				entry, err := os.Stat(path)
				if err != nil {
					return c, fmt.Errorf("stat: %w", err)
				}
				if entry.IsDir() {
					log.Debug().
						Str("dir", dir).
						Str("symlinkSubdir", entry.Name()).
						Msg("skipping subdir")
					continue
				}
				name = entry.Name()
			}

			err = c.LoadFile(filepath.Join(dir, name))
			if err != nil {
				return c, fmt.Errorf("load from dir: %w", err)
			}
		}
	}

	for _, f := range s.Files {
		err := c.LoadFile(s.normalizePath(f))
		if err != nil {
			return c, fmt.Errorf("load from file: %w", err)
		}
	}

	for _, m := range s.Memory {
		err := c.LoadBytes([]byte(m))
		if err != nil {
			return c, fmt.Errorf("load from memory: %w", err)
		}
	}

	return c, nil
}

func (s ConfigSource) normalizePath(path string) string {
	if path[0] == '~' {
		homeDir, err := os.UserHomeDir()
		if err != nil {
			log.Trace().Err(err).Msg("User home directory not defined")
			return path
		}
		path = filepath.Join(homeDir, path[1:])
	}
	return path
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

	res, _ := httputil.DumpResponse(resp, dumpBody)
	log.Trace().Err(err).Bytes("response", res).Msg("response")

	return resp, err
}
