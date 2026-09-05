package app

import (
	"bytes"
	"fmt"
	"net"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestVersion(t *testing.T) {
	var stdout bytes.Buffer
	var stderr bytes.Buffer
	code := Run([]string{"version"}, strings.NewReader(""), &stdout, &stderr, "1.2.3")
	if code != 0 {
		t.Fatalf("code = %d, stderr = %q", code, stderr.String())
	}
	if stdout.String() != "prox 1.2.3\n" {
		t.Fatalf("stdout = %q", stdout.String())
	}
}

func TestUnknownCommand(t *testing.T) {
	var stdout bytes.Buffer
	var stderr bytes.Buffer
	code := Run([]string{"unknown"}, strings.NewReader(""), &stdout, &stderr, "test")
	if code != 2 {
		t.Fatalf("code = %d", code)
	}
	if !strings.Contains(stderr.String(), "unknown command") {
		t.Fatalf("stderr = %q", stderr.String())
	}
}

func TestOnWithoutHookExplainsInitialization(t *testing.T) {
	var stdout bytes.Buffer
	var stderr bytes.Buffer
	code := Run([]string{"on"}, strings.NewReader(""), &stdout, &stderr, "test")
	if code != 2 {
		t.Fatalf("code = %d", code)
	}
	if !strings.Contains(stderr.String(), "prox init bash") {
		t.Fatalf("stderr = %q", stderr.String())
	}
}

func TestInitBash(t *testing.T) {
	var stdout bytes.Buffer
	var stderr bytes.Buffer
	code := Run([]string{"init", "bash"}, strings.NewReader(""), &stdout, &stderr, "1.2.3")
	if code != 0 {
		t.Fatalf("code = %d, stderr = %q", code, stderr.String())
	}
	if !strings.Contains(stdout.String(), "__PROX_HOOK_VERSION='1.2.3'") {
		t.Fatal("generated hook does not contain the version")
	}
}

func TestCheckLocal(t *testing.T) {
	listener, err := net.Listen("tcp", "127.0.0.1:0")
	if err != nil {
		t.Fatal(err)
	}
	defer listener.Close()

	configPath := filepath.Join(t.TempDir(), "config.json")
	configData := fmt.Sprintf(`{"proxy_url":%q}`, "http://"+listener.Addr().String())
	if err := os.WriteFile(configPath, []byte(configData), 0o600); err != nil {
		t.Fatal(err)
	}

	var stdout bytes.Buffer
	var stderr bytes.Buffer
	code := Run(
		[]string{"--config", configPath, "check", "--local"},
		strings.NewReader(""),
		&stdout,
		&stderr,
		"test",
	)
	if code != 0 {
		t.Fatalf("code = %d, stdout = %q, stderr = %q", code, stdout.String(), stderr.String())
	}
	if !strings.Contains(stdout.String(), "Proxy endpoint is reachable") {
		t.Fatalf("stdout = %q", stdout.String())
	}
}

func TestConfigCommands(t *testing.T) {
	configHome := t.TempDir()
	t.Setenv("XDG_CONFIG_HOME", configHome)
	path := filepath.Join(configHome, "prox", "config.json")

	var stdout bytes.Buffer
	var stderr bytes.Buffer
	code := Run([]string{"config", "path"}, strings.NewReader(""), &stdout, &stderr, "test")
	if code != 0 || strings.TrimSpace(stdout.String()) != path {
		t.Fatalf("path: code = %d, stdout = %q, stderr = %q", code, stdout.String(), stderr.String())
	}

	stdout.Reset()
	stderr.Reset()
	code = Run(
		[]string{"config", "init", "--proxy", "http://localhost:8080"},
		strings.NewReader(""),
		&stdout,
		&stderr,
		"test",
	)
	if code != 0 {
		t.Fatalf("init: code = %d, stdout = %q, stderr = %q", code, stdout.String(), stderr.String())
	}

	stdout.Reset()
	stderr.Reset()
	code = Run([]string{"config", "show"}, strings.NewReader(""), &stdout, &stderr, "test")
	if code != 0 || !strings.Contains(stdout.String(), `"proxy_url": "http://localhost:8080"`) {
		t.Fatalf("show: code = %d, stdout = %q, stderr = %q", code, stdout.String(), stderr.String())
	}

	stdout.Reset()
	stderr.Reset()
	code = Run([]string{"config", "validate"}, strings.NewReader(""), &stdout, &stderr, "test")
	if code != 0 || !strings.Contains(stdout.String(), "Config is valid") {
		t.Fatalf("validate: code = %d, stdout = %q, stderr = %q", code, stdout.String(), stderr.String())
	}

	stdout.Reset()
	stderr.Reset()
	code = Run([]string{"config", "init"}, strings.NewReader(""), &stdout, &stderr, "test")
	if code != 1 || !strings.Contains(stderr.String(), "--force") {
		t.Fatalf("duplicate init: code = %d, stdout = %q, stderr = %q", code, stdout.String(), stderr.String())
	}
}
