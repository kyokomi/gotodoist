-- gotodoist ローカルストレージ用SQLiteスキーマ
-- 
-- このファイルはローカルデータベースの構造を定義します
-- Todoist API v1のデータ構造に基づいて設計されています

-- プロジェクト
CREATE TABLE IF NOT EXISTS projects (
    id TEXT PRIMARY KEY,
    name TEXT NOT NULL,
    color TEXT,
    parent_id TEXT,
    child_order INTEGER DEFAULT 0,
    collapsed BOOLEAN DEFAULT FALSE,
    shared BOOLEAN DEFAULT FALSE,
    is_deleted BOOLEAN DEFAULT FALSE,
    is_archived BOOLEAN DEFAULT FALSE,
    is_favorite BOOLEAN DEFAULT FALSE,
    inbox_project BOOLEAN DEFAULT FALSE,
    team_inbox BOOLEAN DEFAULT FALSE,
    sync_id TEXT,
    created_at INTEGER DEFAULT (strftime('%s', 'now')),
    updated_at INTEGER DEFAULT (strftime('%s', 'now'))
);

-- セクション
CREATE TABLE IF NOT EXISTS sections (
    id TEXT PRIMARY KEY,
    name TEXT NOT NULL,
    project_id TEXT NOT NULL,
    section_order INTEGER DEFAULT 0,
    collapsed BOOLEAN DEFAULT FALSE,
    is_deleted BOOLEAN DEFAULT FALSE,
    sync_id TEXT,
    date_added INTEGER,
    date_archived INTEGER,
    created_at INTEGER DEFAULT (strftime('%s', 'now')),
    updated_at INTEGER DEFAULT (strftime('%s', 'now')),
    FOREIGN KEY (project_id) REFERENCES projects(id) ON DELETE CASCADE
);

-- タスク
CREATE TABLE IF NOT EXISTS tasks (
    id TEXT PRIMARY KEY,
    user_id TEXT,
    project_id TEXT NOT NULL,
    section_id TEXT,
    parent_id TEXT,
    content TEXT NOT NULL,
    description TEXT,
    priority INTEGER DEFAULT 1,
    child_order INTEGER DEFAULT 0,
    day_order INTEGER DEFAULT 0,
    is_collapsed BOOLEAN DEFAULT FALSE,
    is_completed BOOLEAN DEFAULT FALSE,
    is_deleted BOOLEAN DEFAULT FALSE,
    assigned_by_uid TEXT,
    responsible_uid TEXT,
    sync_id TEXT,
    -- 期限関連
    due_date TEXT,
    due_string TEXT,
    due_lang TEXT,
    due_is_recurring BOOLEAN DEFAULT FALSE,
    due_timezone TEXT,
    -- タイムスタンプ
    added_at INTEGER,
    completed_at INTEGER,
    created_at INTEGER DEFAULT (strftime('%s', 'now')),
    updated_at INTEGER DEFAULT (strftime('%s', 'now')),
    FOREIGN KEY (project_id) REFERENCES projects(id) ON DELETE CASCADE
    -- section_idとparent_idの外部キー制約を一時的にコメントアウト
    -- FOREIGN KEY (section_id) REFERENCES sections(id) ON DELETE SET NULL,
    -- FOREIGN KEY (parent_id) REFERENCES tasks(id) ON DELETE CASCADE
);

-- ラベル
CREATE TABLE IF NOT EXISTS labels (
    id TEXT PRIMARY KEY,
    name TEXT NOT NULL UNIQUE,
    color TEXT,
    item_order INTEGER DEFAULT 0,
    is_deleted BOOLEAN DEFAULT FALSE,
    is_favorite BOOLEAN DEFAULT FALSE,
    created_at INTEGER DEFAULT (strftime('%s', 'now')),
    updated_at INTEGER DEFAULT (strftime('%s', 'now'))
);

-- タスクとラベルの関連（多対多）
CREATE TABLE IF NOT EXISTS task_labels (
    task_id TEXT NOT NULL,
    label_name TEXT NOT NULL,
    created_at INTEGER DEFAULT (strftime('%s', 'now')),
    PRIMARY KEY (task_id, label_name),
    FOREIGN KEY (task_id) REFERENCES tasks(id) ON DELETE CASCADE
    -- label_nameはラベル文字列をそのまま保存（外部キー制約なし）
);

-- 同期状態管理
CREATE TABLE IF NOT EXISTS sync_state (
    key TEXT PRIMARY KEY,
    value TEXT NOT NULL,
    updated_at INTEGER DEFAULT (strftime('%s', 'now'))
);

-- パフォーマンス向上のためのインデックス
CREATE INDEX IF NOT EXISTS idx_tasks_project_id ON tasks(project_id);
CREATE INDEX IF NOT EXISTS idx_tasks_section_id ON tasks(section_id);
CREATE INDEX IF NOT EXISTS idx_tasks_parent_id ON tasks(parent_id);
CREATE INDEX IF NOT EXISTS idx_tasks_completed ON tasks(is_completed);
CREATE INDEX IF NOT EXISTS idx_tasks_deleted ON tasks(is_deleted);
CREATE INDEX IF NOT EXISTS idx_tasks_priority ON tasks(priority);
CREATE INDEX IF NOT EXISTS idx_sections_project_id ON sections(project_id);
CREATE INDEX IF NOT EXISTS idx_sections_deleted ON sections(is_deleted);
CREATE INDEX IF NOT EXISTS idx_projects_deleted ON projects(is_deleted);
CREATE INDEX IF NOT EXISTS idx_projects_archived ON projects(is_archived);

-- 初期データ
INSERT OR REPLACE INTO sync_state (key, value) VALUES 
    ('sync_token', '*'),
    ('last_sync_time', '0'),
    ('initial_sync_done', 'false'),
    ('schema_version', '1');