package api

import (
	"context"
	"net/http"
	"net/http/httptest"
	"testing"
)

func TestGetAllData(t *testing.T) {
	response := `{
		"sync_token": "test-all-data-token",
		"full_sync": true,
		"items": [],
		"projects": [],
		"labels": []
	}`

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		w.Write([]byte(response))
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
	resp, err := client.GetAllData(ctx)
	if err != nil {
		t.Errorf("unexpected error: %v", err)
		return
	}

	if resp == nil {
		t.Error("expected response but got nil")
		return
	}

	if resp.SyncToken != "test-all-data-token" {
		t.Errorf("expected sync token 'test-all-data-token', got %s", resp.SyncToken)
	}

	if !resp.FullSync {
		t.Error("expected full_sync to be true")
	}
}

func TestGetItems(t *testing.T) {
	response := `{
		"sync_token": "items-sync-token",
		"full_sync": false,
		"items": [
			{
				"id": "item-1",
				"content": "Test item",
				"project_id": "project-1",
				"priority": 1,
				"is_deleted": false,
				"added_at": "2024-01-15T10:30:00Z"
			}
		]
	}`

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		w.Write([]byte(response))
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
	resp, err := client.GetItems(ctx, "test-sync-token")
	if err != nil {
		t.Errorf("unexpected error: %v", err)
		return
	}

	if resp == nil {
		t.Error("expected response but got nil")
		return
	}

	if resp.SyncToken != "items-sync-token" {
		t.Errorf("expected sync token 'items-sync-token', got %s", resp.SyncToken)
	}

	if len(resp.Items) != 1 {
		t.Errorf("expected 1 item, got %d", len(resp.Items))
	}
}

func TestGetProjects(t *testing.T) {
	response := `{
		"sync_token": "projects-sync-token",
		"full_sync": false,
		"projects": [
			{
				"id": "project-1",
				"name": "Test Project",
				"color": "blue",
				"is_deleted": false,
				"is_archived": false,
				"is_favorite": false,
				"shared": false,
				"inbox_project": false,
				"team_inbox": false,
				"child_order": 1
			}
		]
	}`

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		w.Write([]byte(response))
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
	resp, err := client.GetProjects(ctx, "test-sync-token")
	if err != nil {
		t.Errorf("unexpected error: %v", err)
		return
	}

	if resp == nil {
		t.Error("expected response but got nil")
		return
	}

	if resp.SyncToken != "projects-sync-token" {
		t.Errorf("expected sync token 'projects-sync-token', got %s", resp.SyncToken)
	}

	if len(resp.Projects) != 1 {
		t.Errorf("expected 1 project, got %d", len(resp.Projects))
	}
}

func TestAddItem(t *testing.T) {
	tests := []struct {
		name      string
		content   string
		projectID string
		wantError bool
	}{
		{
			name:      "with project ID",
			content:   "Test task",
			projectID: "project-123",
			wantError: false,
		},
		{
			name:      "without project ID",
			content:   "Test task without project",
			projectID: "",
			wantError: false,
		},
	}

	response := `{
		"sync_token": "add-item-token",
		"full_sync": false,
		"temp_id_mapping": {"temp-123": "real-456"}
	}`

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				w.Header().Set("Content-Type", "application/json")
				w.WriteHeader(http.StatusOK)
				w.Write([]byte(response))
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
			resp, err := client.AddItem(ctx, tt.content, tt.projectID)

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

			if resp.SyncToken != "add-item-token" {
				t.Errorf("expected sync token 'add-item-token', got %s", resp.SyncToken)
			}
		})
	}
}

func TestUpdateItem(t *testing.T) {
	response := `{
		"sync_token": "update-item-token",
		"full_sync": false
	}`

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		w.Write([]byte(response))
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
	updates := map[string]interface{}{
		"content":  "Updated content",
		"priority": 2,
	}

	resp, err := client.UpdateItem(ctx, "item-123", updates)
	if err != nil {
		t.Errorf("unexpected error: %v", err)
		return
	}

	if resp == nil {
		t.Error("expected response but got nil")
		return
	}

	if resp.SyncToken != "update-item-token" {
		t.Errorf("expected sync token 'update-item-token', got %s", resp.SyncToken)
	}
}

func TestCompleteItem(t *testing.T) {
	response := `{
		"sync_token": "complete-item-token",
		"full_sync": false
	}`

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		w.Write([]byte(response))
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
	resp, err := client.CompleteItem(ctx, "item-123")
	if err != nil {
		t.Errorf("unexpected error: %v", err)
		return
	}

	if resp == nil {
		t.Error("expected response but got nil")
		return
	}

	if resp.SyncToken != "complete-item-token" {
		t.Errorf("expected sync token 'complete-item-token', got %s", resp.SyncToken)
	}
}

func TestDeleteItem(t *testing.T) {
	response := `{
		"sync_token": "delete-item-token",
		"full_sync": false
	}`

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		w.Write([]byte(response))
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
	resp, err := client.DeleteItem(ctx, "item-123")
	if err != nil {
		t.Errorf("unexpected error: %v", err)
		return
	}

	if resp == nil {
		t.Error("expected response but got nil")
		return
	}

	if resp.SyncToken != "delete-item-token" {
		t.Errorf("expected sync token 'delete-item-token', got %s", resp.SyncToken)
	}
}

func TestAddProject(t *testing.T) {
	response := `{
		"sync_token": "add-project-token",
		"full_sync": false,
		"temp_id_mapping": {"temp-project-123": "real-project-456"}
	}`

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		w.Write([]byte(response))
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
	resp, err := client.AddProject(ctx, "Test Project")
	if err != nil {
		t.Errorf("unexpected error: %v", err)
		return
	}

	if resp == nil {
		t.Error("expected response but got nil")
		return
	}

	if resp.SyncToken != "add-project-token" {
		t.Errorf("expected sync token 'add-project-token', got %s", resp.SyncToken)
	}

	if resp.TempIDMapping["temp-project-123"] != "real-project-456" {
		t.Error("expected temp ID mapping to be set")
	}
}

func TestUpdateProjectSync(t *testing.T) {
	response := `{
		"sync_token": "update-project-token",
		"full_sync": false
	}`

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		w.Write([]byte(response))
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
	updates := map[string]interface{}{
		"name":  "Updated Project Name",
		"color": "red",
	}

	resp, err := client.UpdateProjectSync(ctx, "project-123", updates)
	if err != nil {
		t.Errorf("unexpected error: %v", err)
		return
	}

	if resp == nil {
		t.Error("expected response but got nil")
		return
	}

	if resp.SyncToken != "update-project-token" {
		t.Errorf("expected sync token 'update-project-token', got %s", resp.SyncToken)
	}
}

func TestDeleteProjectSync(t *testing.T) {
	response := `{
		"sync_token": "delete-project-token",
		"full_sync": false
	}`

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		w.Write([]byte(response))
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
	resp, err := client.DeleteProjectSync(ctx, "project-123")
	if err != nil {
		t.Errorf("unexpected error: %v", err)
		return
	}

	if resp == nil {
		t.Error("expected response but got nil")
		return
	}

	if resp.SyncToken != "delete-project-token" {
		t.Errorf("expected sync token 'delete-project-token', got %s", resp.SyncToken)
	}
}

func TestExecuteCommands(t *testing.T) {
	tests := []struct {
		name      string
		commands  []Command
		wantError bool
	}{
		{
			name: "valid commands",
			commands: []Command{
				{
					Type: CommandItemAdd,
					UUID: "uuid-1",
					Args: map[string]interface{}{
						"content": "Task 1",
					},
				},
				{
					Type: CommandItemAdd,
					UUID: "uuid-2",
					Args: map[string]interface{}{
						"content": "Task 2",
					},
				},
			},
			wantError: false,
		},
		{
			name:      "empty commands",
			commands:  []Command{},
			wantError: true,
		},
	}

	response := `{
		"sync_token": "execute-commands-token",
		"full_sync": false
	}`

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				w.Header().Set("Content-Type", "application/json")
				w.WriteHeader(http.StatusOK)
				w.Write([]byte(response))
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
			resp, err := client.ExecuteCommands(ctx, tt.commands)

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

			if resp.SyncToken != "execute-commands-token" {
				t.Errorf("expected sync token 'execute-commands-token', got %s", resp.SyncToken)
			}
		})
	}
}
