//go:build !windows

package tray

type TrayManager struct{}

func Init(onShow, onQuit func()) (*TrayManager, error) {
	return &TrayManager{}, nil
}

func UpdateLabels(showText, quitText string) {}

func Destroy() {}
