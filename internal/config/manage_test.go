package config

import (
	"errors"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestMarshalEffective(t *testing.T) {
	data, err := MarshalEffective(Defaults())
	if err != nil {
		t.Fatal(err)
	}
	text := string(data)
	if !strings.Contains(text, `"connect_timeout": "2s"`) {
		t.Fatalf("output = %q", text)
	}
	if !strings.HasSuffix(text, "\n") {
		t.Fatal("output does not end with a newline")
	}
}

func TestInitialize(t *testing.T) {
	path := filepath.Join(t.TempDir(), "prox", "config.json")
	if err := Initialize(path, "http://localhost:8080", false); err != nil {
		t.Fatal(err)
	}

	info, err := os.Stat(path)
	if err != nil {
		t.Fatal(err)
	}
	if info.Mode().Perm() != 0o600 {
		t.Fatalf("mode = %o", info.Mode().Perm())
	}

	cfg, err := Load(path)
	if err != nil {
		t.Fatal(err)
	}
	if cfg.ProxyURL != "http://localhost:8080" {
		t.Fatalf("proxy_url = %q", cfg.ProxyURL)
	}
	if cfg.CheckURL != DefaultCheckURL {
		t.Fatalf("check_url = %q", cfg.CheckURL)
	}

	err = Initialize(path, "http://localhost:9090", false)
	if !errors.Is(err, ErrConfigExists) {
		t.Fatalf("error = %v", err)
	}
	if err := Initialize(path, "http://localhost:9090", true); err != nil {
		t.Fatal(err)
	}
	cfg, err = Load(path)
	if err != nil {
		t.Fatal(err)
	}
	if cfg.ProxyURL != "http://localhost:9090" {
		t.Fatalf("proxy_url after force = %q", cfg.ProxyURL)
	}
}

func TestInitializeRejectsInvalidProxy(t *testing.T) {
	path := filepath.Join(t.TempDir(), "config.json")
	if err := Initialize(path, "socks5://localhost:1080", false); err == nil {
		t.Fatal("Initialize unexpectedly accepted an unsupported proxy")
	}
	if _, err := os.Stat(path); !errors.Is(err, os.ErrNotExist) {
		t.Fatalf("config was created: %v", err)
	}
}
