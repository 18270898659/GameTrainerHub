package game

import (
	"context"
	"os"
	"path/filepath"
	"regexp"
	"strconv"
	"strings"

	"GameTrainerHub/internal/database/repository"
	"GameTrainerHub/internal/logger"
)

type DiscoveredSteamGame struct {
	AppID       int    `json:"app_id"`
	Name        string `json:"name"`
	InstallDir  string `json:"install_dir"`
	InstallPath string `json:"install_path"`
}

type SteamScanner struct {
	customSteamPaths []string
}

func NewSteamScanner(customPaths ...string) *SteamScanner {
	return &SteamScanner{
		customSteamPaths: customPaths,
	}
}

// FindSteamRoot 探测 Windows 系统中 Steam 的根目录
func (s *SteamScanner) FindSteamRoot() string {
	candidates := append(s.customSteamPaths,
		`C:\Program Files (x86)\Steam`,
		`C:\Program Files\Steam`,
		`D:\Steam`,
		`D:\SteamLibrary`,
		`E:\Steam`,
		`E:\SteamLibrary`,
	)

	for _, p := range candidates {
		if p != "" {
			if fi, err := os.Stat(filepath.Join(p, "steam.exe")); err == nil && !fi.IsDir() {
				return p
			}
			if fi, err := os.Stat(filepath.Join(p, "steamapps")); err == nil && fi.IsDir() {
				return p
			}
		}
	}
	return ""
}

// GetLibraryFolders 解析 libraryfolders.vdf 获取所有 Steam 库目录
func (s *SteamScanner) GetLibraryFolders(steamRoot string) []string {
	if steamRoot == "" {
		return nil
	}

	vdfPath := filepath.Join(steamRoot, "steamapps", "libraryfolders.vdf")
	data, err := os.ReadFile(vdfPath)
	if err != nil {
		// 默认将自身根目录作为唯一库目录
		return []string{steamRoot}
	}

	return s.ParseLibraryFoldersVDF(string(data), steamRoot)
}

// ParseLibraryFoldersVDF 解析 VDF 文本中的 path 字段
func (s *SteamScanner) ParseLibraryFoldersVDF(content string, defaultRoot string) []string {
	re := regexp.MustCompile(`"path"\s+"([^"]+)"`)
	matches := re.FindAllStringSubmatch(content, -1)

	var folders []string
	seen := make(map[string]bool)

	for _, m := range matches {
		if len(m) > 1 {
			// VDF 中的双反斜杠替换为单反斜杠
			cleanPath := filepath.Clean(strings.ReplaceAll(m[1], `\\`, `\`))
			if !seen[strings.ToLower(cleanPath)] {
				seen[strings.ToLower(cleanPath)] = true
				folders = append(folders, cleanPath)
			}
		}
	}

	if len(folders) == 0 && defaultRoot != "" {
		folders = append(folders, defaultRoot)
	}
	return folders
}

// ScanInstalledGames 扫描所有 Steam 库目录下的已安装游戏
func (s *SteamScanner) ScanInstalledGames() ([]*DiscoveredSteamGame, error) {
	root := s.FindSteamRoot()
	if root == "" {
		return nil, nil
	}

	folders := s.GetLibraryFolders(root)
	var games []*DiscoveredSteamGame
	seenAppIDs := make(map[int]bool)

	for _, folder := range folders {
		steamapps := filepath.Join(folder, "steamapps")
		entries, err := os.ReadDir(steamapps)
		if err != nil {
			continue
		}

		for _, entry := range entries {
			if strings.HasPrefix(entry.Name(), "appmanifest_") && strings.HasSuffix(entry.Name(), ".acf") {
				acfPath := filepath.Join(steamapps, entry.Name())
				data, err := os.ReadFile(acfPath)
				if err != nil {
					continue
				}

				game := s.ParseAppManifestACF(string(data), folder)
				if game != nil && !seenAppIDs[game.AppID] && game.Name != "" {
					seenAppIDs[game.AppID] = true
					games = append(games, game)
				}
			}
		}
	}

	return games, nil
}

// ParseAppManifestACF 解析单个 appmanifest 文件
func (s *SteamScanner) ParseAppManifestACF(content string, libraryFolder string) *DiscoveredSteamGame {
	reAppID := regexp.MustCompile(`"appid"\s+"(\d+)"`)
	reName := regexp.MustCompile(`"name"\s+"([^"]+)"`)
	reInstallDir := regexp.MustCompile(`"installdir"\s+"([^"]+)"`)

	mAppID := reAppID.FindStringSubmatch(content)
	mName := reName.FindStringSubmatch(content)
	mInstallDir := reInstallDir.FindStringSubmatch(content)

	if len(mAppID) < 2 || len(mName) < 2 {
		return nil
	}

	appID, _ := strconv.Atoi(mAppID[1])
	name := mName[1]
	installDir := ""
	if len(mInstallDir) > 1 {
		installDir = mInstallDir[1]
	}

	fullPath := filepath.Join(libraryFolder, "steamapps", "common", installDir)

	return &DiscoveredSteamGame{
		AppID:       appID,
		Name:        name,
		InstallDir:  installDir,
		InstallPath: fullPath,
	}
}

// SyncWithGameRepo 将扫描到的 Steam 游戏同步匹配到本地游戏库
func (s *SteamScanner) SyncWithGameRepo(ctx context.Context, repo *repository.GameRepository) (int, error) {
	games, err := s.ScanInstalledGames()
	if err != nil || len(games) == 0 {
		return 0, err
	}

	matchedCount := 0
	for _, sg := range games {
		slug := slugify(sg.Name)
		dbGame, err := repo.GetBySlug(ctx, slug)
		if err == nil && dbGame != nil {
			dbGame.SteamAppID = sg.AppID
			dbGame.InstallPath = sg.InstallPath
			_ = repo.Update(ctx, dbGame)
			matchedCount++
			logger.Info("Matched local Steam game to library", "name", sg.Name, "path", sg.InstallPath)
		}
	}

	return matchedCount, nil
}

func slugify(name string) string {
	lower := strings.ToLower(name)
	reg := regexp.MustCompile(`[^a-z0-9]+`)
	slug := reg.ReplaceAllString(lower, "-")
	return strings.Trim(slug, "-")
}

