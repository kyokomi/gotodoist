// Package storage はローカルデータストレージの実装を提供する
package storage

import (
	"database/sql"
	"embed"
	"fmt"
	"os"
	"path/filepath"
	"time"

	_ "github.com/mattn/go-sqlite3" // SQLite driver
)

//go:embed schema.sql
var schemaSQL embed.FS

// SQLiteDB はSQLiteデータベースのラッパー
type SQLiteDB struct {
	db *sql.DB
}

// NewSQLiteDB は新しいSQLiteDBインスタンスを作成する
func NewSQLiteDB(dbPath string) (*SQLiteDB, error) {
	// データベースファイルのディレクトリを作成
	if err := os.MkdirAll(filepath.Dir(dbPath), 0750); err != nil {
		return nil, fmt.Errorf("failed to create database directory: %w", err)
	}

	// SQLite接続
	db, err := sql.Open("sqlite3", dbPath+"?_foreign_keys=on&_journal_mode=WAL")
	if err != nil {
		return nil, fmt.Errorf("failed to open database: %w", err)
	}

	// 接続テスト
	if err := db.Ping(); err != nil {
		return nil, fmt.Errorf("failed to ping database: %w", err)
	}

	sqliteDB := &SQLiteDB{db: db}

	// スキーマ初期化
	if err := sqliteDB.initializeSchema(); err != nil {
		return nil, fmt.Errorf("failed to initialize schema: %w", err)
	}

	return sqliteDB, nil
}

// Close はデータベース接続を閉じる
func (s *SQLiteDB) Close() error {
	return s.db.Close()
}

// initializeSchema はデータベーススキーマを初期化する
func (s *SQLiteDB) initializeSchema() error {
	// 埋め込まれたスキーマファイルを読み込み
	schemaContent, err := schemaSQL.ReadFile("schema.sql")
	if err != nil {
		return fmt.Errorf("failed to read schema.sql: %w", err)
	}

	// スキーマを実行
	if _, err := s.db.Exec(string(schemaContent)); err != nil {
		return fmt.Errorf("failed to execute schema: %w", err)
	}

	return nil
}

// GetSyncToken は現在の同期トークンを取得する
func (s *SQLiteDB) GetSyncToken() (string, error) {
	var token string
	err := s.db.QueryRow("SELECT value FROM sync_state WHERE key = 'sync_token'").Scan(&token)
	if err != nil {
		return "*", err // デフォルトは全同期
	}
	return token, nil
}

// SetSyncToken は同期トークンを設定する
func (s *SQLiteDB) SetSyncToken(token string) error {
	_, err := s.db.Exec(`
		INSERT OR REPLACE INTO sync_state (key, value, updated_at) 
		VALUES ('sync_token', ?, strftime('%s', 'now'))
	`, token)
	return err
}

// GetLastSyncTime は最後の同期時刻を取得する
func (s *SQLiteDB) GetLastSyncTime() (time.Time, error) {
	var timestamp int64
	err := s.db.QueryRow("SELECT value FROM sync_state WHERE key = 'last_sync_time'").Scan(&timestamp)
	if err != nil {
		return time.Time{}, err
	}
	return time.Unix(timestamp, 0), nil
}

// SetLastSyncTime は最後の同期時刻を設定する
func (s *SQLiteDB) SetLastSyncTime(t time.Time) error {
	_, err := s.db.Exec(`
		INSERT OR REPLACE INTO sync_state (key, value, updated_at) 
		VALUES ('last_sync_time', ?, strftime('%s', 'now'))
	`, t.Unix())
	return err
}

// IsInitialSyncDone は初期同期が完了しているかチェックする
func (s *SQLiteDB) IsInitialSyncDone() (bool, error) {
	var value string
	err := s.db.QueryRow("SELECT value FROM sync_state WHERE key = 'initial_sync_done'").Scan(&value)
	if err != nil {
		return false, err
	}
	return value == "true", nil
}

// SetInitialSyncDone は初期同期完了フラグを設定する
func (s *SQLiteDB) SetInitialSyncDone(done bool) error {
	value := "false"
	if done {
		value = "true"
	}
	_, err := s.db.Exec(`
		INSERT OR REPLACE INTO sync_state (key, value, updated_at) 
		VALUES ('initial_sync_done', ?, strftime('%s', 'now'))
	`, value)
	return err
}

// BeginTx はトランザクションを開始する
func (s *SQLiteDB) BeginTx() (*sql.Tx, error) {
	return s.db.Begin()
}

// GetDB は内部のsql.DBインスタンスを返す（テスト用）
func (s *SQLiteDB) GetDB() *sql.DB {
	return s.db
}

// ResetAllData はローカルストレージのすべてのデータを削除する
func (s *SQLiteDB) ResetAllData() error {
	// トランザクション内ですべてのテーブルをクリア
	tx, err := s.BeginTx()
	if err != nil {
		return fmt.Errorf("failed to begin transaction: %w", err)
	}
	defer func() {
		if err != nil {
			if rollbackErr := tx.Rollback(); rollbackErr != nil {
				// ロールバックのエラーはログに記録するが、元のエラーを優先
				fmt.Printf("Failed to rollback transaction: %v\n", rollbackErr)
			}
		}
	}()

	// 削除クエリのリスト
	queries := []string{
		"DELETE FROM tasks",
		"DELETE FROM projects",
		"DELETE FROM sections",
		"DELETE FROM sync_state",
	}

	// 各テーブルをクリア
	for _, query := range queries {
		if _, err := tx.Exec(query); err != nil {
			return fmt.Errorf("failed to execute query '%s': %w", query, err)
		}
	}

	if err := tx.Commit(); err != nil {
		return fmt.Errorf("failed to commit transaction: %w", err)
	}

	// VACUUMはトランザクション外で実行
	if _, err := s.db.Exec("VACUUM"); err != nil {
		return fmt.Errorf("failed to vacuum database: %w", err)
	}

	return nil
}
