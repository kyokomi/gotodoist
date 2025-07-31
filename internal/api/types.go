package api

import (
	"time"
)

// SyncRequest はSync APIのリクエスト構造体
type SyncRequest struct {
	SyncToken     string    `json:"sync_token"`
	ResourceTypes []string  `json:"resource_types"`
	Commands      []Command `json:"commands,omitempty"`
}

// SyncResponse はSync APIのレスポンス構造体
type SyncResponse struct {
	SyncToken     string            `json:"sync_token"`
	FullSync      bool              `json:"full_sync"`
	Items         []Item            `json:"items,omitempty"`
	Projects      []Project         `json:"projects,omitempty"`
	Sections      []Section         `json:"sections,omitempty"`
	Labels        []Label           `json:"labels,omitempty"`
	Notes         []Note            `json:"notes,omitempty"`
	TempIDMapping map[string]string `json:"temp_id_mapping,omitempty"`
	SyncStatus    map[string]interface{} `json:"sync_status,omitempty"`
}

// Command はSync APIのコマンド構造体
type Command struct {
	Type   string                 `json:"type"`
	UUID   string                 `json:"uuid"`
	TempID string                 `json:"temp_id,omitempty"`
	Args   map[string]interface{} `json:"args"`
}

// Item はTodoistのタスク（アイテム）を表す
type Item struct {
	ID            string     `json:"id"`
	UserID        string     `json:"user_id"`
	ProjectID     string     `json:"project_id"`
	SectionID     string     `json:"section_id,omitempty"`
	Content       string     `json:"content"`
	Description   string     `json:"description,omitempty"`
	Priority      int        `json:"priority"`
	ParentID      string     `json:"parent_id,omitempty"`
	ChildOrder    int        `json:"child_order"`
	DayOrder      int        `json:"day_order"`
	Collapsed     bool       `json:"collapsed"`
	Labels        []string   `json:"labels,omitempty"`
	AssignedBy    string     `json:"assigned_by,omitempty"`
	Responsible   string     `json:"responsible,omitempty"`
	DateAdded     time.Time  `json:"date_added"`
	DateCompleted *time.Time `json:"date_completed,omitempty"`
	IsDeleted     bool       `json:"is_deleted"`
	SyncID        string     `json:"sync_id,omitempty"`
	Due           *Due       `json:"due,omitempty"`
}

// Due はタスクの期限を表す
type Due struct {
	Date        string `json:"date"`
	String      string `json:"string"`
	Lang        string `json:"lang,omitempty"`
	IsRecurring bool   `json:"is_recurring"`
	Timezone    string `json:"timezone,omitempty"`
}

// Project はTodoistのプロジェクトを表す
type Project struct {
	ID           string `json:"id"`
	Name         string `json:"name"`
	Color        string `json:"color"`
	ParentID     string `json:"parent_id,omitempty"`
	ChildOrder   int    `json:"child_order"`
	Collapsed    bool   `json:"collapsed"`
	Shared       bool   `json:"shared"`
	IsDeleted    bool   `json:"is_deleted"`
	IsArchived   bool   `json:"is_archived"`
	IsFavorite   bool   `json:"is_favorite"`
	SyncID       string `json:"sync_id,omitempty"`
	InboxProject bool   `json:"inbox_project"`
	TeamInbox    bool   `json:"team_inbox"`
}

// Section はTodoistのセクションを表す
type Section struct {
	ID           string     `json:"id"`
	Name         string     `json:"name"`
	ProjectID    string     `json:"project_id"`
	SectionOrder int        `json:"section_order"`
	Collapsed    bool       `json:"collapsed"`
	SyncID       string     `json:"sync_id,omitempty"`
	IsDeleted    bool       `json:"is_deleted"`
	DateAdded    time.Time  `json:"date_added"`
	DateArchived *time.Time `json:"date_archived,omitempty"`
}

// Label はTodoistのラベルを表す
type Label struct {
	ID         string `json:"id"`
	Name       string `json:"name"`
	Color      string `json:"color"`
	ItemOrder  int    `json:"item_order"`
	IsDeleted  bool   `json:"is_deleted"`
	IsFavorite bool   `json:"is_favorite"`
}

// Note はTodoistのコメント・ノートを表す
type Note struct {
	ID             string                 `json:"id"`
	PostedUID      string                 `json:"posted_uid"`
	ProjectID      string                 `json:"project_id,omitempty"`
	ItemID         string                 `json:"item_id,omitempty"`
	Content        string                 `json:"content"`
	FileAttachment map[string]interface{} `json:"file_attachment,omitempty"`
	UidsToNotify   []string               `json:"uids_to_notify,omitempty"`
	IsDeleted      bool                   `json:"is_deleted"`
	Posted         time.Time              `json:"posted"`
	Reactions      map[string]interface{} `json:"reactions,omitempty"`
}

// ResourceTypes は同期するリソースタイプの定数
const (
	ResourceAll       = "all"
	ResourceItems     = "items"
	ResourceProjects  = "projects"
	ResourceSections  = "sections"
	ResourceLabels    = "labels"
	ResourceNotes     = "notes"
	ResourceFilters   = "filters"
	ResourceReminders = "reminders"
)

// Command types for Sync API
const (
	CommandItemAdd        = "item_add"
	CommandItemUpdate     = "item_update"
	CommandItemDelete     = "item_delete"
	CommandItemComplete   = "item_complete"
	CommandItemUncomplete = "item_uncomplete"
	CommandItemMove       = "item_move"

	CommandProjectAdd       = "project_add"
	CommandProjectUpdate    = "project_update"
	CommandProjectDelete    = "project_delete"
	CommandProjectArchive   = "project_archive"
	CommandProjectUnarchive = "project_unarchive"

	CommandSectionAdd     = "section_add"
	CommandSectionUpdate  = "section_update"
	CommandSectionDelete  = "section_delete"
	CommandSectionMove    = "section_move"
	CommandSectionArchive = "section_archive"

	CommandLabelAdd    = "label_add"
	CommandLabelUpdate = "label_update"
	CommandLabelDelete = "label_delete"

	CommandNoteAdd    = "note_add"
	CommandNoteUpdate = "note_update"
	CommandNoteDelete = "note_delete"
)
