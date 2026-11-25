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

func TestNormalizeHotkeyString(t *testing.T) {
	tests := []struct {
		input       string
		expected    string
		expectError bool
	}{
		{"Alt+Ctrl+Q", "Alt + Ctrl + Q", false},
		{"Ctrl+Alt+q", "Alt + Ctrl + Q", false},
		{"ctrl + alt + j", "Alt + Ctrl + J", false},
		{"Shift+Ctrl+Alt+Q", "Alt + Ctrl + Shift + Q", false},
		{"F12", "F12", false},
		{"ctrl+f1", "Ctrl + F1", false},
		{"", "", true},
		{"Q", "", true},                // 单独普通字母必须包含修饰键
		{"Ctrl", "", true},            // 仅修饰键无主键
		{"Ctrl+Alt", "", true},        // 仅修饰键无主键
		{"Ctrl+Q+J", "", true},        // 多个主按键
	}

	for _, tc := range tests {
		actual, err := NormalizeHotkeyString(tc.input)
		if tc.expectError {
			if err == nil {
				t.Errorf("NormalizeHotkeyString(%q) expected error, but got nil (result: %q)", tc.input, actual)
			}
		} else {
			if err != nil {
				t.Errorf("NormalizeHotkeyString(%q) unexpected error: %v", tc.input, err)
			} else if actual != tc.expected {
				t.Errorf("NormalizeHotkeyString(%q) = %q; want %q", tc.input, actual, tc.expected)
			}
		}
	}
}

func TestInternalHotkeyConflict(t *testing.T) {
	tm, err := Init(func() {}, func() {})
	if err != nil {
		t.Fatalf("Init failed: %v", err)
	}
	defer Destroy()
	_ = tm

	// 相同的快捷键组合应判定为内部冲突
	err = ApplyHotkeys(true, "Alt+Ctrl+Q", "Ctrl+Alt+q")
	if err == nil {
		t.Errorf("expected internal conflict error when setting identical hotkeys, got nil")
	}

	// 不同的快捷键应能正常处理
	err = ApplyHotkeys(true, "Alt + Ctrl + Q", "Alt + Ctrl + J")
	if err != nil {
		t.Logf("ApplyHotkeys notice: %v", err)
	}
}
