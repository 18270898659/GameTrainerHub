package repository

import (
	"context"
	"fmt"
	"os"
	"time"

	"GameTrainerHub/internal/database"
	"GameTrainerHub/internal/model"
)

type DownloadRepository struct {
	db *database.DB
}

func NewDownloadRepository(db *database.DB) *DownloadRepository {
	return &DownloadRepository{db: db}
}

// CreateTask 创建下载任务
func (r *DownloadRepository) CreateTask(ctx context.Context, t *model.DownloadTask) error {
	now := time.Now()
	res, err := r.db.ExecContext(ctx, `
		INSERT INTO download_tasks (version_id, file_name, save_path, total_bytes, downloaded_bytes, status, error_msg, retries, created_at, updated_at, game_name, trainer_name, version_name)
		VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?)
	`, t.VersionID, t.FileName, t.SavePath, t.TotalBytes, t.DownloadedBytes, t.Status, t.ErrorMsg, t.Retries, now, now, t.GameName, t.TrainerName, t.VersionName)
	if err != nil {
		return fmt.Errorf("failed to create download task: %w", err)
	}

	id, err := res.LastInsertId()
	if err != nil {
		return err
	}
	t.ID = id
	t.CreatedAt = now
	t.UpdatedAt = now
	return nil
}

// UpdateTask 更新下载进度与状态
func (r *DownloadRepository) UpdateTask(ctx context.Context, t *model.DownloadTask) error {
	now := time.Now()
	_, err := r.db.ExecContext(ctx, `
		UPDATE download_tasks
		SET total_bytes = ?, downloaded_bytes = ?, status = ?, error_msg = ?, retries = ?, updated_at = ?, game_name = ?, trainer_name = ?, version_name = ?
		WHERE id = ?
	`, t.TotalBytes, t.DownloadedBytes, t.Status, t.ErrorMsg, t.Retries, now, t.GameName, t.TrainerName, t.VersionName, t.ID)
	if err != nil {
		return fmt.Errorf("failed to update download task: %w", err)
	}
	t.UpdatedAt = now
	return nil
}

// GetTaskByID 获取下载任务
func (r *DownloadRepository) GetTaskByID(ctx context.Context, id int64) (*model.DownloadTask, error) {
	row := r.db.QueryRowContext(ctx, `
		SELECT dt.id, dt.version_id, dt.file_name, dt.save_path, dt.total_bytes, dt.downloaded_bytes, 
		       dt.status, dt.error_msg, dt.retries, dt.created_at, dt.updated_at,
		       COALESCE(NULLIF(dt.game_name, ''), g.name_cn, g.name_en, '') AS game_name,
		       COALESCE(NULLIF(dt.trainer_name, ''), tr.title_en, '') AS trainer_name,
		       COALESCE(NULLIF(dt.version_name, ''), tv.version_name, '') AS version_name
		FROM download_tasks dt
		LEFT JOIN trainer_versions tv ON dt.version_id = tv.id
		LEFT JOIN trainers tr ON tv.trainer_id = tr.id
		LEFT JOIN games g ON tr.game_id = g.id
		WHERE dt.id = ?
	`, id)

	var t model.DownloadTask
	err := row.Scan(&t.ID, &t.VersionID, &t.FileName, &t.SavePath, &t.TotalBytes, &t.DownloadedBytes, &t.Status, &t.ErrorMsg, &t.Retries, &t.CreatedAt, &t.UpdatedAt, &t.GameName, &t.TrainerName, &t.VersionName)
	if err != nil {
		return nil, err
	}
	if t.TotalBytes > 0 {
		t.Progress = float64(t.DownloadedBytes) / float64(t.TotalBytes) * 100
	}
	return &t, nil
}

// ListTasks 获取指定状态或全部下载任务
func (r *DownloadRepository) ListTasks(ctx context.Context, status string) ([]*model.DownloadTask, error) {
	query := `
		SELECT dt.id, dt.version_id, dt.file_name, dt.save_path, dt.total_bytes, dt.downloaded_bytes, 
		       dt.status, dt.error_msg, dt.retries, dt.created_at, dt.updated_at,
		       COALESCE(NULLIF(dt.game_name, ''), g.name_cn, g.name_en, '') AS game_name,
		       COALESCE(NULLIF(dt.trainer_name, ''), tr.title_en, '') AS trainer_name,
		       COALESCE(NULLIF(dt.version_name, ''), tv.version_name, '') AS version_name
		FROM download_tasks dt
		LEFT JOIN trainer_versions tv ON dt.version_id = tv.id
		LEFT JOIN trainers tr ON tv.trainer_id = tr.id
		LEFT JOIN games g ON tr.game_id = g.id
	`
	var args []any
	if status == "active" {
		query += " WHERE dt.status != 'completed'"
	} else if status != "" {
		query += " WHERE dt.status = ?"
		args = append(args, status)
	}
	query += " ORDER BY dt.updated_at DESC"

	rows, err := r.db.QueryContext(ctx, query, args...)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var tasks []*model.DownloadTask
	for rows.Next() {
		var t model.DownloadTask
		if err := rows.Scan(&t.ID, &t.VersionID, &t.FileName, &t.SavePath, &t.TotalBytes, &t.DownloadedBytes, &t.Status, &t.ErrorMsg, &t.Retries, &t.CreatedAt, &t.UpdatedAt, &t.GameName, &t.TrainerName, &t.VersionName); err != nil {
			return nil, err
		}
		if t.TotalBytes > 0 {
			t.Progress = float64(t.DownloadedBytes) / float64(t.TotalBytes) * 100
		}
		tasks = append(tasks, &t)
	}
	return tasks, rows.Err()
}

// ListActiveTasks 仅查询未完成的活动下载任务 (pending, downloading, paused, retrying, failed)
func (r *DownloadRepository) ListActiveTasks(ctx context.Context) ([]*model.DownloadTask, error) {
	return r.ListTasks(ctx, "active")
}

// DeleteTask 删除下载任务
func (r *DownloadRepository) DeleteTask(ctx context.Context, id int64) error {
	_, err := r.db.ExecContext(ctx, "DELETE FROM download_tasks WHERE id = ?", id)
	return err
}

// AddHistory 记录下载历史
func (r *DownloadRepository) AddHistory(ctx context.Context, h *model.DownloadHistory) error {
	now := time.Now()
	res, err := r.db.ExecContext(ctx, `
		INSERT INTO download_history (task_id, trainer_id, version_id, game_name, trainer_name, version_name, file_name, file_size, sha256, save_path, completed_at)
		VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?)
	`, h.TaskID, h.TrainerID, h.VersionID, h.GameName, h.TrainerName, h.VersionName, h.FileName, h.FileSize, h.SHA256, h.SavePath, now)
	if err != nil {
		return err
	}
	id, err := res.LastInsertId()
	if err != nil {
		return err
	}
	h.ID = id
	h.CompletedAt = now
	return nil
}

// ListHistory 查询下载历史
func (r *DownloadRepository) ListHistory(ctx context.Context, limit int) ([]*model.DownloadHistory, error) {
	rows, err := r.db.QueryContext(ctx, `
		SELECT id, task_id, trainer_id, version_id, game_name, 
		       COALESCE(trainer_name, '') AS trainer_name,
		       COALESCE(version_name, '') AS version_name,
		       file_name, file_size, sha256, save_path, completed_at
		FROM download_history
		ORDER BY completed_at DESC
		LIMIT ?
	`, limit)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var list []*model.DownloadHistory
	for rows.Next() {
		var h model.DownloadHistory
		if err := rows.Scan(&h.ID, &h.TaskID, &h.TrainerID, &h.VersionID, &h.GameName, &h.TrainerName, &h.VersionName, &h.FileName, &h.FileSize, &h.SHA256, &h.SavePath, &h.CompletedAt); err != nil {
			return nil, err
		}
		if _, statErr := os.Stat(h.SavePath); statErr == nil {
			h.FileExists = true
		} else {
			h.FileExists = false
		}
		list = append(list, &h)
	}
	return list, rows.Err()
}

// DeleteHistory 删除指定下载历史记录
func (r *DownloadRepository) DeleteHistory(ctx context.Context, id int64) error {
	_, err := r.db.ExecContext(ctx, "DELETE FROM download_history WHERE id = ?", id)
	return err
}

// ClearHistory 清空所有下载历史
func (r *DownloadRepository) ClearHistory(ctx context.Context) error {
	_, err := r.db.ExecContext(ctx, "DELETE FROM download_history")
	return err
}
