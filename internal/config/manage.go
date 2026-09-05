package config

import (
	"encoding/json"
	"errors"
	"fmt"
	"os"
	"path/filepath"
)

var ErrConfigExists = errors.New("configuration file already exists")

type effectiveConfig struct {
	ProxyURL       string   `json:"proxy_url"`
	NoProxy        []string `json:"no_proxy"`
	CheckURL       string   `json:"check_url"`
	ExpectedStatus int      `json:"expected_status"`
	ConnectTimeout string   `json:"connect_timeout"`
	RequestTimeout string   `json:"request_timeout"`
}

// MarshalEffective returns the complete resolved configuration as indented JSON.
func MarshalEffective(cfg Config) ([]byte, error) {
	data, err := json.MarshalIndent(effectiveConfig{
		ProxyURL:       cfg.ProxyURL,
		NoProxy:        cfg.NoProxy,
		CheckURL:       cfg.CheckURL,
		ExpectedStatus: cfg.ExpectedStatus,
		ConnectTimeout: cfg.ConnectTimeout.String(),
		RequestTimeout: cfg.RequestTimeout.String(),
	}, "", "  ")
	if err != nil {
		return nil, fmt.Errorf("encode configuration: %w", err)
	}
	return append(data, '\n'), nil
}

// Initialize writes a minimal user configuration without overwriting an
// existing file unless force is true.
func Initialize(path, proxyURL string, force bool) error {
	cfg := Defaults()
	cfg.ProxyURL = proxyURL
	if err := cfg.Validate(); err != nil {
		return err
	}

	data, err := json.MarshalIndent(struct {
		ProxyURL string `json:"proxy_url"`
	}{ProxyURL: proxyURL}, "", "  ")
	if err != nil {
		return fmt.Errorf("encode configuration: %w", err)
	}
	data = append(data, '\n')

	directory := filepath.Dir(path)
	if err := os.MkdirAll(directory, 0o700); err != nil {
		return fmt.Errorf("create config directory %s: %w", directory, err)
	}

	temporary, err := os.CreateTemp(directory, ".config.json.tmp-*")
	if err != nil {
		return fmt.Errorf("create temporary config: %w", err)
	}
	temporaryPath := temporary.Name()
	defer os.Remove(temporaryPath)

	if err := temporary.Chmod(0o600); err != nil {
		temporary.Close()
		return fmt.Errorf("set config permissions: %w", err)
	}
	if _, err := temporary.Write(data); err != nil {
		temporary.Close()
		return fmt.Errorf("write config: %w", err)
	}
	if err := temporary.Sync(); err != nil {
		temporary.Close()
		return fmt.Errorf("sync config: %w", err)
	}
	if err := temporary.Close(); err != nil {
		return fmt.Errorf("close config: %w", err)
	}

	if force {
		if err := os.Rename(temporaryPath, path); err != nil {
			return fmt.Errorf("install config %s: %w", path, err)
		}
		return nil
	}

	if err := os.Link(temporaryPath, path); err != nil {
		if errors.Is(err, os.ErrExist) {
			return fmt.Errorf("%w: %s", ErrConfigExists, path)
		}
		return fmt.Errorf("install config %s: %w", path, err)
	}
	return nil
}
