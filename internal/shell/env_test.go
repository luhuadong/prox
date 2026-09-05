package shell

import (
	"slices"
	"strings"
	"testing"

	"prox/internal/config"
)

func TestQuoteBash(t *testing.T) {
	got := QuoteBash("one'two;$(three)")
	want := `'one'\''two;$(three)'`
	if got != want {
		t.Fatalf("QuoteBash() = %q, want %q", got, want)
	}
}

func TestMergeEnvironment(t *testing.T) {
	cfg := config.Defaults()
	current := []string{
		"PATH=/usr/bin",
		"http_proxy=http://old.example:8080",
		"HTTP_PROXY=http://leave-me.example:8080",
	}
	merged := MergeEnvironment(current, cfg)

	wants := []string{
		"PATH=/usr/bin",
		"http_proxy=" + cfg.ProxyURL,
		"HTTP_PROXY=http://leave-me.example:8080",
		"https_proxy=" + cfg.ProxyURL,
		"NO_PROXY=localhost,127.0.0.1,::1,.local",
	}
	for _, want := range wants {
		if !slices.Contains(merged, want) {
			t.Errorf("merged environment does not contain %q: %#v", want, merged)
		}
	}
}

func TestBashEnvironmentScriptEscapesValues(t *testing.T) {
	cfg := config.Defaults()
	cfg.NoProxy = []string{"safe'$(touch /tmp/should-not-exist)"}
	script := BashEnvironmentScript(cfg)
	want := "export no_proxy='safe'\\''$(touch /tmp/should-not-exist)'\n"
	if !strings.Contains(script, want) {
		t.Fatalf("script does not contain safely quoted value %q: %q", want, script)
	}
}
