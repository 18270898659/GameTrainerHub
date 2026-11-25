//go:build windows

package tray

import (
	"fmt"
	"os"
	"runtime"
	"strings"
	"sync"
	"syscall"
	"unsafe"

	"golang.org/x/sys/windows"
)

var (
	modUser32   = windows.NewLazySystemDLL("user32.dll")
	modShell32  = windows.NewLazySystemDLL("shell32.dll")
	modKernel32 = windows.NewLazySystemDLL("kernel32.dll")

	procRegisterClassExW    = modUser32.NewProc("RegisterClassExW")
	procCreateWindowExW     = modUser32.NewProc("CreateWindowExW")
	procDefWindowProcW      = modUser32.NewProc("DefWindowProcW")
	procDestroyWindow       = modUser32.NewProc("DestroyWindow")
	procGetMessageW         = modUser32.NewProc("GetMessageW")
	procTranslateMessage    = modUser32.NewProc("TranslateMessage")
	procDispatchMessageW    = modUser32.NewProc("DispatchMessageW")
	procPostMessageW        = modUser32.NewProc("PostMessageW")
	procSendMessageW        = modUser32.NewProc("SendMessageW")
	procPostQuitMessage     = modUser32.NewProc("PostQuitMessage")
	procCreatePopupMenu     = modUser32.NewProc("CreatePopupMenu")
	procAppendMenuW         = modUser32.NewProc("AppendMenuW")
	procTrackPopupMenuEx    = modUser32.NewProc("TrackPopupMenuEx")
	procDestroyMenu         = modUser32.NewProc("DestroyMenu")
	procGetCursorPos        = modUser32.NewProc("GetCursorPos")
	procSetForegroundWindow = modUser32.NewProc("SetForegroundWindow")
	procLoadIconW           = modUser32.NewProc("LoadIconW")
	procDestroyIcon         = modUser32.NewProc("DestroyIcon")

	procShellNotifyIconW = modShell32.NewProc("Shell_NotifyIconW")
	procExtractIconExW   = modShell32.NewProc("ExtractIconExW")

	procGetModuleHandleW    = modKernel32.NewProc("GetModuleHandleW")
	procGetCurrentProcessId = modKernel32.NewProc("GetCurrentProcessId")

	procRegisterHotKey   = modUser32.NewProc("RegisterHotKey")
	procUnregisterHotKey = modUser32.NewProc("UnregisterHotKey")

	procEnumWindows              = modUser32.NewProc("EnumWindows")
	procGetWindowThreadProcessId = modUser32.NewProc("GetWindowThreadProcessId")
	procFindWindowW              = modUser32.NewProc("FindWindowW")
	procIsWindowVisible          = modUser32.NewProc("IsWindowVisible")
	procIsIconic                 = modUser32.NewProc("IsIconic")
	procShowWindow               = modUser32.NewProc("ShowWindow")
	procGetForegroundWindow      = modUser32.NewProc("GetForegroundWindow")
)

const (
	WM_APP                 = 0x8000
	WM_TRAYICON            = WM_APP + 100
	WM_APP_APPLY_HOTKEYS   = WM_APP + 201
	WM_APP_TEST_HOTKEY     = WM_APP + 202
	WM_APP_SUSPEND_HOTKEYS = WM_APP + 203
	WM_CLOSE               = 0x0010
	WM_DESTROY             = 0x0002
	WM_LBUTTONUP           = 0x0202
	WM_LBUTTONDBLCLK       = 0x0203
	WM_RBUTTONUP           = 0x0205
	WM_HOTKEY              = 0x0312

	NIM_ADD    = 0x00000000
	NIM_MODIFY = 0x00000001
	NIM_DELETE = 0x00000002

	NIF_MESSAGE = 0x00000001
	NIF_ICON    = 0x00000002
	NIF_TIP     = 0x00000004
	NIF_INFO    = 0x00000010

	NIIF_INFO    = 0x00000001
	NIIF_WARNING = 0x00000002
	NIIF_ERROR   = 0x00000003

	MOD_ALT      = 0x0001
	MOD_CONTROL  = 0x0002
	MOD_SHIFT    = 0x0004
	MOD_WIN      = 0x0008
	MOD_NOREPEAT = 0x4000

	HOTKEY_ID_TOGGLE    = 1001
	HOTKEY_ID_DOWNLOADS = 1002

	MF_STRING    = 0x00000000
	MF_SEPARATOR = 0x00000800

	TPM_RETURNCMD = 0x0100

	CMD_SHOW = 1
	CMD_QUIT = 2
)

type WNDCLASSEXW struct {
	CbSize        uint32
	Style         uint32
	LpfnWndProc   uintptr
	CbClsExtra    int32
	CbWndExtra    int32
	HInstance     windows.Handle
	HIcon         windows.Handle
	HCursor       windows.Handle
	HbrBackground windows.Handle
	LpszMenuName  *uint16
	LpszClassName *uint16
	HIconSm       windows.Handle
}

type NOTIFYICONDATAW struct {
	CbSize            uint32
	HWnd              windows.Handle
	UID               uint32
	UFlags            uint32
	UCallbackMessage  uint32
	HIcon             windows.Handle
	SzTip             [128]uint16
	DwState           uint32
	DwStateMask       uint32
	SzInfo            [256]uint16
	UTimeoutOrVersion uint32
	SzInfoTitle       [64]uint16
	DwInfoFlags       uint32
	GuidItem          windows.GUID
	HBalloonIcon      windows.Handle
}

type POINT struct {
	X int32
	Y int32
}

type MSG struct {
	HWnd    windows.Handle
	Message uint32
	WParam  uintptr
	LParam  uintptr
	Time    uint32
	Pt      POINT
}

type applyHotkeyReq struct {
	enabled    bool
	normToggle string
	mod1       uint32
	vk1        uint32
	normDl     string
	mod2       uint32
	vk2        uint32
	success    bool
	errCode    uint32
}

type testHotkeyReq struct {
	mod      uint32
	vk       uint32
	conflict bool
}

// TrayManager Windows 原生系统托盘管理器
type TrayManager struct {
	hwnd              windows.Handle
	hIcon             windows.Handle
	nid               NOTIFYICONDATAW
	onShow            func()
	onQuit            func()
	onHotkeyToggle    func()
	onHotkeyDownloads func()
	showText          string
	quitText          string
	isAdded           bool
	currentToggleStr  string
	currentDlStr      string
	currentEnabled    bool
	isSuspended       bool
	pendingApplyReq   *applyHotkeyReq
	pendingTestReq    *testHotkeyReq
	mu                sync.Mutex
	readyChan         chan struct{}
	stoppedChan       chan struct{}
}

var (
	instance *TrayManager
	instMu   sync.Mutex
)

// Init 初始化并运行 Windows 系统托盘
func Init(onShow, onQuit func()) (*TrayManager, error) {
	instMu.Lock()
	defer instMu.Unlock()

	if instance != nil {
		return instance, nil
	}

	tm := &TrayManager{
		onShow:      onShow,
		onQuit:      onQuit,
		showText:    "显示主窗口",
		quitText:    "退出程序",
		readyChan:   make(chan struct{}),
		stoppedChan: make(chan struct{}),
	}
	instance = tm

	go tm.messageLoop()
	<-tm.readyChan

	return tm, nil
}

// UpdateLabels 更新托盘右键菜单文字（多语言适配）
func UpdateLabels(showText, quitText string) {
	instMu.Lock()
	defer instMu.Unlock()
	if instance != nil {
		instance.mu.Lock()
		if showText != "" {
			instance.showText = showText
		}
		if quitText != "" {
			instance.quitText = quitText
		}
		instance.mu.Unlock()
	}
}

// Destroy 移除托盘图标并关闭消息窗口
func Destroy() {
	instMu.Lock()
	defer instMu.Unlock()
	if instance != nil {
		instance.mu.Lock()
		if instance.hwnd != 0 {
			_, _, _ = procUnregisterHotKey.Call(uintptr(instance.hwnd), uintptr(HOTKEY_ID_TOGGLE))
			_, _, _ = procUnregisterHotKey.Call(uintptr(instance.hwnd), uintptr(HOTKEY_ID_DOWNLOADS))
		}
		if instance.isAdded && instance.nid.HWnd != 0 {
			_, _, _ = procShellNotifyIconW.Call(uintptr(NIM_DELETE), uintptr(unsafe.Pointer(&instance.nid)))
			instance.isAdded = false
		}
		if instance.hwnd != 0 {
			_, _, _ = procPostMessageW.Call(uintptr(instance.hwnd), uintptr(WM_CLOSE), 0, 0)
		}
		if instance.hIcon != 0 {
			_, _, _ = procDestroyIcon.Call(uintptr(instance.hIcon))
			instance.hIcon = 0
		}
		instance.mu.Unlock()
		<-instance.stoppedChan
		instance = nil
	}
}

// ShowNotification 发送 Windows 系统托盘气泡/横幅通知
func ShowNotification(title, message string, isError bool) error {
	instMu.Lock()
	defer instMu.Unlock()
	if instance == nil || !instance.isAdded || instance.hwnd == 0 {
		return fmt.Errorf("tray is not initialized")
	}

	instance.mu.Lock()
	defer instance.mu.Unlock()

	nid := instance.nid
	nid.UFlags = NIF_INFO

	titleU16, _ := windows.UTF16FromString(title)
	msgU16, _ := windows.UTF16FromString(message)

	copy(nid.SzInfoTitle[:], titleU16)
	copy(nid.SzInfo[:], msgU16)

	if isError {
		nid.DwInfoFlags = NIIF_ERROR
	} else {
		nid.DwInfoFlags = NIIF_INFO
	}

	r, _, _ := procShellNotifyIconW.Call(uintptr(NIM_MODIFY), uintptr(unsafe.Pointer(&nid)))
	if r == 0 {
		return fmt.Errorf("failed to show tray notification")
	}
	return nil
}

// SetHotkeyHandlers 注册快捷键事件回调
func SetHotkeyHandlers(onToggle, onDownloads func()) {
	instMu.Lock()
	defer instMu.Unlock()
	if instance != nil {
		instance.mu.Lock()
		instance.onHotkeyToggle = onToggle
		instance.onHotkeyDownloads = onDownloads
		instance.mu.Unlock()
	}
}

// NormalizeHotkeyString 将任意快捷键组合规范化为标准格式 "Alt + Ctrl + Q"
func NormalizeHotkeyString(s string) (string, error) {
	s = strings.TrimSpace(s)
	if s == "" {
		return "", fmt.Errorf("快捷键不能为空")
	}

	parts := strings.Split(s, "+")
	var hasAlt, hasCtrl, hasShift, hasWin bool
	var mainKey string

	for _, p := range parts {
		trimmed := strings.TrimSpace(p)
		lower := strings.ToLower(trimmed)
		switch lower {
		case "alt":
			hasAlt = true
		case "ctrl", "control":
			hasCtrl = true
		case "shift":
			hasShift = true
		case "win", "super", "meta", "windows":
			hasWin = true
		default:
			if trimmed == "" {
				continue
			}
			if mainKey != "" {
				return "", fmt.Errorf("快捷键只能包含一个主按键，检测到多个: %s, %s", mainKey, trimmed)
			}
			mainKey = formatMainKeyName(trimmed)
		}
	}

	if mainKey == "" {
		return "", fmt.Errorf("快捷键必须包含一个主按键 (如 Q, J, 1, F1 等)")
	}

	// 校验修饰键规则：普通字符或按键必须搭配至少一个修饰键；F1-F24 功能键允许独立设置
	isFnKey := strings.HasPrefix(strings.ToUpper(mainKey), "F") && len(mainKey) >= 2
	if !hasAlt && !hasCtrl && !hasShift && !hasWin && !isFnKey {
		return "", fmt.Errorf("快捷键必须包含至少一个修饰键 (Ctrl / Alt / Shift / Win)")
	}

	// 统一排序修饰键：Alt, Ctrl, Shift, Win，满足 "Alt + Ctrl + Q" 规范要求
	var mods []string
	if hasAlt {
		mods = append(mods, "Alt")
	}
	if hasCtrl {
		mods = append(mods, "Ctrl")
	}
	if hasShift {
		mods = append(mods, "Shift")
	}
	if hasWin {
		mods = append(mods, "Win")
	}

	combo := append(mods, mainKey)
	return strings.Join(combo, " + "), nil
}

// formatMainKeyName 格式化主按键名称规范显示
func formatMainKeyName(s string) string {
	upper := strings.ToUpper(strings.TrimSpace(s))
	switch upper {
	case "ESC", "ESCAPE":
		return "Esc"
	case "SPACE":
		return "Space"
	case "TAB":
		return "Tab"
	case "ENTER", "RETURN":
		return "Enter"
	case "BACKSPACE", "BKSP":
		return "Backspace"
	case "DELETE", "DEL":
		return "Delete"
	case "INSERT", "INS":
		return "Insert"
	case "HOME":
		return "Home"
	case "END":
		return "End"
	case "PAGEUP", "PGUP":
		return "PageUp"
	case "PAGEDOWN", "PGDN":
		return "PageDown"
	case "UP":
		return "Up"
	case "DOWN":
		return "Down"
	case "LEFT":
		return "Left"
	case "RIGHT":
		return "Right"
	default:
		if strings.HasPrefix(upper, "F") && len(upper) >= 2 {
			return upper
		}
		if len(upper) == 1 {
			return upper
		}
		return upper
	}
}

// mainKeyToVK 将主按键转换为 Windows 虚拟键码 (Virtual Key Code)
func mainKeyToVK(key string) (uint32, error) {
	upper := strings.ToUpper(strings.TrimSpace(key))
	if len(upper) == 1 {
		ch := upper[0]
		if (ch >= 'A' && ch <= 'Z') || (ch >= '0' && ch <= '9') {
			return uint32(ch), nil
		}
		switch ch {
		case ';':
			return 0xBA, nil
		case '=':
			return 0xBB, nil
		case ',':
			return 0xBC, nil
		case '-':
			return 0xBD, nil
		case '.':
			return 0xBE, nil
		case '/':
			return 0xBF, nil
		case '`':
			return 0xC0, nil
		case '[':
			return 0xDB, nil
		case '\\':
			return 0xDC, nil
		case ']':
			return 0xDD, nil
		case '\'':
			return 0xDE, nil
		}
	}

	if strings.HasPrefix(upper, "F") && len(upper) >= 2 {
		num := 0
		for _, ch := range upper[1:] {
			if ch >= '0' && ch <= '9' {
				num = num*10 + int(ch-'0')
			}
		}
		if num >= 1 && num <= 24 {
			return uint32(0x70 + num - 1), nil
		}
	}

	switch upper {
	case "SPACE":
		return 0x20, nil
	case "TAB":
		return 0x09, nil
	case "ENTER", "RETURN":
		return 0x0D, nil
	case "ESC", "ESCAPE":
		return 0x1B, nil
	case "BACKSPACE", "BKSP":
		return 0x08, nil
	case "DELETE", "DEL":
		return 0x2E, nil
	case "INSERT", "INS":
		return 0x2D, nil
	case "HOME":
		return 0x24, nil
	case "END":
		return 0x23, nil
	case "PAGEUP", "PGUP":
		return 0x21, nil
	case "PAGEDOWN", "PGDN":
		return 0x22, nil
	case "UP":
		return 0x26, nil
	case "DOWN":
		return 0x28, nil
	case "LEFT":
		return 0x25, nil
	case "RIGHT":
		return 0x27, nil
	}

	return 0, fmt.Errorf("不支持的按键: %s", key)
}

func parseHotkeyString(s string) (uint32, uint32, error) {
	norm, err := NormalizeHotkeyString(s)
	if err != nil {
		return 0, 0, err
	}

	parts := strings.Split(norm, "+")
	var mod uint32 = MOD_NOREPEAT
	var vk uint32

	for _, p := range parts {
		trimmed := strings.TrimSpace(p)
		lower := strings.ToLower(trimmed)
		switch lower {
		case "alt":
			mod |= MOD_ALT
		case "ctrl":
			mod |= MOD_CONTROL
		case "shift":
			mod |= MOD_SHIFT
		case "win":
			mod |= MOD_WIN
		default:
			vkCode, err := mainKeyToVK(trimmed)
			if err != nil {
				return 0, 0, err
			}
			vk = vkCode
		}
	}

	if vk == 0 {
		return 0, 0, fmt.Errorf("无效的主按键: %s", s)
	}
	return mod, vk, nil
}

// UpdateHotkeys 动态注册或注销全局快捷键（保留兼容）
func UpdateHotkeys(enabled bool, toggleStr, dlStr string) error {
	return ApplyHotkeys(enabled, toggleStr, dlStr)
}

// ApplyHotkeys 强制注册并应用全局快捷键（具备校验、冲突检测与失败回滚保障）
func ApplyHotkeys(enabled bool, toggleStr, dlStr string) error {
	instMu.Lock()
	defer instMu.Unlock()
	if instance == nil || instance.hwnd == 0 {
		return nil
	}

	tm := instance
	tm.mu.Lock()
	defer tm.mu.Unlock()

	if !enabled {
		req := &applyHotkeyReq{enabled: false}
		tm.pendingApplyReq = req
		procSendMessageW.Call(uintptr(tm.hwnd), uintptr(WM_APP_APPLY_HOTKEYS), 0, 0)
		tm.pendingApplyReq = nil
		tm.currentEnabled = false
		return nil
	}

	normToggle, err := NormalizeHotkeyString(toggleStr)
	if err != nil {
		return fmt.Errorf("“呼出/隐藏客户端”快捷键格式非法: %w", err)
	}

	normDl, err := NormalizeHotkeyString(dlStr)
	if err != nil {
		return fmt.Errorf("“直达下载”快捷键格式非法: %w", err)
	}

	// 内部重复/冲突检测
	if normToggle == normDl {
		return fmt.Errorf("快捷键冲突：“呼出/隐藏客户端”与“直达下载”不能设置为相同的组合键 (%s)", normToggle)
	}

	mod1, vk1, err := parseHotkeyString(normToggle)
	if err != nil {
		return fmt.Errorf("“呼出/隐藏客户端”快捷键解析失败: %w", err)
	}
	mod2, vk2, err := parseHotkeyString(normDl)
	if err != nil {
		return fmt.Errorf("“直达下载”快捷键解析失败: %w", err)
	}

	req := &applyHotkeyReq{
		enabled:    true,
		normToggle: normToggle,
		mod1:       mod1,
		vk1:        vk1,
		normDl:     normDl,
		mod2:       mod2,
		vk2:        vk2,
	}

	// 通过 SendMessage 同步投递到窗口创建线程执行注册，符合 Win32 对窗口所有者线程的硬性要求
	tm.pendingApplyReq = req
	procSendMessageW.Call(uintptr(tm.hwnd), uintptr(WM_APP_APPLY_HOTKEYS), 0, 0)
	tm.pendingApplyReq = nil
	if !req.success {
		if req.errCode == 1 {
			return fmt.Errorf("快捷键“呼出/隐藏客户端”(%s)注册失败：已被系统或其他程序占用，请更换组合键", normToggle)
		} else if req.errCode == 2 {
			return fmt.Errorf("快捷键“直达下载”(%s)注册失败：已被系统或其他程序占用，请更换组合键", normDl)
		}
		return fmt.Errorf("全局快捷键注册失败，请检查按键组合")
	}

	tm.currentToggleStr = normToggle
	tm.currentDlStr = normDl
	tm.currentEnabled = true
	return nil
}

// SuspendHotkeys 挂起或恢复快捷键响应（用于前端录入快捷键时避免快捷键被抢占吞掉）
func SuspendHotkeys(suspend bool) {
	instMu.Lock()
	defer instMu.Unlock()
	if instance == nil || instance.hwnd == 0 {
		return
	}

	tm := instance
	tm.mu.Lock()
	defer tm.mu.Unlock()

	var flag uintptr
	if suspend {
		flag = 1
	}
	procSendMessageW.Call(uintptr(tm.hwnd), uintptr(WM_APP_SUSPEND_HOTKEYS), flag, 0)
}

// CheckHotkeyConflict 检测快捷键是否与当前系统产生冲突
func CheckHotkeyConflict(hotkeyStr string) (bool, error) {
	instMu.Lock()
	defer instMu.Unlock()
	if instance == nil || instance.hwnd == 0 {
		return false, nil
	}

	tm := instance
	tm.mu.Lock()
	defer tm.mu.Unlock()

	norm, err := NormalizeHotkeyString(hotkeyStr)
	if err != nil {
		return false, err
	}

	// 如果正是本程序当前生效的快捷键，判定为无冲突
	if tm.currentEnabled && (norm == tm.currentToggleStr || norm == tm.currentDlStr) {
		return false, nil
	}

	mod, vk, err := parseHotkeyString(norm)
	if err != nil {
		return false, err
	}

	req := &testHotkeyReq{
		mod: mod,
		vk:  vk,
	}
	tm.pendingTestReq = req
	procSendMessageW.Call(uintptr(tm.hwnd), uintptr(WM_APP_TEST_HOTKEY), 0, 0)
	tm.pendingTestReq = nil
	return req.conflict, nil
}

// GetMainWindowHandle 获取主窗口句柄
func GetMainWindowHandle() windows.Handle {
	title, _ := windows.UTF16PtrFromString("Game Trainer Hub - PC 游戏修改器聚合管理客户端")
	hwnd, _, _ := procFindWindowW.Call(0, uintptr(unsafe.Pointer(title)))
	if hwnd != 0 {
		return windows.Handle(hwnd)
	}

	// 备选方案：按进程 ID 枚举窗口
	var foundHwnd windows.Handle
	currPid, _, _ := procGetCurrentProcessId.Call()
	cb := syscall.NewCallback(func(h windows.Handle, lparam uintptr) uintptr {
		var pid uint32
		_, _, _ = procGetWindowThreadProcessId.Call(uintptr(h), uintptr(unsafe.Pointer(&pid)))
		if uintptr(pid) == currPid {
			if instance == nil || h != instance.hwnd {
				foundHwnd = h
				return 0 // 找到停止
			}
		}
		return 1 // 继续
	})
	_, _, _ = procEnumWindows.Call(cb, 0)
	return foundHwnd
}

// IsMainWindowActiveAndVisible 判断主窗口是否可见且处于前台激活状态
func IsMainWindowActiveAndVisible() bool {
	hwnd := GetMainWindowHandle()
	if hwnd == 0 {
		return false
	}
	vis, _, _ := procIsWindowVisible.Call(uintptr(hwnd))
	if vis == 0 {
		return false
	}
	iconic, _, _ := procIsIconic.Call(uintptr(hwnd))
	if iconic != 0 {
		return false
	}
	fg, _, _ := procGetForegroundWindow.Call()
	return fg == uintptr(hwnd)
}

// BringMainWindowToForeground 将主窗口恢复并置于前台激活
func BringMainWindowToForeground() {
	hwnd := GetMainWindowHandle()
	if hwnd != 0 {
		_, _, _ = procShowWindow.Call(uintptr(hwnd), 9) // SW_RESTORE = 9
		_, _, _ = procSetForegroundWindow.Call(uintptr(hwnd))
	}
}

func (tm *TrayManager) messageLoop() {
	runtime.LockOSThread()
	defer runtime.UnlockOSThread()
	defer close(tm.stoppedChan)

	className, _ := windows.UTF16PtrFromString("GameTrainerHub_TrayWindow")
	windowName, _ := windows.UTF16PtrFromString("GameTrainerHub_Tray")

	hInst, _, _ := procGetModuleHandleW.Call(0)

	// 提取或载入应用图标
	hIcon := tm.loadAppIcon()
	tm.hIcon = hIcon

	wndProcCallback := syscall.NewCallback(func(hWnd windows.Handle, msg uint32, wParam, lParam uintptr) uintptr {
		switch msg {
		case WM_APP_APPLY_HOTKEYS:
			req := tm.pendingApplyReq
			if req == nil {
				return 0
			}
			if !req.enabled {
				_, _, _ = procUnregisterHotKey.Call(uintptr(hWnd), uintptr(HOTKEY_ID_TOGGLE))
				_, _, _ = procUnregisterHotKey.Call(uintptr(hWnd), uintptr(HOTKEY_ID_DOWNLOADS))
				tm.currentEnabled = false
				req.success = true
				return 1
			}

			oldToggle := tm.currentToggleStr
			oldDl := tm.currentDlStr
			oldEnabled := tm.currentEnabled

			// 先注销已有按键
			_, _, _ = procUnregisterHotKey.Call(uintptr(hWnd), uintptr(HOTKEY_ID_TOGGLE))
			_, _, _ = procUnregisterHotKey.Call(uintptr(hWnd), uintptr(HOTKEY_ID_DOWNLOADS))

			// 注册 Toggle
			r1, _, _ := procRegisterHotKey.Call(uintptr(hWnd), uintptr(HOTKEY_ID_TOGGLE), uintptr(req.mod1), uintptr(req.vk1))
			if r1 == 0 {
				req.success = false
				req.errCode = 1
				// 发生冲突失败，执行原子回滚
				if oldEnabled && oldToggle != "" {
					if m1, v1, e := parseHotkeyString(oldToggle); e == nil {
						_, _, _ = procRegisterHotKey.Call(uintptr(hWnd), uintptr(HOTKEY_ID_TOGGLE), uintptr(m1), uintptr(v1))
					}
					if oldDl != "" {
						if m2, v2, e := parseHotkeyString(oldDl); e == nil {
							_, _, _ = procRegisterHotKey.Call(uintptr(hWnd), uintptr(HOTKEY_ID_DOWNLOADS), uintptr(m2), uintptr(v2))
						}
					}
				}
				return 0
			}

			// 注册 Downloads
			r2, _, _ := procRegisterHotKey.Call(uintptr(hWnd), uintptr(HOTKEY_ID_DOWNLOADS), uintptr(req.mod2), uintptr(req.vk2))
			if r2 == 0 {
				req.success = false
				req.errCode = 2
				// 注销刚才注册成功的 Toggle
				_, _, _ = procUnregisterHotKey.Call(uintptr(hWnd), uintptr(HOTKEY_ID_TOGGLE))
				// 执行原子回滚
				if oldEnabled && oldToggle != "" {
					if m1, v1, e := parseHotkeyString(oldToggle); e == nil {
						_, _, _ = procRegisterHotKey.Call(uintptr(hWnd), uintptr(HOTKEY_ID_TOGGLE), uintptr(m1), uintptr(v1))
					}
					if oldDl != "" {
						if m2, v2, e := parseHotkeyString(oldDl); e == nil {
							_, _, _ = procRegisterHotKey.Call(uintptr(hWnd), uintptr(HOTKEY_ID_DOWNLOADS), uintptr(m2), uintptr(v2))
						}
					}
				}
				return 0
			}

			req.success = true
			return 1

		case WM_APP_TEST_HOTKEY:
			req := tm.pendingTestReq
			if req == nil {
				return 0
			}
			const testID = 9999
			r, _, _ := procRegisterHotKey.Call(uintptr(hWnd), uintptr(testID), uintptr(req.mod), uintptr(req.vk))
			if r == 0 {
				req.conflict = true
			} else {
				req.conflict = false
				_, _, _ = procUnregisterHotKey.Call(uintptr(hWnd), uintptr(testID))
			}
			return 1

		case WM_APP_SUSPEND_HOTKEYS:
			suspend := wParam != 0
			if suspend {
				if !tm.isSuspended && tm.currentEnabled {
					_, _, _ = procUnregisterHotKey.Call(uintptr(hWnd), uintptr(HOTKEY_ID_TOGGLE))
					_, _, _ = procUnregisterHotKey.Call(uintptr(hWnd), uintptr(HOTKEY_ID_DOWNLOADS))
					tm.isSuspended = true
				}
			} else {
				if tm.isSuspended && tm.currentEnabled {
					if tm.currentToggleStr != "" {
						if m, v, err := parseHotkeyString(tm.currentToggleStr); err == nil {
							_, _, _ = procRegisterHotKey.Call(uintptr(hWnd), uintptr(HOTKEY_ID_TOGGLE), uintptr(m), uintptr(v))
						}
					}
					if tm.currentDlStr != "" {
						if m, v, err := parseHotkeyString(tm.currentDlStr); err == nil {
							_, _, _ = procRegisterHotKey.Call(uintptr(hWnd), uintptr(HOTKEY_ID_DOWNLOADS), uintptr(m), uintptr(v))
						}
					}
					tm.isSuspended = false
				}
			}
			return 1

		case WM_HOTKEY:
			switch wParam {
			case HOTKEY_ID_TOGGLE:
				if tm.onHotkeyToggle != nil {
					go tm.onHotkeyToggle()
				}
				return 0
			case HOTKEY_ID_DOWNLOADS:
				if tm.onHotkeyDownloads != nil {
					go tm.onHotkeyDownloads()
				}
				return 0
			}
		case WM_TRAYICON:
			switch lParam {
			case WM_LBUTTONUP, WM_LBUTTONDBLCLK:
				if tm.onShow != nil {
					go tm.onShow()
				}
				return 0
			case WM_RBUTTONUP:
				tm.showContextMenu(hWnd)
				return 0
			}
		case WM_DESTROY:
			_, _, _ = procPostQuitMessage.Call(0)
			return 0
		}
		r, _, _ := procDefWindowProcW.Call(uintptr(hWnd), uintptr(msg), wParam, lParam)
		return r
	})

	var wcx WNDCLASSEXW
	wcx.CbSize = uint32(unsafe.Sizeof(wcx))
	wcx.LpfnWndProc = wndProcCallback
	wcx.HInstance = windows.Handle(hInst)
	wcx.LpszClassName = className
	wcx.HIcon = hIcon
	wcx.HIconSm = hIcon

	_, _, _ = procRegisterClassExW.Call(uintptr(unsafe.Pointer(&wcx)))

	hwnd, _, _ := procCreateWindowExW.Call(
		0,
		uintptr(unsafe.Pointer(className)),
		uintptr(unsafe.Pointer(windowName)),
		0,
		0, 0, 0, 0,
		0, // HWND_MESSAGE
		0,
		hInst,
		0,
	)
	tm.hwnd = windows.Handle(hwnd)

	// 注册 NotifyIcon
	tm.nid.CbSize = uint32(unsafe.Sizeof(tm.nid))
	tm.nid.HWnd = tm.hwnd
	tm.nid.UID = 1001
	tm.nid.UFlags = NIF_MESSAGE | NIF_ICON | NIF_TIP
	tm.nid.UCallbackMessage = WM_TRAYICON
	tm.nid.HIcon = hIcon

	tipText := "Game Trainer Hub"
	tipW, _ := windows.UTF16FromString(tipText)
	copy(tm.nid.SzTip[:], tipW)

	r, _, _ := procShellNotifyIconW.Call(uintptr(NIM_ADD), uintptr(unsafe.Pointer(&tm.nid)))
	if r != 0 {
		tm.isAdded = true
	}

	close(tm.readyChan)

	var msg MSG
	for {
		res, _, _ := procGetMessageW.Call(uintptr(unsafe.Pointer(&msg)), 0, 0, 0)
		if int32(res) <= 0 {
			break
		}
		_, _, _ = procTranslateMessage.Call(uintptr(unsafe.Pointer(&msg)))
		_, _, _ = procDispatchMessageW.Call(uintptr(unsafe.Pointer(&msg)))
	}

	if tm.isAdded {
		_, _, _ = procShellNotifyIconW.Call(uintptr(NIM_DELETE), uintptr(unsafe.Pointer(&tm.nid)))
		tm.isAdded = false
	}
}

func (tm *TrayManager) showContextMenu(hWnd windows.Handle) {
	hMenu, _, _ := procCreatePopupMenu.Call()
	if hMenu == 0 {
		return
	}
	defer func() {
		_, _, _ = procDestroyMenu.Call(hMenu)
	}()

	tm.mu.Lock()
	showLabel := tm.showText
	quitLabel := tm.quitText
	tm.mu.Unlock()

	showW, _ := windows.UTF16PtrFromString(showLabel)
	quitW, _ := windows.UTF16PtrFromString(quitLabel)

	_, _, _ = procAppendMenuW.Call(hMenu, uintptr(MF_STRING), uintptr(CMD_SHOW), uintptr(unsafe.Pointer(showW)))
	_, _, _ = procAppendMenuW.Call(hMenu, uintptr(MF_SEPARATOR), 0, 0)
	_, _, _ = procAppendMenuW.Call(hMenu, uintptr(MF_STRING), uintptr(CMD_QUIT), uintptr(unsafe.Pointer(quitW)))

	var pt POINT
	_, _, _ = procGetCursorPos.Call(uintptr(unsafe.Pointer(&pt)))

	_, _, _ = procSetForegroundWindow.Call(uintptr(hWnd))
	cmd, _, _ := procTrackPopupMenuEx.Call(
		hMenu,
		uintptr(TPM_RETURNCMD),
		uintptr(pt.X),
		uintptr(pt.Y),
		uintptr(hWnd),
		0,
	)
	_, _, _ = procPostMessageW.Call(uintptr(hWnd), 0, 0, 0)

	switch cmd {
	case CMD_SHOW:
		if tm.onShow != nil {
			go tm.onShow()
		}
	case CMD_QUIT:
		if tm.onQuit != nil {
			go tm.onQuit()
		}
	}
}

func (tm *TrayManager) loadAppIcon() windows.Handle {
	exePath, err := os.Executable()
	if err == nil {
		exePathW, _ := windows.UTF16PtrFromString(exePath)
		var hLarge, hSmall windows.Handle
		count, _, _ := procExtractIconExW.Call(
			uintptr(unsafe.Pointer(exePathW)),
			0,
			uintptr(unsafe.Pointer(&hLarge)),
			uintptr(unsafe.Pointer(&hSmall)),
			1,
		)
		if count > 0 {
			if hSmall != 0 {
				if hLarge != 0 {
					_, _, _ = procDestroyIcon.Call(uintptr(hLarge))
				}
				return hSmall
			}
			if hLarge != 0 {
				return hLarge
			}
		}
	}

	// 回退：加载标准默认应用程序图标
	hIcon, _, _ := procLoadIconW.Call(0, uintptr(32512)) // IDI_APPLICATION
	return windows.Handle(hIcon)
}
