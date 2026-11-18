package logger

import (
	"os"
	"path/filepath"
	"testing"
)

func TestLoggerInitAndWrite(t *testing.T) {
	tempDir, err := os.MkdirTemp("", "gth_log_test_*")
	if err != nil {
		t.Fatalf("failed to create temp dir: %v", err)
	}
	defer os.RemoveAll(tempDir)

	l, err := InitLogger(tempDir, "debug")
	if err != nil {
		t.Fatalf("failed to init logger: %v", err)
	}
	if l == nil {
		t.Fatal("expected non-nil logger")
	}

	Debug("debug test message", "key1", "val1")
	Info("info test message", "key2", "val2")
	Warn("warn test message", "key3", "val3")
	Error("error test message", "key4", "val4")

	CloseLogger()

	files, err := os.ReadDir(tempDir)
	if err != nil {
		t.Fatalf("failed to read log dir: %v", err)
	}
	if len(files) == 0 {
		t.Fatal("expected log file to be created")
	}

	content, err := os.ReadFile(filepath.Join(tempDir, files[0].Name()))
	if err != nil {
		t.Fatalf("failed to read log file: %v", err)
	}
	if len(content) == 0 {
		t.Fatal("expected log file to have content")
	}
}

