package main

import (
	"encoding/json"
	"fmt"
	"io"
	"log/slog"
	"net"
	"net/http"
	"strconv"
	"strings"
	"syscall"
	"testing"
	"time"
)

func discardLogger() *slog.Logger {
	return slog.New(slog.NewTextHandler(io.Discard, nil))
}

// freePort asks the kernel for an unused port and releases it, so the server under test binds
// somewhere that is not already claimed on the developer's machine.
func freePort(t *testing.T) int {
	t.Helper()

	listener, err := net.Listen("tcp", ":0")
	if err != nil {
		t.Fatalf("could not reserve a port: %v", err)
	}
	defer listener.Close()

	return listener.Addr().(*net.TCPAddr).Port
}

func waitForServer(t *testing.T, port int) {
	t.Helper()

	deadline := time.Now().Add(5 * time.Second)
	for time.Now().Before(deadline) {
		conn, err := net.DialTimeout("tcp", net.JoinHostPort("127.0.0.1", strconv.Itoa(port)), 100*time.Millisecond)
		if err == nil {
			conn.Close()
			return
		}
		time.Sleep(10 * time.Millisecond)
	}

	t.Fatalf("server did not start listening on port %d", port)
}

func TestRun_ServesRequestsAndShutsDownOnSignal(t *testing.T) {
	port := freePort(t)
	t.Setenv("PORT", strconv.Itoa(port))

	stopped := make(chan error, 1)
	go func() { stopped <- run(discardLogger()) }()

	waitForServer(t, port)

	client := &http.Client{Transport: &http.Transport{DisableKeepAlives: true}}
	endpoint := fmt.Sprintf("http://127.0.0.1:%d/api/v1/calculate", port)

	resp, err := client.Post(endpoint, "application/json", strings.NewReader(`{"operation":"add","a":2,"b":3}`))
	if err != nil {
		t.Fatalf("request failed: %v", err)
	}

	var body struct {
		Result float64 `json:"result"`
	}
	decodeErr := json.NewDecoder(resp.Body).Decode(&body)
	resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		t.Fatalf("status = %d, want 200", resp.StatusCode)
	}
	if decodeErr != nil {
		t.Fatalf("response is not JSON: %v", decodeErr)
	}
	if body.Result != 5 {
		t.Errorf("result = %v, want 5", body.Result)
	}

	// signal.NotifyContext intercepts this, so the test process survives it.
	if err := syscall.Kill(syscall.Getpid(), syscall.SIGTERM); err != nil {
		t.Fatalf("could not signal the process: %v", err)
	}

	select {
	case err := <-stopped:
		if err != nil {
			t.Errorf("run() returned %v, want a clean shutdown", err)
		}
	case <-time.After(5 * time.Second):
		t.Fatal("run() did not return after SIGTERM")
	}
}

func TestRun_ReturnsConfigurationErrors(t *testing.T) {
	t.Setenv("PORT", "not-a-port")

	if err := run(discardLogger()); err == nil {
		t.Error("run() returned no error for an invalid PORT")
	}
}

func TestRun_ReturnsListenErrors(t *testing.T) {
	occupied, err := net.Listen("tcp", ":0")
	if err != nil {
		t.Fatalf("could not occupy a port: %v", err)
	}
	defer occupied.Close()

	t.Setenv("PORT", strconv.Itoa(occupied.Addr().(*net.TCPAddr).Port))

	if err := run(discardLogger()); err == nil {
		t.Error("run() returned no error when the port was already in use")
	}
}
