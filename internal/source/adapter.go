package source

import (
	"context"
	"GameTrainerHub/internal/model"
)

// CatalogItem 网站收录索引简项
type CatalogItem struct {
	Title      string `json:"title"`
	DetailURL  string `json:"detail_url"`
	ExternalID string `json:"external_id"`
}

// DownloadInfo 解析后的真实下载信息
type DownloadInfo struct {
	FinalURL      string `json:"final_url"`
	FileName      string `json:"file_name"`
	FileSize      int64  `json:"file_size"`
	SupportsRange bool   `json:"supports_range"`
	ContentType   string `json:"content_type"`
}

// SourceAdapter 统一修改器资源网站适配器接口
type SourceAdapter interface {
	// ID 适配器唯一标识 (如 "fling")
	ID() string

	// Name 适配器显示名称 (如 "FLiNG Trainer")
	Name() string

	// BaseURL 网站主入口根地址
	BaseURL() string

	// Search 在原网站在线检索游戏修改器
	Search(ctx context.Context, keyword string, page int) ([]*model.SearchResult, error)

	// GetDetail 获取指定修改器详情页信息（包括所有版本、说明与下载入口）
	GetDetail(ctx context.Context, detailURL string) (*model.Trainer, []*model.TrainerVersion, error)

	// GetLatest 获取原网首页或最新发布的修改器列表
	GetLatest(ctx context.Context, page int) ([]*model.SearchResult, error)

	// GetCatalog 获取原网全量 A-Z 索引目录（用于冷启动与基础收录）
	GetCatalog(ctx context.Context) ([]*CatalogItem, error)

	// ResolveDownload 将原网下载入口或 Token 解析为真实可断点续传的直链
	ResolveDownload(ctx context.Context, downloadURL string, referer string) (*DownloadInfo, error)
}

