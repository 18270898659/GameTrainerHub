package download

import (
	"archive/zip"
	"bytes"
	"os"
	"path/filepath"
	"testing"
)

func TestExtractTrainerZipSecurity(t *testing.T) {
	tempDir, err := os.MkdirTemp("", "gth_extract_test_*")
	if err != nil {
		t.Fatalf("failed to create temp dir: %v", err)
	}
	defer os.RemoveAll(tempDir)

	// 创建一个包含合法文件和恶意 Zip Slip 穿越文件的测试压缩包
	zipBuffer := new(bytes.Buffer)
	zipWriter := zip.NewWriter(zipBuffer)

	// 1. 合法文件
	validFile, err := zipWriter.Create("EldenRing_v1.0_Trainer.exe")
	if err != nil {
		t.Fatalf("failed to create zip entry: %v", err)
	}
	_, _ = validFile.Write([]byte("MZ_NORMAL_EXE_CONTENT"))

	// 2. 恶意 Zip Slip 穿越文件
	maliciousSlip, err := zipWriter.Create("../../evil_escaped.exe")
	if err != nil {
		t.Fatalf("failed to create malicious zip entry: %v", err)
	}
	_, _ = maliciousSlip.Write([]byte("MALICIOUS_PAYLOAD"))

	// 3. 根目录绝对路径穿越模拟
	maliciousAbs, err := zipWriter.Create("/Windows/System32/evil.dll")
	if err != nil {
		t.Fatalf("failed to create abs zip entry: %v", err)
	}
	_, _ = maliciousAbs.Write([]byte("ABS_MALICIOUS_PAYLOAD"))

	_ = zipWriter.Close()

	testZipPath := filepath.Join(tempDir, "test_trainer.zip")
	if err := os.WriteFile(testZipPath, zipBuffer.Bytes(), 0644); err != nil {
		t.Fatalf("failed to write test zip: %v", err)
	}

	destDir := filepath.Join(tempDir, "extracted")

	exePath, err := ExtractTrainerZip(testZipPath, destDir)
	if err != nil {
		t.Fatalf("ExtractTrainerZip failed: %v", err)
	}

	// 验证合法文件成功解压
	expectedExe := filepath.Join(destDir, "EldenRing_v1.0_Trainer.exe")
	if exePath != expectedExe {
		t.Errorf("expected main exe path %s, got %s", expectedExe, exePath)
	}

	// 验证恶意穿越文件未逃逸出 destDir
	escapedFile := filepath.Join(tempDir, "evil_escaped.exe")
	if _, err := os.Stat(escapedFile); !os.IsNotExist(err) {
		t.Errorf("CRITICAL SECURITY FLAW: Zip Slip file was written to escaped location: %s", escapedFile)
	}
}
