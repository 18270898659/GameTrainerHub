package search

import (
	"context"
	"sort"
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

type SearchService struct {
	ftsRepo     *fts.FTSRepository
	gameRepo    *repository.GameRepository
	trainerRepo *repository.TrainerRepository
	translator  *translate.Translator
	adapters    map[string]source.SourceAdapter
	mu          sync.RWMutex
}

func NewSearchService(
	ftsRepo *fts.FTSRepository,
	gameRepo *repository.GameRepository,
	trainerRepo *repository.TrainerRepository,
	translator *translate.Translator,
) *SearchService {
	if translator == nil {
		translator = translate.NewTranslator()
	}
	return &SearchService{
		ftsRepo:     ftsRepo,
		gameRepo:    gameRepo,
		trainerRepo: trainerRepo,
		translator:  translator,
		adapters:    make(map[string]source.SourceAdapter),
	}
}

// RegisterAdapter 注册资源网站适配器
func (s *SearchService) RegisterAdapter(adapter source.SourceAdapter) {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.adapters[adapter.ID()] = adapter
}

// Search 执行本地与在线混合搜索
func (s *SearchService) Search(ctx context.Context, filter *model.SearchFilter) ([]*model.SearchResult, error) {
	cleanQuery := strings.TrimSpace(filter.Query)
	if cleanQuery == "" {
		return nil, nil
	}

	// 1. 关键词语义与跨语言扩展
	expandedTerms := s.translator.ExpandQueryTerms(cleanQuery)
	logger.Debug("Search terms expanded", "raw", cleanQuery, "expanded", expandedTerms)

	resultMap := make(map[string]*model.SearchResult)
	var mapMu sync.Mutex

	// 2. 本地 SQLite FTS5 检索
	localMatches, err := s.searchLocal(ctx, cleanQuery, expandedTerms)
	if err != nil {
		logger.Warn("Local FTS search error", "err", err)
	} else {
		for _, item := range localMatches {
			resultMap[item.Slug] = item
		}
	}

	// 3. 若非仅搜本地，并发启动在线原网站检索
	if !filter.LocalOnly {
		var wg sync.WaitGroup
		onlineCtx, cancel := context.WithTimeout(ctx, 8*time.Second)
		defer cancel()

		s.mu.RLock()
		activeAdapters := make([]source.SourceAdapter, 0, len(s.adapters))
		for _, adp := range s.adapters {
			if filter.Source == "" || filter.Source == "all" || filter.Source == adp.ID() {
				activeAdapters = append(activeAdapters, adp)
			}
		}
		s.mu.RUnlock()

		// 决定送往原网检索的关键词集：原网主要为英文，优先使用翻译后的英文候选词或原始词
		onlineKeywords := s.extractOnlineKeywords(cleanQuery, expandedTerms)

		for _, adp := range activeAdapters {
			for _, kw := range onlineKeywords {
				wg.Add(1)
				go func(adapter source.SourceAdapter, keyword string) {
					defer wg.Done()

					items, searchErr := adapter.Search(onlineCtx, keyword, 1)
					if searchErr != nil {
						logger.Warn("Online search adapter error", "adapter", adapter.ID(), "keyword", keyword, "err", searchErr)
						return
					}

					mapMu.Lock()
					defer mapMu.Unlock()
					for _, item := range items {
						if existing, ok := resultMap[item.Slug]; ok {
							// 本地已有该游戏，保留本地的下载状态和中文名，补充在线信息
							if existing.CoverURL == "" && item.CoverURL != "" {
								existing.CoverURL = item.CoverURL
							}
							if existing.TrainerTitle == "" {
								existing.TrainerTitle = item.TrainerTitle
							}
						} else {
							// 自动推导并填充中文名
							cnCandidates := s.translator.TranslateEnglishToChinese(item.GameNameEN)
							if len(cnCandidates) > 0 {
								item.GameNameCN = cnCandidates[0]
							}
							resultMap[item.Slug] = item
						}
					}
				}(adp, kw)
			}
		}

		wg.Wait()
	}

	// 4. 计算相关度打分与排序
	var finalResults []*model.SearchResult
	for _, item := range resultMap {
		item.Score = s.calculateRelevance(item, cleanQuery, expandedTerms)
		finalResults = append(finalResults, item)
	}

	sort.Slice(finalResults, func(i, j int) bool {
		return finalResults[i].Score > finalResults[j].Score
	})

	// 分页截取
	if filter.PageSize > 0 {
		start := (filter.Page - 1) * filter.PageSize
		if start < 0 {
			start = 0
		}
		if start >= len(finalResults) {
			return []*model.SearchResult{}, nil
		}
		end := start + filter.PageSize
		if end > len(finalResults) {
			end = len(finalResults)
		}
		return finalResults[start:end], nil
	}

	return finalResults, nil
}

// searchLocal 检索本地库并拼装修改器信息
func (s *SearchService) searchLocal(ctx context.Context, rawQuery string, expandedTerms []string) ([]*model.SearchResult, error) {
	var matchedGameIDs []int64
	seenIDs := make(map[int64]bool)

	// 对每一个扩展词进行 FTS5 匹配
	for _, term := range expandedTerms {
		ids, err := s.ftsRepo.Search(ctx, term, 20)
		if err == nil {
			for _, id := range ids {
				if !seenIDs[id] {
					seenIDs[id] = true
					matchedGameIDs = append(matchedGameIDs, id)
				}
			}
		}
	}

	var results []*model.SearchResult
	for _, gameID := range matchedGameIDs {
		game, err := s.gameRepo.GetByID(ctx, gameID)
		if err != nil || game == nil {
			continue
		}

		// 查询修改器及下载状态
		trainers, _ := s.trainerRepo.ListTrainersByGameID(ctx, gameID)
		var isDl bool
		var trainerTitle, versionName, detailURL string
		var updatedAt = game.UpdatedAt

		if len(trainers) > 0 {
			tr := trainers[0]
			trainerTitle = tr.TitleEN
			detailURL = tr.DetailURL
			updatedAt = tr.UpdatedAt
			if len(tr.Versions) > 0 {
				versionName = tr.Versions[0].VersionName
				for _, v := range tr.Versions {
					if v.IsDownloaded {
						isDl = true
						break
					}
				}
			}
		}

		results = append(results, &model.SearchResult{
			GameID:       game.ID,
			Slug:         game.Slug,
			GameNameEN:   game.NameEN,
			GameNameCN:   game.NameCN,
			CoverURL:     game.CoverURL,
			SourceID:     "local",
			TrainerTitle: trainerTitle,
			DetailURL:    detailURL,
			VersionName:  versionName,
			UpdatedAt:    updatedAt,
			IsDownloaded: isDl,
			IsOnlineOnly: false,
		})
	}

	return results, nil
}

// extractOnlineKeywords 提取适合向原网发起的英文/规范化搜索关键词
func (s *SearchService) extractOnlineKeywords(rawQuery string, expandedTerms []string) []string {
	var kws []string
	seen := make(map[string]bool)

	add := func(w string) {
		w = strings.TrimSpace(w)
		if w != "" && !seen[strings.ToLower(w)] {
			seen[strings.ToLower(w)] = true
			kws = append(kws, w)
		}
	}

	// 如果包含中文，英译候选集排在最前
	if s.translator.IsChinese(rawQuery) {
		enCandidates := s.translator.TranslateChineseToEnglish(rawQuery)
		for _, en := range enCandidates {
			add(en)
		}
	} else {
		add(rawQuery)
	}

	// 限制最多 2 个关键词发起并发在线请求，避免对源站造成压力
	if len(kws) > 2 {
		return kws[:2]
	}
	return kws
}

// calculateRelevance 计算相关度打分
func (s *SearchService) calculateRelevance(item *model.SearchResult, query string, terms []string) float64 {
	score := 0.0
	lowerQuery := strings.ToLower(query)
	lowerEN := strings.ToLower(item.GameNameEN)
	lowerCN := strings.ToLower(item.GameNameCN)

	// 完全匹配
	if lowerEN == lowerQuery || lowerCN == lowerQuery {
		score += 100.0
	} else if strings.Contains(lowerEN, lowerQuery) || strings.Contains(lowerCN, lowerQuery) {
		score += 60.0
	}

	// 词项匹配加分
	for _, term := range terms {
		lt := strings.ToLower(term)
		if lt == lowerEN || lt == lowerCN {
			score += 40.0
		} else if strings.Contains(lowerEN, lt) || strings.Contains(lowerCN, lt) {
			score += 20.0
		}
	}

	// 本地收录与已下载加分
	if !item.IsOnlineOnly {
		score += 15.0
	}
	if item.IsDownloaded {
		score += 10.0
	}

	return score
}
