package storage

import (
	"database/sql"
	"fmt"

	"github.com/kyokomi/gotodoist/internal/api"
)

// InsertProject はプロジェクトをローカルDBに挿入する
func (s *SQLiteDB) InsertProject(project api.Project) error {
	query := `
		INSERT OR REPLACE INTO projects (
			id, name, color, parent_id, child_order, collapsed, shared,
			is_deleted, is_archived, is_favorite, inbox_project, team_inbox,
			sync_id, updated_at
		) VALUES (
			?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, strftime('%s', 'now')
		)
	`

	_, err := s.db.Exec(query,
		project.ID, project.Name, project.Color,
		nullString(project.ParentID),
		project.ChildOrder, project.Collapsed, project.Shared,
		project.IsDeleted, project.IsArchived, project.IsFavorite,
		project.InboxProject, project.TeamInbox,
		nullString(project.SyncID),
	)

	if err != nil {
		return fmt.Errorf("failed to insert project: %w", err)
	}

	return nil
}

// GetAllProjects は全てのアクティブなプロジェクトを取得する
func (s *SQLiteDB) GetAllProjects() ([]api.Project, error) {
	query := `
		SELECT 
			id, name, color, parent_id, child_order, collapsed, shared,
			is_deleted, is_archived, is_favorite, inbox_project, team_inbox, sync_id
		FROM projects
		WHERE is_deleted = FALSE
		ORDER BY child_order, name
	`

	rows, err := s.db.Query(query)
	if err != nil {
		return nil, fmt.Errorf("failed to query projects: %w", err)
	}
	defer rows.Close()

	var projects []api.Project
	for rows.Next() {
		project, err := s.scanProject(rows)
		if err != nil {
			return nil, fmt.Errorf("failed to scan project: %w", err)
		}
		projects = append(projects, project)
	}

	return projects, nil
}

// GetProjectByID はIDでプロジェクトを取得する
func (s *SQLiteDB) GetProjectByID(projectID string) (*api.Project, error) {
	query := `
		SELECT 
			id, name, color, parent_id, child_order, collapsed, shared,
			is_deleted, is_archived, is_favorite, inbox_project, team_inbox, sync_id
		FROM projects
		WHERE id = ? AND is_deleted = FALSE
	`

	row := s.db.QueryRow(query, projectID)
	project, err := s.scanProject(row)
	if err != nil {
		if err == sql.ErrNoRows {
			return nil, nil
		}
		return nil, fmt.Errorf("failed to get project by ID: %w", err)
	}

	return &project, nil
}

// FindProjectsByName は名前でプロジェクトを検索する（部分一致）
func (s *SQLiteDB) FindProjectsByName(name string) ([]api.Project, error) {
	query := `
		SELECT 
			id, name, color, parent_id, child_order, collapsed, shared,
			is_deleted, is_archived, is_favorite, inbox_project, team_inbox, sync_id
		FROM projects
		WHERE is_deleted = FALSE AND (
			LOWER(name) LIKE LOWER(?) OR 
			LOWER(name) = LOWER(?)
		)
		ORDER BY 
			CASE WHEN LOWER(name) = LOWER(?) THEN 0 ELSE 1 END,  -- 完全一致を優先
			child_order, name
	`

	searchPattern := "%" + name + "%"
	rows, err := s.db.Query(query, searchPattern, name, name)
	if err != nil {
		return nil, fmt.Errorf("failed to search projects by name: %w", err)
	}
	defer rows.Close()

	var projects []api.Project
	for rows.Next() {
		project, err := s.scanProject(rows)
		if err != nil {
			return nil, fmt.Errorf("failed to scan project: %w", err)
		}
		projects = append(projects, project)
	}

	return projects, nil
}

// DeleteProject はプロジェクトを削除する（論理削除）
func (s *SQLiteDB) DeleteProject(projectID string) error {
	query := "UPDATE projects SET is_deleted = TRUE, updated_at = strftime('%s', 'now') WHERE id = ?"
	_, err := s.db.Exec(query, projectID)
	if err != nil {
		return fmt.Errorf("failed to delete project: %w", err)
	}
	return nil
}

// scanProject は行からProjectオブジェクトをスキャンする
func (s *SQLiteDB) scanProject(row interface {
	Scan(dest ...interface{}) error
}) (api.Project, error) {
	var project api.Project
	var parentID, syncID sql.NullString

	err := row.Scan(
		&project.ID, &project.Name, &project.Color, &parentID,
		&project.ChildOrder, &project.Collapsed, &project.Shared,
		&project.IsDeleted, &project.IsArchived, &project.IsFavorite,
		&project.InboxProject, &project.TeamInbox, &syncID,
	)
	if err != nil {
		return project, err
	}

	// NULL値の処理
	project.ParentID = parentID.String
	project.SyncID = syncID.String

	return project, nil
}
