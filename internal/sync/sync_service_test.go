package sync

import (
	"context"
	"os"
	"path/filepath"
	"testing"

	"GameTrainerHub/internal/database"
	"GameTrainerHub/internal/database/fts"
	"GameTrainerHub/internal/database/repository"
	"GameTrainerHub/internal/model"
	"GameTrainerHub/internal/source"
	"GameTrainerHub/internal/translate"
)

type MockSyncAdapter struct {
	catalogItems []*source.CatalogItem
	latestItems  []*model.SearchResult
}

func (m *MockSyncAdapter) ID() string      { return "fling" }
func (m *MockSyncAdapter) Name() string    { return "FLiNG Trainer" }
func (m *MockSyncAdapter) BaseURL() string { return "https://flingtrainer.com" }
func (m *MockSyncAdapter) Search(ctx context.Context, keyword string, page int) ([]*model.SearchResult, error) {
	return nil, nil
}
func (m *MockSyncAdapter) GetDetail(ctx context.Context, detailURL string) (*model.Trainer, []*model.TrainerVersion, error) {
	tr := &model.Trainer{SourceID: "fling", ExternalID: "test-tr", TitleEN: "Test Trainer"}
	ver := &model.TrainerVersion{VersionName: "v1.0", DownloadURL: "https://dl.com"}
	return tr, []*model.TrainerVersion{ver}, nil
}
func (m *MockSyncAdapter) GetLatest(ctx context.Context, page int) ([]*model.SearchResult, error) {
	return m.latestItems, nil
}
func (m *MockSyncAdapter) GetCatalog(ctx context.Context) ([]*source.CatalogItem, error) {
	return m.catalogItems, nil
}
func (m *MockSyncAdapter) ResolveDownload(ctx context.Context, downloadURL string, referer string) (*source.DownloadInfo, error) {
	return nil, nil
}

func TestSyncServiceCatalogAndLatest(t *testing.T) {
	tempDir, err := os.MkdirTemp("", "gth_sync_test_*")
	if err != nil {
		t.Fatalf("failed to create temp dir: %v", err)
	}
	defer os.RemoveAll(tempDir)

	dbPath := filepath.Join(tempDir, "test.db")
	db, err := database.Open(dbPath)
	if err != nil {
		t.Fatalf("failed to open database: %v", err)
	}
	defer db.Close()

	ctx := context.Background()
	gameRepo := repository.NewGameRepository(db)
	trainerRepo := repository.NewTrainerRepository(db)
	syncRepo := repository.NewSyncRepository(db)
	ftsRepo := fts.NewFTSRepository(db)
	trans := translate.NewTranslator()

	mockAdp := &MockSyncAdapter{
		catalogItems: []*source.CatalogItem{
			{Title: "Cyberpunk 2077 Trainer", ExternalID: "cp2077-tr", DetailURL: "https://fling.com/cp"},
			{Title: "ELDEN RING Trainer", ExternalID: "elden-tr", DetailURL: "https://fling.com/elden"},
		},
		latestItems: []*model.SearchResult{
			{GameNameEN: "Cyberpunk 2077", Slug: "cyberpunk-2077", TrainerTitle: "Cyberpunk 2077 Trainer", DetailURL: "https://fling.com/cp"},
		},
	}

	syncSvc := NewSyncService(gameRepo, trainerRepo, syncRepo, ftsRepo, trans, mockAdp)

	// 1. 测试 Catalog 同步
	rec, err := syncSvc.SyncCatalog(ctx)
	if err != nil {
		t.Fatalf("SyncCatalog failed: %v", err)
	}
	if rec.ItemsAdded != 2 {
		t.Errorf("expected 2 items added, got %d", rec.ItemsAdded)
	}

	// 2. 验证 FTS 索引中包含了自动翻译的中文名
	cpGame, err := gameRepo.GetBySlug(ctx, "cyberpunk-2077")
	if err != nil || cpGame == nil {
		t.Fatalf("expected cyberpunk in db: %v", err)
	}
	if cpGame.NameCN != "赛博朋克2077" {
		t.Errorf("expected auto translated CN name '赛博朋克2077', got '%s'", cpGame.NameCN)
	}

	// 3. 测试 Latest 增量同步
	recLatest, err := syncSvc.SyncLatest(ctx, 1)
	if err != nil {
		t.Fatalf("SyncLatest failed: %v", err)
	}
	if recLatest.ItemsUpdated != 1 {
		t.Errorf("expected 1 item updated, got %d", recLatest.ItemsUpdated)
	}

	// 4. 验证 SyncLatest 中维护的 FTS 别名和拼音索引可正常召回
	ftsHits, err := ftsRepo.Search(ctx, "saibopengke", 10)
	if err != nil || len(ftsHits) == 0 {
		t.Fatalf("expected fts search for pinyin 'saibopengke' to hit after SyncLatest, err: %v", err)
	}

	// 5. 验证资源下架标记 (MarkTrainerTakedown)
	err = syncSvc.MarkTrainerTakedown(ctx, "test-tr")
	if err != nil {
		t.Fatalf("MarkTrainerTakedown failed: %v", err)
	}
	tr, err := trainerRepo.GetTrainerByExternalID(ctx, "fling", "test-tr")
	if err != nil || tr == nil {
		t.Fatalf("failed to query takedown trainer: %v", err)
	}
	if !tr.IsDeleted {
		t.Errorf("expected trainer is_deleted to be true after takedown, got %v", tr.IsDeleted)
	}
}

func TestSeedAndFallback(t *testing.T) {
	latest, err := LoadSeedLatest()
	if err != nil {
		t.Fatalf("LoadSeedLatest failed: %v", err)
	}
	t.Logf("LoadSeedLatest returned %d items", len(latest))
	if len(latest) == 0 {
		t.Fatal("expected non-empty seed latest")
	}

	catalog, err := LoadSeedCatalog()
	if err != nil {
		t.Fatalf("LoadSeedCatalog failed: %v", err)
	}
	t.Logf("LoadSeedCatalog returned %d items", len(catalog))
	if len(catalog) < 760 {
		t.Fatalf("expected >= 760 items, got %d", len(catalog))
	}
}

