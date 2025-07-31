package api

import "testing"

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
			if (err != nil) != tt.wantErr {
				t.Errorf("validateCreateProjectRequest() error = %v, wantErr %v", err, tt.wantErr)
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
			if (err != nil) != tt.wantErr {
				t.Errorf("validateUpdateProjectRequest() error = %v, wantErr %v", err, tt.wantErr)
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
			if (err != nil) != tt.wantErr {
				t.Errorf("validateProjectID() error = %v, wantErr %v", err, tt.wantErr)
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
			if (err != nil) != tt.wantErr {
				t.Errorf("validateCreateTaskRequest() error = %v, wantErr %v", err, tt.wantErr)
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
			if (err != nil) != tt.wantErr {
				t.Errorf("validateUpdateTaskRequest() error = %v, wantErr %v", err, tt.wantErr)
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
			if (err != nil) != tt.wantErr {
				t.Errorf("validateTaskID() error = %v, wantErr %v", err, tt.wantErr)
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
			if (err != nil) != tt.wantErr {
				t.Errorf("validateAPIToken() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}
