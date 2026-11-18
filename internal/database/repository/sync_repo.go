package repository

import (
	"context"
	"database/sql"
	"fmt"
	"time"

	"GameTrainerHub/internal/database"
	"GameTrainerHub/internal/model"
)

type SyncRepository struct {
	db *database.DB
}

func NewSyncRepository(db *database.DB) *SyncRepository {
	return &SyncRepository{db: db}
}

// CreateRecord 创建同步记录
func (r *SyncRepository) CreateRecord(ctx context.Context, rec *model.SyncRecord) error {
	now := time.Now()
	res, err := r.db.ExecContext(ctx, `
		INSERT INTO sync_records (source_id, sync_type, status, items_added, items_updated, error_msg, started_at)
		VALUES (?, ?, ?, ?, ?, ?, ?)
	`, rec.SourceID, rec.SyncType, rec.Status, rec.ItemsAdded, rec.ItemsUpdated, rec.ErrorMsg, now)
	if err != nil {
		return fmt.Errorf("failed to create sync record: %w", err)
	}

	id, err := res.LastInsertId()
	if err != nil {
		return err
	}
	rec.ID = id
	rec.StartedAt = now
	return nil
}

// UpdateRecord 更新同步记录（完成/失败）
func (r *SyncRepository) UpdateRecord(ctx context.Context, rec *model.SyncRecord) error {
	now := time.Now()
	_, err := r.db.ExecContext(ctx, `
		UPDATE sync_records
		SET status = ?, items_added = ?, items_updated = ?, error_msg = ?, finished_at = ?
		WHERE id = ?
	`, rec.Status, rec.ItemsAdded, rec.ItemsUpdated, rec.ErrorMsg, now, rec.ID)
	if err != nil {
		return fmt.Errorf("failed to update sync record: %w", err)
	}
	rec.FinishedAt = &now
	return nil
}

// GetLatestRecord 获取指定源最近一次成功的同步记录
func (r *SyncRepository) GetLatestRecord(ctx context.Context, sourceID string) (*model.SyncRecord, error) {
	row := r.db.QueryRowContext(ctx, `
		SELECT id, source_id, sync_type, status, items_added, items_updated, error_msg, started_at, finished_at
		FROM sync_records
		WHERE source_id = ? AND status = 'success'
		ORDER BY finished_at DESC
		LIMIT 1
	`, sourceID)

	var rec model.SyncRecord
	var fin sql.NullTime
	err := row.Scan(&rec.ID, &rec.SourceID, &rec.SyncType, &rec.Status, &rec.ItemsAdded, &rec.ItemsUpdated, &rec.ErrorMsg, &rec.StartedAt, &fin)
	if err != nil {
		return nil, err
	}
	if fin.Valid {
		rec.FinishedAt = &fin.Time
	}
	return &rec, nil
}

