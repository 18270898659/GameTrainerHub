package repository

import (
	"context"
	"database/sql"
	"time"

	"GameTrainerHub/internal/database"
)

type SettingRepository struct {
	db *database.DB
}

func NewSettingRepository(db *database.DB) *SettingRepository {
	return &SettingRepository{db: db}
}

// Get 获取指定设置项
func (r *SettingRepository) Get(ctx context.Context, key string) (string, error) {
	var val string
	err := r.db.QueryRowContext(ctx, "SELECT value FROM app_settings WHERE key = ?", key).Scan(&val)
	if err == sql.ErrNoRows {
		return "", nil
	}
	return val, err
}

// Set 保存设置项（插入或替换）
func (r *SettingRepository) Set(ctx context.Context, key, val string) error {
	now := time.Now()
	_, err := r.db.ExecContext(ctx, `
		INSERT INTO app_settings (key, value, updated_at)
		VALUES (?, ?, ?)
		ON CONFLICT(key) DO UPDATE SET value = excluded.value, updated_at = excluded.updated_at
	`, key, val, now)
	return err
}

// GetAll 获取所有设置
func (r *SettingRepository) GetAll(ctx context.Context) (map[string]string, error) {
	rows, err := r.db.QueryContext(ctx, "SELECT key, value FROM app_settings")
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	settings := make(map[string]string)
	for rows.Next() {
		var k, v string
		if err := rows.Scan(&k, &v); err != nil {
			return nil, err
		}
		settings[k] = v
	}
	return settings, rows.Err()
}

