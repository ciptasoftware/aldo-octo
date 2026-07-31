package testutil

import (
	"net/http"
	"net/http/httptest"
	"testing"
)

// AssertStatus checks if the HTTP status code matches the expected one
func AssertStatus(t *testing.T, expected, got int) {
	t.Helper()
	if expected != got {
		t.Errorf("expected status %d, got %d", expected, got)
	}
}

// PerformRequest is a helper to run a request against a handler
func PerformRequest(h http.Handler, method, path string) *httptest.ResponseRecorder {
	req := httptest.NewRequest(method, path, nil)
	w := httptest.NewRecorder()
	h.ServeHTTP(w, req)
	return w
}
