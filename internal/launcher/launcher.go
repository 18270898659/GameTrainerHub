package launcher

import (
	"context"
	"crypto/sha256"
	"encoding/hex"
	"fmt"
	"io"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"sync"

	"GameTrainerHub/internal/database/repository"
	"GameTrainerHub/internal/logger"
	"GameTrainerHub/internal/model"

	"golang.org/x/sys/windows"
)

// SecurityCardInfo 启动前安全确认卡片信息
type SecurityCardInfo struct {
	TrainerID      int64  `json:"trainer_id"`
	GameName       string `json:"game_name"`
	FileName       string `json:"file_name"`
	FilePath       string `json:"file_path"`
	FileSize       int64  `json:"file_size"`
	SHA256         string `json:"sha256"`
	SourceName     string `json:"source_name"`
	IsCustom       bool   `json:"is_custom"`
	SecurityNotice string `json:"security_notice"`
}

type Launcher struct {
	localRepo       *repository.LocalTrainerRepository
	gameRepo        *repository.GameRepository
	activeProcesses sync.Map // map[int64]*os.Process
}

func NewLauncher(localRepo *repository.LocalTrainerRepository, gameRepo *repository.GameRepository) *Launcher {
	return &Launcher{
		localRepo: localRepo,
		gameRepo:  gameRepo,
	}
}

// IsTrainerRunning 检查指定本地修改器是否正在运行中
func (l *Launcher) IsTrainerRunning(localTrainerID int64) bool {
	_, ok := l.activeProcesses.Load(localTrainerID)
	return ok
}

// StopTrainer 终止指定本地修改器进程
func (l *Launcher) StopTrainer(localTrainerID int64) error {
	val, ok := l.activeProcesses.Load(localTrainerID)
	if !ok {
		return fmt.Errorf("trainer is not currently running")
	}
	proc, ok := val.(*os.Process)
	if ok && proc != nil && proc.Pid > 0 {
		if err := proc.Kill(); err != nil {
			return fmt.Errorf("failed to kill process: %w", err)
		}
	}
	l.activeProcesses.Delete(localTrainerID)
	return nil
}

// GetSecurityCard 获取修改器启动前必须公开展示的安全卡片
func (l *Launcher) GetSecurityCard(ctx context.Context, localTrainerID int64) (*SecurityCardInfo, error) {
	all, err := l.localRepo.ListAll(ctx)
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
		return nil, fmt.Errorf("local trainer not found")
	}

	gameName := "未知游戏"
	if game, err := l.gameRepo.GetByID(ctx, target.GameID); err == nil && game != nil {
		if game.NameCN != "" {
			gameName = game.NameCN
		} else {
			gameName = game.NameEN
		}
	}

	// 动态复核文件当前真实哈希与大小
	fi, err := os.Stat(target.FilePath)
	if err != nil {
		return nil, fmt.Errorf("trainer executable does not exist at %s: %w", target.FilePath, err)
	}

	currentHash, _ := computeFileSHA256(target.FilePath)
	if currentHash == "" {
		currentHash = target.SHA256
	}

	notice := "【安全警示与合规声明】\n" +
		"1. 游戏修改器通过读取和写入游戏进程内存（Memory Hooking）实现数值调整，极易触发 Windows Defender 及杀毒软件的启发式报毒（误报）。\n" +
		"2. 本客户端坚决不自动关闭系统安全防护、不擅自修改杀毒白名单。请核对上方文件 SHA-256 哈希值确认无篡改。\n" +
		"3. 本修改器仅供单机离线模式娱乐使用，严禁用于任何在线多人竞技游戏作弊。"

	return &SecurityCardInfo{
		TrainerID:      target.ID,
		GameName:       gameName,
		FileName:       filepath.Base(target.FilePath),
		FilePath:       target.FilePath,
		FileSize:       fi.Size(),
		SHA256:         currentHash,
		SourceName:     target.SourceName,
		IsCustom:       target.IsCustom,
		SecurityNotice: notice,
	}, nil
}

// LaunchExecutable 在用户确认后安全拉起修改器进程（非阻塞）
func (l *Launcher) LaunchExecutable(ctx context.Context, localTrainerID int64, runAsAdmin bool) error {
	if l.IsTrainerRunning(localTrainerID) {
		return fmt.Errorf("trainer process is already running (ID: %d)", localTrainerID)
	}

	card, err := l.GetSecurityCard(ctx, localTrainerID)
	if err != nil {
		return err
	}

	filePath := card.FilePath
	if !strings.HasSuffix(strings.ToLower(filePath), ".exe") {
		return fmt.Errorf("invalid executable format: %s", filePath)
	}

	workDir := filepath.Dir(filePath)

	if runAsAdmin {
		// 显式以管理员权限请求 Windows UAC 提权，采用原生 ShellExecute 系统调用，杜绝命令注入与脚本逃逸
		logger.Info("Launching trainer with elevated administrator privilege via ShellExecute", "path", filePath)
		verbPtr, err := windows.UTF16PtrFromString("runas")
		if err != nil {
			return fmt.Errorf("invalid verb: %w", err)
		}
		filePtr, err := windows.UTF16PtrFromString(filePath)
		if err != nil {
			return fmt.Errorf("invalid file path: %w", err)
		}
		dirPtr, err := windows.UTF16PtrFromString(workDir)
		if err != nil {
			return fmt.Errorf("invalid working directory: %w", err)
		}
		if err := windows.ShellExecute(0, verbPtr, filePtr, nil, dirPtr, windows.SW_SHOWNORMAL); err != nil {
			return fmt.Errorf("failed to launch elevated trainer: %w", err)
		}
		l.activeProcesses.Store(localTrainerID, &os.Process{Pid: 0})
	} else {
		// 默认普通用户权限执行，工作目录隔离，直接调用可执行文件路径
		logger.Info("Launching trainer with standard privilege", "path", filePath)
		cmd := exec.Command(filePath)
		cmd.Dir = workDir
		if err := cmd.Start(); err != nil {
			return fmt.Errorf("failed to start trainer process: %w", err)
		}
		l.activeProcesses.Store(localTrainerID, cmd.Process)
		go func() {
			_ = cmd.Wait()
			l.activeProcesses.Delete(localTrainerID)
		}()
	}

	// 记录启动次数与最后启动时间
	_ = l.localRepo.RecordRun(ctx, localTrainerID)
	return nil
}

// OpenContainingFolder 在 Windows 资源管理器中打开并选中该文件
func (l *Launcher) OpenContainingFolder(filePath string) error {
	if _, err := os.Stat(filePath); err != nil {
		return fmt.Errorf("target file does not exist: %w", err)
	}

	cleanPath := filepath.Clean(filePath)
	cmd := exec.Command("explorer.exe", fmt.Sprintf("/select,%s", cleanPath))
	return cmd.Start()
}

func computeFileSHA256(path string) (string, error) {
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
