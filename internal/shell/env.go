package shell

import (
	"fmt"
	"strings"

	"prox/internal/config"
)

// ManagedVariables is the complete set changed by prox V0.1.
var ManagedVariables = []string{
	"http_proxy",
	"https_proxy",
	"all_proxy",
	"HTTPS_PROXY",
	"ALL_PROXY",
	"no_proxy",
	"NO_PROXY",
}

// EnvironmentValues builds the proxy environment for a configuration.
func EnvironmentValues(cfg config.Config) map[string]string {
	noProxy := strings.Join(cfg.NoProxy, ",")
	return map[string]string{
		"http_proxy":  cfg.ProxyURL,
		"https_proxy": cfg.ProxyURL,
		"all_proxy":   cfg.ProxyURL,
		"HTTPS_PROXY": cfg.ProxyURL,
		"ALL_PROXY":   cfg.ProxyURL,
		"no_proxy":    noProxy,
		"NO_PROXY":    noProxy,
	}
}

// BashEnvironmentScript returns shell code containing only fixed-name assignments.
func BashEnvironmentScript(cfg config.Config) string {
	values := EnvironmentValues(cfg)
	var builder strings.Builder
	for _, name := range ManagedVariables {
		fmt.Fprintf(&builder, "export %s=%s\n", name, QuoteBash(values[name]))
	}
	fmt.Fprintf(&builder, "__PROX_PROXY_DISPLAY=%s\n", QuoteBash(cfg.RedactedProxyURL()))
	return builder.String()
}

// QuoteBash returns one shell word using single-quote escaping.
func QuoteBash(value string) string {
	return "'" + strings.ReplaceAll(value, "'", "'\\''") + "'"
}

// MergeEnvironment overlays prox-managed values onto an environment list.
func MergeEnvironment(current []string, cfg config.Config) []string {
	managed := EnvironmentValues(cfg)
	seen := make(map[string]bool, len(managed))
	result := make([]string, 0, len(current)+len(managed))

	for _, entry := range current {
		index := strings.IndexByte(entry, '=')
		if index <= 0 {
			result = append(result, entry)
			continue
		}
		name := entry[:index]
		value, ok := managed[name]
		if !ok {
			result = append(result, entry)
			continue
		}
		if seen[name] {
			continue
		}
		result = append(result, name+"="+value)
		seen[name] = true
	}

	for _, name := range ManagedVariables {
		if !seen[name] {
			result = append(result, name+"="+managed[name])
		}
	}
	return result
}
