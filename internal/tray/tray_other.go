//go:build !windows

package tray

type TrayManager struct{}

func Init(onShow, onQuit func()) (*TrayManager, error) {
	return &TrayManager{}, nil
}

func UpdateLabels(showText, quitText string) {}

func ShowNotification(title, message string, isError bool) error { return nil }

func SetHotkeyHandlers(onToggle, onDownloads func()) {}

func UpdateHotkeys(enabled bool, toggleStr, dlStr string) error { return nil }

func CheckHotkeyConflict(hotkeyStr string) (bool, error) { return false, nil }

func Destroy() {}
