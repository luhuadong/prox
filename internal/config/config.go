package config

import (
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net"
	"net/url"
	"os"
	"path/filepath"
	"strconv"
	"strings"
	"time"
)

const (
	DefaultProxyURL       = "http://127.0.0.1:7890"
	DefaultCheckURL       = "https://www.google.com/generate_204"
	DefaultExpectedStatus = 204
	DefaultConnectTimeout = 2 * time.Second
	DefaultRequestTimeout = 8 * time.Second
)

var DefaultNoProxy = []string{"localhost", "127.0.0.1", "::1", ".local"}

// Config is the resolved, validated configuration used by prox.
type Config struct {
	ProxyURL       string
	NoProxy        []string
	CheckURL       string
	ExpectedStatus int
	ConnectTimeout time.Duration
	RequestTimeout time.Duration
	Source         string
}

type fileConfig struct {
	ProxyURL       *string   `json:"proxy_url"`
	NoProxy        *[]string `json:"no_proxy"`
	CheckURL       *string   `json:"check_url"`
	ExpectedStatus *int      `json:"expected_status"`
	ConnectTimeout *string   `json:"connect_timeout"`
	RequestTimeout *string   `json:"request_timeout"`
}

// Defaults returns a fresh copy of the built-in configuration.
func Defaults() Config {
	return Config{
		ProxyURL:       DefaultProxyURL,
		NoProxy:        append([]string(nil), DefaultNoProxy...),
		CheckURL:       DefaultCheckURL,
		ExpectedStatus: DefaultExpectedStatus,
		ConnectTimeout: DefaultConnectTimeout,
		RequestTimeout: DefaultRequestTimeout,
		Source:         "built-in defaults",
	}
}

// DefaultPath resolves the user configuration path according to XDG conventions.
func DefaultPath() (string, error) {
	if xdg := os.Getenv("XDG_CONFIG_HOME"); xdg != "" {
		if !filepath.IsAbs(xdg) {
			return "", fmt.Errorf("XDG_CONFIG_HOME must be an absolute path")
		}
		return filepath.Join(xdg, "prox", "config.json"), nil
	}

	home, err := os.UserHomeDir()
	if err != nil {
		return "", fmt.Errorf("resolve home directory: %w", err)
	}
	return filepath.Join(home, ".config", "prox", "config.json"), nil
}

// Load reads an optional JSON configuration. An empty explicitPath uses the
// default XDG location; a missing default file is not an error.
func Load(explicitPath string) (Config, error) {
	cfg := Defaults()
	path := explicitPath
	if path == "" {
		var err error
		path, err = DefaultPath()
		if err != nil {
			return Config{}, err
		}
	}

	data, err := os.ReadFile(path)
	if err != nil {
		if errors.Is(err, os.ErrNotExist) && explicitPath == "" {
			if err := cfg.Validate(); err != nil {
				return Config{}, err
			}
			return cfg, nil
		}
		return Config{}, fmt.Errorf("read config %s: %w", path, err)
	}

	var raw fileConfig
	decoder := json.NewDecoder(bytes.NewReader(data))
	decoder.DisallowUnknownFields()
	if err := decoder.Decode(&raw); err != nil {
		return Config{}, fmt.Errorf("parse config %s: %w", path, err)
	}
	if err := ensureJSONEOF(decoder); err != nil {
		return Config{}, fmt.Errorf("parse config %s: %w", path, err)
	}

	if raw.ProxyURL != nil {
		cfg.ProxyURL = *raw.ProxyURL
	}
	if raw.NoProxy != nil {
		cfg.NoProxy = append([]string(nil), (*raw.NoProxy)...)
	}
	if raw.CheckURL != nil {
		cfg.CheckURL = *raw.CheckURL
	}
	if raw.ExpectedStatus != nil {
		cfg.ExpectedStatus = *raw.ExpectedStatus
	}
	if raw.ConnectTimeout != nil {
		cfg.ConnectTimeout, err = parseDuration("connect_timeout", *raw.ConnectTimeout)
		if err != nil {
			return Config{}, err
		}
	}
	if raw.RequestTimeout != nil {
		cfg.RequestTimeout, err = parseDuration("request_timeout", *raw.RequestTimeout)
		if err != nil {
			return Config{}, err
		}
	}
	cfg.Source = path

	if err := cfg.Validate(); err != nil {
		return Config{}, fmt.Errorf("validate config %s: %w", path, err)
	}
	return cfg, nil
}

func ensureJSONEOF(decoder *json.Decoder) error {
	var extra any
	if err := decoder.Decode(&extra); !errors.Is(err, io.EOF) {
		if err == nil {
			return errors.New("multiple JSON values are not allowed")
		}
		return err
	}
	return nil
}

func parseDuration(name, value string) (time.Duration, error) {
	duration, err := time.ParseDuration(value)
	if err != nil {
		return 0, fmt.Errorf("parse %s: %w", name, err)
	}
	if duration <= 0 {
		return 0, fmt.Errorf("%s must be greater than zero", name)
	}
	return duration, nil
}

// Validate checks the complete resolved configuration.
func (c Config) Validate() error {
	if _, err := ParseProxyURL(c.ProxyURL); err != nil {
		return err
	}
	if err := ValidateTargetURL(c.CheckURL); err != nil {
		return fmt.Errorf("invalid check_url: %w", err)
	}
	if c.ExpectedStatus < 100 || c.ExpectedStatus > 599 {
		return fmt.Errorf("expected_status must be between 100 and 599")
	}
	if c.ConnectTimeout <= 0 {
		return fmt.Errorf("connect_timeout must be greater than zero")
	}
	if c.RequestTimeout <= 0 {
		return fmt.Errorf("request_timeout must be greater than zero")
	}
	for _, entry := range c.NoProxy {
		if strings.TrimSpace(entry) == "" {
			return fmt.Errorf("no_proxy entries must not be empty")
		}
		if entry != strings.TrimSpace(entry) {
			return fmt.Errorf("no_proxy entry %q must not contain surrounding whitespace", entry)
		}
		if strings.Contains(entry, ",") {
			return fmt.Errorf("no_proxy entry %q must not contain a comma", entry)
		}
	}
	return nil
}

// ParseProxyURL parses a supported HTTP proxy URL.
func ParseProxyURL(raw string) (*url.URL, error) {
	u, err := url.Parse(raw)
	if err != nil {
		return nil, fmt.Errorf("invalid proxy_url: %w", err)
	}
	if u.Scheme != "http" {
		return nil, fmt.Errorf("unsupported proxy scheme %q; prox currently supports http only", u.Scheme)
	}
	if u.Hostname() == "" {
		return nil, fmt.Errorf("proxy_url must include a host")
	}
	if u.User != nil {
		return nil, fmt.Errorf("proxy authentication is not currently supported")
	}
	if u.Path != "" && u.Path != "/" {
		return nil, fmt.Errorf("proxy_url must not include a path")
	}
	if u.RawQuery != "" || u.Fragment != "" {
		return nil, fmt.Errorf("proxy_url must not include a query or fragment")
	}

	port := u.Port()
	if port != "" {
		number, err := strconv.Atoi(port)
		if err != nil || number < 1 || number > 65535 {
			return nil, fmt.Errorf("proxy_url contains an invalid port")
		}
	}
	return u, nil
}

// ValidateTargetURL validates a health-check target.
func ValidateTargetURL(raw string) error {
	u, err := url.Parse(raw)
	if err != nil {
		return err
	}
	if u.Scheme != "http" && u.Scheme != "https" {
		return fmt.Errorf("scheme must be http or https")
	}
	if u.Hostname() == "" {
		return fmt.Errorf("URL must include a host")
	}
	if u.User != nil {
		return fmt.Errorf("URL must not include user information")
	}
	return nil
}

// ProxyEndpoint returns a host:port suitable for a TCP connection.
func (c Config) ProxyEndpoint() (string, error) {
	u, err := ParseProxyURL(c.ProxyURL)
	if err != nil {
		return "", err
	}
	port := u.Port()
	if port == "" {
		port = "80"
	}
	return net.JoinHostPort(u.Hostname(), port), nil
}

// RedactedProxyURL returns a safe display form of the proxy URL.
func (c Config) RedactedProxyURL() string {
	u, err := url.Parse(c.ProxyURL)
	if err != nil {
		return "<invalid>"
	}
	if u.User != nil {
		username := u.User.Username()
		u.User = url.UserPassword(username, "****")
	}
	return u.String()
}
