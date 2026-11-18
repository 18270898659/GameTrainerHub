package repository

import (
	"context"
	"database/sql"
	"fmt"
	"time"

	"GameTrainerHub/internal/database"
	"GameTrainerHub/internal/model"
)

type TrainerRepository struct {
	db *database.DB
}

func NewTrainerRepository(db *database.DB) *TrainerRepository {
	return &TrainerRepository{db: db}
}

// CreateTrainer 插入修改器
func (r *TrainerRepository) CreateTrainer(ctx context.Context, t *model.Trainer) error {
	now := time.Now()
	res, err := r.db.ExecContext(ctx, `
		INSERT INTO trainers (game_id, source_id, external_id, title_en, title_cn, detail_url, author, latest_version, published_at, updated_at, is_deleted)
		VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?)
	`, t.GameID, t.SourceID, t.ExternalID, t.TitleEN, t.TitleCN, t.DetailURL, t.Author, t.LatestVersion, t.PublishedAt, now, t.IsDeleted)
	if err != nil {
		return fmt.Errorf("failed to insert trainer: %w", err)
	}

	id, err := res.LastInsertId()
	if err != nil {
		return err
	}
	t.ID = id
	t.UpdatedAt = now
	return nil
}

// UpdateTrainer 更新修改器
func (r *TrainerRepository) UpdateTrainer(ctx context.Context, t *model.Trainer) error {
	now := time.Now()
	_, err := r.db.ExecContext(ctx, `
		UPDATE trainers
		SET title_en = ?, title_cn = ?, detail_url = ?, author = ?, latest_version = ?, updated_at = ?, is_deleted = ?
		WHERE id = ?
	`, t.TitleEN, t.TitleCN, t.DetailURL, t.Author, t.LatestVersion, now, t.IsDeleted, t.ID)
	if err != nil {
		return fmt.Errorf("failed to update trainer: %w", err)
	}
	t.UpdatedAt = now
	return nil
}

// GetTrainerByID 根据 ID 获取修改器
func (r *TrainerRepository) GetTrainerByID(ctx context.Context, id int64) (*model.Trainer, error) {
	row := r.db.QueryRowContext(ctx, `
		SELECT id, game_id, source_id, external_id, title_en, title_cn, detail_url, author, latest_version, published_at, updated_at, is_deleted
		FROM trainers
		WHERE id = ?
	`, id)

	var t model.Trainer
	err := row.Scan(&t.ID, &t.GameID, &t.SourceID, &t.ExternalID, &t.TitleEN, &t.TitleCN, &t.DetailURL, &t.Author, &t.LatestVersion, &t.PublishedAt, &t.UpdatedAt, &t.IsDeleted)
	if err != nil {
		return nil, err
	}

	// 载入版本列表
	versions, err := r.ListVersionsByTrainerID(ctx, t.ID)
	if err == nil {
		t.Versions = versions
	}
	return &t, nil
}

// GetTrainerByExternalID 根据来源与外部ID查询修改器
func (r *TrainerRepository) GetTrainerByExternalID(ctx context.Context, sourceID, externalID string) (*model.Trainer, error) {
	row := r.db.QueryRowContext(ctx, `
		SELECT id, game_id, source_id, external_id, title_en, title_cn, detail_url, author, latest_version, published_at, updated_at, is_deleted
		FROM trainers
		WHERE source_id = ? AND external_id = ?
	`, sourceID, externalID)

	var t model.Trainer
	err := row.Scan(&t.ID, &t.GameID, &t.SourceID, &t.ExternalID, &t.TitleEN, &t.TitleCN, &t.DetailURL, &t.Author, &t.LatestVersion, &t.PublishedAt, &t.UpdatedAt, &t.IsDeleted)
	if err != nil {
		return nil, err
	}

	versions, err := r.ListVersionsByTrainerID(ctx, t.ID)
	if err == nil {
		t.Versions = versions
	}
	return &t, nil
}

// ListTrainersByGameID 查询指定游戏下的所有修改器
func (r *TrainerRepository) ListTrainersByGameID(ctx context.Context, gameID int64) ([]*model.Trainer, error) {
	rows, err := r.db.QueryContext(ctx, `
		SELECT id, game_id, source_id, external_id, title_en, title_cn, detail_url, author, latest_version, published_at, updated_at, is_deleted
		FROM trainers
		WHERE game_id = ? AND is_deleted = 0
		ORDER BY updated_at DESC
	`, gameID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var trainers []*model.Trainer
	for rows.Next() {
		var t model.Trainer
		if err := rows.Scan(&t.ID, &t.GameID, &t.SourceID, &t.ExternalID, &t.TitleEN, &t.TitleCN, &t.DetailURL, &t.Author, &t.LatestVersion, &t.PublishedAt, &t.UpdatedAt, &t.IsDeleted); err != nil {
			return nil, err
		}
		versions, err := r.ListVersionsByTrainerID(ctx, t.ID)
		if err == nil {
			t.Versions = versions
		}
		trainers = append(trainers, &t)
	}
	return trainers, rows.Err()
}

// CreateVersion 插入修改器版本
func (r *TrainerRepository) CreateVersion(ctx context.Context, v *model.TrainerVersion) error {
	now := time.Now()
	res, err := r.db.ExecContext(ctx, `
		INSERT INTO trainer_versions (trainer_id, version_name, target_game_version, features_text, download_url, file_name, file_size, sha256, published_at, created_at)
		VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?)
	`, v.TrainerID, v.VersionName, v.TargetGameVersion, v.FeaturesText, v.DownloadURL, v.FileName, v.FileSize, v.SHA256, v.PublishedAt, now)
	if err != nil {
		return fmt.Errorf("failed to insert trainer version: %w", err)
	}

	id, err := res.LastInsertId()
	if err != nil {
		return err
	}
	v.ID = id
	v.CreatedAt = now
	return nil
}

// GetVersionByID 获取指定修改器版本
func (r *TrainerRepository) GetVersionByID(ctx context.Context, id int64) (*model.TrainerVersion, error) {
	row := r.db.QueryRowContext(ctx, `
		SELECT tv.id, tv.trainer_id, tv.version_name, tv.target_game_version, tv.features_text, tv.download_url,
		       tv.file_name, tv.file_size, tv.sha256, tv.published_at, tv.created_at,
		       EXISTS(SELECT 1 FROM local_trainers WHERE version_id = tv.id) AS is_downloaded,
		       COALESCE((SELECT file_path FROM local_trainers WHERE version_id = tv.id LIMIT 1), '') AS local_path
		FROM trainer_versions tv
		WHERE tv.id = ?
	`, id)

	var v model.TrainerVersion
	var isDl int
	err := row.Scan(&v.ID, &v.TrainerID, &v.VersionName, &v.TargetGameVersion, &v.FeaturesText, &v.DownloadURL, &v.FileName, &v.FileSize, &v.SHA256, &v.PublishedAt, &v.CreatedAt, &isDl, &v.LocalPath)
	if err != nil {
		return nil, err
	}
	v.IsDownloaded = (isDl == 1)
	return &v, nil
}

// ListVersionsByTrainerID 获取修改器的所有版本
func (r *TrainerRepository) ListVersionsByTrainerID(ctx context.Context, trainerID int64) ([]*model.TrainerVersion, error) {
	rows, err := r.db.QueryContext(ctx, `
		SELECT tv.id, tv.trainer_id, tv.version_name, tv.target_game_version, tv.features_text, tv.download_url,
		       tv.file_name, tv.file_size, tv.sha256, tv.published_at, tv.created_at,
		       EXISTS(SELECT 1 FROM local_trainers WHERE version_id = tv.id) AS is_downloaded,
		       COALESCE((SELECT file_path FROM local_trainers WHERE version_id = tv.id LIMIT 1), '') AS local_path
		FROM trainer_versions tv
		WHERE tv.trainer_id = ?
		ORDER BY tv.published_at DESC, tv.id DESC
	`, trainerID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var versions []*model.TrainerVersion
	for rows.Next() {
		var v model.TrainerVersion
		var isDl int
		if err := rows.Scan(&v.ID, &v.TrainerID, &v.VersionName, &v.TargetGameVersion, &v.FeaturesText, &v.DownloadURL, &v.FileName, &v.FileSize, &v.SHA256, &v.PublishedAt, &v.CreatedAt, &isDl, &v.LocalPath); err != nil {
			return nil, err
		}
		v.IsDownloaded = (isDl == 1)
		versions = append(versions, &v)
	}
	return versions, rows.Err()
}

// UpsertTrainerWithVersions 事务保存修改器及其版本列表
func (r *TrainerRepository) UpsertTrainerWithVersions(ctx context.Context, t *model.Trainer, versions []*model.TrainerVersion) error {
	return r.db.WithTx(ctx, func(tx *sql.Tx) error {
		// 查询是否存在
		var existingID int64
		err := tx.QueryRowContext(ctx, `
			SELECT id FROM trainers WHERE source_id = ? AND external_id = ?
		`, t.SourceID, t.ExternalID).Scan(&existingID)

		now := time.Now()
		if err == sql.ErrNoRows {
			// 插入
			res, err := tx.ExecContext(ctx, `
				INSERT INTO trainers (game_id, source_id, external_id, title_en, title_cn, detail_url, author, latest_version, published_at, updated_at, is_deleted)
				VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?, 0)
			`, t.GameID, t.SourceID, t.ExternalID, t.TitleEN, t.TitleCN, t.DetailURL, t.Author, t.LatestVersion, t.PublishedAt, now)
			if err != nil {
				return err
			}
			t.ID, err = res.LastInsertId()
			if err != nil {
				return err
			}
		} else if err == nil {
			// 更新
			t.ID = existingID
			_, err = tx.ExecContext(ctx, `
				UPDATE trainers 
				SET title_en = ?, title_cn = ?, detail_url = ?, author = ?, latest_version = ?, updated_at = ?
				WHERE id = ?
			`, t.TitleEN, t.TitleCN, t.DetailURL, t.Author, t.LatestVersion, now, t.ID)
			if err != nil {
				return err
			}
		} else {
			return err
		}

		// 处理版本列表
		for _, v := range versions {
			v.TrainerID = t.ID
			var vID int64
			vErr := tx.QueryRowContext(ctx, `
				SELECT id FROM trainer_versions WHERE trainer_id = ? AND version_name = ?
			`, t.ID, v.VersionName).Scan(&vID)

			if vErr == sql.ErrNoRows {
				res, err := tx.ExecContext(ctx, `
					INSERT INTO trainer_versions (trainer_id, version_name, target_game_version, features_text, download_url, file_name, file_size, sha256, published_at, created_at)
					VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?)
				`, v.TrainerID, v.VersionName, v.TargetGameVersion, v.FeaturesText, v.DownloadURL, v.FileName, v.FileSize, v.SHA256, v.PublishedAt, now)
				if err != nil {
					return err
				}
				v.ID, _ = res.LastInsertId()
			} else if vErr == nil {
				v.ID = vID
				// 更新下载链接或说明
				_, err := tx.ExecContext(ctx, `
					UPDATE trainer_versions 
					SET download_url = ?, file_name = ?, features_text = ?, target_game_version = ?
					WHERE id = ?
				`, v.DownloadURL, v.FileName, v.FeaturesText, v.TargetGameVersion, v.ID)
				if err != nil {
					return err
				}
			} else {
				return vErr
			}
		}

		return nil
	})
}

