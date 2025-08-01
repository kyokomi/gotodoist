package storage

import (
	"database/sql"
	"fmt"
	"time"

	"github.com/kyokomi/gotodoist/internal/api"
)

// InsertSection はセクションをローカルDBに挿入する
func (s *SQLiteDB) InsertSection(section api.Section) error {
	query := `
		INSERT OR REPLACE INTO sections (
			id, name, project_id, section_order, collapsed, is_deleted,
			sync_id, date_added, date_archived, updated_at
		) VALUES (
			?, ?, ?, ?, ?, ?, ?, ?, ?, strftime('%s', 'now')
		)
	`

	var dateAdded, dateArchived sql.NullInt64
	if !section.DateAdded.IsZero() {
		dateAdded = sql.NullInt64{Int64: section.DateAdded.Unix(), Valid: true}
	}
	if section.DateArchived != nil && !section.DateArchived.IsZero() {
		dateArchived = sql.NullInt64{Int64: section.DateArchived.Unix(), Valid: true}
	}

	_, err := s.db.Exec(query,
		section.ID, section.Name, section.ProjectID,
		section.SectionOrder, section.Collapsed, section.IsDeleted,
		nullString(section.SyncID), dateAdded, dateArchived,
	)

	if err != nil {
		return fmt.Errorf("failed to insert section: %w", err)
	}

	return nil
}

// GetAllSections は全てのアクティブなセクションを取得する
func (s *SQLiteDB) GetAllSections() ([]api.Section, error) {
	query := `
		SELECT 
			id, name, project_id, section_order, collapsed, is_deleted,
			sync_id, date_added, date_archived
		FROM sections
		WHERE is_deleted = FALSE
		ORDER BY project_id, section_order, name
	`

	rows, err := s.db.Query(query)
	if err != nil {
		return nil, fmt.Errorf("failed to query sections: %w", err)
	}
	defer func() {
		if err := rows.Close(); err != nil {
			fmt.Printf("Warning: failed to close rows: %v\n", err)
		}
	}()

	var sections []api.Section
	for rows.Next() {
		section, err := s.scanSection(rows)
		if err != nil {
			return nil, fmt.Errorf("failed to scan section: %w", err)
		}
		sections = append(sections, section)
	}

	return sections, nil
}

// GetSectionsByProject はプロジェクト指定でセクションを取得する
func (s *SQLiteDB) GetSectionsByProject(projectID string) ([]api.Section, error) {
	query := `
		SELECT 
			id, name, project_id, section_order, collapsed, is_deleted,
			sync_id, date_added, date_archived
		FROM sections
		WHERE project_id = ? AND is_deleted = FALSE
		ORDER BY section_order, name
	`

	rows, err := s.db.Query(query, projectID)
	if err != nil {
		return nil, fmt.Errorf("failed to query sections by project: %w", err)
	}
	defer func() {
		if err := rows.Close(); err != nil {
			fmt.Printf("Warning: failed to close rows: %v\n", err)
		}
	}()

	var sections []api.Section
	for rows.Next() {
		section, err := s.scanSection(rows)
		if err != nil {
			return nil, fmt.Errorf("failed to scan section: %w", err)
		}
		sections = append(sections, section)
	}

	return sections, nil
}

// GetSectionByID はIDでセクションを取得する
func (s *SQLiteDB) GetSectionByID(sectionID string) (*api.Section, error) {
	query := `
		SELECT 
			id, name, project_id, section_order, collapsed, is_deleted,
			sync_id, date_added, date_archived
		FROM sections
		WHERE id = ? AND is_deleted = FALSE
	`

	row := s.db.QueryRow(query, sectionID)
	section, err := s.scanSection(row)
	if err != nil {
		if err == sql.ErrNoRows {
			return nil, nil
		}
		return nil, fmt.Errorf("failed to get section by ID: %w", err)
	}

	return &section, nil
}

// DeleteSection はセクションを削除する（論理削除）
func (s *SQLiteDB) DeleteSection(sectionID string) error {
	query := "UPDATE sections SET is_deleted = TRUE, updated_at = strftime('%s', 'now') WHERE id = ?"
	_, err := s.db.Exec(query, sectionID)
	if err != nil {
		return fmt.Errorf("failed to delete section: %w", err)
	}
	return nil
}

// scanSection は行からSectionオブジェクトをスキャンする
func (s *SQLiteDB) scanSection(row interface {
	Scan(dest ...interface{}) error
}) (api.Section, error) {
	var section api.Section
	var syncID sql.NullString
	var dateAdded, dateArchived sql.NullInt64

	err := row.Scan(
		&section.ID, &section.Name, &section.ProjectID,
		&section.SectionOrder, &section.Collapsed, &section.IsDeleted,
		&syncID, &dateAdded, &dateArchived,
	)
	if err != nil {
		return section, err
	}

	// NULL値の処理
	section.SyncID = syncID.String

	// 日時の処理
	if dateAdded.Valid {
		section.DateAdded = api.TodoistTime{Time: time.Unix(dateAdded.Int64, 0)}
	}
	if dateArchived.Valid {
		archivedTime := api.TodoistTime{Time: time.Unix(dateArchived.Int64, 0)}
		section.DateArchived = &archivedTime
	}

	return section, nil
}
