// Package testhelper provides helper functions and constants for testing
package testhelper

import (
	"context"
	"net/http"
	"testing"
)

// TestUpdateOperation tests a generic update operation with table-driven tests
func TestUpdateOperation(t *testing.T, tests []UpdateTestCase, operationFunc func(*http.ServeMux, *testing.T, UpdateTestCase)) {
	t.Helper()

	for _, tt := range tests {
		t.Run(tt.Name, func(t *testing.T) {
			mux := http.NewServeMux()
			operationFunc(mux, t, tt)
		})
	}
}

// UpdateTestCase represents a test case for update operations
type UpdateTestCase struct {
	Name      string
	ID        string
	Request   interface{}
	Response  string
	WantError bool
}

// TestArchiveOperation tests archive/unarchive operations with common pattern
func TestArchiveOperation(
	t *testing.T,
	operationName, syncToken string,
	operationFunc func(context.Context, string) (SyncResponseInterface, error),
) {
	t.Helper()

	tests := []struct {
		name      string
		id        string
		wantError bool
	}{
		{
			name:      "valid ID",
			id:        "test-123",
			wantError: false,
		},
		{
			name:      "empty ID",
			id:        "",
			wantError: true,
		},
	}

	response := SimpleSyncResponse
	if syncToken != "" {
		response = `{
			"sync_token": "` + syncToken + `",
			"full_sync": false
		}`
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			server := NewHTTPTestServer(func(w http.ResponseWriter, r *http.Request) {
				JSONResponse(t, w, http.StatusOK, response)
			})
			defer server.Close()

			ctx := context.Background()
			resp, err := operationFunc(ctx, tt.id)

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

			if resp == nil {
				t.Error("expected response but got nil")
				return
			}

			if resp.GetSyncToken() != syncToken {
				t.Errorf("expected sync token '%s', got %s", syncToken, resp.GetSyncToken())
			}
		})
	}
}

// SyncResponseInterface defines interface for sync responses to avoid import cycles
type SyncResponseInterface interface {
	GetSyncToken() string
}

// SyncResponseAdapter adapts api.SyncResponse to SyncResponseInterface
type SyncResponseAdapter struct {
	SyncToken string
	FullSync  bool
}

// GetSyncToken returns the sync token from the adapter
func (s *SyncResponseAdapter) GetSyncToken() string {
	return s.SyncToken
}

// TestGetOperation tests get operations with common pattern
func TestGetOperation(
	t *testing.T,
	operationName, response string,
	operationFunc func(context.Context, string) (GetResponseInterface, error),
) {
	t.Helper()

	server := NewHTTPTestServer(func(w http.ResponseWriter, r *http.Request) {
		JSONResponse(t, w, http.StatusOK, response)
	})
	defer server.Close()

	ctx := context.Background()
	resp, err := operationFunc(ctx, TestSyncToken)
	if err != nil {
		t.Errorf("unexpected error: %v", err)
		return
	}

	if resp == nil {
		t.Error("expected response but got nil")
		return
	}

	// Each specific test should verify the response details
}

// GetResponseInterface defines interface for get responses
type GetResponseInterface interface {
	GetSyncToken() string
	GetItemCount() int
}

// TestCase defines a common test case structure
type TestCase struct {
	Name      string
	WantError bool
}

// SyncTestCase defines test cases for sync operations with single ID parameter
type SyncTestCase struct {
	TestCase
	ID string
}

// RunSyncOperationTest runs tests for operations that take an ID and return SyncResponse
func RunSyncOperationTest(
	t *testing.T,
	tests []SyncTestCase,
	response string,
	expectedToken string,
	operationFunc func(client interface{}, ctx context.Context, id string) (SyncResponseInterface, error),
	clientFactory func(token string) (interface{}, error),
	setBaseURL func(client interface{}, url string) error,
) {
	t.Helper()

	for _, tt := range tests {
		t.Run(tt.Name, func(t *testing.T) {
			server := NewHTTPTestServer(func(w http.ResponseWriter, r *http.Request) {
				JSONResponse(t, w, http.StatusOK, response)
			})
			defer server.Close()

			client, err := clientFactory(TestAPIToken)
			if err != nil {
				t.Fatalf("failed to create client: %v", err)
			}
			if err := setBaseURL(client, server.URL); err != nil {
				t.Fatalf("failed to set base URL: %v", err)
			}

			ctx := context.Background()
			resp, err := operationFunc(client, ctx, tt.ID)

			if tt.WantError {
				if err == nil {
					t.Error("expected error but got nil")
				}
				return
			}

			if err != nil {
				t.Errorf("unexpected error: %v", err)
				return
			}

			if resp == nil {
				t.Error("expected response but got nil")
				return
			}

			if resp.GetSyncToken() != expectedToken {
				t.Errorf("expected sync token '%s', got %s", expectedToken, resp.GetSyncToken())
			}
		})
	}
}
