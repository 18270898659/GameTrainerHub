package fts

import (
	"context"
	"os"
	"path/filepath"
	"testing"

	"GameTrainerHub/internal/database"
)

func setupTestDB(t *testing.T) (*database.DB, func()) {
	tempDir, err := os.MkdirTemp("", "gth_fts_test_*")
	if err != nil {
		t.Fatalf("failed to create temp dir: %v", err)
	}

	dbPath := filepath.Join(tempDir, "test.db")
	db, err := database.Open(dbPath)
	if err != nil {
		t.Fatalf("failed to open database: %v", err)
	}

	cleanup := func() {
		_ = db.Close()
		_ = os.RemoveAll(tempDir)
	}
	return db, cleanup
}

func TestFTSSearchScenarios(t *testing.T) {
	db, cleanup := setupTestDB(t)
	defer cleanup()

	ftsRepo := NewFTSRepository(db)
	ctx := context.Background()

	// 准备测试数据:
	// 1. Cyberpunk 2077
	err := ftsRepo.IndexGame(ctx, 101, "Cyberpunk 2077", "赛博朋克2077", "2077 夜之城", "saibopengke", "sbpk")
	if err != nil {
		t.Fatalf("failed to index Cyberpunk: %v", err)
	}

	// 2. ELDEN RING
	err = ftsRepo.IndexGame(ctx, 102, "ELDEN RING", "艾尔登法环", "老头环 法环", "aierdengfahuan", "aedfh")
	if err != nil {
		t.Fatalf("failed to index Elden Ring: %v", err)
	}

	// 3. Red Dead Redemption 2
	err = ftsRepo.IndexGame(ctx, 103, "Red Dead Redemption 2", "荒野大镖客2", "大镖客2 RDR2", "huangyedabiaoke", "hydbk")
	if err != nil {
		t.Fatalf("failed to index RDR2: %v", err)
	}

	// 测试场景 1: 英文全名及前缀检索
	res, err := ftsRepo.Search(ctx, "Cyberpunk", 10)
	if err != nil || len(res) == 0 || res[0] != 101 {
		t.Fatalf("expected to find game 101 with 'Cyberpunk', got %v, err: %v", res, err)
	}

	resPrefix, err := ftsRepo.Search(ctx, "Cyber", 10)
	if err != nil || len(resPrefix) == 0 || resPrefix[0] != 101 {
		t.Fatalf("expected to find game 101 with prefix 'Cyber', got %v, err: %v", resPrefix, err)
	}

	// 测试场景 2: 中文检索
	resCN, err := ftsRepo.Search(ctx, "艾尔登法环", 10)
	if err != nil || len(resCN) == 0 || resCN[0] != 102 {
		t.Fatalf("expected to find game 102 with '艾尔登法环', got %v, err: %v", resCN, err)
	}

	// 测试场景 3: 别名检索 (如 "老头环")
	resAlias, err := ftsRepo.Search(ctx, "老头环", 10)
	if err != nil || len(resAlias) == 0 || resAlias[0] != 102 {
		t.Fatalf("expected to find game 102 with alias '老头环', got %v, err: %v", resAlias, err)
	}

	// 测试场景 4: 拼音全拼检索 (如 "saibopengke")
	resPinyin, err := ftsRepo.Search(ctx, "saibopengke", 10)
	if err != nil || len(resPinyin) == 0 || resPinyin[0] != 101 {
		t.Fatalf("expected to find game 101 with pinyin 'saibopengke', got %v, err: %v", resPinyin, err)
	}

	// 测试场景 5: 拼音首字母检索 (如 "hydbk")
	resInitials, err := ftsRepo.Search(ctx, "hydbk", 10)
	if err != nil || len(resInitials) == 0 || resInitials[0] != 103 {
		t.Fatalf("expected to find game 103 with initials 'hydbk', got %v, err: %v", resInitials, err)
	}

	// 测试场景 6: 删除索引
	if err := ftsRepo.RemoveGameIndex(ctx, 101); err != nil {
		t.Fatalf("failed to remove index: %v", err)
	}
	resAfterDel, err := ftsRepo.Search(ctx, "Cyberpunk", 10)
	if err != nil || len(resAfterDel) != 0 {
		t.Fatalf("expected empty search result after deletion, got %v", resAfterDel)
	}
}

