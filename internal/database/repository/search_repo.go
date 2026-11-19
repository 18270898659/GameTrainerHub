package repository

import (
	"context"
	"strings"
	"time"

	"GameTrainerHub/internal/database"
	"GameTrainerHub/internal/model"
)

type SearchRepository struct {
	db *database.DB
}

func NewSearchRepository(db *database.DB) *SearchRepository {
	return &SearchRepository{db: db}
}

// AddHistory 记录或更新搜索关键词与时间
func (r *SearchRepository) AddHistory(ctx context.Context, keyword string) error {
	trimmed := strings.TrimSpace(keyword)
	if trimmed == "" {
		return nil
	}
	_, err := r.db.ExecContext(ctx, `
		INSERT INTO search_history (keyword, created_at)
		VALUES (?, CURRENT_TIMESTAMP)
		ON CONFLICT(keyword) DO UPDATE SET created_at = CURRENT_TIMESTAMP
	`, trimmed)
	return err
}

// ListHistory 获取最近的搜索关键词列表（按时间倒序）
func (r *SearchRepository) ListHistory(ctx context.Context, limit int) ([]string, error) {
	if limit <= 0 {
		limit = 20
	}
	rows, err := r.db.QueryContext(ctx, `
		SELECT keyword
		FROM search_history
		ORDER BY created_at DESC
		LIMIT ?
	`, limit)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var keywords []string
	for rows.Next() {
		var kw string
		if err := rows.Scan(&kw); err != nil {
			return nil, err
		}
		keywords = append(keywords, kw)
	}
	return keywords, rows.Err()
}

// RemoveHistory 删除单条搜索历史
func (r *SearchRepository) RemoveHistory(ctx context.Context, keyword string) error {
	_, err := r.db.ExecContext(ctx, "DELETE FROM search_history WHERE keyword = ?", keyword)
	return err
}

// ClearHistory 清空全部搜索历史
func (r *SearchRepository) ClearHistory(ctx context.Context) error {
	_, err := r.db.ExecContext(ctx, "DELETE FROM search_history")
	return err
}

// ListDetailedHistory 获取带时间戳的完整搜索历史列表
func (r *SearchRepository) ListDetailedHistory(ctx context.Context, limit int) ([]*model.SearchHistoryItem, error) {
	if limit <= 0 {
		limit = 20
	}
	rows, err := r.db.QueryContext(ctx, `
		SELECT id, keyword, created_at
		FROM search_history
		ORDER BY created_at DESC
		LIMIT ?
	`, limit)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var items []*model.SearchHistoryItem
	for rows.Next() {
		var it model.SearchHistoryItem
		var t time.Time
		if err := rows.Scan(&it.ID, &it.Keyword, &t); err != nil {
			return nil, err
		}
		it.CreatedAt = t
		items = append(items, &it)
	}
	return items, rows.Err()
}
