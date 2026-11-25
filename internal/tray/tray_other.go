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

func ApplyHotkeys(enabled bool, toggleStr, dlStr string) error { return nil }

func SuspendHotkeys(suspend bool) {}

func NormalizeHotkeyString(s string) (string, error) { return s, nil }

func CheckHotkeyConflict(hotkeyStr string) (bool, error) { return false, nil }

func IsMainWindowActiveAndVisible() bool { return false }

func BringMainWindowToForeground() {}

func Destroy() {}
