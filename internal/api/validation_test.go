package api

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestValidateCreateProjectRequest(t *testing.T) {
	tests := []struct {
		name    string
		req     *CreateProjectRequest
		wantErr bool
	}{
		{
			name: "valid request",
			req: &CreateProjectRequest{
				Name: "Test Project",
			},
			wantErr: false,
		},
		{
			name:    "nil request",
			req:     nil,
			wantErr: true,
		},
		{
			name: "empty name",
			req: &CreateProjectRequest{
				Name: "",
			},
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := validateCreateProjectRequest(tt.req)
			if tt.wantErr {
				assert.Error(t, err, "validateCreateProjectRequestでエラーが期待されます")
			} else {
				assert.NoError(t, err, "validateCreateProjectRequestでエラーが発生しました")
			}
		})
	}
}

func TestValidateUpdateProjectRequest(t *testing.T) {
	tests := []struct {
		name    string
		req     *UpdateProjectRequest
		wantErr bool
	}{
		{
			name: "valid request",
			req: &UpdateProjectRequest{
				Name: "Updated Project",
			},
			wantErr: false,
		},
		{
			name:    "nil request",
			req:     nil,
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := validateUpdateProjectRequest(tt.req)
			if tt.wantErr {
				assert.Error(t, err, "validateUpdateProjectRequestでエラーが期待されます")
			} else {
				assert.NoError(t, err, "validateUpdateProjectRequestでエラーが発生しました")
			}
		})
	}
}

func TestValidateProjectID(t *testing.T) {
	tests := []struct {
		name      string
		projectID string
		wantErr   bool
	}{
		{
			name:      "valid project ID",
			projectID: "project-123",
			wantErr:   false,
		},
		{
			name:      "empty project ID",
			projectID: "",
			wantErr:   true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := validateProjectID(tt.projectID)
			if tt.wantErr {
				assert.Error(t, err, "validateProjectIDでエラーが期待されます")
			} else {
				assert.NoError(t, err, "validateProjectIDでエラーが発生しました")
			}
		})
	}
}

func TestValidateCreateTaskRequest(t *testing.T) {
	tests := []struct {
		name    string
		req     *CreateTaskRequest
		wantErr bool
	}{
		{
			name: "valid request",
			req: &CreateTaskRequest{
				Content: "Test task",
			},
			wantErr: false,
		},
		{
			name:    "nil request",
			req:     nil,
			wantErr: true,
		},
		{
			name: "empty content",
			req: &CreateTaskRequest{
				Content: "",
			},
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := validateCreateTaskRequest(tt.req)
			if tt.wantErr {
				assert.Error(t, err, "validateCreateTaskRequestでエラーが期待されます")
			} else {
				assert.NoError(t, err, "validateCreateTaskRequestでエラーが発生しました")
			}
		})
	}
}

func TestValidateUpdateTaskRequest(t *testing.T) {
	tests := []struct {
		name    string
		req     *UpdateTaskRequest
		wantErr bool
	}{
		{
			name: "valid request",
			req: &UpdateTaskRequest{
				Content: "Updated task",
			},
			wantErr: false,
		},
		{
			name:    "nil request",
			req:     nil,
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := validateUpdateTaskRequest(tt.req)
			if tt.wantErr {
				assert.Error(t, err, "validateUpdateTaskRequestでエラーが期待されます")
			} else {
				assert.NoError(t, err, "validateUpdateTaskRequestでエラーが発生しました")
			}
		})
	}
}

func TestValidateTaskID(t *testing.T) {
	tests := []struct {
		name    string
		taskID  string
		wantErr bool
	}{
		{
			name:    "valid task ID",
			taskID:  "task-123",
			wantErr: false,
		},
		{
			name:    "empty task ID",
			taskID:  "",
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := validateTaskID(tt.taskID)
			if tt.wantErr {
				assert.Error(t, err, "validateTaskIDでエラーが期待されます")
			} else {
				assert.NoError(t, err, "validateTaskIDでエラーが発生しました")
			}
		})
	}
}

func TestValidateAPIToken(t *testing.T) {
	tests := []struct {
		name    string
		token   string
		wantErr bool
	}{
		{
			name:    "valid token",
			token:   "valid-token-123",
			wantErr: false,
		},
		{
			name:    "empty token",
			token:   "",
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := validateAPIToken(tt.token)
			if tt.wantErr {
				assert.Error(t, err, "validateAPITokenでエラーが期待されます")
			} else {
				assert.NoError(t, err, "validateAPITokenでエラーが発生しました")
			}
		})
	}
}
