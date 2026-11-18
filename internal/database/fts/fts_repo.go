package fts

import (
	"context"
	"database/sql"
	"fmt"
	"strings"

	"GameTrainerHub/internal/database"
)

type FTSRepository struct {
	db *database.DB
}

func NewFTSRepository(db *database.DB) *FTSRepository {
	return &FTSRepository{db: db}
}

// IndexGame 索引单个游戏及其别名与拼音
func (r *FTSRepository) IndexGame(ctx context.Context, gameID int64, nameEN, nameCN, aliases, pinyin, pinyinInitials string) error {
	// 先删除旧索引
	_ = r.RemoveGameIndex(ctx, gameID)

	_, err := r.db.ExecContext(ctx, `
		INSERT INTO game_search_fts (game_id, name_en, name_cn, aliases, pinyin, pinyin_initials)
		VALUES (?, ?, ?, ?, ?, ?)
	`, gameID, nameEN, nameCN, aliases, pinyin, pinyinInitials)
	return err
}

// RemoveGameIndex 删除指定游戏的全文检索索引
func (r *FTSRepository) RemoveGameIndex(ctx context.Context, gameID int64) error {
	_, err := r.db.ExecContext(ctx, "DELETE FROM game_search_fts WHERE game_id = ?", gameID)
	return err
}

// Search 执行 FTS5 全文搜索，返回匹配的 game_id 列表（按 BM25 相关度排序）
func (r *FTSRepository) Search(ctx context.Context, rawQuery string, limit int) ([]int64, error) {
	cleanQuery := strings.TrimSpace(rawQuery)
	if cleanQuery == "" {
		return nil, nil
	}

	// 清洗 FTS5 特殊保留字符 (避免 SQL 语法注入或 FTS5 语法报错)
	replacer := strings.NewReplacer(
		"'", " ",
		"\"", " ",
		"*", " ",
		"^", " ",
		":", " ",
		"(", " ",
		")", " ",
		"-", " ",
		"+", " ",
		"NOT", " ",
		"AND", " ",
		"OR", " ",
	)
	safeQuery := replacer.Replace(cleanQuery)
	words := strings.Fields(safeQuery)
	if len(words) == 0 {
		return nil, nil
	}

	// 构造多词前缀匹配表达式: "word1"* OR "word2"*
	var matchParts []string
	for _, w := range words {
		matchParts = append(matchParts, fmt.Sprintf("\"%s\"*", w))
	}
	matchExpr := strings.Join(matchParts, " OR ")

	query := `
		SELECT game_id
		FROM game_search_fts
		WHERE game_search_fts MATCH ?
		ORDER BY bm25(game_search_fts) ASC
		LIMIT ?
	`

	rows, err := r.db.QueryContext(ctx, query, matchExpr, limit)
	if err != nil {
		return nil, fmt.Errorf("FTS5 match error: %w", err)
	}
	defer rows.Close()

	var gameIDs []int64
	for rows.Next() {
		var id int64
		if err := rows.Scan(&id); err != nil {
			return nil, err
		}
		gameIDs = append(gameIDs, id)
	}
	return gameIDs, rows.Err()
}

// RebuildIndex 从 games 和 game_aliases 重建全文检索表
func (r *FTSRepository) RebuildIndex(ctx context.Context) error {
	return r.db.WithTx(ctx, func(tx *sql.Tx) error {
		// 清空 FTS 表
		if _, err := tx.ExecContext(ctx, "DELETE FROM game_search_fts"); err != nil {
			return err
		}

		// 重新拉取并插入所有游戏
		rows, err := tx.QueryContext(ctx, `
			SELECT g.id, g.name_en, g.name_cn,
			       COALESCE(GROUP_CONCAT(ga.alias, ' '), '') AS aliases
			FROM games g
			LEFT JOIN game_aliases ga ON g.id = ga.game_id
			GROUP BY g.id
		`)
		if err != nil {
			return err
		}
		defer rows.Close()

		type item struct {
			id      int64
			nameEN  string
			nameCN  string
			aliases string
		}
		var items []item
		for rows.Next() {
			var it item
			if err := rows.Scan(&it.id, &it.nameEN, &it.nameCN, &it.aliases); err != nil {
				return err
			}
			items = append(items, it)
		}

		for _, it := range items {
			_, err := tx.ExecContext(ctx, `
				INSERT INTO game_search_fts (game_id, name_en, name_cn, aliases, pinyin, pinyin_initials)
				VALUES (?, ?, ?, ?, '', '')
			`, it.id, it.nameEN, it.nameCN, it.aliases)
			if err != nil {
				return err
			}
		}

		return nil
	})
}
