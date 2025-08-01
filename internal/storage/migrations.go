package storage

import (
	"fmt"
	"strconv"
)

const (
	// CurrentSchemaVersion は現在のスキーマバージョン
	CurrentSchemaVersion = 1
)

// Migration はデータベースマイグレーションを表す
type Migration struct {
	Version int
	Name    string
	SQL     string
}

// migrations は実行可能なマイグレーション一覧
var migrations = []Migration{
	{
		Version: 1,
		Name:    "initial_schema",
		SQL: `
-- 初期スキーマは既にschema.sqlで定義済み
-- この時点では何もしない
`,
	},
	// 将来のマイグレーションはここに追加
	// {
	//     Version: 2,
	//     Name:    "add_task_tags",
	//     SQL: `ALTER TABLE tasks ADD COLUMN tags TEXT;`,
	// },
}

// RunMigrations はデータベースマイグレーションを実行する
func (s *SQLiteDB) RunMigrations() error {
	// 現在のスキーマバージョンを取得
	currentVersion, err := s.getCurrentSchemaVersion()
	if err != nil {
		return fmt.Errorf("failed to get current schema version: %w", err)
	}

	// 必要なマイグレーションを実行
	for _, migration := range migrations {
		if migration.Version > currentVersion {
			if err := s.runMigration(migration); err != nil {
				return fmt.Errorf("failed to run migration %d (%s): %w",
					migration.Version, migration.Name, err)
			}
		}
	}

	return nil
}

// getCurrentSchemaVersion は現在のスキーマバージョンを取得する
func (s *SQLiteDB) getCurrentSchemaVersion() (int, error) {
	var versionStr string
	err := s.db.QueryRow("SELECT value FROM sync_state WHERE key = 'schema_version'").Scan(&versionStr)
	if err != nil {
		// schema_versionが存在しない場合は0とする
		return 0, nil
	}

	version, err := strconv.Atoi(versionStr)
	if err != nil {
		return 0, fmt.Errorf("invalid schema version: %s", versionStr)
	}

	return version, nil
}

// runMigration は単一のマイグレーションを実行する
func (s *SQLiteDB) runMigration(migration Migration) error {
	// トランザクション開始
	tx, err := s.db.Begin()
	if err != nil {
		return fmt.Errorf("failed to begin transaction: %w", err)
	}
	defer func() {
		if err != nil {
			tx.Rollback()
		}
	}()

	// マイグレーションSQLを実行
	if migration.SQL != "" {
		if _, err = tx.Exec(migration.SQL); err != nil {
			return fmt.Errorf("failed to execute migration SQL: %w", err)
		}
	}

	// スキーマバージョンを更新
	_, err = tx.Exec(`
		INSERT OR REPLACE INTO sync_state (key, value, updated_at) 
		VALUES ('schema_version', ?, strftime('%s', 'now'))
	`, migration.Version)
	if err != nil {
		return fmt.Errorf("failed to update schema version: %w", err)
	}

	// トランザクションコミット
	if err = tx.Commit(); err != nil {
		return fmt.Errorf("failed to commit migration transaction: %w", err)
	}

	fmt.Printf("Migration %d (%s) completed successfully\n", migration.Version, migration.Name)
	return nil
}

// GetSchemaVersion は現在のスキーマバージョンを返す（外部向け）
func (s *SQLiteDB) GetSchemaVersion() (int, error) {
	return s.getCurrentSchemaVersion()
}

// ResetDatabase はデータベースを初期化する（開発/テスト用）
func (s *SQLiteDB) ResetDatabase() error {
	tables := []string{
		"task_labels",
		"tasks",
		"sections",
		"projects",
		"labels",
		"sync_state",
	}

	// 全テーブルを削除
	for _, table := range tables {
		query := fmt.Sprintf("DROP TABLE IF EXISTS %s", table)
		if _, err := s.db.Exec(query); err != nil {
			return fmt.Errorf("failed to drop table %s: %w", table, err)
		}
	}

	// スキーマを再初期化
	if err := s.initializeSchema(); err != nil {
		return fmt.Errorf("failed to reinitialize schema: %w", err)
	}

	fmt.Println("Database reset completed")
	return nil
}
