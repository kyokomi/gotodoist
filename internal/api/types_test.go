package api

import (
	"encoding/json"
	"strings"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestTodoistTime_UnmarshalJSON(t *testing.T) {
	tests := []struct {
		name     string
		input    string
		expected time.Time
		wantErr  bool
	}{
		{
			name:     "null value",
			input:    `"null"`,
			expected: time.Time{},
			wantErr:  false,
		},
		{
			name:     "empty string",
			input:    `""`,
			expected: time.Time{},
			wantErr:  false,
		},
		{
			name:     "RFC3339 format",
			input:    `"2024-01-15T10:30:00Z"`,
			expected: time.Date(2024, 1, 15, 10, 30, 0, 0, time.UTC),
			wantErr:  false,
		},
		{
			name:     "RFC3339Nano format",
			input:    `"2024-01-15T10:30:00.123456Z"`,
			expected: time.Date(2024, 1, 15, 10, 30, 0, 123456000, time.UTC),
			wantErr:  false,
		},
		{
			name:     "Todoist format with microseconds",
			input:    `"2024-01-15T10:30:00.123456Z"`,
			expected: time.Date(2024, 1, 15, 10, 30, 0, 123456000, time.UTC),
			wantErr:  false,
		},
		{
			name:     "Todoist format without microseconds",
			input:    `"2024-01-15T10:30:00Z"`,
			expected: time.Date(2024, 1, 15, 10, 30, 0, 0, time.UTC),
			wantErr:  false,
		},
		{
			name:     "without timezone",
			input:    `"2024-01-15T10:30:00"`,
			expected: time.Date(2024, 1, 15, 10, 30, 0, 0, time.UTC),
			wantErr:  false,
		},
		{
			name:     "invalid format - should not error",
			input:    `"invalid-date"`,
			expected: time.Time{}, // パースに失敗してもゼロ値を返す
			wantErr:  false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var todoistTime TodoistTime
			err := json.Unmarshal([]byte(tt.input), &todoistTime)

			if tt.wantErr {
				assert.Error(t, err, "エラーが期待されますが、nilが返されました")
				return
			}

			require.NoError(t, err, "予期しないエラーが発生しました")
			assert.True(t, todoistTime.Equal(tt.expected),
				"期待される時間 %v と実際の時間 %v が一致しません", tt.expected, todoistTime.Time)
		})
	}
}

func TestTodoistTime_MarshalJSON(t *testing.T) {
	tests := []struct {
		name     string
		time     TodoistTime
		expected string
	}{
		{
			name:     "zero time",
			time:     TodoistTime{},
			expected: `""`,
		},
		{
			name:     "valid time",
			time:     TodoistTime{time.Date(2024, 1, 15, 10, 30, 0, 0, time.UTC)},
			expected: `"2024-01-15T10:30:00Z"`,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result, err := json.Marshal(tt.time)
			require.NoError(t, err, "JSONマーシャリングで予期しないエラーが発生しました")
			assert.Equal(t, tt.expected, string(result), "マーシャリング結果が期待値と異なります")
		})
	}
}

func TestItem_JSONUnmarshaling(t *testing.T) {
	jsonData := `{
		"id": "item-123",
		"user_id": "user-456",
		"project_id": "project-789",
		"content": "Test task",
		"description": "Test description",
		"priority": 2,
		"child_order": 1,
		"day_order": 2,
		"is_collapsed": false,
		"labels": ["label1", "label2"],
		"assigned_by_uid": "user-assigned",
		"responsible_uid": "user-responsible",
		"added_at": "2024-01-15T10:30:00Z",
		"completed_at": "2024-01-16T15:45:00Z",
		"is_deleted": false,
		"sync_id": "sync-999",
		"due": {
			"date": "2024-01-20",
			"string": "next Friday",
			"lang": "en",
			"is_recurring": false,
			"timezone": "UTC"
		}
	}`

	var item Item
	err := json.Unmarshal([]byte(jsonData), &item)
	if err != nil {
		t.Fatalf("failed to unmarshal item: %v", err)
	}

	// フィールドの検証
	if item.ID != "item-123" {
		t.Errorf("expected ID 'item-123', got %s", item.ID)
	}

	if item.Content != "Test task" {
		t.Errorf("expected content 'Test task', got %s", item.Content)
	}

	if item.Priority != 2 {
		t.Errorf("expected priority 2, got %d", item.Priority)
	}

	if len(item.Labels) != 2 {
		t.Errorf("expected 2 labels, got %d", len(item.Labels))
	}

	if item.Labels[0] != "label1" || item.Labels[1] != "label2" {
		t.Errorf("expected labels [label1, label2], got %v", item.Labels)
	}

	expectedAddedAt := time.Date(2024, 1, 15, 10, 30, 0, 0, time.UTC)
	if !item.DateAdded.Equal(expectedAddedAt) {
		t.Errorf("expected added_at %v, got %v", expectedAddedAt, item.DateAdded.Time)
	}

	if item.DateCompleted == nil {
		t.Error("expected completed_at to be set")
	} else {
		expectedCompletedAt := time.Date(2024, 1, 16, 15, 45, 0, 0, time.UTC)
		if !item.DateCompleted.Equal(expectedCompletedAt) {
			t.Errorf("expected completed_at %v, got %v", expectedCompletedAt, item.DateCompleted.Time)
		}
	}

	if item.Due == nil {
		t.Error("expected due to be set")
	} else {
		if item.Due.Date != "2024-01-20" {
			t.Errorf("expected due date '2024-01-20', got %s", item.Due.Date)
		}
		if item.Due.String != "next Friday" {
			t.Errorf("expected due string 'next Friday', got %s", item.Due.String)
		}
	}
}

func TestProject_JSONUnmarshaling(t *testing.T) {
	jsonData := `{
		"id": "project-123",
		"name": "Test Project",
		"color": "blue",
		"parent_id": "parent-456",
		"child_order": 1,
		"collapsed": false,
		"shared": true,
		"is_deleted": false,
		"is_archived": false,
		"is_favorite": true,
		"sync_id": "sync-789",
		"inbox_project": false,
		"team_inbox": false
	}`

	var project Project
	err := json.Unmarshal([]byte(jsonData), &project)
	if err != nil {
		t.Fatalf("failed to unmarshal project: %v", err)
	}

	// フィールドの検証
	if project.ID != "project-123" {
		t.Errorf("expected ID 'project-123', got %s", project.ID)
	}

	if project.Name != "Test Project" {
		t.Errorf("expected name 'Test Project', got %s", project.Name)
	}

	if project.Color != "blue" {
		t.Errorf("expected color 'blue', got %s", project.Color)
	}

	if project.ParentID != "parent-456" {
		t.Errorf("expected parent_id 'parent-456', got %s", project.ParentID)
	}

	if !project.Shared {
		t.Error("expected shared to be true")
	}

	if !project.IsFavorite {
		t.Error("expected is_favorite to be true")
	}

	if project.IsDeleted {
		t.Error("expected is_deleted to be false")
	}

	if project.IsArchived {
		t.Error("expected is_archived to be false")
	}
}

func TestSyncRequest_JSONMarshaling(t *testing.T) {
	req := SyncRequest{
		SyncToken:     "test-token",
		ResourceTypes: []string{ResourceItems, ResourceProjects},
		Commands: []Command{
			{
				Type: CommandItemAdd,
				UUID: "uuid-123",
				Args: map[string]interface{}{
					"content":    "Test task",
					"project_id": "project-456",
				},
			},
		},
	}

	data, err := json.Marshal(req)
	if err != nil {
		t.Fatalf("failed to marshal sync request: %v", err)
	}

	// JSONに含まれるべき要素を確認
	jsonStr := string(data)
	if !containsString(jsonStr, "test-token") {
		t.Error("JSON should contain sync token")
	}

	if !containsString(jsonStr, ResourceItems) {
		t.Error("JSON should contain items resource type")
	}

	if !containsString(jsonStr, CommandItemAdd) {
		t.Error("JSON should contain item_add command")
	}
}

func TestSyncResponse_JSONUnmarshaling(t *testing.T) {
	jsonData := `{
		"sync_token": "response-token",
		"full_sync": true,
		"items": [
			{
				"id": "item-1",
				"content": "Task 1",
				"project_id": "project-1",
				"priority": 1,
				"is_deleted": false,
				"added_at": "2024-01-15T10:30:00Z"
			}
		],
		"projects": [
			{
				"id": "project-1",
				"name": "Project 1",
				"color": "blue",
				"is_deleted": false,
				"is_archived": false,
				"is_favorite": false,
				"shared": false,
				"inbox_project": false,
				"team_inbox": false,
				"child_order": 1
			}
		],
		"temp_id_mapping": {
			"temp-123": "real-456"
		},
		"sync_status": {
			"status": "ok"
		}
	}`

	var resp SyncResponse
	err := json.Unmarshal([]byte(jsonData), &resp)
	if err != nil {
		t.Fatalf("failed to unmarshal sync response: %v", err)
	}

	if resp.SyncToken != "response-token" {
		t.Errorf("expected sync token 'response-token', got %s", resp.SyncToken)
	}

	if !resp.FullSync {
		t.Error("expected full_sync to be true")
	}

	if len(resp.Items) != 1 {
		t.Errorf("expected 1 item, got %d", len(resp.Items))
	}

	if len(resp.Projects) != 1 {
		t.Errorf("expected 1 project, got %d", len(resp.Projects))
	}

	if resp.TempIDMapping["temp-123"] != "real-456" {
		t.Errorf("expected temp ID mapping temp-123 -> real-456, got %s", resp.TempIDMapping["temp-123"])
	}
}

// containsString はテスト用のヘルパー関数
func containsString(s, substr string) bool {
	return strings.Contains(s, substr)
}
