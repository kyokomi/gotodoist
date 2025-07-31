package api

import (
	"context"
	"fmt"
	"net/http"
	"testing"

	"github.com/kyokomi/gotodoist/internal/testhelper"
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
				Color:      testhelper.TestColorBlue,
				IsFavorite: true,
			},
			response: fmt.Sprintf(`{
				"sync_token": "%s",
				"full_sync": false,
				"temp_id_mapping": {"temp-123": "real-123"}
			}`, testhelper.TestSyncToken),
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
			server := testhelper.NewHTTPTestServer(func(w http.ResponseWriter, r *http.Request) {
				if tt.wantError && tt.req != nil && tt.req.Name == "" {
					testhelper.ErrorResponse(t, w, http.StatusBadRequest, "Name is required")
					return
				}

				testhelper.JSONResponse(t, w, http.StatusOK, tt.response)
			})
			defer server.Close()

			client, err := NewClient(testhelper.TestAPIToken)
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

			if resp.SyncToken != testhelper.TestSyncToken {
				t.Errorf("expected sync token '%s', got %s", testhelper.TestSyncToken, resp.SyncToken)
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
			server := testhelper.NewHTTPTestServer(func(w http.ResponseWriter, r *http.Request) {
				testhelper.JSONResponse(t, w, http.StatusOK, tt.response)
			})
			defer server.Close()

			client, err := NewClient(testhelper.TestAPIToken)
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
				"id": testhelper.TestProjectID,
				"name": "Project 1",
				"color": testhelper.TestColorBlue,
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

	server := testhelper.NewHTTPTestServer(func(w http.ResponseWriter, r *http.Request) {
		testhelper.JSONResponse(t, w, http.StatusOK, response)
	})
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

	if projects[0].ID != testhelper.TestProjectID {
		t.Errorf("expected project ID 'project-1', got %s", projects[0].ID)
	}

	if projects[0].Name != "Project 1" {
		t.Errorf("expected project name 'Project 1', got %s", projects[0].Name)
	}

	if projects[0].Color != testhelper.TestColorBlue {
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
	response := testhelper.DeleteSyncResponse

	server := testhelper.NewHTTPTestServer(func(w http.ResponseWriter, r *http.Request) {
		testhelper.JSONResponse(t, w, http.StatusOK, response)
	})
	defer server.Close()

	client, err := NewClient(testhelper.TestAPIToken)
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

	if resp.SyncToken != testhelper.TestSyncTokenDelete {
		t.Errorf("expected sync token '%s', got %s", testhelper.TestSyncTokenDelete, resp.SyncToken)
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
			server := testhelper.NewHTTPTestServer(func(w http.ResponseWriter, r *http.Request) {
				testhelper.JSONResponse(t, w, http.StatusOK, response)
			})
			defer server.Close()

			client, err := NewClient(testhelper.TestAPIToken)
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
			server := testhelper.NewHTTPTestServer(func(w http.ResponseWriter, r *http.Request) {
				testhelper.JSONResponse(t, w, http.StatusOK, response)
			})
			defer server.Close()

			client, err := NewClient(testhelper.TestAPIToken)
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
	testFilteredProjects(t, "favorite", true, false, func(client *Client, ctx context.Context) ([]Project, error) {
		return client.GetFavoriteProjects(ctx)
	})
}

func TestGetSharedProjects(t *testing.T) {
	testFilteredProjects(t, "shared", false, true, func(client *Client, ctx context.Context) ([]Project, error) {
		return client.GetSharedProjects(ctx)
	})
}

// Helper function to test filtered projects
func testFilteredProjects(t *testing.T, filterType string, isFavorite, isShared bool,
	getFunc func(*Client, context.Context) ([]Project, error)) {
	response := fmt.Sprintf(`{
		"sync_token": "test-sync-token",
		"full_sync": true,
		"projects": [
			{
				"id": testhelper.TestProjectID,
				"name": "%s Project",
				"color": testhelper.TestColorBlue,
				"is_deleted": false,
				"is_archived": false,
				"is_favorite": %t,
				"shared": %t,
				"inbox_project": false,
				"team_inbox": false,
				"child_order": 1
			},
			{
				"id": "project-2",
				"name": "Other Project", 
				"color": "red",
				"is_deleted": false,
				"is_archived": false,
				"is_favorite": %t,
				"shared": %t,
				"inbox_project": false,
				"team_inbox": false,
				"child_order": 2
			}
		]
	}`, filterType, isFavorite, isShared, !isFavorite, !isShared)

	server := testhelper.NewHTTPTestServer(func(w http.ResponseWriter, r *http.Request) {
		testhelper.JSONResponse(t, w, http.StatusOK, response)
	})
	defer server.Close()

	client, err := NewClient("test-token")
	if err != nil {
		t.Fatalf("failed to create client: %v", err)
	}

	if err := client.SetBaseURL(server.URL); err != nil {
		t.Fatalf("failed to set base URL: %v", err)
	}

	ctx := context.Background()
	projects, err := getFunc(client, ctx)
	if err != nil {
		t.Errorf("unexpected error: %v", err)
		return
	}

	if len(projects) != 1 {
		t.Errorf("expected 1 %s project, got %d", filterType, len(projects))
		return
	}

	if projects[0].ID != testhelper.TestProjectID {
		t.Errorf("expected project ID 'project-1', got %s", projects[0].ID)
	}

	if isFavorite && !projects[0].IsFavorite {
		t.Error("expected project to be favorite")
	}

	if isShared && !projects[0].Shared {
		t.Error("expected project to be shared")
	}
}
