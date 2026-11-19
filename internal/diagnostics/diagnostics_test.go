package diagnostics

import (
	"os"
	"path/filepath"
	"testing"
	"time"

	"GameTrainerHub/internal/config"
	"GameTrainerHub/internal/model"
)

func TestDiagnostics(t *testing.T) {
	tempDir, err := os.MkdirTemp("", "gth_diag_test_*")
	if err != nil {
		t.Fatalf("failed to create temp dir: %v", err)
	}
	defer os.RemoveAll(tempDir)

	cfg := &config.AppConfig{
		BaseDir:     tempDir,
		DownloadDir: filepath.Join(tempDir, "downloads"),
		DataDir:     filepath.Join(tempDir, "data"),
	}
	_ = cfg.EnsureDirectories()

	rep, err := DiagnoseEnvironment(cfg, nil, 1, 2, 3)
	if err != nil {
		t.Fatalf("DiagnoseEnvironment failed: %v", err)
	}
	if rep == nil {
		t.Fatal("expected non-nil report")
	}
	if !rep.DownloadDirWritable {
		t.Errorf("expected download dir writable")
	}

	logs := []model.LogEntry{
		{
			Timestamp: time.Now(),
			Level:     "INFO",
			Category:  "download",
			Message:   "Starting test download",
			Details:   "token=secret123",
		},
	}

	exportPath, err := ExportDiagnosticsReport(rep, logs, tempDir)
	if err != nil {
		t.Fatalf("ExportDiagnosticsReport failed: %v", err)
	}
	if _, err := os.Stat(exportPath); os.IsNotExist(err) {
		t.Fatalf("expected export file %s to exist", exportPath)
	}

	content, _ := os.ReadFile(exportPath)
	if !containsStr(string(content), "token=***") {
		t.Errorf("expected token to be masked in export, got:\n%s", string(content))
	}
}

func containsStr(s, sub string) bool {
	return len(s) >= len(sub) && (s == sub || len(sub) == 0 || (len(s) > 0 && len(sub) > 0 && stringContains(s, sub)))
}

func stringContains(s, sub string) bool {
	for i := 0; i+len(sub) <= len(s); i++ {
		if s[i:i+len(sub)] == sub {
			return true
		}
	}
	return false
}
