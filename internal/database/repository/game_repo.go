package repository

import (
	"context"
	"database/sql"
	"fmt"
	"time"

	"GameTrainerHub/internal/database"
	"GameTrainerHub/internal/model"
)

type GameRepository struct {
	db *database.DB
}

func NewGameRepository(db *database.DB) *GameRepository {
	return &GameRepository{db: db}
}

// Create 插入新游戏
func (r *GameRepository) Create(ctx context.Context, g *model.Game) error {
	now := time.Now()
	res, err := r.db.ExecContext(ctx, `
		INSERT INTO games (slug, name_en, name_cn, cover_url, steam_app_id, install_path, created_at, updated_at)
		VALUES (?, ?, ?, ?, ?, ?, ?, ?)
	`, g.Slug, g.NameEN, g.NameCN, g.CoverURL, g.SteamAppID, g.InstallPath, now, now)
	if err != nil {
		return fmt.Errorf("failed to insert game: %w", err)
	}

	id, err := res.LastInsertId()
	if err != nil {
		return err
	}
	g.ID = id
	g.CreatedAt = now
	g.UpdatedAt = now
	return nil
}

// Update 更新游戏信息
func (r *GameRepository) Update(ctx context.Context, g *model.Game) error {
	now := time.Now()
	_, err := r.db.ExecContext(ctx, `
		UPDATE games 
		SET name_en = ?, name_cn = ?, cover_url = ?, steam_app_id = ?, install_path = ?, updated_at = ?
		WHERE id = ?
	`, g.NameEN, g.NameCN, g.CoverURL, g.SteamAppID, g.InstallPath, now, g.ID)
	if err != nil {
		return fmt.Errorf("failed to update game: %w", err)
	}
	g.UpdatedAt = now
	return nil
}

// UpsertBySlug 根据 slug 插入或更新游戏
func (r *GameRepository) UpsertBySlug(ctx context.Context, g *model.Game) (*model.Game, error) {
	existing, err := r.GetBySlug(ctx, g.Slug)
	if err != nil && err != sql.ErrNoRows {
		return nil, err
	}

	if existing == nil {
		if err := r.Create(ctx, g); err != nil {
			return nil, err
		}
		return g, nil
	}

	// 更新已有游戏字段
	if g.NameEN != "" {
		existing.NameEN = g.NameEN
	}
	if g.NameCN != "" {
		existing.NameCN = g.NameCN
	}
	if g.CoverURL != "" {
		existing.CoverURL = g.CoverURL
	}
	if g.SteamAppID != 0 {
		existing.SteamAppID = g.SteamAppID
	}
	if g.InstallPath != "" {
		existing.InstallPath = g.InstallPath
	}

	if err := r.Update(ctx, existing); err != nil {
		return nil, err
	}
	return existing, nil
}

// GetByID 通过 ID 查询游戏
func (r *GameRepository) GetByID(ctx context.Context, id int64) (*model.Game, error) {
	row := r.db.QueryRowContext(ctx, `
		SELECT g.id, g.slug, g.name_en, g.name_cn, g.cover_url, g.steam_app_id, g.install_path, 
		       g.created_at, g.updated_at,
		       EXISTS(SELECT 1 FROM favorites WHERE game_id = g.id) AS is_fav
		FROM games g
		WHERE g.id = ?
	`, id)

	var g model.Game
	var isFav int
	err := row.Scan(&g.ID, &g.Slug, &g.NameEN, &g.NameCN, &g.CoverURL, &g.SteamAppID, &g.InstallPath, &g.CreatedAt, &g.UpdatedAt, &isFav)
	if err != nil {
		return nil, err
	}
	g.IsFavorite = (isFav == 1)

	// 加载别名
	aliases, _ := r.GetAliases(ctx, g.ID)
	for _, a := range aliases {
		g.Aliases = append(g.Aliases, a.Alias)
	}
	return &g, nil
}

// GetBySlug 通过 Slug 查询游戏
func (r *GameRepository) GetBySlug(ctx context.Context, slug string) (*model.Game, error) {
	row := r.db.QueryRowContext(ctx, `
		SELECT g.id, g.slug, g.name_en, g.name_cn, g.cover_url, g.steam_app_id, g.install_path, 
		       g.created_at, g.updated_at,
		       EXISTS(SELECT 1 FROM favorites WHERE game_id = g.id) AS is_fav
		FROM games g
		WHERE g.slug = ?
	`, slug)

	var g model.Game
	var isFav int
	err := row.Scan(&g.ID, &g.Slug, &g.NameEN, &g.NameCN, &g.CoverURL, &g.SteamAppID, &g.InstallPath, &g.CreatedAt, &g.UpdatedAt, &isFav)
	if err != nil {
		return nil, err
	}
	g.IsFavorite = (isFav == 1)

	aliases, _ := r.GetAliases(ctx, g.ID)
	for _, a := range aliases {
		g.Aliases = append(g.Aliases, a.Alias)
	}
	return &g, nil
}

// List 分页查询游戏列表
func (r *GameRepository) List(ctx context.Context, offset, limit int) ([]*model.Game, error) {
	rows, err := r.db.QueryContext(ctx, `
		SELECT g.id, g.slug, g.name_en, g.name_cn, g.cover_url, g.steam_app_id, g.install_path, 
		       g.created_at, g.updated_at,
		       EXISTS(SELECT 1 FROM favorites WHERE game_id = g.id) AS is_fav
		FROM games g
		ORDER BY g.updated_at DESC
		LIMIT ? OFFSET ?
	`, limit, offset)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var games []*model.Game
	for rows.Next() {
		var g model.Game
		var isFav int
		if err := rows.Scan(&g.ID, &g.Slug, &g.NameEN, &g.NameCN, &g.CoverURL, &g.SteamAppID, &g.InstallPath, &g.CreatedAt, &g.UpdatedAt, &isFav); err != nil {
			return nil, err
		}
		g.IsFavorite = (isFav == 1)
		games = append(games, &g)
	}
	return games, rows.Err()
}

// Count 获取游戏总数
func (r *GameRepository) Count(ctx context.Context) (int, error) {
	var count int
	err := r.db.QueryRowContext(ctx, "SELECT COUNT(*) FROM games").Scan(&count)
	return count, err
}

// AddAlias 添加游戏别名
func (r *GameRepository) AddAlias(ctx context.Context, gameID int64, alias, lang string) error {
	_, err := r.db.ExecContext(ctx, `
		INSERT INTO game_aliases (game_id, alias, lang)
		VALUES (?, ?, ?)
	`, gameID, alias, lang)
	return err
}

// GetAliases 获取游戏所有别名
func (r *GameRepository) GetAliases(ctx context.Context, gameID int64) ([]*model.GameAlias, error) {
	rows, err := r.db.QueryContext(ctx, `
		SELECT id, game_id, alias, lang, created_at 
		FROM game_aliases 
		WHERE game_id = ?
	`, gameID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var aliases []*model.GameAlias
	for rows.Next() {
		var a model.GameAlias
		if err := rows.Scan(&a.ID, &a.GameID, &a.Alias, &a.Lang, &a.CreatedAt); err != nil {
			return nil, err
		}
		aliases = append(aliases, &a)
	}
	return aliases, rows.Err()
}

// IsFavorite 查询是否已收藏
func (r *GameRepository) IsFavorite(ctx context.Context, gameID int64) (bool, error) {
	var count int
	err := r.db.QueryRowContext(ctx, "SELECT COUNT(*) FROM favorites WHERE game_id = ?", gameID).Scan(&count)
	return count > 0, err
}

// SetFavorite 设置/取消收藏
func (r *GameRepository) SetFavorite(ctx context.Context, gameID int64, isFav bool) error {
	if isFav {
		_, err := r.db.ExecContext(ctx, "INSERT OR IGNORE INTO favorites (game_id) VALUES (?)", gameID)
		return err
	}
	_, err := r.db.ExecContext(ctx, "DELETE FROM favorites WHERE game_id = ?", gameID)
	return err
}

// ListFavorites 获取所有收藏游戏
func (r *GameRepository) ListFavorites(ctx context.Context) ([]*model.Game, error) {
	rows, err := r.db.QueryContext(ctx, `
		SELECT g.id, g.slug, g.name_en, g.name_cn, g.cover_url, g.steam_app_id, g.install_path, 
		       g.created_at, g.updated_at, 1 AS is_fav
		FROM games g
		JOIN favorites f ON g.id = f.game_id
		ORDER BY f.created_at DESC
	`)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var games []*model.Game
	for rows.Next() {
		var g model.Game
		var isFav int
		if err := rows.Scan(&g.ID, &g.Slug, &g.NameEN, &g.NameCN, &g.CoverURL, &g.SteamAppID, &g.InstallPath, &g.CreatedAt, &g.UpdatedAt, &isFav); err != nil {
			return nil, err
		}
		g.IsFavorite = true
		games = append(games, &g)
	}
	return games, rows.Err()
}
