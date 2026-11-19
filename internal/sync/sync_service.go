package sync

import (
	"context"
	"fmt"
	"strings"
	"sync"
	"time"

	"GameTrainerHub/internal/database/fts"
	"GameTrainerHub/internal/database/repository"
	"GameTrainerHub/internal/logger"
	"GameTrainerHub/internal/model"
	"GameTrainerHub/internal/source"
	"GameTrainerHub/internal/translate"
)

type SyncService struct {
	gameRepo    *repository.GameRepository
	trainerRepo *repository.TrainerRepository
	syncRepo    *repository.SyncRepository
	ftsRepo     *fts.FTSRepository
	translator  *translate.Translator
	adapter     source.SourceAdapter

	isRunning bool
	mu        sync.Mutex
	stopChan  chan struct{}
}

func NewSyncService(
	gameRepo *repository.GameRepository,
	trainerRepo *repository.TrainerRepository,
	syncRepo *repository.SyncRepository,
	ftsRepo *fts.FTSRepository,
	translator *translate.Translator,
	adapter source.SourceAdapter,
) *SyncService {
	if translator == nil {
		translator = translate.NewTranslator()
	}
	return &SyncService{
		gameRepo:    gameRepo,
		trainerRepo: trainerRepo,
		syncRepo:    syncRepo,
		ftsRepo:     ftsRepo,
		translator:  translator,
		adapter:     adapter,
		stopChan:    make(chan struct{}),
	}
}

// SyncCatalog 执行全站基准目录同步（用于冷启动或全量更新）
func (s *SyncService) SyncCatalog(ctx context.Context) (*model.SyncRecord, error) {
	s.mu.Lock()
	if s.isRunning {
		s.mu.Unlock()
		return nil, fmt.Errorf("sync is already running")
	}
	s.isRunning = true
	s.mu.Unlock()

	defer func() {
		s.mu.Lock()
		s.isRunning = false
		s.mu.Unlock()
	}()

	rec := &model.SyncRecord{
		SourceID:  s.adapter.ID(),
		SyncType:  "catalog",
		Status:    model.SyncStatusRunning,
		StartedAt: time.Now(),
	}
	if err := s.syncRepo.CreateRecord(ctx, rec); err != nil {
		return nil, err
	}

	items, err := s.adapter.GetCatalog(ctx)
	if err != nil {
		rec.Status = model.SyncStatusFailed
		rec.ErrorMsg = err.Error()
		_ = s.syncRepo.UpdateRecord(ctx, rec)
		return rec, err
	}

	addedCount := 0
	updatedCount := 0

	for _, it := range items {
		gameNameEN := strings.TrimSuffix(it.Title, " Trainer")
		slug := slugify(gameNameEN)

		// 匹配或推导中文名
		var gameNameCN string
		cnCandidates := s.translator.TranslateEnglishToChinese(gameNameEN)
		if len(cnCandidates) > 0 {
			gameNameCN = cnCandidates[0]
		}

		game := &model.Game{
			Slug:   slug,
			NameEN: gameNameEN,
			NameCN: gameNameCN,
		}

		existingGame, _ := s.gameRepo.GetBySlug(ctx, slug)
		var gameID int64
		if existingGame == nil {
			if err := s.gameRepo.Create(ctx, game); err == nil {
				gameID = game.ID
				addedCount++
			}
		} else {
			gameID = existingGame.ID
			if existingGame.NameCN == "" && gameNameCN != "" {
				existingGame.NameCN = gameNameCN
				_ = s.gameRepo.Update(ctx, existingGame)
			}
			updatedCount++
		}

		if gameID > 0 {
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
			_ = s.ftsRepo.IndexGame(ctx, gameID, gameNameEN, gameNameCN, aliasStr, py, initials)

			// 登记修改器主记录
			tr := &model.Trainer{
				GameID:     gameID,
				SourceID:   s.adapter.ID(),
				ExternalID: it.ExternalID,
				TitleEN:    it.Title,
				TitleCN:    gameNameCN + " 修改器",
				DetailURL:  it.DetailURL,
				Author:     "FLiNG",
			}
			existingTr, _ := s.trainerRepo.GetTrainerByExternalID(ctx, s.adapter.ID(), it.ExternalID)
			if existingTr == nil {
				_ = s.trainerRepo.CreateTrainer(ctx, tr)
			}
		}
	}

	rec.Status = model.SyncStatusSuccess
	rec.ItemsAdded = addedCount
	rec.ItemsUpdated = updatedCount
	_ = s.syncRepo.UpdateRecord(ctx, rec)

	logger.Info("Catalog sync finished", "added", addedCount, "updated", updatedCount)
	return rec, nil
}

// SyncLatest 同步最近发布的修改器
func (s *SyncService) SyncLatest(ctx context.Context, pages int) (*model.SyncRecord, error) {
	s.mu.Lock()
	if s.isRunning {
		s.mu.Unlock()
		return nil, fmt.Errorf("sync is already running")
	}
	s.isRunning = true
	s.mu.Unlock()

	defer func() {
		s.mu.Lock()
		s.isRunning = false
		s.mu.Unlock()
	}()

	rec := &model.SyncRecord{
		SourceID:  s.adapter.ID(),
		SyncType:  "incremental",
		Status:    model.SyncStatusRunning,
		StartedAt: time.Now(),
	}
	_ = s.syncRepo.CreateRecord(ctx, rec)

	if pages <= 0 {
		pages = 1
	}

	addedCount := 0
	updatedCount := 0

	for p := 1; p <= pages; p++ {
		results, err := s.adapter.GetLatest(ctx, p)
		if err != nil {
			logger.Warn("Failed to fetch latest page", "page", p, "err", err)
			continue
		}

		for _, item := range results {
			var gameNameCN string
			cnCandidates := s.translator.TranslateEnglishToChinese(item.GameNameEN)
			if len(cnCandidates) > 0 {
				gameNameCN = cnCandidates[0]
			}

			game := &model.Game{
				Slug:     item.Slug,
				NameEN:   item.GameNameEN,
				NameCN:   gameNameCN,
				CoverURL: item.CoverURL,
			}

			upsertedGame, err := s.gameRepo.UpsertBySlug(ctx, game)
			if err != nil {
				continue
			}

			// 维护 FTS 索引 (包含别名与拼音全拼、声母)
			aliases := s.translator.TranslateEnglishToChinese(upsertedGame.NameEN)
			aliasStr := strings.Join(aliases, " ")
			var pinyins []string
			var initialsList []string
			if upsertedGame.NameCN != "" {
				pinyins = append(pinyins, s.translator.ToPinyin(upsertedGame.NameCN))
				initialsList = append(initialsList, s.translator.ToPinyinInitials(upsertedGame.NameCN))
			}
			for _, al := range aliases {
				pinyins = append(pinyins, s.translator.ToPinyin(al))
				initialsList = append(initialsList, s.translator.ToPinyinInitials(al))
			}
			py := strings.Join(pinyins, " ")
			initials := strings.Join(initialsList, " ")
			_ = s.ftsRepo.IndexGame(ctx, upsertedGame.ID, upsertedGame.NameEN, upsertedGame.NameCN, aliasStr, py, initials)

			// 获取并更新版本详情
			detailTrainer, versions, err := s.adapter.GetDetail(ctx, item.DetailURL)
			if err == nil && detailTrainer != nil {
				detailTrainer.GameID = upsertedGame.ID
				detailTrainer.TitleCN = gameNameCN + " 修改器"
				_ = s.trainerRepo.UpsertTrainerWithVersions(ctx, detailTrainer, versions)
				updatedCount++
			}
		}
	}

	rec.Status = model.SyncStatusSuccess
	rec.ItemsAdded = addedCount
	rec.ItemsUpdated = updatedCount
	_ = s.syncRepo.UpdateRecord(ctx, rec)

	return rec, nil
}

// MarkTrainerTakedown 标记指定来源的修改器已下架
func (s *SyncService) MarkTrainerTakedown(ctx context.Context, externalID string) error {
	logger.Info("Marking trainer as delisted/takedown", "source", s.adapter.ID(), "external_id", externalID)
	return s.trainerRepo.MarkAsDeleted(ctx, s.adapter.ID(), externalID)
}

// StartPeriodicSync 启动后台定期增量同步
func (s *SyncService) StartPeriodicSync(intervalHours int) {
	if intervalHours <= 0 {
		intervalHours = 24
	}

	go func() {
		ticker := time.NewTicker(time.Duration(intervalHours) * time.Hour)
		defer ticker.Stop()

		for {
			select {
			case <-s.stopChan:
				return
			case <-ticker.C:
				logger.Info("Starting periodic background sync...")
				_, _ = s.SyncLatest(context.Background(), 2)
			}
		}
	}()
}

// Stop 停止后台同步
func (s *SyncService) Stop() {
	close(s.stopChan)
}

func slugify(name string) string {
	lower := strings.ToLower(name)
	var sb strings.Builder
	for _, r := range lower {
		if (r >= 'a' && r <= 'z') || (r >= '0' && r <= '9') {
			sb.WriteRune(r)
		} else if sb.Len() > 0 && !strings.HasSuffix(sb.String(), "-") {
			sb.WriteRune('-')
		}
	}
	return strings.Trim(sb.String(), "-")
}

