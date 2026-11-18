package performance

import (
	"context"
	"database/sql"
	"fmt"
	"os"
	"path/filepath"
	"runtime"
	"testing"
	"time"

	"GameTrainerHub/internal/database"
	"GameTrainerHub/internal/database/fts"
	"GameTrainerHub/internal/database/repository"
	"GameTrainerHub/internal/model"
	"GameTrainerHub/internal/search"
	"GameTrainerHub/internal/translate"
)

func TestPerformanceMetrics(t *testing.T) {
	tempDir, err := os.MkdirTemp("", "gth_perf_*")
	if err != nil {
		t.Fatalf("failed to create temp dir: %v", err)
	}
	defer os.RemoveAll(tempDir)

	dbPath := filepath.Join(tempDir, "perf.db")

	// 1. 测量冷启动与数据库初始化时间
	startInit := time.Now()
	db, err := database.Open(dbPath)
	if err != nil {
		t.Fatalf("failed to open database: %v", err)
	}
	defer db.Close()
	initDuration := time.Since(startInit)
	t.Logf("[METRIC] Cold Startup & SQLite WAL Initialization: %v", initDuration)

	ctx := context.Background()
	gameRepo := repository.NewGameRepository(db)
	trainerRepo := repository.NewTrainerRepository(db)
	ftsRepo := fts.NewFTSRepository(db)
	translator := translate.NewTranslator()
	searchSvc := search.NewSearchService(ftsRepo, gameRepo, trainerRepo, translator)

	// 2. 测量大规模数据批量写入性能（1,000 条游戏与 FTS 索引记录）
	numGames := 1000
	startInsert := time.Now()
	_ = db.WithTx(ctx, func(tx *sql.Tx) error {
		return nil
	})

	for i := 1; i <= numGames; i++ {
		slug := fmt.Sprintf("game-benchmark-%d", i)
		nameEN := fmt.Sprintf("Benchmark Game %d", i)
		nameCN := fmt.Sprintf("性能基准测试游戏 %d", i)
		g := &model.Game{Slug: slug, NameEN: nameEN, NameCN: nameCN}
		_ = gameRepo.Create(ctx, g)

		py := translator.ToPinyin(nameCN)
		initials := translator.ToPinyinInitials(nameCN)
		_ = ftsRepo.IndexGame(ctx, g.ID, nameEN, nameCN, "", py, initials)

		tr := &model.Trainer{
			GameID:     g.ID,
			SourceID:   "fling",
			ExternalID: fmt.Sprintf("perf_ext_%d", i),
			TitleEN:    nameEN + " Trainer",
			TitleCN:    nameCN + " 修改器",
			DetailURL:  "https://example.com/trainer/" + slug,
			UpdatedAt:  time.Now(),
		}
		_ = trainerRepo.CreateTrainer(ctx, tr)
	}
	insertDuration := time.Since(startInsert)
	t.Logf("[METRIC] Batch insert & FTS indexing of %d records: %v (avg %.2f ms/record)",
		numGames, insertDuration, float64(insertDuration.Milliseconds())/float64(numGames))

	// 3. 测量不同关键词的本地毫秒级搜索响应时间
	searchQueries := []string{
		"Benchmark Game 500",
		"性能基准测试游戏 888",
		"xingnengjizhun",
		"xnjz",
		"Game 999",
	}

	for _, q := range searchQueries {
		startSearch := time.Now()
		res, err := searchSvc.Search(ctx, &model.SearchFilter{Query: q, LocalOnly: true})
		searchDuration := time.Since(startSearch)
		if err != nil {
			t.Fatalf("search '%s' failed: %v", q, err)
		}
		t.Logf("[METRIC] Local FTS Search for '%s': %v (matched %d results)", q, searchDuration, len(res))
	}

	// 4. 测量 SQLite 单条与聚合读取延迟
	startQuery := time.Now()
	for i := 0; i < 100; i++ {
		_, _ = gameRepo.GetByID(ctx, int64(i+1))
	}
	query100Duration := time.Since(startQuery)
	t.Logf("[METRIC] 100 individual primary key reads: %v (avg %.2f us/op)",
		query100Duration, float64(query100Duration.Microseconds())/100.0)

	// 5. 测量内存开销 (Runtime MemStats)
	var m runtime.MemStats
	runtime.ReadMemStats(&m)
	t.Logf("[METRIC] Memory Alloc: %.2f MB, TotalAlloc: %.2f MB, Sys: %.2f MB, NumGC: %d",
		float64(m.Alloc)/1024/1024, float64(m.TotalAlloc)/1024/1024, float64(m.Sys)/1024/1024, m.NumGC)
}
