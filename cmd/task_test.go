package cmd

import (
	"strings"
	"testing"
	"time"

	"github.com/spf13/cobra"

	"github.com/kyokomi/gotodoist/internal/api"
)

func TestBuildUpdateTaskRequestFromFlags(t *testing.T) {
	tests := []struct {
		name      string
		flags     map[string]string
		want      *api.UpdateTaskRequest
		wantError bool
		errorMsg  string
	}{
		{
			name:      "no flags specified",
			flags:     map[string]string{},
			want:      nil,
			wantError: true,
			errorMsg:  "at least one update field must be specified",
		},
		{
			name: "content only",
			flags: map[string]string{
				"content": "Updated task content",
			},
			want: &api.UpdateTaskRequest{
				Content: "Updated task content",
			},
			wantError: false,
		},
		{
			name: "valid priority",
			flags: map[string]string{
				"priority": "3",
			},
			want: &api.UpdateTaskRequest{
				Priority: 3,
			},
			wantError: false,
		},
		{
			name: "invalid priority format",
			flags: map[string]string{
				"priority": "invalid",
			},
			want:      nil,
			wantError: true,
			errorMsg:  "invalid priority: invalid",
		},
		{
			name: "priority too low",
			flags: map[string]string{
				"priority": "0",
			},
			want:      nil,
			wantError: true,
			errorMsg:  "priority must be between 1 and 4",
		},
		{
			name: "priority too high",
			flags: map[string]string{
				"priority": "5",
			},
			want:      nil,
			wantError: true,
			errorMsg:  "priority must be between 1 and 4",
		},
		{
			name: "labels parsing",
			flags: map[string]string{
				"labels": "work, personal,  urgent  , home",
			},
			want: &api.UpdateTaskRequest{
				Labels: []string{"work", "personal", "urgent", "home"},
			},
			wantError: false,
		},
		{
			name: "all fields",
			flags: map[string]string{
				"content":     "Complete project",
				"description": "Finish all remaining tasks",
				"priority":    "2",
				"due":         "today",
				"labels":      "work, important",
			},
			want: &api.UpdateTaskRequest{
				Content:     "Complete project",
				Description: "Finish all remaining tasks",
				Priority:    2,
				DueString:   "today",
				Labels:      []string{"work", "important"},
			},
			wantError: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Create a mock cobra command with flags
			cmd := &cobra.Command{}
			cmd.Flags().String("content", "", "task content")
			cmd.Flags().String("description", "", "task description")
			cmd.Flags().String("priority", "", "task priority")
			cmd.Flags().String("due", "", "due date")
			cmd.Flags().String("labels", "", "labels")

			// Set flag values
			for key, value := range tt.flags {
				err := cmd.Flags().Set(key, value)
				if err != nil {
					t.Fatalf("failed to set flag %s: %v", key, err)
				}
			}

			// Get parameters from command flags
			params := getTaskUpdateParams(cmd, []string{"test-id"})

			// Create a dummy executor for testing
			executor := &taskExecutor{}

			// Call the function
			got, err := executor.buildUpdateTaskRequest(params)

			// Check error expectations
			if tt.wantError {
				if err == nil {
					t.Errorf("expected error but got nil")
					return
				}
				if tt.errorMsg != "" && !strings.Contains(err.Error(), tt.errorMsg) {
					t.Errorf("expected error containing '%s', got '%s'", tt.errorMsg, err.Error())
				}
				return
			}

			if err != nil {
				t.Errorf("unexpected error: %v", err)
				return
			}

			// Check result
			if !equalUpdateTaskRequest(got, tt.want) {
				t.Errorf("got %+v, want %+v", got, tt.want)
			}
		})
	}
}

// Helper function to compare UpdateTaskRequest structs
func equalUpdateTaskRequest(a, b *api.UpdateTaskRequest) bool {
	if a == nil && b == nil {
		return true
	}
	if a == nil || b == nil {
		return false
	}

	if a.Content != b.Content ||
		a.Description != b.Description ||
		a.Priority != b.Priority ||
		a.DueString != b.DueString {
		return false
	}

	// Compare Labels slices
	if len(a.Labels) != len(b.Labels) {
		return false
	}
	for i, label := range a.Labels {
		if label != b.Labels[i] {
			return false
		}
	}

	return true
}

func TestFilterActiveTasks(t *testing.T) {
	now := time.Now()
	completedTask := api.Item{
		ID:            "task1",
		Content:       "Completed task",
		DateCompleted: &api.TodoistTime{Time: now},
	}
	activeTask := api.Item{
		ID:            "task2",
		Content:       "Active task",
		DateCompleted: nil,
	}

	tests := []struct {
		name    string
		tasks   []api.Item
		showAll bool
		want    []api.Item
	}{
		{
			name:    "showAll=true returns all tasks",
			tasks:   []api.Item{completedTask, activeTask},
			showAll: true,
			want:    []api.Item{completedTask, activeTask},
		},
		{
			name:    "showAll=false filters out completed tasks",
			tasks:   []api.Item{completedTask, activeTask},
			showAll: false,
			want:    []api.Item{activeTask},
		},
		{
			name:    "empty task list",
			tasks:   []api.Item{},
			showAll: false,
			want:    []api.Item{},
		},
		{
			name:    "all tasks completed",
			tasks:   []api.Item{completedTask},
			showAll: false,
			want:    []api.Item{},
		},
		{
			name:    "all tasks active",
			tasks:   []api.Item{activeTask},
			showAll: false,
			want:    []api.Item{activeTask},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := filterActiveTasks(tt.tasks, tt.showAll)

			if len(got) != len(tt.want) {
				t.Errorf("got %d tasks, want %d tasks", len(got), len(tt.want))
				return
			}

			for i, task := range got {
				if task.ID != tt.want[i].ID {
					t.Errorf("task[%d]: got ID %s, want ID %s", i, task.ID, tt.want[i].ID)
				}
			}
		})
	}
}
