package fling

import (
	"context"
	"fmt"
	"net/http"
	"net/url"
	"regexp"
	"strconv"
	"strings"
	"time"

	"GameTrainerHub/internal/model"
	"GameTrainerHub/internal/source"
)

type FlingAdapter struct {
	client  *source.HttpClient
	baseURL string
}

func NewFlingAdapter(client *source.HttpClient) *FlingAdapter {
	if client == nil {
		client, _ = source.NewHttpClient("", 500)
	}
	return &FlingAdapter{
		client:  client,
		baseURL: "https://flingtrainer.com",
	}
}

func (a *FlingAdapter) ID() string {
	return "fling"
}

func (a *FlingAdapter) Name() string {
	return "FLiNG Trainer"
}

func (a *FlingAdapter) BaseURL() string {
	return a.baseURL
}

// Search 在原网站在线检索修改器
func (a *FlingAdapter) Search(ctx context.Context, keyword string, page int) ([]*model.SearchResult, error) {
	escaped := url.QueryEscape(strings.TrimSpace(keyword))
	searchURL := fmt.Sprintf("%s/?s=%s", a.baseURL, escaped)
	if page > 1 {
		searchURL = fmt.Sprintf("%s/page/%d/?s=%s", a.baseURL, page, escaped)
	}

	htmlBytes, err := a.client.Get(ctx, searchURL, a.baseURL)
	if err != nil {
		return nil, fmt.Errorf("FLiNG search request failed: %w", err)
	}

	return a.ParseSearchResults(string(htmlBytes))
}

// GetLatest 获取原网最新更新的修改器
func (a *FlingAdapter) GetLatest(ctx context.Context, page int) ([]*model.SearchResult, error) {
	targetURL := a.baseURL + "/"
	if page > 1 {
		targetURL = fmt.Sprintf("%s/page/%d/", a.baseURL, page)
	}

	htmlBytes, err := a.client.Get(ctx, targetURL, a.baseURL)
	if err != nil {
		return nil, fmt.Errorf("FLiNG latest request failed: %w", err)
	}

	return a.ParseSearchResults(string(htmlBytes))
}

// GetCatalog 获取全站 A-Z 索引目录 (760+ 游戏修改器种子)
func (a *FlingAdapter) GetCatalog(ctx context.Context) ([]*source.CatalogItem, error) {
	catalogURL := a.baseURL + "/all-trainers-a-z/"
	htmlBytes, err := a.client.Get(ctx, catalogURL, a.baseURL)
	if err != nil {
		return nil, fmt.Errorf("FLiNG catalog request failed: %w", err)
	}

	return a.ParseCatalog(string(htmlBytes))
}

// GetDetail 获取详情页信息与版本列表
func (a *FlingAdapter) GetDetail(ctx context.Context, detailURL string) (*model.Trainer, []*model.TrainerVersion, error) {
	htmlBytes, err := a.client.Get(ctx, detailURL, a.baseURL)
	if err != nil {
		return nil, nil, fmt.Errorf("FLiNG detail request failed: %w", err)
	}

	return a.ParseDetail(detailURL, string(htmlBytes))
}

// ResolveDownload 解析真实下载直链与文件属性
func (a *FlingAdapter) ResolveDownload(ctx context.Context, downloadURL string, referer string) (*source.DownloadInfo, error) {
	if referer == "" {
		referer = a.baseURL
	}

	req, err := http.NewRequestWithContext(ctx, http.MethodGet, downloadURL, nil)
	if err != nil {
		return nil, err
	}
	req.Header.Set("User-Agent", "Mozilla/5.0 (Windows NT 10.0; Win64; x64) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/120.0.0.0 Safari/537.36")
	req.Header.Set("Referer", referer)
	req.Header.Set("Range", "bytes=0-0") // 探测 Range 支持与文件大小

	resp, err := a.client.RawClient().Do(req)
	if err != nil {
		return nil, fmt.Errorf("download resolution failed: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK && resp.StatusCode != http.StatusPartialContent {
		return nil, fmt.Errorf("download endpoint returned status %d", resp.StatusCode)
	}

	info := &source.DownloadInfo{
		FinalURL:      resp.Request.URL.String(),
		SupportsRange: (resp.StatusCode == http.StatusPartialContent || resp.Header.Get("Accept-Ranges") == "bytes"),
		ContentType:   resp.Header.Get("Content-Type"),
	}

	// 解析文件总大小
	contentRange := resp.Header.Get("Content-Range")
	if contentRange != "" {
		// 格式: bytes 0-0/1008725
		parts := strings.Split(contentRange, "/")
		if len(parts) == 2 {
			total, parseErr := strconv.ParseInt(strings.TrimSpace(parts[1]), 10, 64)
			if parseErr == nil {
				info.FileSize = total
			}
		}
	}
	if info.FileSize == 0 && resp.ContentLength > 0 {
		info.FileSize = resp.ContentLength
	}

	// 解析文件名
	contentDisp := resp.Header.Get("Content-Disposition")
	if contentDisp != "" && strings.Contains(contentDisp, "filename=") {
		re := regexp.MustCompile(`filename=["']?([^"';]+)["']?`)
		m := re.FindStringSubmatch(contentDisp)
		if len(m) > 1 {
			info.FileName = m[1]
		}
	}
	if info.FileName == "" {
		// 从 URL 路径获取
		u, _ := url.Parse(info.FinalURL)
		if u != nil {
			pathSegs := strings.Split(strings.Trim(u.Path, "/"), "/")
			if len(pathSegs) > 0 {
				lastSeg, _ := url.PathUnescape(pathSegs[len(pathSegs)-1])
				info.FileName = lastSeg
			}
		}
	}
	if info.FileName == "" {
		info.FileName = "trainer.zip"
	}

	return info, nil
}

// ----------------- HTML 解析器方法 -----------------

var (
	reArticle     = regexp.MustCompile(`(?s)<article[^>]*id="post-(\d+)"[^>]*>([\s\S]*?)</article>`)
	reTitle       = regexp.MustCompile(`(?s)<h2[^>]*class="[^"]*post-title[^"]*"[^>]*>\s*<a\s+href="([^"]+)"[^>]*>([\s\S]*?)</a>`)
	reThumb       = regexp.MustCompile(`<div class="post-details-thumb">\s*<a[^>]*>\s*<img[^>]*src="([^"]+)"`)
	reDateParts   = regexp.MustCompile(`(?s)<div class="post-details-day">(\d+)</div>\s*<div class="post-details-month">([A-Za-z]+)</div>\s*<div class="post-details-year">(\d+)</div>`)
	reDetailTitle = regexp.MustCompile(`(?si)<h1[^>]*>([\s\S]*?)</h1>`)
	reDownloads   = regexp.MustCompile(`(?s)<a\s+[^>]*href="(https?://flingtrainer\.com/downloads/[^"]+)"[^>]*>([\s\S]*?)</a>`)
	reCleanHTML   = regexp.MustCompile(`<[^>]+>`)
	reCatalogItem = regexp.MustCompile(`(?s)<a\s+href="(https?://flingtrainer\.com/trainer/[^"/]+/?)"[^>]*>\s*([\s\S]*?)\s*</a>`)
	reScript      = regexp.MustCompile(`(?si)<script[^>]*>[\s\S]*?</script>`)
	reStyle       = regexp.MustCompile(`(?si)<style[^>]*>[\s\S]*?</style>`)
	reOGImage     = regexp.MustCompile(`(?i)<meta\s+property=["']og:image["']\s+content=["']([^"']+)["']`)
	rePostImg     = regexp.MustCompile(`(?i)<img[^>]*class=["'][^"']*(?:attachment-|wp-post-image)[^"']*["'][^>]*src=["']([^"']+)["']`)
)

func (a *FlingAdapter) ParseSearchResults(html string) ([]*model.SearchResult, error) {
	articles := reArticle.FindAllStringSubmatch(html, -1)
	var results []*model.SearchResult

	for _, art := range articles {
		body := art[2]

		titleMatch := reTitle.FindStringSubmatch(body)
		if len(titleMatch) < 3 {
			continue
		}
		detailURL := titleMatch[1]
		rawTitle := cleanHTML(titleMatch[2])

		var coverURL string
		thumbMatch := reThumb.FindStringSubmatch(body)
		if len(thumbMatch) > 1 {
			coverURL = thumbMatch[1]
		}

		updatedAt := time.Now()
		dateMatch := reDateParts.FindStringSubmatch(body)
		if len(dateMatch) == 4 {
			// day, month, year
			dateStr := fmt.Sprintf("%s %s %s", dateMatch[1], dateMatch[2], dateMatch[3])
			parsedTime, parseErr := time.Parse("02 Jan 2006", dateStr)
			if parseErr == nil {
				updatedAt = parsedTime
			}
		}

		// 规范化游戏名与 Slug
		gameName := strings.TrimSuffix(rawTitle, " Trainer")
		slug := slugify(gameName)

		results = append(results, &model.SearchResult{
			Slug:         slug,
			GameNameEN:   gameName,
			TrainerTitle: rawTitle,
			DetailURL:    detailURL,
			CoverURL:     coverURL,
			SourceID:     a.ID(),
			UpdatedAt:    updatedAt,
			IsOnlineOnly: true,
		})
	}

	return results, nil
}

func (a *FlingAdapter) ParseCatalog(html string) ([]*source.CatalogItem, error) {
	matches := reCatalogItem.FindAllStringSubmatch(html, -1)
	var items []*source.CatalogItem
	seen := make(map[string]bool)

	for _, m := range matches {
		detailURL := m[1]
		title := cleanHTML(m[2])

		if title == "" || seen[detailURL] || strings.HasPrefix(title, "http") {
			continue
		}
		seen[detailURL] = true

		// 提取 externalID 从 URL 路径
		parts := strings.Split(strings.Trim(detailURL, "/"), "/")
		externalID := parts[len(parts)-1]

		items = append(items, &source.CatalogItem{
			Title:      title,
			DetailURL:  detailURL,
			ExternalID: externalID,
		})
	}

	return items, nil
}

func (a *FlingAdapter) ParseDetail(detailURL string, html string) (*model.Trainer, []*model.TrainerVersion, error) {
	// 提取标题，优先 h1，若缺失则从 title 或 URL 回退
	var rawTitle string
	titleMatch := reDetailTitle.FindStringSubmatch(html)
	if len(titleMatch) >= 2 {
		rawTitle = cleanHTML(titleMatch[1])
	}
	if rawTitle == "" {
		reHTMLTitle := regexp.MustCompile(`(?si)<title>([\s\S]*?)</title>`)
		tMatch := reHTMLTitle.FindStringSubmatch(html)
		if len(tMatch) >= 2 {
			titleText := cleanHTML(tMatch[1])
			if idx := strings.Index(titleText, " - "); idx != -1 {
				titleText = strings.TrimSpace(titleText[:idx])
			}
			rawTitle = titleText
		}
	}
	if rawTitle == "" {
		urlParts := strings.Split(strings.Trim(detailURL, "/"), "/")
		if len(urlParts) > 0 {
			slug := urlParts[len(urlParts)-1]
			slug = strings.TrimSuffix(slug, "-trainer")
			rawTitle = strings.Title(strings.ReplaceAll(slug, "-", " ")) + " Trainer"
		}
	}
	if rawTitle == "" {
		return nil, nil, fmt.Errorf("failed to parse entry title in %s", detailURL)
	}
	gameName := strings.TrimSuffix(rawTitle, " Trainer")

	// 提取外部ID
	urlParts := strings.Split(strings.Trim(detailURL, "/"), "/")
	externalID := urlParts[len(urlParts)-1]

	var coverURL string
	if m := reOGImage.FindStringSubmatch(html); len(m) > 1 {
		coverURL = m[1]
	} else if m := rePostImg.FindStringSubmatch(html); len(m) > 1 {
		coverURL = m[1]
	} else if m := reThumb.FindStringSubmatch(html); len(m) > 1 {
		coverURL = m[1]
	}

	trainer := &model.Trainer{
		SourceID:      a.ID(),
		ExternalID:    externalID,
		TitleEN:       rawTitle,
		CoverURL:      coverURL,
		DetailURL:     detailURL,
		Author:        "FLiNG",
		LatestVersion: "",
		PublishedAt:   time.Now(),
		UpdatedAt:     time.Now(),
	}

	// 提取下载链接与版本
	dlMatches := reDownloads.FindAllStringSubmatch(html, -1)
	var versions []*model.TrainerVersion

	for _, dl := range dlMatches {
		dlURL := dl[1]
		linkText := cleanHTML(dl[2])
		if linkText == "" {
			continue
		}

		// 整理版本名称
		versionName := extractVersionName(linkText, gameName)
		fileName := linkText + ".zip"
		if strings.HasSuffix(strings.ToLower(linkText), ".exe") || strings.HasSuffix(strings.ToLower(linkText), ".zip") {
			fileName = linkText
		}

		v := &model.TrainerVersion{
			VersionName: versionName,
			DownloadURL: dlURL,
			FileName:    fileName,
			PublishedAt: time.Now(),
		}
		versions = append(versions, v)
	}

	if len(versions) > 0 {
		trainer.LatestVersion = versions[0].VersionName
	}

	// 提取修改项说明内容 (支持 class="entry" 与 class="entry-content")
	contentIdx := strings.Index(html, `class="entry"`)
	if contentIdx == -1 {
		contentIdx = strings.Index(html, `class="entry-content`)
	}
	if contentIdx != -1 {
		endIdx := strings.Index(html[contentIdx:], `</div><!--/.entry`)
		if endIdx != -1 {
			rawContent := html[contentIdx : contentIdx+endIdx]
			features := cleanHTML(rawContent)
			featuresText := strings.TrimSpace(features)
			for _, v := range versions {
				v.FeaturesText = featuresText
			}
		}
	}

	trainer.Versions = versions
	return trainer, versions, nil
}

// 辅助清洗方法
func cleanHTML(s string) string {
	s = reScript.ReplaceAllString(s, "")
	s = reStyle.ReplaceAllString(s, "")
	s = reCleanHTML.ReplaceAllString(s, "")
	s = strings.ReplaceAll(s, "&nbsp;", " ")
	s = strings.ReplaceAll(s, "&amp;", "&")
	s = strings.ReplaceAll(s, "&#8211;", "-")
	s = strings.ReplaceAll(s, "&#8217;", "'")
	s = strings.ReplaceAll(s, "&#038;", "&")
	s = strings.ReplaceAll(s, "&#046;", ".")
	return strings.TrimSpace(s)
}

func slugify(name string) string {
	lower := strings.ToLower(name)
	reg := regexp.MustCompile(`[^a-z0-9]+`)
	slug := reg.ReplaceAllString(lower, "-")
	return strings.Trim(slug, "-")
}

func extractVersionName(text string, gameName string) string {
	// 去除作者后缀 -FLiNG
	t := strings.TrimSuffix(text, "-FLiNG")
	t = strings.ReplaceAll(t, ".", " ")
	t = strings.ReplaceAll(t, "_", " ")
	t = strings.TrimSpace(t)

	// 如果包含 Plus，截取版本段
	plusIdx := strings.Index(strings.ToLower(t), "plus")
	if plusIdx != -1 {
		vIdx := strings.Index(strings.ToLower(t), "v")
		if vIdx != -1 && vIdx < plusIdx {
			return strings.TrimSpace(t[vIdx:])
		}
	}
	return t
}

