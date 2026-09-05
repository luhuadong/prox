package health

import (
	"context"
	"errors"
	"fmt"
	"net"
	"net/http"
	"net/url"
	"strings"
	"time"

	"prox/internal/config"
)

// FailureKind identifies the stage that failed without exposing implementation details.
type FailureKind string

const (
	FailureNone                FailureKind = ""
	FailureEndpointUnreachable FailureKind = "endpoint_unreachable"
	FailureProxyRejected       FailureKind = "proxy_rejected"
	FailureTarget              FailureKind = "target_failed"
)

// Step is one human-readable health-check result.
type Step struct {
	Name     string
	OK       bool
	Detail   string
	Duration time.Duration
	Failure  FailureKind
	Err      error
}

// CheckEndpoint verifies that the configured proxy host and port accept a TCP connection.
func CheckEndpoint(ctx context.Context, cfg config.Config) Step {
	started := time.Now()
	endpoint, err := cfg.ProxyEndpoint()
	if err != nil {
		return Step{
			Name:    "Endpoint",
			Failure: FailureEndpointUnreachable,
			Err:     err,
			Detail:  err.Error(),
		}
	}

	dialer := net.Dialer{Timeout: cfg.ConnectTimeout}
	conn, err := dialer.DialContext(ctx, "tcp", endpoint)
	duration := time.Since(started)
	if err != nil {
		return Step{
			Name:     "Endpoint",
			Detail:   fmt.Sprintf("%s: %s", endpoint, conciseError(err)),
			Duration: duration,
			Failure:  FailureEndpointUnreachable,
			Err:      err,
		}
	}
	_ = conn.Close()

	return Step{
		Name:     "Endpoint",
		OK:       true,
		Detail:   endpoint + " reachable",
		Duration: duration,
	}
}

// CheckInternet performs a request through the configured proxy. When expectedStatus
// is nil, any HTTP status from 200 through 399 is accepted.
func CheckInternet(
	ctx context.Context,
	cfg config.Config,
	target string,
	expectedStatus *int,
) Step {
	started := time.Now()
	if err := config.ValidateTargetURL(target); err != nil {
		return Step{
			Name:    "Internet",
			Detail:  "invalid check URL: " + err.Error(),
			Failure: FailureTarget,
			Err:     err,
		}
	}

	proxyURL, err := config.ParseProxyURL(cfg.ProxyURL)
	if err != nil {
		return Step{
			Name:    "Internet",
			Detail:  err.Error(),
			Failure: FailureProxyRejected,
			Err:     err,
		}
	}

	dialer := &net.Dialer{
		Timeout:   cfg.ConnectTimeout,
		KeepAlive: 30 * time.Second,
	}
	transport := &http.Transport{
		Proxy:                 http.ProxyURL(proxyURL),
		DialContext:           dialer.DialContext,
		ForceAttemptHTTP2:     true,
		DisableKeepAlives:     true,
		TLSHandshakeTimeout:   cfg.ConnectTimeout,
		ResponseHeaderTimeout: cfg.RequestTimeout,
	}
	defer transport.CloseIdleConnections()

	client := &http.Client{
		Transport: transport,
		Timeout:   cfg.RequestTimeout,
	}
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, target, nil)
	if err != nil {
		return Step{
			Name:    "Internet",
			Detail:  err.Error(),
			Failure: FailureTarget,
			Err:     err,
		}
	}
	req.Header.Set("User-Agent", "prox-health-check")

	response, err := client.Do(req)
	duration := time.Since(started)
	if err != nil {
		failure := classifyRequestError(err)
		return Step{
			Name:     "Internet",
			Detail:   conciseError(err),
			Duration: duration,
			Failure:  failure,
			Err:      err,
		}
	}
	defer response.Body.Close()

	host := response.Request.URL.Hostname()
	if host == "" {
		host = target
	}
	if response.StatusCode == http.StatusProxyAuthRequired {
		return Step{
			Name:     "Internet",
			Detail:   fmt.Sprintf("proxy authentication required (HTTP %d)", response.StatusCode),
			Duration: duration,
			Failure:  FailureProxyRejected,
		}
	}

	statusOK := response.StatusCode >= 200 && response.StatusCode <= 399
	if expectedStatus != nil {
		statusOK = response.StatusCode == *expectedStatus
	}
	if !statusOK {
		expectation := "HTTP 200-399"
		if expectedStatus != nil {
			expectation = fmt.Sprintf("HTTP %d", *expectedStatus)
		}
		return Step{
			Name:     "Internet",
			Detail:   fmt.Sprintf("%s returned HTTP %d; expected %s", host, response.StatusCode, expectation),
			Duration: duration,
			Failure:  FailureTarget,
		}
	}

	return Step{
		Name:     "Internet",
		OK:       true,
		Detail:   fmt.Sprintf("%s returned HTTP %d", host, response.StatusCode),
		Duration: duration,
	}
}

func classifyRequestError(err error) FailureKind {
	message := strings.ToLower(err.Error())
	if strings.Contains(message, "proxyconnect") ||
		strings.Contains(message, "proxy authentication") ||
		strings.Contains(message, "status code 407") {
		return FailureProxyRejected
	}
	return FailureTarget
}

func conciseError(err error) string {
	var urlError *url.Error
	if errors.As(err, &urlError) && urlError.Err != nil {
		err = urlError.Err
	}
	var networkError net.Error
	if errors.As(err, &networkError) && networkError.Timeout() {
		return "request timed out"
	}
	return err.Error()
}
