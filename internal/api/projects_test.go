package api

import (
	"context"
	"net/http"
	"net/http/httptest"
	"testing"
)

func TestCreateProject(t *testing.T) {
	tests := []struct {
		name      string
		req       *CreateProjectRequest
		response  string
		wantError bool
	}{
		{
			name: "valid request",
			req: &CreateProjectRequest{
				Name:       "Test Project",
				Color:      "blue",
				IsFavorite: true,
			},
			response: `{
				"sync_token": "test-sync-token",
				"full_sync": false,
				"temp_id_mapping": {"temp-123": "real-123"}
			}`,
			wantError: false,
		},
		{
			name:      "nil request",
			req:       nil,
			wantError: true,
		},
		{
			name: "empty name",
			req: &CreateProjectRequest{
				Name: "",
			},
			wantError: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				if tt.wantError && tt.req != nil && tt.req.Name == "" {
					w.WriteHeader(http.StatusBadRequest)
					w.Write([]byte(`{"error": "Name is required"}`))
					return
				}

				w.Header().Set("Content-Type", "application/json")
				w.WriteHeader(http.StatusOK)
				w.Write([]byte(tt.response))
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
			resp, err := client.CreateProject(ctx, tt.req)

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

			if resp.SyncToken != "test-sync-token" {
				t.Errorf("expected sync token 'test-sync-token', got %s", resp.SyncToken)
			}
		})
	}
}

func TestUpdateProject(t *testing.T) {
	tests := []struct {
		name      string
		projectID string
		req       *UpdateProjectRequest
		response  string
		wantError bool
	}{
		{
			name:      "valid request",
			projectID: "project-123",
			req: &UpdateProjectRequest{
				Name:       "Updated Project",
				Color:      "red",
				IsFavorite: false,
			},
			response: `{
				"sync_token": "test-sync-token-updated",
				"full_sync": false
			}`,
			wantError: false,
		},
		{
			name:      "empty project ID",
			projectID: "",
			req:       &UpdateProjectRequest{Name: "Updated Project"},
			wantError: true,
		},
		{
			name:      "nil request",
			projectID: "project-123",
			req:       nil,
			wantError: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				w.Header().Set("Content-Type", "application/json")
				w.WriteHeader(http.StatusOK)
				w.Write([]byte(tt.response))
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
			resp, err := client.UpdateProject(ctx, tt.projectID, tt.req)

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
			}
		})
	}
}

func TestGetAllProjects(t *testing.T) {
	response := `{
		"sync_token": "test-sync-token",
		"full_sync": true,
		"projects": [
			{
				"id": "project-1",
				"name": "Project 1",
				"color": "blue",
				"is_deleted": false,
				"is_archived": false,
				"is_favorite": true,
				"shared": false,
				"inbox_project": false,
				"team_inbox": false,
				"child_order": 1
			},
			{
				"id": "project-2",
				"name": "Project 2",
				"color": "red",
				"is_deleted": true,
				"is_archived": false,
				"is_favorite": false,
				"shared": false,
				"inbox_project": false,
				"team_inbox": false,
				"child_order": 2
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
	projects, err := client.GetAllProjects(ctx)
	if err != nil {
		t.Errorf("unexpected error: %v", err)
		return
	}

	// 削除されていないプロジェクトのみが返されるべき
	if len(projects) != 1 {
		t.Errorf("expected 1 project, got %d", len(projects))
		return
	}

	if projects[0].ID != "project-1" {
		t.Errorf("expected project ID 'project-1', got %s", projects[0].ID)
	}

	if projects[0].Name != "Project 1" {
		t.Errorf("expected project name 'Project 1', got %s", projects[0].Name)
	}

	if projects[0].Color != "blue" {
		t.Errorf("expected color 'blue', got %s", projects[0].Color)
	}

	if !projects[0].IsFavorite {
		t.Error("expected project to be favorite")
	}

	if projects[0].IsDeleted {
		t.Error("expected project to not be deleted")
	}
}

func TestDeleteProject(t *testing.T) {
	response := `{
		"sync_token": "test-sync-token-delete",
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
	resp, err := client.DeleteProject(ctx, "project-123")
	if err != nil {
		t.Errorf("unexpected error: %v", err)
		return
	}

	if resp == nil {
		t.Error("expected response but got nil")
		return
	}

	if resp.SyncToken != "test-sync-token-delete" {
		t.Errorf("expected sync token 'test-sync-token-delete', got %s", resp.SyncToken)
	}
}

func TestArchiveProject(t *testing.T) {
	tests := []struct {
		name      string
		projectID string
		wantError bool
	}{
		{
			name:      "valid project ID",
			projectID: "project-123",
			wantError: false,
		},
		{
			name:      "empty project ID",
			projectID: "",
			wantError: true,
		},
	}

	response := `{
		"sync_token": "test-sync-token-archive",
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
			resp, err := client.ArchiveProject(ctx, tt.projectID)

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

			if resp.SyncToken != "test-sync-token-archive" {
				t.Errorf("expected sync token 'test-sync-token-archive', got %s", resp.SyncToken)
			}
		})
	}
}

func TestUnarchiveProject(t *testing.T) {
	tests := []struct {
		name      string
		projectID string
		wantError bool
	}{
		{
			name:      "valid project ID",
			projectID: "project-123",
			wantError: false,
		},
		{
			name:      "empty project ID",
			projectID: "",
			wantError: true,
		},
	}

	response := `{
		"sync_token": "test-sync-token-unarchive",
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
			resp, err := client.UnarchiveProject(ctx, tt.projectID)

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

			if resp.SyncToken != "test-sync-token-unarchive" {
				t.Errorf("expected sync token 'test-sync-token-unarchive', got %s", resp.SyncToken)
			}
		})
	}
}

func TestGetFavoriteProjects(t *testing.T) {
	response := `{
		"sync_token": "test-sync-token",
		"full_sync": true,
		"projects": [
			{
				"id": "project-1",
				"name": "Favorite Project",
				"color": "blue",
				"is_deleted": false,
				"is_archived": false,
				"is_favorite": true,
				"shared": false,
				"inbox_project": false,
				"team_inbox": false,
				"child_order": 1
			},
			{
				"id": "project-2",
				"name": "Regular Project",
				"color": "red",
				"is_deleted": false,
				"is_archived": false,
				"is_favorite": false,
				"shared": false,
				"inbox_project": false,
				"team_inbox": false,
				"child_order": 2
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
	projects, err := client.GetFavoriteProjects(ctx)
	if err != nil {
		t.Errorf("unexpected error: %v", err)
		return
	}

	// お気に入りプロジェクトのみが返されるべき
	if len(projects) != 1 {
		t.Errorf("expected 1 favorite project, got %d", len(projects))
		return
	}

	if projects[0].ID != "project-1" {
		t.Errorf("expected project ID 'project-1', got %s", projects[0].ID)
	}

	if !projects[0].IsFavorite {
		t.Error("expected project to be favorite")
	}
}

func TestGetSharedProjects(t *testing.T) {
	response := `{
		"sync_token": "test-sync-token",
		"full_sync": true,
		"projects": [
			{
				"id": "project-1",
				"name": "Shared Project",
				"color": "blue",
				"is_deleted": false,
				"is_archived": false,
				"is_favorite": false,
				"shared": true,
				"inbox_project": false,
				"team_inbox": false,
				"child_order": 1
			},
			{
				"id": "project-2",
				"name": "Private Project",
				"color": "red",
				"is_deleted": false,
				"is_archived": false,
				"is_favorite": false,
				"shared": false,
				"inbox_project": false,
				"team_inbox": false,
				"child_order": 2
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
	projects, err := client.GetSharedProjects(ctx)
	if err != nil {
		t.Errorf("unexpected error: %v", err)
		return
	}

	// 共有プロジェクトのみが返されるべき
	if len(projects) != 1 {
		t.Errorf("expected 1 shared project, got %d", len(projects))
		return
	}

	if projects[0].ID != "project-1" {
		t.Errorf("expected project ID 'project-1', got %s", projects[0].ID)
	}

	if !projects[0].Shared {
		t.Error("expected project to be shared")
	}
}
