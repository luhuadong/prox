package health

import (
	"context"
	"net"
	"net/http"
	"net/http/httptest"
	"sync/atomic"
	"testing"
	"time"

	"prox/internal/config"
)

func TestCheckEndpoint(t *testing.T) {
	listener, err := net.Listen("tcp", "127.0.0.1:0")
	if err != nil {
		t.Fatal(err)
	}
	defer listener.Close()

	cfg := config.Defaults()
	cfg.ProxyURL = "http://" + listener.Addr().String()
	step := CheckEndpoint(context.Background(), cfg)
	if !step.OK {
		t.Fatalf("CheckEndpoint() = %#v", step)
	}
}

func TestCheckEndpointUnavailable(t *testing.T) {
	listener, err := net.Listen("tcp", "127.0.0.1:0")
	if err != nil {
		t.Fatal(err)
	}
	address := listener.Addr().String()
	_ = listener.Close()

	cfg := config.Defaults()
	cfg.ProxyURL = "http://" + address
	cfg.ConnectTimeout = 100 * time.Millisecond
	step := CheckEndpoint(context.Background(), cfg)
	if step.OK || step.Failure != FailureEndpointUnreachable {
		t.Fatalf("CheckEndpoint() = %#v", step)
	}
}

func TestCheckInternetUsesExplicitProxyAndIgnoresNoProxyEnvironment(t *testing.T) {
	var requests atomic.Int32
	proxy := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		requests.Add(1)
		w.WriteHeader(http.StatusNoContent)
	}))
	defer proxy.Close()

	t.Setenv("NO_PROXY", "*")
	t.Setenv("no_proxy", "*")

	cfg := config.Defaults()
	cfg.ProxyURL = proxy.URL
	target := "http://target.invalid/generate_204"
	expected := http.StatusNoContent
	step := CheckInternet(context.Background(), cfg, target, &expected)
	if !step.OK {
		t.Fatalf("CheckInternet() = %#v", step)
	}
	if requests.Load() != 1 {
		t.Fatalf("proxy received %d requests, want 1", requests.Load())
	}
}

func TestCheckInternetRejectsUnexpectedStatus(t *testing.T) {
	proxy := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
	}))
	defer proxy.Close()

	cfg := config.Defaults()
	cfg.ProxyURL = proxy.URL
	expected := http.StatusNoContent
	step := CheckInternet(context.Background(), cfg, "http://target.invalid/", &expected)
	if step.OK || step.Failure != FailureTarget {
		t.Fatalf("CheckInternet() = %#v", step)
	}
}
