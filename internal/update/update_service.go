package update

import (
	"context"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"
	"path/filepath"
	"strconv"
	"strings"
	"time"

	"GameTrainerHub/internal/logger"
)

// ReleaseInfo 发布更新信息实体
type ReleaseInfo struct {
	Version      string    `json:"version"`
	DownloadURL  string    `json:"download_url"`
	SHA256       string    `json:"sha256"`
	FileSize     int64     `json:"file_size"`
	ReleaseNotes string    `json:"release_notes"`
	Mandatory    bool      `json:"mandatory"`
	PublishedAt  time.Time `json:"published_at"`
}

type UpdateService struct {
	currentVersion string
	updateCheckURL string
	httpClient     *http.Client
}

func NewUpdateService(currentVersion string, updateCheckURL string, client *http.Client) *UpdateService {
	if client == nil {
		client = &http.Client{Timeout: 15 * time.Second}
	}
	return &UpdateService{
		currentVersion: currentVersion,
		updateCheckURL: updateCheckURL,
		httpClient:     client,
	}
}

// CheckUpdate 检查是否有新版本
func (s *UpdateService) CheckUpdate(ctx context.Context) (*ReleaseInfo, bool, error) {
	if s.updateCheckURL == "" {
		return nil, false, fmt.Errorf("update check URL is empty")
	}

	req, err := http.NewRequestWithContext(ctx, http.MethodGet, s.updateCheckURL, nil)
	if err != nil {
		return nil, false, err
	}
	req.Header.Set("User-Agent", "GameTrainerHub-Updater/"+s.currentVersion)

	resp, err := s.httpClient.Do(req)
	if err != nil {
		return nil, false, fmt.Errorf("failed to fetch update info: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, false, fmt.Errorf("update server returned status: %d", resp.StatusCode)
	}

	var release ReleaseInfo
	if err := json.NewDecoder(resp.Body).Decode(&release); err != nil {
		return nil, false, fmt.Errorf("failed to decode release payload: %w", err)
	}

	hasNewVersion := CompareVersions(release.Version, s.currentVersion) > 0
	return &release, hasNewVersion, nil
}

// VerifyUpdatePackage 校验下载的更新安装包完整性
func (s *UpdateService) VerifyUpdatePackage(packagePath string, expectedSHA256 string, expectedSize int64) (bool, error) {
	fi, err := os.Stat(packagePath)
	if err != nil {
		return false, fmt.Errorf("cannot stat package file: %w", err)
	}

	if expectedSize > 0 && fi.Size() != expectedSize {
		return false, fmt.Errorf("package size mismatch: got %d, expected %d", fi.Size(), expectedSize)
	}

	f, err := os.Open(packagePath)
	if err != nil {
		return false, err
	}
	defer f.Close()

	h := sha256.New()
	if _, err := io.Copy(h, f); err != nil {
		return false, err
	}

	actualHash := hex.EncodeToString(h.Sum(nil))
	if expectedSHA256 != "" && !strings.EqualFold(actualHash, expectedSHA256) {
		logger.Warn("Update package SHA-256 verification failed", "expected", expectedSHA256, "actual", actualHash)
		return false, fmt.Errorf("sha256 mismatch: expected %s, got %s", expectedSHA256, actualHash)
	}

	return true, nil
}

// CompareVersions 语义化版本号比较 (v1 > v2: 1, v1 == v2: 0, v1 < v2: -1)
func CompareVersions(v1, v2 string) int {
	clean1 := strings.TrimPrefix(strings.TrimSpace(v1), "v")
	clean2 := strings.TrimPrefix(strings.TrimSpace(v2), "v")

	parts1 := strings.Split(clean1, ".")
	parts2 := strings.Split(clean2, ".")

	maxLen := len(parts1)
	if len(parts2) > maxLen {
		maxLen = len(parts2)
	}

	for i := 0; i < maxLen; i++ {
		var n1, n2 int
		if i < len(parts1) {
			n1, _ = strconv.Atoi(parts1[i])
		}
		if i < len(parts2) {
			n2, _ = strconv.Atoi(parts2[i])
		}

		if n1 > n2 {
			return 1
		}
		if n1 < n2 {
			return -1
		}
	}

	return 0
}

// GetPackageTempPath 获取安装包缓存路径
func (s *UpdateService) GetPackageTempPath(cacheDir string, version string) string {
	return filepath.Join(cacheDir, fmt.Sprintf("GameTrainerHub-Setup-%s.exe", version))
}

