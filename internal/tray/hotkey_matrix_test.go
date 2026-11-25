//go:build windows

package tray

import (
	"strings"
	"testing"
	"time"
)

// TestHotkeyMatrixNormalization 测试快捷键不同顺序、大小写、空白的识别与规范化
func TestHotkeyMatrixNormalization(t *testing.T) {
	cases := []struct {
		input       string
		expected    string
		expectError bool
		description string
	}{
		{"Alt + Ctrl + Q", "Alt + Ctrl + Q", false, "标准顺序 Alt + Ctrl + Q"},
		{"Ctrl + Alt + Q", "Alt + Ctrl + Q", false, "乱序 Ctrl + Alt + Q 自动统一为 Alt + Ctrl + Q"},
		{"ctrl+alt+q", "Alt + Ctrl + Q", false, "小写无空格 ctrl+alt+q 规范化"},
		{"ALT+CTRL+Q", "Alt + Ctrl + Q", false, "大写无空格 ALT+CTRL+Q 规范化"},
		{"Alt + Ctrl + J", "Alt + Ctrl + J", false, "标准顺序 Alt + Ctrl + J"},
		{"Ctrl + Alt + J", "Alt + Ctrl + J", false, "乱序 Ctrl + Alt + J 自动统一为 Alt + Ctrl + J"},
		{"ctrl+alt+j", "Alt + Ctrl + J", false, "小写无空格 ctrl+alt+j 规范化"},
		{"Shift + Ctrl + Alt + K", "Alt + Ctrl + Shift + K", false, "三键修饰组合规范化"},
		{"Win + Alt + Ctrl + D", "Alt + Ctrl + Win + D", false, "包含 Win 键的四键修饰组合"},
		{"F11", "F11", false, "F 功能键允许无修饰键"},
		{"Ctrl + F5", "Ctrl + F5", false, "修饰键 + F 功能键"},
		{"Ctrl + Space", "Ctrl + Space", false, "修饰键 + 空格键"},
		{"", "", true, "空字符串应报错"},
		{"   ", "", true, "纯空格应报错"},
		{"Q", "", true, "单个普通字符无修饰键应报错"},
		{"Ctrl + Alt", "", true, "仅有修饰键无主键应报错"},
		{"Alt + Ctrl + Q + J", "", true, "包含多个主键应报错"},
	}

	for _, c := range cases {
		actual, err := NormalizeHotkeyString(c.input)
		if c.expectError {
			if err == nil {
				t.Errorf("[%s] 输入 %q 预期应返回错误，但实际返回成功: %q", c.description, c.input, actual)
			}
		} else {
			if err != nil {
				t.Errorf("[%s] 输入 %q 解析报错: %v", c.description, c.input, err)
			} else if actual != c.expected {
				t.Errorf("[%s] 输入 %q 结果 = %q; 预期 = %q", c.description, c.input, actual, c.expected)
			}
		}
	}
}

// TestHotkeyMatrixRegistrationAndRollback 测试真实 Windows 下快捷键注册、占用检测与原子回滚
func TestHotkeyMatrixRegistrationAndRollback(t *testing.T) {
	tm, err := Init(func() {}, func() {})
	if err != nil {
		t.Fatalf("Init TrayManager failed: %v", err)
	}
	defer Destroy()

	// 1. 内部快捷键重复/冲突检测 (两组功能不能相同)
	err = ApplyHotkeys(true, "Alt + Ctrl + J", "Alt + Ctrl + J")
	if err == nil {
		t.Errorf("预期两组相同快捷键应报错，但未报错")
	} else if !strings.Contains(err.Error(), "不能设置为相同的组合键") {
		t.Errorf("冲突报错信息不符合预期: %v", err)
	}

	// 2. 乱序输入检测内部冲突 (Ctrl+Alt+J 与 Alt+Ctrl+J 本质相同)
	err = ApplyHotkeys(true, "Ctrl + Alt + J", "Alt + Ctrl + J")
	if err == nil {
		t.Errorf("预期乱序相同快捷键应报错冲突，但未报错")
	}

	// 3. 占用检测：当按键已被外部程序 (如系统或 QQ 占用的 Alt+Ctrl+Q) 占用时，应正确识别已被占用
	isConflict, err := CheckHotkeyConflict("Alt + Ctrl + Q")
	if err != nil {
		t.Fatalf("CheckHotkeyConflict 报错: %v", err)
	}
	t.Logf("Alt + Ctrl + Q 外部程序占用状态: %v", isConflict)

	// 4. 成功注册未被占用的有效全局快捷键 (例如 Alt + Shift + Y, Alt + Shift + U)
	err = ApplyHotkeys(true, "Alt + Shift + Y", "Alt + Shift + U")
	if err != nil {
		t.Fatalf("ApplyHotkeys 注册有效组合失败: %v", err)
	}

	// 5. 自查当前已注册快捷键不应误报外部冲突 (本程序自己持有的按键判定为无冲突)
	selfConflict, err := CheckHotkeyConflict("Alt + Shift + Y")
	if err != nil {
		t.Errorf("CheckHotkeyConflict 报错: %v", err)
	}
	if selfConflict {
		t.Errorf("自己已注册的快捷键 Alt + Shift + Y 不应被误判为冲突")
	}

	// 6. 挂起与恢复测试 (模拟前端录入状态)
	SuspendHotkeys(true)
	time.Sleep(50 * time.Millisecond)
	SuspendHotkeys(false)
	time.Sleep(50 * time.Millisecond)

	// 7. 原子回滚测试：尝试注册一个冲突的快捷键组合，原有有效快捷键应被原子保留
	if isConflict {
		failErr := ApplyHotkeys(true, "Alt + Ctrl + Q", "Alt + Shift + U")
		if failErr == nil {
			t.Errorf("预期注册被占用的 Alt + Ctrl + Q 应报错，但返回成功")
		} else {
			t.Logf("成功捕获真实 Windows 快捷键占用异常并安全回滚: %v", failErr)
		}
	}

	// 8. 禁用全局快捷键测试
	err = ApplyHotkeys(false, "Alt + Shift + Y", "Alt + Shift + U")
	if err != nil {
		t.Errorf("禁用快捷键不应报错: %v", err)
	}

	_ = tm
}
