// Package testhelper provides helper functions and constants for testing
package testhelper

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
)

// HTTPTestServer creates a test HTTP server with the given handler
type HTTPTestServer struct {
	*httptest.Server
}

// NewHTTPTestServer creates a new test HTTP server
func NewHTTPTestServer(handler http.HandlerFunc) *HTTPTestServer {
	return &HTTPTestServer{
		Server: httptest.NewServer(handler),
	}
}

// JSONResponse writes a JSON response with proper headers
func JSONResponse(t *testing.T, w http.ResponseWriter, statusCode int, response interface{}) {
	t.Helper()

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(statusCode)

	var data []byte
	var err error

	switch v := response.(type) {
	case string:
		data = []byte(v)
	case []byte:
		data = v
	default:
		data, err = json.Marshal(response)
		if err != nil {
			t.Fatalf("failed to marshal response: %v", err)
		}
	}

	if _, err := w.Write(data); err != nil {
		t.Logf("failed to write response: %v", err)
	}
}

// ErrorResponse writes an error JSON response
func ErrorResponse(t *testing.T, w http.ResponseWriter, statusCode int, errorMessage string) {
	t.Helper()
	JSONResponse(t, w, statusCode, map[string]string{"error": errorMessage})
}
