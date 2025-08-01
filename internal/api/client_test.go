package api

import (
	"context"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"
)

func TestNewClient(t *testing.T) {
	tests := []struct {
		name      string
		token     string
		wantError bool
	}{
		{
			name:      "valid token",
			token:     "test-token",
			wantError: false,
		},
		{
			name:      "empty token",
			token:     "",
			wantError: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			client, err := NewClient(tt.token)
			if tt.wantError {
				if err == nil {
					t.Error("expected error but got nil")
				}
				return
			}
			if err != nil {
				t.Errorf("unexpected error: %v", err)
				return
			}
			if client == nil {
				t.Error("expected client but got nil")
				return
			}
			if client.token != tt.token {
				t.Errorf("expected token %s, got %s", tt.token, client.token)
			}
		})
	}
}

func TestClient_SetTimeout(t *testing.T) {
	client, err := NewClient("test-token")
	if err != nil {
		t.Fatalf("failed to create client: %v", err)
	}

	timeout := 10 * time.Second
	client.SetTimeout(timeout)

	if client.httpClient.Timeout != timeout {
		t.Errorf("expected timeout %v, got %v", timeout, client.httpClient.Timeout)
	}
}

func TestClient_SetBaseURL(t *testing.T) {
	client, err := NewClient("test-token")
	if err != nil {
		t.Fatalf("failed to create client: %v", err)
	}

	baseURL := "https://example.com/api"
	err = client.SetBaseURL(baseURL)
	if err != nil {
		t.Errorf("unexpected error: %v", err)
	}

	if client.baseURL.String() != baseURL {
		t.Errorf("expected base URL %s, got %s", baseURL, client.baseURL.String())
	}

	// Test invalid URL (Note: "invalid-url" is actually a valid relative URL in Go)
	err = client.SetBaseURL("ht!tp://invalid")
	if err == nil {
		t.Error("expected error for invalid URL")
	}
}

func TestClient_newRequest(t *testing.T) {
	client, err := NewClient("test-token")
	if err != nil {
		t.Fatalf("failed to create client: %v", err)
	}

	ctx := context.Background()
	req, err := client.newRequest(ctx, http.MethodGet, "/test", nil)
	if err != nil {
		t.Errorf("unexpected error: %v", err)
	}

	if req.Method != http.MethodGet {
		t.Errorf("expected method %s, got %s", http.MethodGet, req.Method)
	}

	expectedURL := DefaultBaseURL + "/test"
	if req.URL.String() != expectedURL {
		t.Errorf("expected URL %s, got %s", expectedURL, req.URL.String())
	}

	auth := req.Header.Get("Authorization")
	expectedAuth := "Bearer test-token"
	if auth != expectedAuth {
		t.Errorf("expected Authorization header %s, got %s", expectedAuth, auth)
	}

	userAgent := req.Header.Get("User-Agent")
	if userAgent != UserAgent {
		t.Errorf("expected User-Agent %s, got %s", UserAgent, userAgent)
	}
}

func TestClient_do_success(t *testing.T) {
	// テスト用HTTPサーバーを作成
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		_, _ = w.Write([]byte(`{"id": "test-id", "name": "test-name"}`))
	}))
	defer server.Close()

	client, err := NewClient("test-token")
	if err != nil {
		t.Fatalf("failed to create client: %v", err)
	}

	// テスト用サーバーのURLを設定
	if err := client.SetBaseURL(server.URL); err != nil {
		t.Fatalf("failed to set base URL: %v", err)
	}

	ctx := context.Background()
	req, err := client.newRequest(ctx, http.MethodGet, "/test", nil)
	if err != nil {
		t.Fatalf("failed to create request: %v", err)
	}

	type testResponse struct {
		ID   string `json:"id"`
		Name string `json:"name"`
	}

	var resp testResponse
	err = client.do(req, &resp)
	if err != nil {
		t.Errorf("unexpected error: %v", err)
	}

	if resp.ID != "test-id" {
		t.Errorf("expected ID test-id, got %s", resp.ID)
	}
	if resp.Name != "test-name" {
		t.Errorf("expected name test-name, got %s", resp.Name)
	}
}

func TestClient_do_error(t *testing.T) {
	// エラーレスポンスを返すテスト用HTTPサーバー
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusBadRequest)
		_, _ = w.Write([]byte(`{"error": "Invalid request"}`))
	}))
	defer server.Close()

	client, err := NewClient("test-token")
	if err != nil {
		t.Fatalf("failed to create client: %v", err)
	}

	if err := client.SetBaseURL(server.URL); err != nil {
		t.Fatalf("failed to set base URL: %v", err)
	}

	ctx := context.Background()
	req, err := client.newRequest(ctx, http.MethodGet, "/test", nil)
	if err != nil {
		t.Fatalf("failed to create request: %v", err)
	}

	err = client.do(req, nil)
	if err == nil {
		t.Error("expected error but got nil")
	}

	apiErr, ok := err.(*Error)
	if !ok {
		t.Errorf("expected Error, got %T", err)
	}

	if apiErr.StatusCode != http.StatusBadRequest {
		t.Errorf("expected status code %d, got %d", http.StatusBadRequest, apiErr.StatusCode)
	}
}

func TestError_Methods(t *testing.T) {
	tests := []struct {
		name       string
		statusCode int
		message    string
		checkFunc  func(*Error) bool
		expected   bool
	}{
		{
			name:       "IsNotFound - 404",
			statusCode: http.StatusNotFound,
			message:    "Not found",
			checkFunc:  (*Error).IsNotFound,
			expected:   true,
		},
		{
			name:       "IsNotFound - 400",
			statusCode: http.StatusBadRequest,
			message:    "Bad request",
			checkFunc:  (*Error).IsNotFound,
			expected:   false,
		},
		{
			name:       "IsUnauthorized - 401",
			statusCode: http.StatusUnauthorized,
			message:    "Unauthorized",
			checkFunc:  (*Error).IsUnauthorized,
			expected:   true,
		},
		{
			name:       "IsForbidden - 403",
			statusCode: http.StatusForbidden,
			message:    "Forbidden",
			checkFunc:  (*Error).IsForbidden,
			expected:   true,
		},
		{
			name:       "IsRateLimited - 429",
			statusCode: http.StatusTooManyRequests,
			message:    "Too many requests",
			checkFunc:  (*Error).IsRateLimited,
			expected:   true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			apiErr := &Error{
				StatusCode: tt.statusCode,
				Message:    tt.message,
			}

			result := tt.checkFunc(apiErr)
			if result != tt.expected {
				t.Errorf("expected %v, got %v", tt.expected, result)
			}

			// Test Error() method
			actualError := apiErr.Error()
			if actualError == "" {
				t.Error("Error() returned empty string")
			}
		})
	}
}
