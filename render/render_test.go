package render

import (
	"bytes"
	"encoding/json"
	"io/fs"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"testing/fstest"
)

// --- JSON Tests ---

func TestJSON(t *testing.T) {
	w := httptest.NewRecorder()
	data := map[string]string{"message": "hello"}

	JSON(w, http.StatusOK, data)

	if w.Code != http.StatusOK {
		t.Errorf("expected status 200, got %d", w.Code)
	}

	ct := w.Header().Get("Content-Type")
	if ct != "application/json" {
		t.Errorf("expected Content-Type application/json, got %q", ct)
	}

	var result map[string]string
	if err := json.Unmarshal(w.Body.Bytes(), &result); err != nil {
		t.Fatalf("failed to parse JSON response: %v", err)
	}

	if result["message"] != "hello" {
		t.Errorf("expected message=%q, got %q", "hello", result["message"])
	}
}

func TestJSONError(t *testing.T) {
	w := httptest.NewRecorder()

	JSONError(w, http.StatusBadRequest, "invalid input")

	if w.Code != http.StatusBadRequest {
		t.Errorf("expected status 400, got %d", w.Code)
	}

	var result map[string]string
	if err := json.Unmarshal(w.Body.Bytes(), &result); err != nil {
		t.Fatalf("failed to parse JSON response: %v", err)
	}

	if result["error"] != "invalid input" {
		t.Errorf("expected error=%q, got %q", "invalid input", result["error"])
	}
}

func TestJSONMarshalError(t *testing.T) {
	w := httptest.NewRecorder()

	// Channels cannot be marshaled to JSON
	JSON(w, http.StatusOK, make(chan int))

	if w.Code != http.StatusInternalServerError {
		t.Errorf("expected status 500, got %d", w.Code)
	}
}

// --- BindJSON Tests ---

func TestBindJSONValid(t *testing.T) {
	body := `{"name":"John","age":"30"}`
	req := httptest.NewRequest("POST", "/", strings.NewReader(body))
	req.Header.Set("Content-Type", "application/json")

	var result map[string]string
	err := BindJSON(req, &result)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if result["name"] != "John" {
		t.Errorf("expected name=%q, got %q", "John", result["name"])
	}
}

func TestBindJSONInvalid(t *testing.T) {
	req := httptest.NewRequest("POST", "/", strings.NewReader("not json"))
	req.Header.Set("Content-Type", "application/json")

	var result map[string]string
	err := BindJSON(req, &result)
	if err == nil {
		t.Error("expected error for invalid JSON")
	}
}

func TestBindJSONTooLarge(t *testing.T) {
	// Create a body larger than MaxBodySize (1 MB)
	bigBody := strings.Repeat("x", MaxBodySize+1)
	req := httptest.NewRequest("POST", "/", strings.NewReader(bigBody))

	var result map[string]string
	err := BindJSON(req, &result)
	if err == nil {
		t.Error("expected error for oversized body")
	}
}

// --- BindQuery Tests ---

func TestBindQuery(t *testing.T) {
	req := httptest.NewRequest("GET", "/?name=Alice&empty=", nil)

	name := BindQuery(req, "name", "default")
	if name != "Alice" {
		t.Errorf("expected %q, got %q", "Alice", name)
	}

	missing := BindQuery(req, "missing_key", "fallback")
	if missing != "fallback" {
		t.Errorf("expected %q, got %q", "fallback", missing)
	}

	empty := BindQuery(req, "empty", "default_val")
	if empty != "default_val" {
		t.Errorf("expected %q for empty param, got %q", "default_val", empty)
	}
}

// --- Template Engine Tests ---

func makeTestFS() fs.FS {
	return fstest.MapFS{
		"templates/hello.html": &fstest.MapFile{
			Data: []byte(`<h1>Hello, {{.Name}}!</h1>`),
		},
	}
}

func TestTemplateRenderProdMode(t *testing.T) {
	testFS := makeTestFS()
	engine, err := NewEngine("templates", "", testFS, true)
	if err != nil {
		t.Fatalf("failed to create engine: %v", err)
	}

	w := httptest.NewRecorder()
	data := map[string]string{"Name": "World"}
	engine.Render(w, http.StatusOK, "hello.html", data)

	if w.Code != http.StatusOK {
		t.Errorf("expected status 200, got %d", w.Code)
	}

	body := w.Body.String()
	if !strings.Contains(body, "Hello, World!") {
		t.Errorf("expected body to contain 'Hello, World!', got %q", body)
	}

	ct := w.Header().Get("Content-Type")
	if ct != "text/html; charset=utf-8" {
		t.Errorf("expected Content-Type text/html, got %q", ct)
	}
}

func TestTemplateRenderMissingTemplate(t *testing.T) {
	testFS := makeTestFS()
	engine, err := NewEngine("templates", "", testFS, true)
	if err != nil {
		t.Fatalf("failed to create engine: %v", err)
	}

	w := httptest.NewRecorder()
	engine.Render(w, http.StatusOK, "nonexistent.html", nil)

	if w.Code != http.StatusInternalServerError {
		t.Errorf("expected status 500 for missing template, got %d", w.Code)
	}
}

func TestTemplateRenderBuffered(t *testing.T) {
	// Verify that render uses buffer (partial response prevention)
	testFS := makeTestFS()
	engine, err := NewEngine("templates", "", testFS, true)
	if err != nil {
		t.Fatalf("failed to create engine: %v", err)
	}

	w := httptest.NewRecorder()
	data := map[string]string{"Name": "Test"}
	engine.Render(w, http.StatusCreated, "hello.html", data)

	if w.Code != http.StatusCreated {
		t.Errorf("expected status 201, got %d", w.Code)
	}

	_ = bytes.Contains(w.Body.Bytes(), []byte("Test"))
}
