package trainer

import (
	"context"
	"crypto/sha256"
	"encoding/hex"
	"fmt"
	"io"
	"os"
	"path/filepath"
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
	LocalTrainer *model.LocalTrainer   `json:"local_trainer"`
	Game         *model.Game           `json:"game"`
	Version      *model.TrainerVersion `json:"version,omitempty"`
	FileExists   bool                  `json:"file_exists"`
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
			}
		}
		details = append(details, d)
	}
	return details, nil
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
