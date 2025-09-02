
// --- Additional tests for healthCheckHandler ---
// Note: Using Go's standard testing package and net/http/httptest.

type forcedErr string
func (e forcedErr) Error() string { return string(e) }
var forcedWriteErr = forcedErr("forced write error")

// alwaysErrorWriter simulates a ResponseWriter that always fails on Write,
// while recording headers and all WriteHeader calls.
type alwaysErrorWriter struct {
	header      http.Header
	statusCalls []int
}
func (w *alwaysErrorWriter) Header() http.Header {
	if w.header == nil {
		w.header = make(http.Header)
	}
	return w.header
}
func (w *alwaysErrorWriter) WriteHeader(code int) { w.statusCalls = append(w.statusCalls, code) }
func (w *alwaysErrorWriter) Write(_ []byte) (int, error) { return 0, forcedWriteErr }

// firstFailWriter fails on the first Write, then succeeds for subsequent writes.
// It records the body bytes and all status codes attempted.
type firstFailWriter struct {
	header      http.Header
	statusCalls []int
	writeCalls  int
	body        []byte
}
func (w *firstFailWriter) Header() http.Header {
	if w.header == nil {
		w.header = make(http.Header)
	}
	return w.header
}
func (w *firstFailWriter) WriteHeader(code int) { w.statusCalls = append(w.statusCalls, code) }
func (w *firstFailWriter) Write(p []byte) (int, error) {
	if w.writeCalls == 0 {
		w.writeCalls++
		return 0, forcedWriteErr
	}
	w.writeCalls++
	w.body = append(w.body, p...)
	return len(p), nil
}

// Ensures strict Content-Type and 200 status on the happy path.
func TestHealthCheckHandler_ContentTypeAndStatus_Strict(t *testing.T) {
	app := newTestApplicationWithEnv(t, "test")
	req := httptest.NewRequest(http.MethodGet, "/v1/health", nil)
	rr := httptest.NewRecorder()

	if err := app.healthCheckHandler(rr, req); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	res := rr.Result()
	defer res.Body.Close()

	if res.StatusCode != http.StatusOK {
		t.Fatalf("expected status %d, got %d", http.StatusOK, res.StatusCode)
	}
	if got := res.Header.Get("Content-Type"); got != "application/json; charset=utf-8" {
		t.Errorf("expected Content-Type %q, got %q", "application/json; charset=utf-8", got)
	}
}

// Validates response schema, key set, and that time is RFC3339 and recent.
func TestHealthCheckHandler_ResponseSchema_And_Time(t *testing.T) {
	app := newTestApplicationWithEnv(t, "test")
	req := httptest.NewRequest(http.MethodGet, "/v1/health", nil)
	rr := httptest.NewRecorder()

	if err := app.healthCheckHandler(rr, req); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	var payload map[string]string
	if err := json.NewDecoder(rr.Body).Decode(&payload); err != nil {
		t.Fatalf("decode failed: %v", err)
	}

	if len(payload) != 4 {
		t.Errorf("expected exactly 4 keys, got %d: %#v", len(payload), payload)
	}
	if payload["status"] != "ok" {
		t.Errorf("status: expected %q, got %q", "ok", payload["status"])
	}
	if payload["env"] != app.config.env {
		t.Errorf("env: expected %q, got %q", app.config.env, payload["env"])
	}
	if payload["version"] != version {
		t.Errorf("version: expected %q, got %q", version, payload["version"])
	}
	ts := payload["time"]
	parsed, err := time.Parse(time.RFC3339, ts)
	if err != nil {
		t.Fatalf("time not RFC3339: %q: %v", ts, err)
	}
	// ensure near now (5s window)
	if d := time.Since(parsed); d < -5*time.Second || d > 5*time.Second {
		t.Errorf("time not within 5s of now; now=%s, payload=%s (delta=%v)", time.Now().UTC().Format(time.RFC3339), ts, d)
	}
}

// When writes always fail, handler should propagate a non-nil error and attempt a 500 via writeJSONError.
func TestHealthCheckHandler_WriteAlwaysFails_ReturnsErrorAndAttempts500(t *testing.T) {
	app := newTestApplicationWithEnv(t, "test")
	req := httptest.NewRequest(http.MethodGet, "/v1/health", nil)
	w := &alwaysErrorWriter{}

	err := app.healthCheckHandler(w, req)
	if err == nil {
		t.Fatalf("expected non-nil error when writer fails")
	}

	// Content-Type should have been set before the write attempt.
	if got := w.Header().Get("Content-Type"); got != "application/json; charset=utf-8" {
		t.Errorf("expected Content-Type %q to be set, got %q", "application/json; charset=utf-8", got)
	}

	// Assert that a 500 attempt occurred (writeJSONError path).
	found500 := false
	for _, sc := range w.statusCalls {
		if sc == http.StatusInternalServerError {
			found500 = true
			break
		}
	}
	if !found500 {
		t.Errorf("expected an attempted 500 status; status calls: %v", w.statusCalls)
	}
}

// If the first write fails but subsequent writes succeed, verify error envelope is produced.
func TestHealthCheckHandler_FirstWriteFails_ThenErrorEnvelope(t *testing.T) {
	app := newTestApplicationWithEnv(t, "test")
	req := httptest.NewRequest(http.MethodGet, "/v1/health", nil)
	w := &firstFailWriter{}

	err := app.healthCheckHandler(w, req)
	if err != nil {
		t.Fatalf("expected nil error after error envelope write succeeds; got %v", err)
	}

	// A 500 should have been attempted at some point.
	found500 := false
	for _, sc := range w.statusCalls {
		if sc == http.StatusInternalServerError {
			found500 = true
			break
		}
	}
	if !found500 {
		t.Errorf("expected an attempted 500 status; status calls: %v", w.statusCalls)
	}

	// Decode error envelope body
	var env struct {
		Error string `json:"error"`
	}
	if derr := json.Unmarshal(w.body, &env); derr != nil {
		t.Fatalf("failed to decode error envelope: %v; body=%q", derr, string(w.body))
	}
	if env.Error == "" {
		t.Errorf("expected non-empty error message in envelope; body=%q", string(w.body))
	}

	// Content-Type should remain JSON
	if got := w.Header().Get("Content-Type"); got != "application/json; charset=utf-8" {
		t.Errorf("expected Content-Type %q, got %q", "application/json; charset=utf-8", got)
	}
}