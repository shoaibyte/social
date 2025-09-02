// Tests for JSON helpers.
//
// Testing framework: Go standard library `testing` with `net/http/httptest`.
// No third-party test dependencies are used.
package main

import (
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
)

type sample struct {
	Name string `json:"name"`
	Age  int    `json:"age"`
}

func TestWriteJSON_SetsHeadersStatusAndBody(t *testing.T) {
	t.Parallel()

	rr := httptest.NewRecorder()
	payload := map[string]interface{}{"message": "ok", "count": 1}

	if err := writeJSON(rr, http.StatusCreated, payload); err != nil {
		t.Fatalf("writeJSON returned error: %v", err)
	}

	if got := rr.Header().Get("Content-Type"); got != "application/json; charset=utf-8" {
		t.Errorf("Content-Type = %q, want %q", got, "application/json; charset=utf-8")
	}
	if rr.Code != http.StatusCreated {
		t.Errorf("status = %d, want %d", rr.Code, http.StatusCreated)
	}

	var got map[string]interface{}
	if err := json.NewDecoder(rr.Body).Decode(&got); err != nil {
		t.Fatalf("decoding response body failed: %v", err)
	}
	if got["message"] != "ok" || intFromInterface(got["count"]) != 1 {
		t.Errorf("unexpected body: %#v", got)
	}
}

func TestWriteJSON_ReturnsErrorOnUnsupportedType(t *testing.T) {
	t.Parallel()

	rr := httptest.NewRecorder()
	ch := make(chan int) // unsupported by JSON encoder

	err := writeJSON(rr, http.StatusOK, ch)
	if err == nil {
		t.Fatalf("expected error for unsupported type, got nil")
	}
	// Headers and status are set before encoding.
	if rr.Header().Get("Content-Type") != "application/json; charset=utf-8" {
		t.Errorf("Content-Type header not set")
	}
	if rr.Code != http.StatusOK {
		t.Errorf("status = %d, want %d", rr.Code, http.StatusOK)
	}
}

func TestWriteJSONError_WrapsMessageAndStatus(t *testing.T) {
	t.Parallel()

	rr := httptest.NewRecorder()
	const msg = "missing name"

	if err := writeJSONError(rr, http.StatusBadRequest, msg); err != nil {
		t.Fatalf("writeJSONError returned error: %v", err)
	}

	if rr.Code != http.StatusBadRequest {
		t.Errorf("status = %d, want %d", rr.Code, http.StatusBadRequest)
	}
	if got := rr.Header().Get("Content-Type"); got != "application/json; charset=utf-8" {
		t.Errorf("Content-Type = %q, want %q", got, "application/json; charset=utf-8")
	}

	var env struct{ Error string `json:"error"` }
	if err := json.NewDecoder(rr.Body).Decode(&env); err != nil {
		t.Fatalf("decoding response body failed: %v", err)
	}
	if env.Error != msg {
		t.Errorf("error message = %q, want %q", env.Error, msg)
	}
}

func TestReadJSON_Success(t *testing.T) {
	t.Parallel()

	payload := `{"name":"Ana","age":30}`
	req := httptest.NewRequest(http.MethodPost, "/", strings.NewReader(payload))
	rr := httptest.NewRecorder()

	var s sample
	if err := readJSON(rr, req, &s); err != nil {
		t.Fatalf("readJSON returned error: %v", err)
	}
	if s.Name != "Ana" || s.Age != 30 {
		t.Errorf("decoded struct = %+v, want {Name:Ana Age:30}", s)
	}
}

func TestReadJSON_UnknownField_ReturnsError(t *testing.T) {
	t.Parallel()

	payload := `{"name":"Ana","age":30,"extra":true}`
	req := httptest.NewRequest(http.MethodPost, "/", strings.NewReader(payload))
	rr := httptest.NewRecorder()

	var s sample
	err := readJSON(rr, req, &s)
	if err == nil {
		t.Fatal("expected error for unknown field, got nil")
	}
	if !strings.Contains(err.Error(), "unknown field") {
		t.Errorf("error = %v, expected to contain %q", err, "unknown field")
	}
}

func TestReadJSON_InvalidJSON_ReturnsError(t *testing.T) {
	t.Parallel()

	payload := `{"name":`
	req := httptest.NewRequest(http.MethodPost, "/", strings.NewReader(payload))
	rr := httptest.NewRecorder()

	var s sample
	if err := readJSON(rr, req, &s); err == nil {
		t.Fatal("expected error for invalid JSON, got nil")
	}
}

func TestReadJSON_EmptyBody_ReturnsEOF(t *testing.T) {
	t.Parallel()

	req := httptest.NewRequest(http.MethodPost, "/", bytes.NewBuffer(nil))
	rr := httptest.NewRecorder()

	var s sample
	err := readJSON(rr, req, &s)
	if err == nil {
		t.Fatal("expected io.EOF for empty body, got nil")
	}
	if !errors.Is(err, io.EOF) {
		t.Errorf("error = %v, want io.EOF", err)
	}
}

func TestReadJSON_TooLargeBody_ReturnsError(t *testing.T) {
	t.Parallel()

	const maxBytes = 1_048_578 // keep in sync with implementation
	veryLong := strings.Repeat("x", maxBytes+10) // exceed limit
	payload := fmt.Sprintf(`{"name":"%s"}`, veryLong)

	req := httptest.NewRequest(http.MethodPost, "/", strings.NewReader(payload))
	rr := httptest.NewRecorder()

	var s sample
	err := readJSON(rr, req, &s)
	if err == nil {
		t.Fatal("expected error due to body size limit, got nil")
	}
	// Message text can vary by Go version; assert broadly.
	if !strings.Contains(err.Error(), "too large") {
		t.Errorf("error %q does not indicate size limit exceeded", err)
	}
}

// Helper to coerce decoded JSON number types.
func intFromInterface(v interface{}) int {
	switch n := v.(type) {
	case float64:
		return int(n)
	case int:
		return n
	default:
		return 0
	}
}