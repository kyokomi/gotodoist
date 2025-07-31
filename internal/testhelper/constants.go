// Package testhelper provides helper functions and constants for testing
package testhelper

// Common test constants
const (
	// Token constants
	TestAPIToken = "test-token"

	// Sync token constants
	TestSyncToken       = "test-sync-token"
	TestSyncTokenDelete = "test-sync-token-delete"
	TestSyncTokenClose  = "test-sync-token-close"

	// ID constants
	TestProjectID = "project-1"
	TestItemID    = "item-123"
	TestTaskID    = "task-123"

	// Color constants
	TestColorBlue = "blue"
	TestColorRed  = "red"

	// Common response templates
	SimpleSyncResponse = `{
		"sync_token": "%s",
		"full_sync": false
	}`

	DeleteSyncResponse = `{
		"sync_token": "test-sync-token-delete",
		"full_sync": false
	}`
)
