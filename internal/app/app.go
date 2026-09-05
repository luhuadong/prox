package app

import (
	"context"
	"errors"
	"flag"
	"fmt"
	"io"
	"os"
	"strings"
	"time"

	"github.com/luhuadong/prox/internal/config"
	"github.com/luhuadong/prox/internal/health"
	"github.com/luhuadong/prox/internal/runner"
	"github.com/luhuadong/prox/internal/shell"
)

const usageText = `prox safely manages proxy environment variables for terminal commands.

Usage:
  prox [--config PATH] <command>

Commands:
  init bash             Print the Bash integration script
  completion bash       Print Bash completion
  config <command>      Inspect or create user configuration
  on                    Enable the proxy in the current Shell (requires init)
  off                   Restore the previous Shell environment (requires init)
  status                Show activation state without accessing the network
  check [options]       Check the configured proxy
  run -- <command>      Run one command with proxy variables
  version               Print the version
  help                  Print this help

Check options:
  --local               Only validate configuration and the proxy endpoint
  --url URL             Override the configured health-check URL
  --quiet               Suppress normal output

The default proxy is http://127.0.0.1:7890.
`

type globalOptions struct {
	configPath string
	arguments  []string
}

// Run executes prox and returns the process exit status.
func Run(arguments []string, stdin io.Reader, stdout, stderr io.Writer, version string) int {
	global, err := parseGlobalOptions(arguments)
	if err != nil {
		fmt.Fprintf(stderr, "prox: %v\n", err)
		return 2
	}
	if len(global.arguments) == 0 {
		fmt.Fprint(stdout, usageText)
		return 0
	}

	command := global.arguments[0]
	commandArguments := global.arguments[1:]
	switch command {
	case "help", "--help", "-h":
		fmt.Fprint(stdout, usageText)
		return 0
	case "version", "--version":
		if len(commandArguments) != 0 {
			fmt.Fprintln(stderr, "prox: version does not accept arguments")
			return 2
		}
		fmt.Fprintf(stdout, "prox %s\n", version)
		return 0
	case "init":
		return runInit(commandArguments, stdout, stderr, version)
	case "completion":
		return runCompletion(commandArguments, stdout, stderr)
	case "config":
		return runConfig(commandArguments, global.configPath, stdout, stderr)
	case "on", "off":
		return shellIntegrationRequired(command, stderr)
	case "status":
		return runStatus(commandArguments, global.configPath, stdout, stderr)
	case "check":
		return runCheck(commandArguments, global.configPath, stdout, stderr)
	case "run":
		return runCommand(commandArguments, global.configPath, stdin, stdout, stderr)
	case "__shell-env":
		return runShellEnvironment(commandArguments, global.configPath, stdout, stderr)
	default:
		fmt.Fprintf(stderr, "prox: unknown command %q\n", command)
		fmt.Fprintln(stderr, "Run `prox help` for usage.")
		return 2
	}
}

func runConfig(arguments []string, explicitPath string, stdout, stderr io.Writer) int {
	if len(arguments) == 0 {
		writeConfigUsage(stdout)
		return 0
	}

	path := explicitPath
	if path == "" {
		var err error
		path, err = config.DefaultPath()
		if err != nil {
			fmt.Fprintf(stderr, "prox: %v\n", err)
			return 2
		}
	}

	switch arguments[0] {
	case "path":
		if len(arguments) != 1 {
			fmt.Fprintln(stderr, "Usage: prox [--config PATH] config path")
			return 2
		}
		fmt.Fprintln(stdout, path)
		return 0
	case "show":
		if len(arguments) != 1 {
			fmt.Fprintln(stderr, "Usage: prox [--config PATH] config show")
			return 2
		}
		cfg, err := config.Load(explicitPath)
		if err != nil {
			fmt.Fprintf(stderr, "prox: %v\n", err)
			return 2
		}
		data, err := config.MarshalEffective(cfg)
		if err != nil {
			fmt.Fprintf(stderr, "prox: %v\n", err)
			return 1
		}
		fmt.Fprintf(stdout, "Config: %s\n", cfg.Source)
		_, _ = stdout.Write(data)
		return 0
	case "validate":
		if len(arguments) != 1 {
			fmt.Fprintln(stderr, "Usage: prox [--config PATH] config validate")
			return 2
		}
		cfg, err := config.Load(explicitPath)
		if err != nil {
			fmt.Fprintf(stderr, "prox: %v\n", err)
			return 2
		}
		fmt.Fprintf(stdout, "Config is valid: %s\n", cfg.Source)
		return 0
	case "init":
		return runConfigInit(arguments[1:], path, stdout, stderr)
	case "help", "--help", "-h":
		writeConfigUsage(stdout)
		return 0
	default:
		fmt.Fprintf(stderr, "prox: unknown config command %q\n", arguments[0])
		writeConfigUsage(stderr)
		return 2
	}
}

func runConfigInit(arguments []string, path string, stdout, stderr io.Writer) int {
	flags := flag.NewFlagSet("config init", flag.ContinueOnError)
	flags.SetOutput(io.Discard)
	proxyURL := flags.String("proxy", config.DefaultProxyURL, "proxy URL")
	force := flags.Bool("force", false, "replace an existing configuration")
	if err := flags.Parse(arguments); err != nil {
		if errors.Is(err, flag.ErrHelp) {
			fmt.Fprintln(stdout, "Usage: prox [--config PATH] config init [--proxy URL] [--force]")
			return 0
		}
		fmt.Fprintf(stderr, "prox: %v\n", err)
		fmt.Fprintln(stderr, "Usage: prox [--config PATH] config init [--proxy URL] [--force]")
		return 2
	}
	if flags.NArg() != 0 {
		fmt.Fprintln(stderr, "prox: config init does not accept positional arguments")
		return 2
	}
	if err := config.Initialize(path, *proxyURL, *force); err != nil {
		if errors.Is(err, config.ErrConfigExists) {
			fmt.Fprintf(stderr, "prox: %v; use --force to replace it\n", err)
			return 1
		}
		fmt.Fprintf(stderr, "prox: initialize config: %v\n", err)
		return 2
	}
	fmt.Fprintf(stdout, "Created config: %s\n", path)
	return 0
}

func writeConfigUsage(writer io.Writer) {
	fmt.Fprintln(writer, `Usage: prox [--config PATH] config <command>

Commands:
  path                  Print the configuration path
  show                  Print the effective configuration
  init [options]        Create a minimal user configuration
  validate              Validate the effective configuration

Init options:
  --proxy URL           Set the proxy URL
  --force               Replace an existing configuration`)
}

func parseGlobalOptions(arguments []string) (globalOptions, error) {
	options := globalOptions{arguments: append([]string(nil), arguments...)}
	for len(options.arguments) > 0 {
		argument := options.arguments[0]
		switch {
		case argument == "--config":
			if len(options.arguments) < 2 || options.arguments[1] == "" {
				return globalOptions{}, errors.New("--config requires a path")
			}
			options.configPath = options.arguments[1]
			options.arguments = options.arguments[2:]
		case strings.HasPrefix(argument, "--config="):
			options.configPath = strings.TrimPrefix(argument, "--config=")
			if options.configPath == "" {
				return globalOptions{}, errors.New("--config requires a path")
			}
			options.arguments = options.arguments[1:]
		default:
			return options, nil
		}
	}
	return options, nil
}

func runInit(arguments []string, stdout, stderr io.Writer, version string) int {
	if len(arguments) != 1 || arguments[0] != "bash" {
		fmt.Fprintln(stderr, "Usage: prox init bash")
		return 2
	}
	if _, err := io.WriteString(stdout, shell.BashHook(version)); err != nil {
		fmt.Fprintf(stderr, "prox: write Bash integration: %v\n", err)
		return 1
	}
	return 0
}

func runCompletion(arguments []string, stdout, stderr io.Writer) int {
	if len(arguments) != 1 || arguments[0] != "bash" {
		fmt.Fprintln(stderr, "Usage: prox completion bash")
		return 2
	}
	if _, err := io.WriteString(stdout, shell.BashCompletion()); err != nil {
		fmt.Fprintf(stderr, "prox: write Bash completion: %v\n", err)
		return 1
	}
	return 0
}

func shellIntegrationRequired(command string, stderr io.Writer) int {
	fmt.Fprintf(stderr, "prox: `%s` requires Bash integration\n", command)
	fmt.Fprintln(stderr, "Run `eval \"$(prox init bash)\"` in the current Shell first.")
	return 2
}

func runStatus(arguments []string, configPath string, stdout, stderr io.Writer) int {
	if len(arguments) != 0 {
		fmt.Fprintln(stderr, "Usage: prox status")
		return 2
	}
	cfg, err := config.Load(configPath)
	if err != nil {
		fmt.Fprintf(stderr, "prox: %v\n", err)
		return 2
	}

	state := os.Getenv("PROX_INTERNAL_STATE")
	if state != "active" && state != "drifted" {
		state = "inactive"
	}
	fmt.Fprintf(stdout, "State: %s\n", state)
	if state != "inactive" {
		display := os.Getenv("PROX_INTERNAL_PROXY")
		if display == "" {
			display = cfg.RedactedProxyURL()
		}
		fmt.Fprintf(stdout, "Proxy: %s\n", display)
		fmt.Fprintln(stdout, "Health: not checked")
	}
	fmt.Fprintf(stdout, "Config: %s\n", cfg.Source)
	if state == "drifted" {
		fmt.Fprintln(stderr, "Warning: managed proxy variables changed after activation.")
	}
	return 0
}

func runCheck(arguments []string, configPath string, stdout, stderr io.Writer) int {
	flags := flag.NewFlagSet("check", flag.ContinueOnError)
	flags.SetOutput(io.Discard)
	localOnly := flags.Bool("local", false, "only check the proxy endpoint")
	overrideURL := flags.String("url", "", "override the check URL")
	quiet := flags.Bool("quiet", false, "suppress normal output")
	if err := flags.Parse(arguments); err != nil {
		if errors.Is(err, flag.ErrHelp) {
			writeCheckUsage(stdout)
			return 0
		}
		fmt.Fprintf(stderr, "prox: %v\n", err)
		writeCheckUsage(stderr)
		return 2
	}
	if flags.NArg() != 0 {
		fmt.Fprintln(stderr, "prox: check does not accept positional arguments")
		return 2
	}
	if *localOnly && *overrideURL != "" {
		fmt.Fprintln(stderr, "prox: --local and --url cannot be used together")
		return 2
	}

	cfg, err := config.Load(configPath)
	if err != nil {
		fmt.Fprintf(stderr, "prox: %v\n", err)
		return 2
	}
	if !*quiet {
		fmt.Fprintf(stdout, "Proxy: %s\n\n", cfg.RedactedProxyURL())
		writeStep(stdout, health.Step{Name: "Config", OK: true, Detail: "valid"})
	}

	endpoint := health.CheckEndpoint(context.Background(), cfg)
	if !*quiet {
		writeStep(stdout, endpoint)
	}
	if !endpoint.OK {
		if *quiet {
			fmt.Fprintf(stderr, "prox: proxy endpoint is unavailable: %s\n", endpoint.Detail)
		} else {
			fmt.Fprintln(stdout, "\nProxy endpoint is unavailable.")
		}
		return 1
	}
	if *localOnly {
		if !*quiet {
			fmt.Fprintln(stdout, "\nProxy endpoint is reachable.")
		}
		return 0
	}

	target := cfg.CheckURL
	expected := &cfg.ExpectedStatus
	if *overrideURL != "" {
		if err := config.ValidateTargetURL(*overrideURL); err != nil {
			fmt.Fprintf(stderr, "prox: invalid --url: %v\n", err)
			return 2
		}
		target = *overrideURL
		expected = nil
	}
	internet := health.CheckInternet(context.Background(), cfg, target, expected)
	if !*quiet {
		writeStep(stdout, internet)
	}
	if !internet.OK {
		if *quiet {
			fmt.Fprintf(stderr, "prox: test request failed: %s\n", internet.Detail)
		} else {
			fmt.Fprintln(stdout, "\nThe proxy endpoint is reachable, but the test request failed.")
		}
		return 1
	}

	if !*quiet {
		fmt.Fprintln(stdout, "\nProxy is healthy.")
	}
	return 0
}

func writeCheckUsage(writer io.Writer) {
	fmt.Fprintln(writer, "Usage: prox check [--local] [--url URL] [--quiet]")
}

func writeStep(writer io.Writer, step health.Step) {
	result := "FAIL"
	if step.OK {
		result = "PASS"
	}
	duration := ""
	if step.Duration > 0 {
		duration = "  " + formatDuration(step.Duration)
	}
	fmt.Fprintf(writer, "%-4s  %-10s  %s%s\n", result, step.Name, step.Detail, duration)
}

func formatDuration(duration time.Duration) string {
	if duration < time.Millisecond {
		return "<1 ms"
	}
	if duration < time.Second {
		return fmt.Sprintf("%d ms", duration.Milliseconds())
	}
	return fmt.Sprintf("%.1f s", duration.Seconds())
}

func runShellEnvironment(arguments []string, configPath string, stdout, stderr io.Writer) int {
	if len(arguments) != 1 || arguments[0] != "bash" {
		fmt.Fprintln(stderr, "prox: internal shell environment supports bash only")
		return 2
	}
	cfg, err := config.Load(configPath)
	if err != nil {
		fmt.Fprintf(stderr, "prox: %v\n", err)
		return 2
	}
	endpoint := health.CheckEndpoint(context.Background(), cfg)
	if !endpoint.OK {
		fmt.Fprintf(stderr, "Proxy endpoint is unavailable: %s\n", endpoint.Detail)
		fmt.Fprintln(stderr, "Environment unchanged.")
		return 1
	}
	if value, exists := os.LookupEnv("HTTP_PROXY"); exists && value != "" {
		fmt.Fprintln(stderr, "Warning: HTTP_PROXY is already set and is not managed by prox V0.1.")
	}
	if _, err := io.WriteString(stdout, shell.BashEnvironmentScript(cfg)); err != nil {
		fmt.Fprintf(stderr, "prox: write Shell environment: %v\n", err)
		return 1
	}
	return 0
}

func runCommand(
	arguments []string,
	configPath string,
	stdin io.Reader,
	stdout, stderr io.Writer,
) int {
	if len(arguments) < 2 || arguments[0] != "--" {
		fmt.Fprintln(stderr, "Usage: prox run -- <command> [arguments...]")
		return 2
	}
	command := arguments[1:]
	cfg, err := config.Load(configPath)
	if err != nil {
		fmt.Fprintf(stderr, "prox: %v\n", err)
		return 125
	}
	endpoint := health.CheckEndpoint(context.Background(), cfg)
	if !endpoint.OK {
		fmt.Fprintf(stderr, "prox: proxy endpoint is unavailable: %s\n", endpoint.Detail)
		return 125
	}

	// stdin/stdout/stderr remain attached because Replace uses execve on Linux.
	_ = stdin
	_ = stdout
	environment := shell.MergeEnvironment(os.Environ(), cfg)
	code, err := runner.Replace(command, environment)
	if err != nil {
		fmt.Fprintf(stderr, "prox: execute %q: %v\n", command[0], err)
		return code
	}
	return code
}
