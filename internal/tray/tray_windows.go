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
	modUser32  = windows.NewLazySystemDLL("user32.dll")
	modShell32 = windows.NewLazySystemDLL("shell32.dll")
	modKernel32 = windows.NewLazySystemDLL("kernel32.dll")

	procRegisterClassExW  = modUser32.NewProc("RegisterClassExW")
	procCreateWindowExW   = modUser32.NewProc("CreateWindowExW")
	procDefWindowProcW    = modUser32.NewProc("DefWindowProcW")
	procDestroyWindow     = modUser32.NewProc("DestroyWindow")
	procGetMessageW       = modUser32.NewProc("GetMessageW")
	procTranslateMessage  = modUser32.NewProc("TranslateMessage")
	procDispatchMessageW  = modUser32.NewProc("DispatchMessageW")
	procPostMessageW      = modUser32.NewProc("PostMessageW")
	procPostQuitMessage   = modUser32.NewProc("PostQuitMessage")
	procCreatePopupMenu   = modUser32.NewProc("CreatePopupMenu")
	procAppendMenuW       = modUser32.NewProc("AppendMenuW")
	procTrackPopupMenuEx  = modUser32.NewProc("TrackPopupMenuEx")
	procDestroyMenu       = modUser32.NewProc("DestroyMenu")
	procGetCursorPos      = modUser32.NewProc("GetCursorPos")
	procSetForegroundWindow = modUser32.NewProc("SetForegroundWindow")
	procLoadIconW         = modUser32.NewProc("LoadIconW")
	procDestroyIcon       = modUser32.NewProc("DestroyIcon")

	procShellNotifyIconW  = modShell32.NewProc("Shell_NotifyIconW")
	procExtractIconExW    = modShell32.NewProc("ExtractIconExW")

	procGetModuleHandleW  = modKernel32.NewProc("GetModuleHandleW")

	procRegisterHotKey   = modUser32.NewProc("RegisterHotKey")
	procUnregisterHotKey = modUser32.NewProc("UnregisterHotKey")
)

const (
	WM_APP           = 0x8000
	WM_TRAYICON      = WM_APP + 100
	WM_CLOSE         = 0x0010
	WM_DESTROY       = 0x0002
	WM_LBUTTONUP     = 0x0202
	WM_LBUTTONDBLCLK = 0x0203
	WM_RBUTTONUP     = 0x0205
	WM_HOTKEY        = 0x0312

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
	CbSize           uint32
	HWnd             windows.Handle
	UID              uint32
	UFlags           uint32
	UCallbackMessage uint32
	HIcon            windows.Handle
	SzTip            [128]uint16
	DwState          uint32
	DwStateMask      uint32
	SzInfo           [256]uint16
	UTimeoutOrVersion uint32
	SzInfoTitle      [64]uint16
	DwInfoFlags      uint32
	GuidItem         windows.GUID
	HBalloonIcon     windows.Handle
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

// UpdateHotkeys 动态注册或注销全局快捷键
func UpdateHotkeys(enabled bool, toggleStr, dlStr string) error {
	instMu.Lock()
	defer instMu.Unlock()
	if instance == nil || instance.hwnd == 0 {
		return nil
	}

	// 先注销已有快捷键
	_, _, _ = procUnregisterHotKey.Call(uintptr(instance.hwnd), uintptr(HOTKEY_ID_TOGGLE))
	_, _, _ = procUnregisterHotKey.Call(uintptr(instance.hwnd), uintptr(HOTKEY_ID_DOWNLOADS))

	if !enabled {
		return nil
	}

	var firstErr error
	if toggleStr != "" {
		mod, vk, err := parseHotkeyString(toggleStr)
		if err != nil {
			firstErr = err
		} else {
			r, _, _ := procRegisterHotKey.Call(uintptr(instance.hwnd), uintptr(HOTKEY_ID_TOGGLE), uintptr(mod), uintptr(vk))
			if r == 0 {
				firstErr = fmt.Errorf("快捷键已被系统或其他应用占用: %s", toggleStr)
			}
		}
	}

	if dlStr != "" {
		mod, vk, err := parseHotkeyString(dlStr)
		if err != nil && firstErr == nil {
			firstErr = err
		} else if err == nil {
			r, _, _ := procRegisterHotKey.Call(uintptr(instance.hwnd), uintptr(HOTKEY_ID_DOWNLOADS), uintptr(mod), uintptr(vk))
			if r == 0 && firstErr == nil {
				firstErr = fmt.Errorf("快捷键已被系统或其他应用占用: %s", dlStr)
			}
		}
	}

	return firstErr
}

// CheckHotkeyConflict 检测快捷键是否与当前系统产生冲突
func CheckHotkeyConflict(hotkeyStr string) (bool, error) {
	instMu.Lock()
	defer instMu.Unlock()
	if instance == nil || instance.hwnd == 0 {
		return false, nil
	}

	mod, vk, err := parseHotkeyString(hotkeyStr)
	if err != nil {
		return false, err
	}

	const testID = 9999
	r, _, _ := procRegisterHotKey.Call(uintptr(instance.hwnd), uintptr(testID), uintptr(mod), uintptr(vk))
	if r == 0 {
		return true, nil // 注册失败表示被其他程序占用了
	}
	_, _, _ = procUnregisterHotKey.Call(uintptr(instance.hwnd), uintptr(testID))
	return false, nil
}

func parseHotkeyString(s string) (uint32, uint32, error) {
	parts := strings.Split(s, "+")
	var mod uint32 = MOD_NOREPEAT
	var vk uint32

	for _, p := range parts {
		trimmed := strings.TrimSpace(strings.ToLower(p))
		switch trimmed {
		case "alt":
			mod |= MOD_ALT
		case "ctrl", "control":
			mod |= MOD_CONTROL
		case "shift":
			mod |= MOD_SHIFT
		case "win", "super":
			mod |= MOD_WIN
		default:
			upper := strings.ToUpper(trimmed)
			if len(upper) == 1 && upper[0] >= 'A' && upper[0] <= 'Z' {
				vk = uint32(upper[0])
			} else if len(upper) == 1 && upper[0] >= '0' && upper[0] <= '9' {
				vk = uint32(upper[0])
			} else if strings.HasPrefix(upper, "F") && len(upper) >= 2 {
				num := 0
				for _, ch := range upper[1:] {
					if ch >= '0' && ch <= '9' {
						num = num*10 + int(ch-'0')
					}
				}
				if num >= 1 && num <= 12 {
					vk = uint32(0x70 + num - 1)
				}
			} else if upper == "SPACE" {
				vk = 0x20
			} else if upper == "ESC" || upper == "ESCAPE" {
				vk = 0x1B
			} else if upper == "TAB" {
				vk = 0x09
			} else {
				return 0, 0, fmt.Errorf("不支持的按键: %s", p)
			}
		}
	}

	if vk == 0 {
		return 0, 0, fmt.Errorf("快捷键必须包含一个主按键: %s", s)
	}
	return mod, vk, nil
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
