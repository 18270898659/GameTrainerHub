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
