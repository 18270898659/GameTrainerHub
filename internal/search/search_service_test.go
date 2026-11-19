package search

import (
	"context"
	"errors"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"GameTrainerHub/internal/database"
	"GameTrainerHub/internal/database/fts"
	"GameTrainerHub/internal/database/repository"
	"GameTrainerHub/internal/model"
	"GameTrainerHub/internal/source"
	"GameTrainerHub/internal/translate"
)

// MockAdapter 用于单元测试的 Mock 网站适配器
type MockAdapter struct {
	id         string
	name       string
	mockItems  []*model.SearchResult
	shouldFail bool
}

func (m *MockAdapter) ID() string      { return m.id }
func (m *MockAdapter) Name() string    { return m.name }
func (m *MockAdapter) BaseURL() string { return "https://mock.com" }

func (m *MockAdapter) Search(ctx context.Context, keyword string, page int) ([]*model.SearchResult, error) {
	if m.shouldFail {
		return nil, errors.New("network timeout")
	}
	var res []*model.SearchResult
	for _, it := range m.mockItems {
		if strings.Contains(strings.ToLower(it.GameNameEN), strings.ToLower(keyword)) ||
			strings.Contains(strings.ToLower(it.TrainerTitle), strings.ToLower(keyword)) {
			res = append(res, it)
		}
	}
	return res, nil
}

func (m *MockAdapter) GetDetail(ctx context.Context, detailURL string) (*model.Trainer, []*model.TrainerVersion, error) {
	return nil, nil, nil
}
func (m *MockAdapter) GetLatest(ctx context.Context, page int) ([]*model.SearchResult, error) {
	return m.mockItems, nil
}
func (m *MockAdapter) GetCatalog(ctx context.Context) ([]*source.CatalogItem, error) {
	return nil, nil
}
func (m *MockAdapter) ResolveDownload(ctx context.Context, downloadURL string, referer string) (*source.DownloadInfo, error) {
	return &source.DownloadInfo{FinalURL: downloadURL, FileName: "test.zip", SupportsRange: true}, nil
}

func setupSearchTest(t *testing.T) (*SearchService, *database.DB, func()) {
	tempDir, err := os.MkdirTemp("", "gth_search_test_*")
	if err != nil {
		t.Fatalf("failed to create temp dir: %v", err)
	}

	dbPath := filepath.Join(tempDir, "test.db")
	db, err := database.Open(dbPath)
	if err != nil {
		t.Fatalf("failed to open database: %v", err)
	}

	ftsRepo := fts.NewFTSRepository(db)
	gameRepo := repository.NewGameRepository(db)
	trainerRepo := repository.NewTrainerRepository(db)
	trans := translate.NewTranslator()

	svc := NewSearchService(ftsRepo, gameRepo, trainerRepo, trans)

	cleanup := func() {
		_ = db.Close()
		_ = os.RemoveAll(tempDir)
	}
	return svc, db, cleanup
}

func TestSearchServiceHybridSearch(t *testing.T) {
	svc, db, cleanup := setupSearchTest(t)
	defer cleanup()

	ctx := context.Background()
	gameRepo := repository.NewGameRepository(db)
	ftsRepo := fts.NewFTSRepository(db)

	// 1. 本地入库游戏
	game := &model.Game{
		Slug:     "cyberpunk-2077",
		NameEN:   "Cyberpunk 2077",
		NameCN:   "赛博朋克2077",
		CoverURL: "https://example.com/cover.jpg",
	}
	_ = gameRepo.Create(ctx, game)
	_ = ftsRepo.IndexGame(ctx, game.ID, game.NameEN, game.NameCN, "2077", "saibopengke", "sbpk")

	// 2. 注册 Mock 在线适配器（带有在网但本地未下载的 Elden Ring）
	mockAdp := &MockAdapter{
		id:   "fling",
		name: "FLiNG Trainer",
		mockItems: []*model.SearchResult{
			{
				Slug:         "elden-ring",
				GameNameEN:   "ELDEN RING",
				TrainerTitle: "ELDEN RING Trainer",
				DetailURL:    "https://flingtrainer.com/trainer/elden-ring/",
				SourceID:     "fling",
				IsOnlineOnly: true,
			},
		},
	}
	svc.RegisterAdapter(mockAdp)

	// 测试场景 1: 用中文 "赛博朋克" 搜索本地收录游戏
	resCN, err := svc.Search(ctx, &model.SearchFilter{Query: "赛博朋克", Page: 1, PageSize: 10})
	if err != nil {
		t.Fatalf("Search failed: %v", err)
	}
	if len(resCN) == 0 {
		t.Fatal("expected at least 1 result for '赛博朋克'")
	}
	if resCN[0].Slug != "cyberpunk-2077" {
		t.Errorf("expected first result 'cyberpunk-2077', got '%s'", resCN[0].Slug)
	}

	// 测试场景 2: 用拼音缩写 "sbpk" 搜索
	resPinyin, err := svc.Search(ctx, &model.SearchFilter{Query: "sbpk", Page: 1, PageSize: 10})
	if err != nil {
		t.Fatalf("Search failed: %v", err)
	}
	if len(resPinyin) == 0 || resPinyin[0].Slug != "cyberpunk-2077" {
		t.Fatalf("expected to find cyberpunk with 'sbpk', got %v", resPinyin)
	}

	// 测试场景 3: 用中文昵称 "老头环" 检索在线游戏 (通过中英翻译引擎与适配器)
	resOnline, err := svc.Search(ctx, &model.SearchFilter{Query: "老头环", Page: 1, PageSize: 10})
	if err != nil {
		t.Fatalf("Search failed: %v", err)
	}
	if len(resOnline) == 0 {
		t.Fatal("expected online result for '老头环' (mapped to ELDEN RING)")
	}
	if resOnline[0].Slug != "elden-ring" {
		t.Errorf("expected 'elden-ring', got '%s'", resOnline[0].Slug)
	}
	if resOnline[0].GameNameCN != "艾尔登法环" {
		t.Errorf("expected auto-translated CN name '艾尔登法环', got '%s'", resOnline[0].GameNameCN)
	}

	// 测试场景 4: 原网网络异常时，优雅降级，本地搜索不受阻
	mockAdp.shouldFail = true
	resFallback, err := svc.Search(ctx, &model.SearchFilter{Query: "2077", Page: 1, PageSize: 10})
	if err != nil {
		t.Fatalf("fallback search should not return error, got: %v", err)
	}
	if len(resFallback) == 0 || resFallback[0].Slug != "cyberpunk-2077" {
		t.Fatalf("expected local results when online search fails, got %v", resFallback)
	}
}

func TestSearchServiceMultiSourceAggregation(t *testing.T) {
	svc, _, cleanup := setupSearchTest(t)
	defer cleanup()

	ctx := context.Background()

	// 注册 2 个不同的数据源适配器
	adpFling := &MockAdapter{
		id:   "fling",
		name: "FLiNG Trainer",
		mockItems: []*model.SearchResult{
			{
				Slug:         "sekiro",
				GameNameEN:   "Sekiro: Shadows Die Twice",
				TrainerTitle: "Sekiro FLiNG Trainer v1.06",
				DetailURL:    "https://flingtrainer.com/trainer/sekiro/",
				SourceID:     "fling",
				IsOnlineOnly: true,
			},
		},
	}
	adpWeMod := &MockAdapter{
		id:   "wemod",
		name: "WeMod",
		mockItems: []*model.SearchResult{
			{
				Slug:         "sekiro",
				GameNameEN:   "Sekiro: Shadows Die Twice",
				TrainerTitle: "Sekiro WeMod Trainer",
				DetailURL:    "https://wemod.com/cheats/sekiro",
				SourceID:     "wemod",
				IsOnlineOnly: true,
			},
		},
	}

	svc.RegisterAdapter(adpFling)
	svc.RegisterAdapter(adpWeMod)

	// 搜索 "sekiro"，两个数据源的修改器都应被检索出，不会因 slug 相同而被丢弃
	results, err := svc.Search(ctx, &model.SearchFilter{Query: "sekiro", Page: 1, PageSize: 10})
	if err != nil {
		t.Fatalf("Search failed: %v", err)
	}

	if len(results) != 2 {
		t.Fatalf("expected 2 results from 2 sources for sekiro, got %d", len(results))
	}

	sources := make(map[string]bool)
	for _, res := range results {
		sources[res.SourceID] = true
	}
	if !sources["fling"] || !sources["wemod"] {
		t.Errorf("expected both fling and wemod sources, got: %+v", sources)
	}
}

