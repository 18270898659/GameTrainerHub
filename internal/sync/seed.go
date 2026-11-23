package sync

import (
	"context"
	_ "embed"
	"encoding/json"
	"fmt"
	"strings"
	"time"

	"GameTrainerHub/internal/logger"
	"GameTrainerHub/internal/model"
)

//go:embed seed_catalog.json
var seedCatalogJSON []byte

//go:embed seed_latest.json
var seedLatestJSON []byte

// SeedCatalogItem 种子目录条目
type SeedCatalogItem struct {
	Title      string `json:"title"`
	GameNameEN string `json:"game_name_en"`
	GameNameCN string `json:"game_name_cn,omitempty"`
	Slug       string `json:"slug"`
	DetailURL  string `json:"detail_url"`
	ExternalID string `json:"external_id"`
	CoverURL   string `json:"cover_url,omitempty"`
}

// LoadSeedCatalog 加载内置全量种子目录
func LoadSeedCatalog() ([]*SeedCatalogItem, error) {
	var items []*SeedCatalogItem
	if err := json.Unmarshal(seedCatalogJSON, &items); err != nil {
		return nil, fmt.Errorf("failed to unmarshal seed catalog: %w", err)
	}
	return items, nil
}

// LoadSeedLatest 加载内置最新推荐修改器列表
func LoadSeedLatest() ([]*model.SearchResult, error) {
	var items []*model.SearchResult
	if err := json.Unmarshal(seedLatestJSON, &items); err != nil {
		return nil, fmt.Errorf("failed to unmarshal seed latest: %w", err)
	}
	return items, nil
}

// SeedInitialDatabase 在全新设备冷启动时，极速批量填充 760+ 游戏基准库并建立全文索引
func (s *SyncService) SeedInitialDatabase(ctx context.Context) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	items, err := LoadSeedCatalog()
	if err != nil {
		return err
	}

	latestItems, _ := LoadSeedLatest()
	latestCoverMap := make(map[string]string)
	for _, it := range latestItems {
		if it.CoverURL != "" {
			latestCoverMap[it.Slug] = it.CoverURL
		}
	}

	rec := &model.SyncRecord{
		SourceID:  s.adapter.ID(),
		SyncType:  "seed_initial",
		Status:    model.SyncStatusRunning,
		StartedAt: time.Now(),
	}
	_ = s.syncRepo.CreateRecord(ctx, rec)

	addedCount := 0
	for _, it := range items {
		gameNameEN := it.GameNameEN
		if gameNameEN == "" {
			gameNameEN = strings.TrimSuffix(it.Title, " Trainer")
		}
		slug := it.Slug
		if slug == "" {
			slug = slugify(gameNameEN)
		}

		// 匹配或推导中文名
		gameNameCN := it.GameNameCN
		if gameNameCN == "" {
			cnCandidates := s.translator.TranslateEnglishToChinese(gameNameEN)
			if len(cnCandidates) > 0 {
				gameNameCN = cnCandidates[0]
			}
		}

		coverURL := it.CoverURL
		if coverURL == "" {
			coverURL = latestCoverMap[slug]
		}

		game := &model.Game{
			Slug:     slug,
			NameEN:   gameNameEN,
			NameCN:   gameNameCN,
			CoverURL: coverURL,
		}

		if err := s.gameRepo.Create(ctx, game); err != nil {
			// 若已存在则跳过
			continue
		}
		addedCount++

		// 创建基础修改器记录
		trainer := &model.Trainer{
			GameID:     game.ID,
			SourceID:   s.adapter.ID(),
			ExternalID: it.ExternalID,
			TitleEN:    it.Title,
			TitleCN:    gameNameCN + " 修改器",
			DetailURL:  it.DetailURL,
			Author:     "FLiNG",
		}
		_ = s.trainerRepo.CreateTrainer(ctx, trainer)

		// 维护 FTS 索引 (包含别名与拼音全拼、声母)
		aliases := s.translator.TranslateEnglishToChinese(gameNameEN)
		aliasStr := strings.Join(aliases, " ")
		var pinyins []string
		var initialsList []string
		if gameNameCN != "" {
			pinyins = append(pinyins, s.translator.ToPinyin(gameNameCN))
			initialsList = append(initialsList, s.translator.ToPinyinInitials(gameNameCN))
		}
		for _, al := range aliases {
			pinyins = append(pinyins, s.translator.ToPinyin(al))
			initialsList = append(initialsList, s.translator.ToPinyinInitials(al))
		}
		py := strings.Join(pinyins, " ")
		initials := strings.Join(initialsList, " ")
		_ = s.ftsRepo.IndexGame(ctx, game.ID, gameNameEN, gameNameCN, aliasStr, py, initials)
	}

	rec.Status = model.SyncStatusSuccess
	rec.ItemsAdded = addedCount
	now := time.Now()
	rec.FinishedAt = &now
	_ = s.syncRepo.UpdateRecord(ctx, rec)

	logger.Info("Initial database seeding finished", "items_seeded", addedCount)
	return nil
}
