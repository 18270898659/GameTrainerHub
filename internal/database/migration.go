package database

import (
	"context"
	"database/sql"
	"fmt"

	"GameTrainerHub/internal/logger"
)

type Migration struct {
	Version int
	Name    string
	SQL     string
}

// migrations 注册所有有序数据库版本迁移
var migrations = []Migration{
	{
		Version: 1,
		Name:    "001_initial_schema",
		SQL: `
-- 1. 资源源表
CREATE TABLE IF NOT EXISTS sources (
    id TEXT PRIMARY KEY,
    name TEXT NOT NULL,
    base_url TEXT NOT NULL,
    is_enabled INTEGER NOT NULL DEFAULT 1,
    rate_limit_ms INTEGER NOT NULL DEFAULT 500,
    created_at DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP
);

-- 插入默认 FLiNG 资源源
INSERT OR IGNORE INTO sources (id, name, base_url, is_enabled, rate_limit_ms)
VALUES ('fling', 'FLiNG Trainer', 'https://flingtrainer.com', 1, 500);

-- 2. 游戏基础信息表
CREATE TABLE IF NOT EXISTS games (
    id INTEGER PRIMARY KEY AUTOINCREMENT,
    slug TEXT NOT NULL UNIQUE,
    name_en TEXT NOT NULL,
    name_cn TEXT NOT NULL DEFAULT '',
    cover_url TEXT NOT NULL DEFAULT '',
    steam_app_id INTEGER DEFAULT 0,
    install_path TEXT DEFAULT '',
    created_at DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP,
    updated_at DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP
);
CREATE INDEX IF NOT EXISTS idx_games_slug ON games(slug);
CREATE INDEX IF NOT EXISTS idx_games_steam_id ON games(steam_app_id);

-- 3. 游戏别名多语言表
CREATE TABLE IF NOT EXISTS game_aliases (
    id INTEGER PRIMARY KEY AUTOINCREMENT,
    game_id INTEGER NOT NULL REFERENCES games(id) ON DELETE CASCADE,
    alias TEXT NOT NULL,
    lang TEXT NOT NULL DEFAULT 'zh',
    created_at DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP
);
CREATE INDEX IF NOT EXISTS idx_aliases_game_id ON game_aliases(game_id);
CREATE INDEX IF NOT EXISTS idx_aliases_alias ON game_aliases(alias);

-- 4. 修改器资源主表
CREATE TABLE IF NOT EXISTS trainers (
    id INTEGER PRIMARY KEY AUTOINCREMENT,
    game_id INTEGER NOT NULL REFERENCES games(id) ON DELETE CASCADE,
    source_id TEXT NOT NULL REFERENCES sources(id),
    external_id TEXT NOT NULL,
    title_en TEXT NOT NULL,
    title_cn TEXT NOT NULL DEFAULT '',
    detail_url TEXT NOT NULL,
    author TEXT NOT NULL DEFAULT '',
    latest_version TEXT NOT NULL DEFAULT '',
    published_at DATETIME,
    updated_at DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP,
    is_deleted INTEGER NOT NULL DEFAULT 0,
    UNIQUE(source_id, external_id)
);
CREATE INDEX IF NOT EXISTS idx_trainers_game_id ON trainers(game_id);

-- 5. 修改器版本与下载表
CREATE TABLE IF NOT EXISTS trainer_versions (
    id INTEGER PRIMARY KEY AUTOINCREMENT,
    trainer_id INTEGER NOT NULL REFERENCES trainers(id) ON DELETE CASCADE,
    version_name TEXT NOT NULL,
    target_game_version TEXT NOT NULL DEFAULT '',
    features_text TEXT NOT NULL DEFAULT '',
    download_url TEXT NOT NULL,
    file_name TEXT NOT NULL,
    file_size INTEGER NOT NULL DEFAULT 0,
    sha256 TEXT NOT NULL DEFAULT '',
    published_at DATETIME,
    created_at DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP,
    UNIQUE(trainer_id, version_name)
);
CREATE INDEX IF NOT EXISTS idx_versions_trainer_id ON trainer_versions(trainer_id);

-- 6. 下载任务表
CREATE TABLE IF NOT EXISTS download_tasks (
    id INTEGER PRIMARY KEY AUTOINCREMENT,
    version_id INTEGER NOT NULL REFERENCES trainer_versions(id),
    file_name TEXT NOT NULL,
    save_path TEXT NOT NULL,
    total_bytes INTEGER NOT NULL DEFAULT 0,
    downloaded_bytes INTEGER NOT NULL DEFAULT 0,
    status TEXT NOT NULL,
    error_msg TEXT DEFAULT '',
    retries INTEGER NOT NULL DEFAULT 0,
    created_at DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP,
    updated_at DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP
);
CREATE INDEX IF NOT EXISTS idx_download_tasks_status ON download_tasks(status);

-- 7. 下载历史表
CREATE TABLE IF NOT EXISTS download_history (
    id INTEGER PRIMARY KEY AUTOINCREMENT,
    task_id INTEGER NOT NULL,
    trainer_id INTEGER NOT NULL,
    version_id INTEGER NOT NULL,
    game_name TEXT NOT NULL,
    file_name TEXT NOT NULL,
    file_size INTEGER NOT NULL DEFAULT 0,
    sha256 TEXT NOT NULL DEFAULT '',
    save_path TEXT NOT NULL,
    completed_at DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP
);

-- 8. 本地修改器库
CREATE TABLE IF NOT EXISTS local_trainers (
    id INTEGER PRIMARY KEY AUTOINCREMENT,
    game_id INTEGER NOT NULL REFERENCES games(id) ON DELETE CASCADE,
    version_id INTEGER REFERENCES trainer_versions(id) ON DELETE SET NULL,
    file_path TEXT NOT NULL UNIQUE,
    file_size INTEGER NOT NULL,
    sha256 TEXT NOT NULL,
    source_name TEXT NOT NULL DEFAULT 'FLiNG',
    is_custom INTEGER NOT NULL DEFAULT 0,
    last_run_at DATETIME,
    run_count INTEGER NOT NULL DEFAULT 0,
    created_at DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP
);
CREATE INDEX IF NOT EXISTS idx_local_trainers_game ON local_trainers(game_id);

-- 9. 游戏收藏表
CREATE TABLE IF NOT EXISTS favorites (
    id INTEGER PRIMARY KEY AUTOINCREMENT,
    game_id INTEGER NOT NULL UNIQUE REFERENCES games(id) ON DELETE CASCADE,
    created_at DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP
);

-- 10. 同步记录表
CREATE TABLE IF NOT EXISTS sync_records (
    id INTEGER PRIMARY KEY AUTOINCREMENT,
    source_id TEXT NOT NULL REFERENCES sources(id),
    sync_type TEXT NOT NULL,
    status TEXT NOT NULL,
    items_added INTEGER NOT NULL DEFAULT 0,
    items_updated INTEGER NOT NULL DEFAULT 0,
    error_msg TEXT DEFAULT '',
    started_at DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP,
    finished_at DATETIME
);

-- 11. 客户端设置表
CREATE TABLE IF NOT EXISTS app_settings (
    key TEXT PRIMARY KEY,
    value TEXT NOT NULL,
    updated_at DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP
);

-- 12. 全文检索 FTS5 虚拟表
CREATE VIRTUAL TABLE IF NOT EXISTS game_search_fts USING fts5(
    game_id UNINDEXED,
    name_en,
    name_cn,
    aliases,
    pinyin,
    pinyin_initials,
    tokenize='unicode61'
);
`,
	},
	{
		Version: 2,
		Name:    "002_phase2_extensions",
		SQL: `
-- 13. 搜索历史表
CREATE TABLE IF NOT EXISTS search_history (
    id INTEGER PRIMARY KEY AUTOINCREMENT,
    keyword TEXT NOT NULL UNIQUE,
    created_at DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP
);
CREATE INDEX IF NOT EXISTS idx_search_history_created ON search_history(created_at DESC);

-- 14. 最近访问游戏表
CREATE TABLE IF NOT EXISTS visited_games (
    id INTEGER PRIMARY KEY AUTOINCREMENT,
    game_id INTEGER NOT NULL UNIQUE REFERENCES games(id) ON DELETE CASCADE,
    visited_at DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP
);
CREATE INDEX IF NOT EXISTS idx_visited_games_time ON visited_games(visited_at DESC);
`,
	},
	{
		Version: 3,
		Name:    "003_download_task_metadata",
		SQL: `
-- 给 download_tasks 与 download_history 补充元数据字段，避免跨表关联丢失或页面显示为组件名称
ALTER TABLE download_tasks ADD COLUMN game_name TEXT NOT NULL DEFAULT '';
ALTER TABLE download_tasks ADD COLUMN trainer_name TEXT NOT NULL DEFAULT '';
ALTER TABLE download_tasks ADD COLUMN version_name TEXT NOT NULL DEFAULT '';
ALTER TABLE download_history ADD COLUMN trainer_name TEXT NOT NULL DEFAULT '';
ALTER TABLE download_history ADD COLUMN version_name TEXT NOT NULL DEFAULT '';
`,
	},
}

// Migrate 执行数据库版本迁移
func (db *DB) Migrate() error {
	ctx := context.Background()

	// 确保迁移记录表存在
	_, err := db.ExecContext(ctx, `
		CREATE TABLE IF NOT EXISTS schema_migrations (
			version INTEGER PRIMARY KEY,
			name TEXT NOT NULL,
			applied_at DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP
		);
	`)
	if err != nil {
		return fmt.Errorf("failed to create schema_migrations table: %w", err)
	}

	for _, m := range migrations {
		var exists bool
		err := db.QueryRowContext(ctx, "SELECT EXISTS(SELECT 1 FROM schema_migrations WHERE version = ?)", m.Version).Scan(&exists)
		if err != nil && err != sql.ErrNoRows {
			return fmt.Errorf("failed to check migration version %d: %w", m.Version, err)
		}

		if !exists {
			logger.Info("Applying database migration", "version", m.Version, "name", m.Name)
			err := db.WithTx(ctx, func(tx *sql.Tx) error {
				if _, execErr := tx.ExecContext(ctx, m.SQL); execErr != nil {
					return fmt.Errorf("failed to exec migration sql: %w", execErr)
				}
				if _, recordErr := tx.ExecContext(ctx, "INSERT INTO schema_migrations (version, name) VALUES (?, ?)", m.Version, m.Name); recordErr != nil {
					return fmt.Errorf("failed to record migration: %w", recordErr)
				}
				return nil
			})
			if err != nil {
				return fmt.Errorf("migration %d failed: %w", m.Version, err)
			}
			logger.Info("Database migration applied successfully", "version", m.Version, "name", m.Name)
		}
	}

	return nil
}

// GetAppliedMigrations 获取已执行的迁移版本列表
func (db *DB) GetAppliedMigrations(ctx context.Context) ([]int, error) {
	rows, err := db.QueryContext(ctx, "SELECT version FROM schema_migrations ORDER BY version ASC")
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var versions []int
	for rows.Next() {
		var v int
		if err := rows.Scan(&v); err != nil {
			return nil, err
		}
		versions = append(versions, v)
	}
	return versions, rows.Err()
}

