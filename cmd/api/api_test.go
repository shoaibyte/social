package main

import (
	"net"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"
)

// helper to build a minimal application instance for testing
func newTestApp() *application {
	return &application{
		config: config{
			addr: "127.0.0.1:0",
			env:  "test",
			db: dbConfig{
				address:      "",
				maxOpenConns: 0,
				maxIdleConns: 0,
				maxIdleTime:  "0s",
			},
		},
	}
}

// Test that the router mounts /v1/health and the handler is reachable.
// We assert non-404 to validate the route wiring, and prefer 200 if the handler returns it.
func TestMount_HealthRoute_WiredAndResponds(t *testing.T) {
	app := newTestApp()
	handler := app.mount()

	req := httptest.NewRequest(http.MethodGet, "/v1/health", nil)
	rec := httptest.NewRecorder()

	handler.ServeHTTP(rec, req)

	if rec.Code == http.StatusNotFound {
		t.Fatalf("expected /v1/health route to be mounted, got 404 Not Found")
	}

	// Prefer 200 OK for a health endpoint; if different, still surface a clear diagnostic.
	if rec.Code != http.StatusOK {
		t.Logf("warning: /v1/health returned status %d; expected 200 OK. Body: %s", rec.Code, strings.TrimSpace(rec.Body.String()))
	}
}

// The chi/middleware.RequestID should add an X-Request-ID header on responses.
// We verify the header is present and non-empty for a successful request.
func TestMount_RequestIDMiddleware_SetsHeader(t *testing.T) {
	app := newTestApp()
	handler := app.mount()

	req := httptest.NewRequest(http.MethodGet, "/v1/health", nil)
	rec := httptest.NewRecorder()

	handler.ServeHTTP(rec, req)

	got := rec.Header().Get("X-Request-ID")
	if got == "" {
		t.Fatalf("expected X-Request-ID header to be set by middleware; got empty")
	}
}

// Validate run() returns promptly with an error when the address is already in use.
// This avoids starting a long-lived server while still exercising the code path.
func TestRun_AddressInUse_ReturnsError(t *testing.T) {
	// Bind to an ephemeral port and keep it open to force EADDRINUSE for the server under test.
	l, err := net.Listen("tcp", "127.0.0.1:0")
	if err != nil {
		t.Fatalf("failed to acquire a test listener: %v", err)
	}
	defer l.Close()

	addr := l.Addr().String()

	app := newTestApp()
	app.config.addr = addr

	// Use a no-op handler; we only care that ListenAndServe fails immediately.
	mux := http.NewServeMux()

	errCh := make(chan error, 1)
	go func() {
		errCh <- app.run(mux)
	}()

	select {
	case runErr := <-errCh:
		if runErr == nil {
			t.Fatalf("expected run() to return an error when address is in use; got nil")
		}
		// Be lenient on exact message, which can vary by OS.
		msg := runErr.Error()
		if !strings.Contains(strings.ToLower(msg), "address already in use") && !strings.Contains(strings.ToLower(msg), "in use") {
			t.Logf("run() returned error (as expected), but message didn't include 'address already in use': %v", runErr)
		}
	case <-time.After(2 * time.Second):
		t.Fatal("run() did not return promptly when address is in use (possible hang)")
	}
}
