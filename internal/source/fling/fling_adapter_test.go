package fling

import (
	"context"
	"strings"
	"testing"
	"time"

	"GameTrainerHub/internal/source"
)

// 测试用真实 HTML 样本
const sampleSearchHTML = `
<article id="post-15664" class="group post-standard post-15664 post type-post status-publish format-standard has-post-thumbnail hentry category-trainer tag-cyberpunk-2077">	
	<div class="post-details">
		<div class="post-details-thumb">
			<a href="https://flingtrainer.com/trainer/cyberpunk-2077-trainer-16095388215/">
				<img width="200" height="200" src="https://flingtrainer.com/wp-content/uploads/2020/12/header-7-200x200.jpg" class="attachment-stylizer-small size-stylizer-small wp-post-image" alt="" />
			</a>
		</div>
		<div class="post-details-day">28</div>
		<div class="post-details-month">Oct</div>
		<div class="post-details-year">2026</div>
	</div>
	<div class="post-content">
		<h2 class="post-title">
			<a href="https://flingtrainer.com/trainer/cyberpunk-2077-trainer-16095388215/" rel="bookmark">Cyberpunk 2077 Trainer</a>
		</h2>
	</div>
</article>
`

const sampleDetailHTML = `
<h1 class="entry-title">Cyberpunk 2077 Trainer</h1>
<div class="entry-content themeform">
	<p>OPTIONS</p>
	<p>Num 1 – Infinite Health<br />Num 2 – Infinite Stamina<br />Num 3 – Infinite Items/Ammo</p>
	<p><a href="https://flingtrainer.com/downloads/_FszD2rT8W8Gcuowp4IMUQ,,">Cyberpunk.2077.v2.0-v2.5.Plus.46.Trainer.Updated-FLiNG</a></p>
	<p><a href="https://flingtrainer.com/downloads/p0svCFCzrONqkS4MVRgoEw,,">Cyberpunk.2077.v2.0-v2.5.Plus.46.Trainer-FLiNG</a></p>
</div><!--/.entry-content-->
`

const sampleCatalogHTML = `
<div id="a-z-listing-1" class="az-listing">
	<div id="a-z-listing-letter-C-1">
		<ul class="az-columns">
			<li><a href="https://flingtrainer.com/trainer/cyberpunk-2077-trainer-16095388215/">Cyberpunk 2077 Trainer</a></li>
			<li><a href="https://flingtrainer.com/trainer/control-trainer/">Control Trainer</a></li>
		</ul>
	</div>
</div>
`

func TestParseSearchResults(t *testing.T) {
	adapter := NewFlingAdapter(nil)

	results, err := adapter.ParseSearchResults(sampleSearchHTML)
	if err != nil {
		t.Fatalf("ParseSearchResults error: %v", err)
	}

	if len(results) != 1 {
		t.Fatalf("expected 1 result, got %d", len(results))
	}

	item := results[0]
	if item.GameNameEN != "Cyberpunk 2077" {
		t.Errorf("expected 'Cyberpunk 2077', got '%s'", item.GameNameEN)
	}
	if item.Slug != "cyberpunk-2077" {
		t.Errorf("expected slug 'cyberpunk-2077', got '%s'", item.Slug)
	}
	if item.DetailURL != "https://flingtrainer.com/trainer/cyberpunk-2077-trainer-16095388215/" {
		t.Errorf("unexpected detail URL: %s", item.DetailURL)
	}
	if item.CoverURL != "https://flingtrainer.com/wp-content/uploads/2020/12/header-7-200x200.jpg" {
		t.Errorf("unexpected cover URL: %s", item.CoverURL)
	}
}

func TestParseCatalog(t *testing.T) {
	adapter := NewFlingAdapter(nil)

	items, err := adapter.ParseCatalog(sampleCatalogHTML)
	if err != nil {
		t.Fatalf("ParseCatalog error: %v", err)
	}

	if len(items) != 2 {
		t.Fatalf("expected 2 items, got %d", len(items))
	}

	if items[0].Title != "Cyberpunk 2077 Trainer" || items[0].ExternalID != "cyberpunk-2077-trainer-16095388215" {
		t.Errorf("unexpected first catalog item: %+v", items[0])
	}
	if items[1].Title != "Control Trainer" || items[1].ExternalID != "control-trainer" {
		t.Errorf("unexpected second catalog item: %+v", items[1])
	}
}

func TestParseDetail(t *testing.T) {
	adapter := NewFlingAdapter(nil)

	detailURL := "https://flingtrainer.com/trainer/cyberpunk-2077-trainer-16095388215/"
	trainer, versions, err := adapter.ParseDetail(detailURL, sampleDetailHTML)
	if err != nil {
		t.Fatalf("ParseDetail error: %v", err)
	}

	if trainer.TitleEN != "Cyberpunk 2077 Trainer" {
		t.Errorf("expected 'Cyberpunk 2077 Trainer', got '%s'", trainer.TitleEN)
	}
	if trainer.Author != "FLiNG" {
		t.Errorf("expected author 'FLiNG', got '%s'", trainer.Author)
	}

	if len(versions) != 2 {
		t.Fatalf("expected 2 versions, got %d", len(versions))
	}

	v1 := versions[0]
	if !strings.Contains(v1.DownloadURL, "_FszD2rT8W8Gcuowp4IMUQ") {
		t.Errorf("unexpected download URL: %s", v1.DownloadURL)
	}
	if !strings.Contains(v1.FeaturesText, "Infinite Health") {
		t.Errorf("expected features to contain 'Infinite Health', got '%s'", v1.FeaturesText)
	}
}

func TestLiveFLiNGSearch(t *testing.T) {
	// 在线集成测试：若网络不可达则跳过，不阻断离线单测
	client, err := source.NewHttpClient("", 500)
	if err != nil {
		t.Skip("failed to create http client")
	}

	adapter := NewFlingAdapter(client)
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	results, err := adapter.Search(ctx, "Cyberpunk", 1)
	if err != nil {
		t.Logf("Live FLiNG search encountered network condition (skipped in CI/offline): %v", err)
		return
	}

	if len(results) > 0 {
		t.Logf("Live FLiNG search succeeded! Found %d items. First: %s", len(results), results[0].TrainerTitle)
		if !strings.Contains(strings.ToLower(results[0].GameNameEN), "cyberpunk") {
			t.Errorf("expected game name to contain 'cyberpunk', got '%s'", results[0].GameNameEN)
		}
	}
}

func TestLiveBlackMythWukongGetDetail(t *testing.T) {
	client, err := source.NewHttpClient("", 500)
	if err != nil {
		t.Skip("failed to create http client")
	}

	adapter := NewFlingAdapter(client)
	ctx, cancel := context.WithTimeout(context.Background(), 15*time.Second)
	defer cancel()

	detailURL := "https://flingtrainer.com/trainer/black-myth-wukong-trainer/"
	trainer, versions, err := adapter.GetDetail(ctx, detailURL)
	if err != nil {
		t.Logf("Live network fetch skipped: %v", err)
		return
	}

	if trainer == nil {
		t.Fatal("expected trainer not to be nil")
	}
	if trainer.TitleEN != "Black Myth: Wukong Trainer" {
		t.Errorf("expected 'Black Myth: Wukong Trainer', got '%s'", trainer.TitleEN)
	}
	if len(versions) == 0 {
		t.Fatal("expected at least 1 version for Black Myth: Wukong")
	}
	t.Logf("Successfully fetched Black Myth: Wukong with %d versions. Latest: %s", len(versions), versions[0].VersionName)
	if !strings.Contains(versions[0].FeaturesText, "God Mode") {
		t.Errorf("expected features to contain 'God Mode', got snippet: %s", versions[0].FeaturesText[:min(100, len(versions[0].FeaturesText))])
	}
}

func TestResolveDownloadBMW(t *testing.T) {
	adapter := NewFlingAdapter(nil)
	ctx, cancel := context.WithTimeout(context.Background(), 20*time.Second)
	defer cancel()

	dlURL := "https://flingtrainer.com/downloads/J0ZX9_hiI3QYHi9EG5A7vw,,"
	start := time.Now()
	info, err := adapter.ResolveDownload(ctx, dlURL, "https://flingtrainer.com/trainer/black-myth-wukong-trainer/")
	dur := time.Since(start)
	if err != nil {
		t.Fatalf("ResolveDownload failed: %v", err)
	}
	if info == nil || info.FinalURL == "" {
		t.Fatalf("expected non-empty FinalURL")
	}
	t.Logf("ResolveDownload completed in %v, finalURL: %s", dur, info.FinalURL)
}

