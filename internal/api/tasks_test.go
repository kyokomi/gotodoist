package api

import (
	"context"
	"net/http"
	"net/http/httptest"
	"testing"
)

func TestCreateTask(t *testing.T) {
	tests := []struct {
		name      string
		req       *CreateTaskRequest
		response  string
		wantError bool
	}{
		{
			name: "valid request",
			req: &CreateTaskRequest{
				Content:     "Test task",
				Description: "Test description",
				ProjectID:   "project-123",
				Priority:    2,
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
			name: "empty content",
			req: &CreateTaskRequest{
				Content: "",
			},
			wantError: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// テスト用HTTPサーバーを作成
			server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				if tt.wantError && tt.req != nil && tt.req.Content == "" {
					// 400エラーを返す
					w.WriteHeader(http.StatusBadRequest)
					w.Write([]byte(`{"error": "Content is required"}`))
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
			resp, err := client.CreateTask(ctx, tt.req)

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

func TestUpdateTask(t *testing.T) {
	tests := []struct {
		name      string
		taskID    string
		req       *UpdateTaskRequest
		response  string
		wantError bool
	}{
		{
			name:   "valid request",
			taskID: "task-123",
			req: &UpdateTaskRequest{
				Content:  "Updated task",
				Priority: 3,
			},
			response: `{
				"sync_token": "test-sync-token-updated",
				"full_sync": false
			}`,
			wantError: false,
		},
		{
			name:      "empty task ID",
			taskID:    "",
			req:       &UpdateTaskRequest{Content: "Updated task"},
			wantError: true,
		},
		{
			name:      "nil request",
			taskID:    "task-123",
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
			resp, err := client.UpdateTask(ctx, tt.taskID, tt.req)

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

func TestGetTasks(t *testing.T) {
	response := `{
		"sync_token": "test-sync-token",
		"full_sync": true,
		"items": [
			{
				"id": "task-1",
				"content": "Task 1",
				"project_id": "project-1",
				"priority": 1,
				"is_deleted": false,
				"added_at": "2024-01-01T10:00:00Z"
			},
			{
				"id": "task-2",
				"content": "Task 2",
				"project_id": "project-2",
				"priority": 2,
				"is_deleted": true,
				"added_at": "2024-01-02T10:00:00Z"
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
	tasks, err := client.GetTasks(ctx)
	if err != nil {
		t.Errorf("unexpected error: %v", err)
		return
	}

	// 削除されていないタスクのみが返されるべき
	if len(tasks) != 1 {
		t.Errorf("expected 1 task, got %d", len(tasks))
		return
	}

	if tasks[0].ID != "task-1" {
		t.Errorf("expected task ID 'task-1', got %s", tasks[0].ID)
	}

	if tasks[0].Content != "Task 1" {
		t.Errorf("expected task content 'Task 1', got %s", tasks[0].Content)
	}

	if tasks[0].Priority != 1 {
		t.Errorf("expected priority 1, got %d", tasks[0].Priority)
	}
}

func TestGetTasksByProject(t *testing.T) {
	tests := []struct {
		name       string
		projectID  string
		wantError  bool
		wantLength int
	}{
		{
			name:       "valid project ID",
			projectID:  "project-1",
			wantError:  false,
			wantLength: 1,
		},
		{
			name:      "empty project ID",
			projectID: "",
			wantError: true,
		},
	}

	response := `{
		"sync_token": "test-sync-token",
		"full_sync": true,
		"items": [
			{
				"id": "task-1",
				"content": "Task 1",
				"project_id": "project-1",
				"priority": 1,
				"is_deleted": false,
				"added_at": "2024-01-01T10:00:00Z"
			},
			{
				"id": "task-2",
				"content": "Task 2",
				"project_id": "project-2",
				"priority": 2,
				"is_deleted": false,
				"added_at": "2024-01-02T10:00:00Z"
			}
		]
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
			tasks, err := client.GetTasksByProject(ctx, tt.projectID)

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

			if len(tasks) != tt.wantLength {
				t.Errorf("expected %d task(s), got %d", tt.wantLength, len(tasks))
			}
		})
	}
}

func TestCloseTask(t *testing.T) {
	response := `{
		"sync_token": "test-sync-token-close",
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
	resp, err := client.CloseTask(ctx, "task-123")
	if err != nil {
		t.Errorf("unexpected error: %v", err)
		return
	}

	if resp == nil {
		t.Error("expected response but got nil")
		return
	}

	if resp.SyncToken != "test-sync-token-close" {
		t.Errorf("expected sync token 'test-sync-token-close', got %s", resp.SyncToken)
	}
}

func TestDeleteTask(t *testing.T) {
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
	resp, err := client.DeleteTask(ctx, "task-123")
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

func TestPriority_String(t *testing.T) {
	tests := []struct {
		priority Priority
		expected string
	}{
		{PriorityNormal, "Normal"},
		{PriorityHigh, "High"},
		{PriorityVeryHigh, "Very High"},
		{PriorityUrgent, "Urgent"},
		{Priority(99), "Priority(99)"},
	}

	for _, tt := range tests {
		t.Run(tt.expected, func(t *testing.T) {
			result := tt.priority.String()
			if result != tt.expected {
				t.Errorf("expected %s, got %s", tt.expected, result)
			}
		})
	}
}

func TestPriority_IsValid(t *testing.T) {
	tests := []struct {
		priority Priority
		expected bool
	}{
		{PriorityNormal, true},
		{PriorityHigh, true},
		{PriorityVeryHigh, true},
		{PriorityUrgent, true},
		{Priority(0), false},
		{Priority(5), false},
		{Priority(99), false},
	}

	for _, tt := range tests {
		t.Run(tt.priority.String(), func(t *testing.T) {
			result := tt.priority.IsValid()
			if result != tt.expected {
				t.Errorf("expected %v, got %v", tt.expected, result)
			}
		})
	}
}

func TestGetTasksByPriority(t *testing.T) {
	response := `{
		"sync_token": "test-sync-token",
		"full_sync": true,
		"items": [
			{
				"id": "task-1",
				"content": "High priority task",
				"project_id": "project-1",
				"priority": 2,
				"is_deleted": false,
				"added_at": "2024-01-01T10:00:00Z"
			},
			{
				"id": "task-2",
				"content": "Normal priority task",
				"project_id": "project-1",
				"priority": 1,
				"is_deleted": false,
				"added_at": "2024-01-02T10:00:00Z"
			}
		]
	}`

	tests := []struct {
		name       string
		priority   Priority
		wantError  bool
		wantLength int
	}{
		{
			name:       "high priority",
			priority:   PriorityHigh,
			wantError:  false,
			wantLength: 1,
		},
		{
			name:       "normal priority",
			priority:   PriorityNormal,
			wantError:  false,
			wantLength: 1,
		},
		{
			name:      "invalid priority",
			priority:  Priority(99),
			wantError: true,
		},
	}

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
			tasks, err := client.GetTasksByPriority(ctx, tt.priority)

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

			if len(tasks) != tt.wantLength {
				t.Errorf("expected %d task(s), got %d", tt.wantLength, len(tasks))
			}

			// 優先度が正しくフィルタリングされているかチェック
			for _, task := range tasks {
				if task.Priority != int(tt.priority) {
					t.Errorf("expected priority %d, got %d", int(tt.priority), task.Priority)
				}
			}
		})
	}
}
