package repository

import (
	"context"
	"database/sql"
	"fmt"
	"time"

	"GameTrainerHub/internal/database"
	"GameTrainerHub/internal/model"
)

type LocalTrainerRepository struct {
	db *database.DB
}

func NewLocalTrainerRepository(db *database.DB) *LocalTrainerRepository {
	return &LocalTrainerRepository{db: db}
}

// Create 登记本地修改器
func (r *LocalTrainerRepository) Create(ctx context.Context, lt *model.LocalTrainer) error {
	now := time.Now()
	res, err := r.db.ExecContext(ctx, `
		INSERT INTO local_trainers (game_id, version_id, file_path, file_size, sha256, source_name, is_custom, run_count, created_at)
		VALUES (?, ?, ?, ?, ?, ?, ?, 0, ?)
	`, lt.GameID, lt.VersionID, lt.FilePath, lt.FileSize, lt.SHA256, lt.SourceName, lt.IsCustom, now)
	if err != nil {
		return fmt.Errorf("failed to insert local trainer: %w", err)
	}

	id, err := res.LastInsertId()
	if err != nil {
		return err
	}
	lt.ID = id
	lt.CreatedAt = now
	return nil
}

// GetByFilePath 通过路径查询本地修改器
func (r *LocalTrainerRepository) GetByFilePath(ctx context.Context, filePath string) (*model.LocalTrainer, error) {
	row := r.db.QueryRowContext(ctx, `
		SELECT id, game_id, version_id, file_path, file_size, sha256, source_name, is_custom, last_run_at, run_count, created_at
		FROM local_trainers
		WHERE file_path = ?
	`, filePath)

	var lt model.LocalTrainer
	var lastRun sql.NullTime
	err := row.Scan(&lt.ID, &lt.GameID, &lt.VersionID, &lt.FilePath, &lt.FileSize, &lt.SHA256, &lt.SourceName, &lt.IsCustom, &lastRun, &lt.RunCount, &lt.CreatedAt)
	if err != nil {
		return nil, err
	}
	if lastRun.Valid {
		lt.LastRunAt = &lastRun.Time
	}
	return &lt, nil
}

// ListByGameID 查询某游戏下的所有本地修改器
func (r *LocalTrainerRepository) ListByGameID(ctx context.Context, gameID int64) ([]*model.LocalTrainer, error) {
	rows, err := r.db.QueryContext(ctx, `
		SELECT id, game_id, version_id, file_path, file_size, sha256, source_name, is_custom, last_run_at, run_count, created_at
		FROM local_trainers
		WHERE game_id = ?
		ORDER BY created_at DESC
	`, gameID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var list []*model.LocalTrainer
	for rows.Next() {
		var lt model.LocalTrainer
		var lastRun sql.NullTime
		if err := rows.Scan(&lt.ID, &lt.GameID, &lt.VersionID, &lt.FilePath, &lt.FileSize, &lt.SHA256, &lt.SourceName, &lt.IsCustom, &lastRun, &lt.RunCount, &lt.CreatedAt); err != nil {
			return nil, err
		}
		if lastRun.Valid {
			lt.LastRunAt = &lastRun.Time
		}
		list = append(list, &lt)
	}
	return list, rows.Err()
}

// ListAll 查询全部本地已下载/导入的修改器
func (r *LocalTrainerRepository) ListAll(ctx context.Context) ([]*model.LocalTrainer, error) {
	rows, err := r.db.QueryContext(ctx, `
		SELECT id, game_id, version_id, file_path, file_size, sha256, source_name, is_custom, last_run_at, run_count, created_at
		FROM local_trainers
		ORDER BY last_run_at DESC NULLS LAST, created_at DESC
	`)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var list []*model.LocalTrainer
	for rows.Next() {
		var lt model.LocalTrainer
		var lastRun sql.NullTime
		if err := rows.Scan(&lt.ID, &lt.GameID, &lt.VersionID, &lt.FilePath, &lt.FileSize, &lt.SHA256, &lt.SourceName, &lt.IsCustom, &lastRun, &lt.RunCount, &lt.CreatedAt); err != nil {
			return nil, err
		}
		if lastRun.Valid {
			lt.LastRunAt = &lastRun.Time
		}
		list = append(list, &lt)
	}
	return list, rows.Err()
}

// Delete 删除本地修改器记录
func (r *LocalTrainerRepository) Delete(ctx context.Context, id int64) error {
	_, err := r.db.ExecContext(ctx, "DELETE FROM local_trainers WHERE id = ?", id)
	return err
}

// Update 更新本地修改器记录（例如重新定位路径、文件大小、哈希）
func (r *LocalTrainerRepository) Update(ctx context.Context, lt *model.LocalTrainer) error {
	_, err := r.db.ExecContext(ctx, `
		UPDATE local_trainers
		SET file_path = ?, file_size = ?, sha256 = ?, source_name = ?, is_custom = ?
		WHERE id = ?
	`, lt.FilePath, lt.FileSize, lt.SHA256, lt.SourceName, lt.IsCustom, lt.ID)
	return err
}

// RecordRun 记录启动一次修改器
func (r *LocalTrainerRepository) RecordRun(ctx context.Context, id int64) error {
	now := time.Now()
	_, err := r.db.ExecContext(ctx, `
		UPDATE local_trainers
		SET last_run_at = ?, run_count = run_count + 1
		WHERE id = ?
	`, now, id)
	return err
}

