package launcher

import (
	"bytes"
	"context"
	"crypto/sha256"
	"encoding/hex"
	"errors"
	"fmt"
	"io"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"sync"
	"syscall"
	"unsafe"

	"GameTrainerHub/internal/database/repository"
	"GameTrainerHub/internal/logger"
	"GameTrainerHub/internal/model"

	"golang.org/x/sys/windows"
)

var (
	shell32            = windows.NewLazySystemDLL("shell32.dll")
	procShellExecuteEx = shell32.NewProc("ShellExecuteExW")
)

const (
	SEE_MASK_NOCLOSEPROCESS = 0x00000040
)

// SHELLEXECUTEINFO 对应 Windows 原生 SHELLEXECUTEINFOW 结构体
type SHELLEXECUTEINFO struct {
	cbSize       uint32
	fMask        uint32
	hwnd         windows.Handle
	lpVerb       *uint16
	lpFile       *uint16
	lpParameters *uint16
	lpDirectory  *uint16
	nShow        int32
	hInstApp     windows.Handle
	lpIDList     uintptr
	lpClass      *uint16
	hkeyClass    windows.Handle
	dwHotKey     uint32
	hIcon        windows.Handle
	hProcess     windows.Handle
}

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
	RequiresAdmin  bool   `json:"requires_admin"`
	SecurityNotice string `json:"security_notice"`
}

type activeProcess struct {
	proc   *os.Process
	handle windows.Handle
}

type Launcher struct {
	localRepo       *repository.LocalTrainerRepository
	gameRepo        *repository.GameRepository
	activeProcesses sync.Map // map[int64]*activeProcess
}

func NewLauncher(localRepo *repository.LocalTrainerRepository, gameRepo *repository.GameRepository) *Launcher {
	return &Launcher{
		localRepo: localRepo,
		gameRepo:  gameRepo,
	}
}

// IsTrainerRunning 检查指定本地修改器是否正在运行中
func (l *Launcher) IsTrainerRunning(localTrainerID int64) bool {
	val, ok := l.activeProcesses.Load(localTrainerID)
	if !ok {
		return false
	}
	ap, ok := val.(*activeProcess)
	if !ok || ap == nil {
		l.activeProcesses.Delete(localTrainerID)
		return false
	}

	// 若存在 Windows 进程句柄，实时探测进程是否存活
	if ap.handle != 0 {
		var exitCode uint32
		if err := windows.GetExitCodeProcess(ap.handle, &exitCode); err == nil {
			if exitCode != 259 /* STILL_ACTIVE */ {
				l.activeProcesses.Delete(localTrainerID)
				return false
			}
		}
	}
	return true
}

// StopTrainer 终止指定本地修改器进程
func (l *Launcher) StopTrainer(localTrainerID int64) error {
	val, ok := l.activeProcesses.Load(localTrainerID)
	if !ok {
		return fmt.Errorf("trainer is not currently running")
	}
	ap, ok := val.(*activeProcess)
	if ok && ap != nil {
		if ap.handle != 0 {
			_ = windows.TerminateProcess(ap.handle, 1)
		}
		if ap.proc != nil && ap.proc.Pid > 0 {
			_ = ap.proc.Kill()
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

	requiresAdmin := HasRequireAdministratorManifest(target.FilePath)

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
		RequiresAdmin:  requiresAdmin,
		SecurityNotice: notice,
	}, nil
}

// LaunchExecutable 在用户确认后安全拉起修改器进程（兼容普通权限与 UAC 提权）
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
		// 用户明确勾选以管理员身份启动，直接触发 UAC 权限提升流程
		logger.Info("Launching trainer with elevated administrator privilege via ShellExecuteEx (User requested)", "path", filePath)
		ap, err := l.launchElevated(filePath, workDir, localTrainerID)
		if err != nil {
			return err
		}
		l.activeProcesses.Store(localTrainerID, ap)
	} else {
		// 默认普通权限尝试启动
		logger.Info("Attempting to launch trainer with standard privilege", "path", filePath)
		cmd := exec.Command(filePath)
		cmd.Dir = workDir
		startErr := cmd.Start()
		if startErr != nil {
			// 检测是否需要 UAC 权限提升 (Windows 740 ERROR_ELEVATION_REQUIRED)
			if isElevationRequiredError(startErr) {
				logger.Warn("Trainer requires administrator elevation (error 740), requesting UAC elevation", "path", filePath)
				ap, err := l.launchElevated(filePath, workDir, localTrainerID)
				if err != nil {
					return err
				}
				l.activeProcesses.Store(localTrainerID, ap)
			} else {
				return fmt.Errorf("failed to start trainer process: %w", startErr)
			}
		} else {
			// 普通权限启动成功
			ap := &activeProcess{proc: cmd.Process}
			l.activeProcesses.Store(localTrainerID, ap)
			go func() {
				_ = cmd.Wait()
				l.activeProcesses.Delete(localTrainerID)
				logger.Info("Standard trainer process has terminated", "trainer_id", localTrainerID)
			}()
		}
	}

	// 记录启动次数与最后启动时间
	_ = l.localRepo.RecordRun(ctx, localTrainerID)
	return nil
}

// launchElevated 使用 Windows 原生 ShellExecuteExW 以 runas 动词请求 UAC 提权启动
func (l *Launcher) launchElevated(filePath, workDir string, localTrainerID int64) (*activeProcess, error) {
	verbPtr, err := windows.UTF16PtrFromString("runas")
	if err != nil {
		return nil, fmt.Errorf("invalid verb: %w", err)
	}
	filePtr, err := windows.UTF16PtrFromString(filePath)
	if err != nil {
		return nil, fmt.Errorf("invalid file path: %w", err)
	}
	dirPtr, err := windows.UTF16PtrFromString(workDir)
	if err != nil {
		return nil, fmt.Errorf("invalid working directory: %w", err)
	}

	var info SHELLEXECUTEINFO
	info.cbSize = uint32(unsafe.Sizeof(info))
	info.fMask = SEE_MASK_NOCLOSEPROCESS
	info.lpVerb = verbPtr
	info.lpFile = filePtr
	info.lpDirectory = dirPtr
	info.nShow = windows.SW_SHOWNORMAL

	ret, _, callErr := procShellExecuteEx.Call(uintptr(unsafe.Pointer(&info)))
	if ret == 0 {
		// 检查是否为用户在 UAC 弹窗中点击“否”取消 (ERROR_CANCELLED = 1223)
		if errno, ok := callErr.(syscall.Errno); ok {
			if errno == 1223 || errno == windows.ERROR_CANCELLED {
				return nil, fmt.Errorf("用户取消了管理员权限授权 (UAC 取消)")
			}
			if errno != 0 {
				return nil, fmt.Errorf("Windows UAC 提权启动失败: %w", errno)
			}
		}
		if callErr != nil {
			return nil, fmt.Errorf("Windows UAC 提权启动失败: %w", callErr)
		}
		return nil, fmt.Errorf("Windows UAC 提权启动失败")
	}

	var proc *os.Process
	if info.hProcess != 0 {
		pid, pErr := windows.GetProcessId(info.hProcess)
		if pErr == nil && pid > 0 {
			p, err := os.FindProcess(int(pid))
			if err == nil {
				proc = p
			}
		}
	}
	if proc == nil {
		proc = &os.Process{Pid: 0}
	}

	ap := &activeProcess{
		proc:   proc,
		handle: info.hProcess,
	}

	if info.hProcess != 0 {
		go func(h windows.Handle, tID int64) {
			defer windows.CloseHandle(h)
			_, _ = windows.WaitForSingleObject(h, windows.INFINITE)
			l.activeProcesses.Delete(tID)
			logger.Info("Elevated trainer process has terminated", "trainer_id", tID)
		}(info.hProcess, localTrainerID)
	}

	return ap, nil
}

// isElevationRequiredError 判断错误是否为 Windows 740 提权请求错误
func isElevationRequiredError(err error) bool {
	if err == nil {
		return false
	}
	var errno syscall.Errno
	if errors.As(err, &errno) && (errno == 740 || errno == windows.ERROR_ELEVATION_REQUIRED) {
		return true
	}
	var pathErr *os.PathError
	if errors.As(err, &pathErr) {
		if errors.As(pathErr.Err, &errno) && (errno == 740 || errno == windows.ERROR_ELEVATION_REQUIRED) {
			return true
		}
	}
	errStr := strings.ToLower(err.Error())
	return strings.Contains(errStr, "elevation") || strings.Contains(errStr, "requires elevation") || strings.Contains(errStr, "740")
}

// HasRequireAdministratorManifest 扫描 PE 文件是否声明了 requireAdministrator 权限级别
func HasRequireAdministratorManifest(path string) bool {
	f, err := os.Open(path)
	if err != nil {
		return false
	}
	defer f.Close()

	fi, err := f.Stat()
	if err != nil || fi.Size() <= 0 {
		return false
	}

	size := fi.Size()
	bufSize := int64(4 * 1024 * 1024)
	if size < bufSize {
		bufSize = size
	}
	buf := make([]byte, bufSize)
	n, _ := f.Read(buf)
	if bytes.Contains(buf[:n], []byte("requireAdministrator")) {
		return true
	}

	if size > bufSize {
		tailSize := int64(2 * 1024 * 1024)
		if tailSize > size {
			tailSize = size
		}
		tailBuf := make([]byte, tailSize)
		_, _ = f.Seek(-tailSize, io.SeekEnd)
		tn, _ := f.Read(tailBuf)
		if bytes.Contains(tailBuf[:tn], []byte("requireAdministrator")) {
			return true
		}
	}
	return false
}

// OpenContainingFolder 在 Windows 资源管理器中打开并选中该文件
// 优先使用 Windows Shell API 与绝对系统路径，杜绝依赖 PATH 环境变量
func (l *Launcher) OpenContainingFolder(filePath string) error {
	cleanPath := filepath.Clean(filePath)
	fi, err := os.Stat(cleanPath)
	if err != nil {
		// 若文件自身不存在但父目录存在，降级直接打开父目录
		parentDir := filepath.Dir(cleanPath)
		if pfi, perr := os.Stat(parentDir); perr == nil && pfi.IsDir() {
			return openFolderDirect(parentDir)
		}
		return fmt.Errorf("目标路径不存在或无法访问: %w", err)
	}

	explorerPath := getExplorerExePath()

	// 1. 若目标是文件，使用 /select,"path" 选中该文件
	if !fi.IsDir() {
		verbPtr, _ := windows.UTF16PtrFromString("open")
		filePtr, _ := windows.UTF16PtrFromString(explorerPath)
		// 严密包裹双引号以兼容中文、空格及特殊符号
		argsPtr, _ := windows.UTF16PtrFromString(fmt.Sprintf(`/select,"%s"`, cleanPath))

		if err := windows.ShellExecute(0, verbPtr, filePtr, argsPtr, nil, windows.SW_SHOWNORMAL); err == nil {
			return nil
		}

		// 回退使用绝对路径执行
		cmd := exec.Command(explorerPath, fmt.Sprintf(`/select,%s`, cleanPath))
		if err := cmd.Start(); err == nil {
			return nil
		}
	}

	// 2. 若目标为目录或文件选中失败，直接打开目标所在目录
	dirToOpen := cleanPath
	if !fi.IsDir() {
		dirToOpen = filepath.Dir(cleanPath)
	}
	return openFolderDirect(dirToOpen)
}

func openFolderDirect(dirPath string) error {
	verbPtr, _ := windows.UTF16PtrFromString("open")
	dirPtr, _ := windows.UTF16PtrFromString(dirPath)
	// 原生调用 ShellExecute 打开文件夹，无需依赖 PATH 环境变量中的 explorer.exe
	if err := windows.ShellExecute(0, verbPtr, dirPtr, nil, nil, windows.SW_SHOWNORMAL); err == nil {
		return nil
	}

	// 回退使用绝对路径 explorer.exe 打开
	explorerPath := getExplorerExePath()
	cmd := exec.Command(explorerPath, dirPath)
	return cmd.Start()
}

func getExplorerExePath() string {
	candidates := []string{
		filepath.Join(os.Getenv("SystemRoot"), "explorer.exe"),
		filepath.Join(os.Getenv("WINDIR"), "explorer.exe"),
		`C:\Windows\explorer.exe`,
	}
	for _, p := range candidates {
		if p != "" {
			if fi, err := os.Stat(p); err == nil && !fi.IsDir() {
				return p
			}
		}
	}
	if p, err := exec.LookPath("explorer.exe"); err == nil {
		return p
	}
	return `C:\Windows\explorer.exe`
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
