package trainer

import (
	"context"
	"crypto/sha256"
	"encoding/hex"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"sort"
	"strings"

	"GameTrainerHub/internal/database/repository"
	"GameTrainerHub/internal/model"
)

type TrainerService struct {
	gameRepo    *repository.GameRepository
	trainerRepo *repository.TrainerRepository
	localRepo   *repository.LocalTrainerRepository
}

func NewTrainerService(
	gameRepo *repository.GameRepository,
	trainerRepo *repository.TrainerRepository,
	localRepo *repository.LocalTrainerRepository,
) *TrainerService {
	return &TrainerService{
		gameRepo:    gameRepo,
		trainerRepo: trainerRepo,
		localRepo:   localRepo,
	}
}

type LocalTrainerDetail struct {
	LocalTrainer      *model.LocalTrainer   `json:"local_trainer"`
	Game              *model.Game           `json:"game"`
	Version           *model.TrainerVersion `json:"version,omitempty"`
	FileExists        bool                  `json:"file_exists"`
	IsRunning         bool                  `json:"is_running"`
	TargetGameVersion string                `json:"target_game_version,omitempty"`
}

// ListLocalTrainers 列出所有已下载或手动导入的本地修改器
func (s *TrainerService) ListLocalTrainers(ctx context.Context) ([]*LocalTrainerDetail, error) {
	locals, err := s.localRepo.ListAll(ctx)
	if err != nil {
		return nil, err
	}

	var details []*LocalTrainerDetail
	for _, lt := range locals {
		d := &LocalTrainerDetail{LocalTrainer: lt}
		if _, statErr := os.Stat(lt.FilePath); statErr == nil {
			d.FileExists = true
		} else {
			d.FileExists = false
		}
		if game, err := s.gameRepo.GetByID(ctx, lt.GameID); err == nil {
			d.Game = game
		}
		if lt.VersionID != nil {
			if ver, err := s.trainerRepo.GetVersionByID(ctx, *lt.VersionID); err == nil {
				d.Version = ver
				d.TargetGameVersion = ver.TargetGameVersion
			}
		}
		details = append(details, d)
	}
	return details, nil
}

// GetRecentTrainers 获取最近实际启动运行过的修改器列表（按 last_run_at 倒序排列，排除未启动过的修改器）
func (s *TrainerService) GetRecentTrainers(ctx context.Context, limit int) ([]*LocalTrainerDetail, error) {
	all, err := s.ListLocalTrainers(ctx)
	if err != nil {
		return nil, err
	}
	var ranTrainers []*LocalTrainerDetail
	for _, d := range all {
		if d.LocalTrainer != nil && d.LocalTrainer.RunCount > 0 && d.LocalTrainer.LastRunAt != nil {
			ranTrainers = append(ranTrainers, d)
		}
	}

	sort.Slice(ranTrainers, func(i, j int) bool {
		t1 := ranTrainers[i].LocalTrainer.LastRunAt
		t2 := ranTrainers[j].LocalTrainer.LastRunAt
		if t1 == nil {
			return false
		}
		if t2 == nil {
			return true
		}
		return t1.After(*t2)
	})

	if limit <= 0 {
		limit = 10
	}
	if len(ranTrainers) > limit {
		return ranTrainers[:limit], nil
	}
	return ranTrainers, nil
}

// VerifyTrainerIntegrity 校验本地修改器文件 SHA-256 完整性
func (s *TrainerService) VerifyTrainerIntegrity(ctx context.Context, localTrainerID int64) (*model.IntegrityResult, error) {
	all, err := s.localRepo.ListAll(ctx)
	if err != nil {
		return nil, err
	}
	var target *model.LocalTrainer
	for _, lt := range all {
		if lt.ID == localTrainerID {
			target = lt
			break
		}
	}
	if target == nil {
		return nil, fmt.Errorf("local trainer with ID %d not found", localTrainerID)
	}

	res := &model.IntegrityResult{
		TrainerID:      localTrainerID,
		FilePath:       target.FilePath,
		ExpectedSHA256: target.SHA256,
	}

	fi, err := os.Stat(target.FilePath)
	if err != nil {
		res.FileExists = false
		res.IsValid = false
		res.Message = "修改器本地文件不存在或已被移动"
		return res, nil
	}

	res.FileExists = true
	res.FileSize = fi.Size()

	actualHash, err := hashFile(target.FilePath)
	if err != nil {
		res.IsValid = false
		res.Message = fmt.Sprintf("计算文件哈希失败: %v", err)
		return res, nil
	}
	res.ActualSHA256 = actualHash

	if strings.EqualFold(res.ActualSHA256, res.ExpectedSHA256) {
		res.IsValid = true
		res.Message = "文件哈希校验完全一致，完整性验证通过"
	} else {
		res.IsValid = false
		res.Message = "文件哈希与记录不一致，文件可能已被篡改或损坏"
	}
	return res, nil
}

// RelocateTrainer 修改器文件重新定位（当原路径失效时更新新路径并重新计算哈希与大小）
func (s *TrainerService) RelocateTrainer(ctx context.Context, localTrainerID int64, newFilePath string) error {
	cleanPath := filepath.Clean(strings.TrimSpace(newFilePath))
	fi, err := os.Stat(cleanPath)
	if err != nil {
		return fmt.Errorf("目标文件不存在: %w", err)
	}
	if !strings.HasSuffix(strings.ToLower(cleanPath), ".exe") {
		return fmt.Errorf("重新定位的目标文件必须是 .exe 可执行文件")
	}

	newHash, err := hashFile(cleanPath)
	if err != nil {
		return fmt.Errorf("计算新文件哈希失败: %w", err)
	}

	all, err := s.localRepo.ListAll(ctx)
	if err != nil {
		return err
	}
	for _, lt := range all {
		if lt.ID == localTrainerID {
			lt.FilePath = cleanPath
			lt.FileSize = fi.Size()
			lt.SHA256 = newHash
			return s.localRepo.Update(ctx, lt)
		}
	}
	return fmt.Errorf("未找到本地修改器 ID %d", localTrainerID)
}

// ImportCustomTrainer 手动导入本地现有的修改器 EXE
func (s *TrainerService) ImportCustomTrainer(ctx context.Context, gameID int64, filePath string) (*model.LocalTrainer, error) {
	fi, err := os.Stat(filePath)
	if err != nil {
		return nil, fmt.Errorf("file does not exist: %w", err)
	}
	if !strings.HasSuffix(strings.ToLower(filePath), ".exe") {
		return nil, fmt.Errorf("only .exe executables are supported for import")
	}

	// 计算哈希
	hStr, err := hashFile(filePath)
	if err != nil {
		return nil, err
	}

	// 检查是否已存在
	existing, err := s.localRepo.GetByFilePath(ctx, filePath)
	if err == nil && existing != nil {
		return existing, nil
	}

	// 若未指定关联游戏 (gameID <= 0)，根据文件名智能解析或创建独立游戏实体，杜绝跨游戏图片与实体串接
	if gameID <= 0 {
		baseName := filepath.Base(filePath)
		cleanName := strings.TrimSuffix(baseName, filepath.Ext(baseName))
		cleanName = strings.ReplaceAll(cleanName, "_", " ")
		cleanName = strings.ReplaceAll(cleanName, ".", " ")
		slug := "custom-" + strings.ToLower(strings.ReplaceAll(cleanName, " ", "-"))
		g, _ := s.gameRepo.GetOrCreateBySlug(ctx, slug, cleanName, cleanName, "")
		if g != nil {
			gameID = g.ID
		}
	}

	lt := &model.LocalTrainer{
		GameID:     gameID,
		FilePath:   filePath,
		FileSize:   fi.Size(),
		SHA256:     hStr,
		SourceName: "Custom",
		IsCustom:   true,
	}

	if err := s.localRepo.Create(ctx, lt); err != nil {
		return nil, err
	}
	return lt, nil
}

// DeleteLocalTrainer 删除本地修改器记录，可选择是否物理删除磁盘文件
func (s *TrainerService) DeleteLocalTrainer(ctx context.Context, id int64, deleteFile bool) error {
	// 查询本地路径
	all, err := s.localRepo.ListAll(ctx)
	if err != nil {
		return err
	}

	for _, lt := range all {
		if lt.ID == id {
			if deleteFile && lt.FilePath != "" {
				_ = os.Remove(lt.FilePath)
			}
			return s.localRepo.Delete(ctx, id)
		}
	}
	return nil
}

// ScanFolderForTrainers 扫描指定目录下的 EXE 文件
func (s *TrainerService) ScanFolderForTrainers(dirPath string) ([]string, error) {
	var exeFiles []string
	err := filepath.Walk(dirPath, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return nil
		}
		if !info.IsDir() && strings.HasSuffix(strings.ToLower(path), ".exe") {
			exeFiles = append(exeFiles, path)
		}
		return nil
	})
	return exeFiles, err
}

// ToggleFavorite 切换游戏收藏状态
func (s *TrainerService) ToggleFavorite(ctx context.Context, gameID int64) (bool, error) {
	isFav, err := s.gameRepo.IsFavorite(ctx, gameID)
	if err != nil {
		return false, err
	}
	newStatus := !isFav
	err = s.gameRepo.SetFavorite(ctx, gameID, newStatus)
	return newStatus, err
}

func hashFile(path string) (string, error) {
	f, err := os.Open(path)
	if err != nil {
		return "", err
	}
	defer f.Close()

	h := sha256.New()
	if _, err := io.Copy(h, f); err != nil {
		return "", err
	}
	return hex.EncodeToString(h.Sum(nil)), nil
}
