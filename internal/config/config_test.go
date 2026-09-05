package config

import (
	"os"
	"path/filepath"
	"testing"
	"time"
)

func TestLoadDefaultsWhenDefaultFileIsMissing(t *testing.T) {
	t.Setenv("XDG_CONFIG_HOME", t.TempDir())

	cfg, err := Load("")
	if err != nil {
		t.Fatalf("Load() error = %v", err)
	}
	if cfg.ProxyURL != DefaultProxyURL {
		t.Fatalf("ProxyURL = %q, want %q", cfg.ProxyURL, DefaultProxyURL)
	}
	if cfg.Source != "built-in defaults" {
		t.Fatalf("Source = %q, want built-in defaults", cfg.Source)
	}
}

func TestLoadOverridesDefaults(t *testing.T) {
	path := filepath.Join(t.TempDir(), "config.json")
	data := []byte(`{
  "proxy_url": "http://localhost:8080",
  "no_proxy": ["localhost"],
  "check_url": "https://example.com/health",
  "expected_status": 200,
  "connect_timeout": "500ms",
  "request_timeout": "3s"
}`)
	if err := os.WriteFile(path, data, 0o600); err != nil {
		t.Fatal(err)
	}

	cfg, err := Load(path)
	if err != nil {
		t.Fatalf("Load() error = %v", err)
	}
	if cfg.ProxyURL != "http://localhost:8080" {
		t.Fatalf("ProxyURL = %q", cfg.ProxyURL)
	}
	if cfg.ConnectTimeout != 500*time.Millisecond {
		t.Fatalf("ConnectTimeout = %s", cfg.ConnectTimeout)
	}
	if cfg.RequestTimeout != 3*time.Second {
		t.Fatalf("RequestTimeout = %s", cfg.RequestTimeout)
	}
	if cfg.Source != path {
		t.Fatalf("Source = %q, want %q", cfg.Source, path)
	}
}

func TestLoadRejectsUnknownFields(t *testing.T) {
	path := filepath.Join(t.TempDir(), "config.json")
	if err := os.WriteFile(path, []byte(`{"proxy_url":"http://localhost:7890","unknown":true}`), 0o600); err != nil {
		t.Fatal(err)
	}
	if _, err := Load(path); err == nil {
		t.Fatal("Load() succeeded with an unknown field")
	}
}

func TestParseProxyURL(t *testing.T) {
	tests := []struct {
		name    string
		value   string
		wantErr bool
	}{
		{name: "http", value: "http://127.0.0.1:7890"},
		{name: "ipv6", value: "http://[::1]:7890"},
		{name: "default port", value: "http://proxy.example.com"},
		{name: "socks unsupported", value: "socks5://127.0.0.1:7891", wantErr: true},
		{name: "auth unsupported", value: "http://user:pass@localhost:7890", wantErr: true},
		{name: "missing host", value: "http://", wantErr: true},
		{name: "invalid port", value: "http://localhost:99999", wantErr: true},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			_, err := ParseProxyURL(test.value)
			if (err != nil) != test.wantErr {
				t.Fatalf("ParseProxyURL(%q) error = %v, wantErr %v", test.value, err, test.wantErr)
			}
		})
	}
}

func TestProxyEndpoint(t *testing.T) {
	cfg := Defaults()
	cfg.ProxyURL = "http://[::1]:7890"
	endpoint, err := cfg.ProxyEndpoint()
	if err != nil {
		t.Fatal(err)
	}
	if endpoint != "[::1]:7890" {
		t.Fatalf("endpoint = %q", endpoint)
	}
}
