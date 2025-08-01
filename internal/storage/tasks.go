package storage

import (
	"database/sql"
	"fmt"
	"time"

	"github.com/kyokomi/gotodoist/internal/api"
)

// InsertTask はタスクをローカルDBに挿入する
func (s *SQLiteDB) InsertTask(task api.Item) error {
	query := `
		INSERT OR REPLACE INTO tasks (
			id, user_id, project_id, section_id, parent_id, content, description,
			priority, child_order, day_order, is_collapsed, is_completed, is_deleted,
			assigned_by_uid, responsible_uid, sync_id,
			due_date, due_string, due_lang, due_is_recurring, due_timezone,
			added_at, completed_at, updated_at
		) VALUES (
			?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, strftime('%s', 'now')
		)
	`

	var dueDate, dueString, dueLang, dueTimezone sql.NullString
	var dueIsRecurring sql.NullBool
	if task.Due != nil {
		dueDate = sql.NullString{String: task.Due.Date, Valid: task.Due.Date != ""}
		dueString = sql.NullString{String: task.Due.String, Valid: task.Due.String != ""}
		dueLang = sql.NullString{String: task.Due.Lang, Valid: task.Due.Lang != ""}
		dueTimezone = sql.NullString{String: task.Due.Timezone, Valid: task.Due.Timezone != ""}
		dueIsRecurring = sql.NullBool{Bool: task.Due.IsRecurring, Valid: true}
	}

	var completedAt sql.NullInt64
	if task.DateCompleted != nil {
		completedAt = sql.NullInt64{Int64: task.DateCompleted.Unix(), Valid: true}
	}

	_, err := s.db.Exec(query,
		task.ID, task.UserID, task.ProjectID,
		nullString(task.SectionID), nullString(task.ParentID),
		task.Content, task.Description,
		task.Priority, task.ChildOrder, task.DayOrder,
		task.Collapsed, task.DateCompleted != nil, task.IsDeleted,
		nullString(task.AssignedByUID), nullString(task.ResponsibleUID),
		nullString(task.SyncID),
		dueDate, dueString, dueLang, dueIsRecurring, dueTimezone,
		task.DateAdded.Unix(), completedAt,
	)

	if err != nil {
		return fmt.Errorf("failed to insert task: %w", err)
	}

	// ラベルがある場合は関連テーブルにも挿入
	if len(task.Labels) > 0 {
		if err := s.insertTaskLabels(task.ID, task.Labels); err != nil {
			return fmt.Errorf("failed to insert task labels: %w", err)
		}
	}

	return nil
}

// GetTasks は全てのアクティブなタスクを取得する
func (s *SQLiteDB) GetTasks() ([]api.Item, error) {
	query := `
		SELECT 
			t.id, t.user_id, t.project_id, t.section_id, t.parent_id, 
			t.content, t.description, t.priority, t.child_order, t.day_order,
			t.is_collapsed, t.is_completed, t.is_deleted,
			t.assigned_by_uid, t.responsible_uid, t.sync_id,
			t.due_date, t.due_string, t.due_lang, t.due_is_recurring, t.due_timezone,
			t.added_at, t.completed_at
		FROM tasks t
		WHERE t.is_deleted = FALSE
		ORDER BY t.child_order, t.id
	`

	rows, err := s.db.Query(query)
	if err != nil {
		return nil, fmt.Errorf("failed to query tasks: %w", err)
	}
	defer func() {
		if err := rows.Close(); err != nil {
			fmt.Printf("Warning: failed to close rows: %v\n", err)
		}
	}()

	var tasks []api.Item
	for rows.Next() {
		task, err := s.scanTask(rows)
		if err != nil {
			return nil, fmt.Errorf("failed to scan task: %w", err)
		}

		// ラベルを取得
		labels, err := s.getTaskLabels(task.ID)
		if err != nil {
			return nil, fmt.Errorf("failed to get task labels: %w", err)
		}
		task.Labels = labels

		tasks = append(tasks, task)
	}

	return tasks, nil
}

// GetTasksByProject はプロジェクト指定でタスクを取得する
func (s *SQLiteDB) GetTasksByProject(projectID string) ([]api.Item, error) {
	query := `
		SELECT 
			t.id, t.user_id, t.project_id, t.section_id, t.parent_id, 
			t.content, t.description, t.priority, t.child_order, t.day_order,
			t.is_collapsed, t.is_completed, t.is_deleted,
			t.assigned_by_uid, t.responsible_uid, t.sync_id,
			t.due_date, t.due_string, t.due_lang, t.due_is_recurring, t.due_timezone,
			t.added_at, t.completed_at
		FROM tasks t
		WHERE t.project_id = ? AND t.is_deleted = FALSE
		ORDER BY t.child_order, t.id
	`

	rows, err := s.db.Query(query, projectID)
	if err != nil {
		return nil, fmt.Errorf("failed to query tasks by project: %w", err)
	}
	defer func() {
		if err := rows.Close(); err != nil {
			fmt.Printf("Warning: failed to close rows: %v\n", err)
		}
	}()

	var tasks []api.Item
	for rows.Next() {
		task, err := s.scanTask(rows)
		if err != nil {
			return nil, fmt.Errorf("failed to scan task: %w", err)
		}

		// ラベルを取得
		labels, err := s.getTaskLabels(task.ID)
		if err != nil {
			return nil, fmt.Errorf("failed to get task labels: %w", err)
		}
		task.Labels = labels

		tasks = append(tasks, task)
	}

	return tasks, nil
}

// DeleteTask はタスクを削除する（論理削除）
func (s *SQLiteDB) DeleteTask(taskID string) error {
	query := "UPDATE tasks SET is_deleted = TRUE, updated_at = strftime('%s', 'now') WHERE id = ?"
	_, err := s.db.Exec(query, taskID)
	if err != nil {
		return fmt.Errorf("failed to delete task: %w", err)
	}
	return nil
}

// scanTask は行からTaskオブジェクトをスキャンする
func (s *SQLiteDB) scanTask(row interface {
	Scan(dest ...interface{}) error
}) (api.Item, error) {
	var task api.Item
	var userID, sectionID, parentID, assignedByUID, responsibleUID, syncID sql.NullString
	var dueDate, dueString, dueLang, dueTimezone sql.NullString
	var dueIsRecurring sql.NullBool
	var addedAt, completedAt sql.NullInt64

	err := row.Scan(
		&task.ID, &userID, &task.ProjectID, &sectionID, &parentID,
		&task.Content, &task.Description, &task.Priority, &task.ChildOrder, &task.DayOrder,
		&task.Collapsed, &task.IsDeleted, &task.IsDeleted,
		&assignedByUID, &responsibleUID, &syncID,
		&dueDate, &dueString, &dueLang, &dueIsRecurring, &dueTimezone,
		&addedAt, &completedAt,
	)
	if err != nil {
		return task, err
	}

	// NULL値の処理
	task.UserID = userID.String
	task.SectionID = sectionID.String
	task.ParentID = parentID.String
	task.AssignedByUID = assignedByUID.String
	task.ResponsibleUID = responsibleUID.String
	task.SyncID = syncID.String

	// 日時の処理
	if addedAt.Valid {
		task.DateAdded = api.TodoistTime{Time: time.Unix(addedAt.Int64, 0)}
	}
	if completedAt.Valid {
		completedTime := api.TodoistTime{Time: time.Unix(completedAt.Int64, 0)}
		task.DateCompleted = &completedTime
	}

	// Due情報の処理
	if dueDate.Valid || dueString.Valid {
		task.Due = &api.Due{
			Date:        dueDate.String,
			String:      dueString.String,
			Lang:        dueLang.String,
			IsRecurring: dueIsRecurring.Bool,
			Timezone:    dueTimezone.String,
		}
	}

	return task, nil
}

// insertTaskLabels はタスクのラベルを挿入する
func (s *SQLiteDB) insertTaskLabels(taskID string, labels []string) error {
	// 既存のラベル関連を削除
	if _, err := s.db.Exec("DELETE FROM task_labels WHERE task_id = ?", taskID); err != nil {
		return err
	}

	// 新しいラベル関連を挿入
	for _, label := range labels {
		_, err := s.db.Exec(`
			INSERT OR IGNORE INTO task_labels (task_id, label_name) 
			VALUES (?, ?)
		`, taskID, label)
		if err != nil {
			return err
		}
	}

	return nil
}

// getTaskLabels はタスクのラベルを取得する
func (s *SQLiteDB) getTaskLabels(taskID string) ([]string, error) {
	rows, err := s.db.Query("SELECT label_name FROM task_labels WHERE task_id = ?", taskID)
	if err != nil {
		return nil, err
	}
	defer func() {
		if err := rows.Close(); err != nil {
			fmt.Printf("Warning: failed to close rows: %v\n", err)
		}
	}()

	var labels []string
	for rows.Next() {
		var label string
		if err := rows.Scan(&label); err != nil {
			return nil, err
		}
		labels = append(labels, label)
	}

	return labels, nil
}

// nullString はstring値をsql.NullStringに変換する
func nullString(s string) sql.NullString {
	return sql.NullString{String: s, Valid: s != ""}
}
