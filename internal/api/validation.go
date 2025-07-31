package api

import "fmt"

// validateCreateProjectRequest はCreateProjectRequestの検証を行う
func validateCreateProjectRequest(req *CreateProjectRequest) error {
	if req == nil {
		return fmt.Errorf("create project request is required")
	}
	if req.Name == "" {
		return fmt.Errorf("project name is required")
	}
	return nil
}

// validateUpdateProjectRequest はUpdateProjectRequestの検証を行う
func validateUpdateProjectRequest(req *UpdateProjectRequest) error {
	if req == nil {
		return fmt.Errorf("update project request is required")
	}
	return nil
}

// validateProjectID はプロジェクトIDの検証を行う
func validateProjectID(projectID string) error {
	if projectID == "" {
		return fmt.Errorf("project ID is required")
	}
	return nil
}

// validateCreateTaskRequest はCreateTaskRequestの検証を行う
func validateCreateTaskRequest(req *CreateTaskRequest) error {
	if req == nil {
		return fmt.Errorf("create task request is required")
	}
	if req.Content == "" {
		return fmt.Errorf("task content is required")
	}
	return nil
}

// validateUpdateTaskRequest はUpdateTaskRequestの検証を行う
func validateUpdateTaskRequest(req *UpdateTaskRequest) error {
	if req == nil {
		return fmt.Errorf("update task request is required")
	}
	return nil
}

// validateTaskID はタスクIDの検証を行う
func validateTaskID(taskID string) error {
	if taskID == "" {
		return fmt.Errorf("task ID is required")
	}
	return nil
}

// validateAPIToken はAPIトークンの検証を行う
func validateAPIToken(token string) error {
	if token == "" {
		return fmt.Errorf("API token is required")
	}
	return nil
}
