package search

import (
	"context"
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"

	"GameTrainerHub/internal/database"
	"GameTrainerHub/internal/database/fts"
	"GameTrainerHub/internal/database/repository"
	"GameTrainerHub/internal/model"
	"GameTrainerHub/internal/translate"
)

func setupTestSearchEnvironment(t *testing.T) (*SearchService, func()) {
	tempDir, err := os.MkdirTemp("", "gth_search_edge_*")
	if err != nil {
		t.Fatalf("failed to create temp dir: %v", err)
	}

	dbPath := filepath.Join(tempDir, "test.db")
	db, err := database.Open(dbPath)
	if err != nil {
		t.Fatalf("failed to open database: %v", err)
	}

	ctx := context.Background()
	gameRepo := repository.NewGameRepository(db)
	trainerRepo := repository.NewTrainerRepository(db)
	ftsRepo := fts.NewFTSRepository(db)
	translator := translate.NewTranslator()

	// 注入种子测试游戏数据
	games := []*model.Game{
		{Slug: "cyberpunk-2077", NameEN: "Cyberpunk 2077", NameCN: "赛博朋克 2077", CoverURL: "https://example.com/cp.jpg"},
		{Slug: "elden-ring", NameEN: "ELDEN RING", NameCN: "艾尔登法环", CoverURL: "https://example.com/er.jpg"},
		{Slug: "red-dead-redemption-2", NameEN: "Red Dead Redemption 2", NameCN: "荒野大镖客：救赎 2", CoverURL: "https://example.com/rdr2.jpg"},
		{Slug: "black-myth-wukong", NameEN: "Black Myth: Wukong", NameCN: "黑神话：悟空", CoverURL: "https://example.com/bmw.jpg"},
		{Slug: "monster-hunter-world", NameEN: "Monster Hunter: World", NameCN: "怪物猎人：世界", CoverURL: "https://example.com/mhw.jpg"},
		{Slug: "resident-evil-4", NameEN: "Resident Evil 4", NameCN: "生化危机 4：重制版", CoverURL: "https://example.com/re4.jpg"},
	}

	for _, g := range games {
		_ = gameRepo.Create(ctx, g)

		// 维护 FTS5 索引与全拼/声母/别名
		aliases := translator.TranslateEnglishToChinese(g.NameEN)
		aliasStr := strings.Join(aliases, " ")
		var pinyins []string
		var initialsList []string
		if g.NameCN != "" {
			pinyins = append(pinyins, translator.ToPinyin(g.NameCN))
			initialsList = append(initialsList, translator.ToPinyinInitials(g.NameCN))
		}
		for _, al := range aliases {
			pinyins = append(pinyins, translator.ToPinyin(al))
			initialsList = append(initialsList, translator.ToPinyinInitials(al))
		}
		py := strings.Join(pinyins, " ")
		initials := strings.Join(initialsList, " ")
		_ = ftsRepo.IndexGame(ctx, g.ID, g.NameEN, g.NameCN, aliasStr, py, initials)

		// 插入修改器主记录
		tr := &model.Trainer{
			GameID:    g.ID,
			SourceID:  "fling",
			TitleEN:   g.NameEN + " Trainer",
			TitleCN:   g.NameCN + " 修改器",
			DetailURL: "https://flingtrainer.com/trainer/" + g.Slug,
			UpdatedAt: time.Now(),
		}
		_ = trainerRepo.CreateTrainer(ctx, tr)

		// 插入修改器版本
		_ = trainerRepo.CreateVersion(ctx, &model.TrainerVersion{
			TrainerID:         tr.ID,
			VersionName:       "v1.0 Plus 25",
			TargetGameVersion: "v1.0+",
			DownloadURL:       "https://flingtrainer.com/download/" + g.Slug,
		})
	}

	svc := NewSearchService(ftsRepo, gameRepo, trainerRepo, translator)

	cleanup := func() {
		_ = db.Close()
		_ = os.RemoveAll(tempDir)
	}

	return svc, cleanup
}

func TestSearchComprehensiveScenarios(t *testing.T) {
	svc, cleanup := setupTestSearchEnvironment(t)
	defer cleanup()

	ctx := context.Background()

	cases := []struct {
		name         string
		keyword      string
		expectedGame string
		minResults   int
	}{
		{"中文精确名称", "赛博朋克 2077", "Cyberpunk 2077", 1},
		{"英文精确名称", "Cyberpunk 2077", "Cyberpunk 2077", 1},
		{"中文俗称/别名1", "老头环", "ELDEN RING", 1},
		{"中文俗称/别名2", "黑猴", "Black Myth: Wukong", 1},
		{"中文俗称/别名3", "怪猎世界", "Monster Hunter: World", 1},
		{"中文俗称/别名4", "大镖客2", "Red Dead Redemption 2", 1},
		{"英文别名/缩写1", "rdr2", "Red Dead Redemption 2", 1},
		{"拼音全拼", "saibopengke", "Cyberpunk 2077", 1},
		{"拼音缩写声母", "sbpk", "Cyberpunk 2077", 1},
		{"拼音全拼老头环", "laotouhuan", "ELDEN RING", 1},
		{"大小写混合乱序", "cYbErPuNk", "Cyberpunk 2077", 1},
		{"模糊部分关键词", "朋克", "Cyberpunk 2077", 1},
		{"模糊法环", "法环", "ELDEN RING", 1},
		{"英文大小写不敏感", "elden ring", "ELDEN RING", 1},
		{"复合关键词", "艾尔登法环 修改器", "ELDEN RING", 1},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			filter := &model.SearchFilter{Query: tc.keyword, LocalOnly: true}
			results, err := svc.Search(ctx, filter)
			if err != nil {
				t.Fatalf("Search failed for keyword '%s': %v", tc.keyword, err)
			}
			if len(results) < tc.minResults {
				t.Fatalf("Expected at least %d results for '%s', got %d", tc.minResults, tc.keyword, len(results))
			}

			found := false
			for _, r := range results {
				if r.GameNameEN == tc.expectedGame || r.GameNameCN == tc.expectedGame || strings.Contains(r.TrainerTitle, tc.expectedGame) {
					found = true
					break
				}
			}
			if !found {
				t.Errorf("Keyword '%s' did not find expected game '%s'. Got: %+v", tc.keyword, tc.expectedGame, results[0])
			}
		})
	}
}

func TestSearchAbnormalAndEdgeCases(t *testing.T) {
	svc, cleanup := setupTestSearchEnvironment(t)
	defer cleanup()

	ctx := context.Background()

	// 1. 空关键词检索：必须安全返回空切片，不崩溃不报错
	resEmpty, err := svc.Search(ctx, &model.SearchFilter{Query: "", LocalOnly: true})
	if err != nil {
		t.Fatalf("empty search failed: %v", err)
	}
	if len(resEmpty) != 0 {
		t.Errorf("expected 0 results for empty query, got %d", len(resEmpty))
	}

	// 2. 纯空格关键词
	resSpaces, err := svc.Search(ctx, &model.SearchFilter{Query: "     ", LocalOnly: true})
	if err != nil {
		t.Fatalf("spaces search failed: %v", err)
	}
	if len(resSpaces) != 0 {
		t.Errorf("expected 0 results for whitespace query, got %d", len(resSpaces))
	}

	// 3. 特殊字符与 SQL/FTS5 注入测试
	evilKeywords := []string{
		"' OR '1'='1",
		"\"; DROP TABLE games; --",
		"***^^^$$$(((",
		"!@#$%^&*()_+=-[]{};':\",./<>?",
		"NOT AND OR NEAR",
		"\"unclosed quote",
	}
	for _, evil := range evilKeywords {
		t.Run("EvilKeyword_"+evil, func(t *testing.T) {
			// 不能 panic，必须安全降级
			res, err := svc.Search(ctx, &model.SearchFilter{Query: evil, LocalOnly: true})
			if err != nil {
				t.Fatalf("special char search crashed or errored: %v", err)
			}
			_ = res
		})
	}

	// 4. 超长输入（10,000 字符）抗 DoS 压力测试
	longQuery := strings.Repeat("Cyberpunk ", 1000)
	resLong, err := svc.Search(ctx, &model.SearchFilter{Query: longQuery, LocalOnly: true})
	if err != nil {
		t.Fatalf("long query search failed: %v", err)
	}
	_ = resLong
}
