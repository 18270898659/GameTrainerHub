package tray

import (
	"testing"
	"time"
)

func TestTrayLifecycle(t *testing.T) {
	shown := false
	quitCalled := false

	tm, err := Init(func() {
		shown = true
	}, func() {
		quitCalled = true
	})

	if err != nil {
		t.Fatalf("Init failed: %v", err)
	}
	if tm == nil {
		t.Fatal("expected non-nil TrayManager")
	}

	UpdateLabels("显示", "退出")
	time.Sleep(50 * time.Millisecond)

	Destroy()

	_ = shown
	_ = quitCalled
}

func TestHotkeyAndNotification(t *testing.T) {
	tm, err := Init(func() {}, func() {})
	if err != nil {
		t.Fatalf("Init failed: %v", err)
	}
	defer Destroy()

	SetHotkeyHandlers(func() {}, func() {})

	// 测试快捷键注册
	if err := UpdateHotkeys(true, "Alt+Q", "Ctrl+J"); err != nil {
		t.Logf("UpdateHotkeys notice: %v", err)
	}

	// 测试快捷键冲突检测
	isConflict, err := CheckHotkeyConflict("Alt+Q")
	if err != nil {
		t.Fatalf("CheckHotkeyConflict failed: %v", err)
	}
	t.Logf("Alt+Q conflict state: %v", isConflict)

	// 测试通知气泡
	err = ShowNotification("测试标题", "测试内容", false)
	if err != nil {
		t.Logf("ShowNotification notice: %v", err)
	}

	_ = tm
}
