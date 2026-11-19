package update

import (
	"context"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"testing"
	"time"
)

func TestCompareVersions(t *testing.T) {
	tests := []struct {
		v1       string
		v2       string
		expected int
	}{
		{"1.0.1", "1.0.0", 1},
		{"v1.2.0", "1.1.9", 1},
		{"1.0.0", "1.0.0", 0},
		{"v2.0.0", "v2.0.0", 0},
		{"0.9.9", "1.0.0", -1},
		{"1.0", "1.0.1", -1},
		{"2.1.0", "2.0.9", 1},
	}

	for _, tc := range tests {
		res := CompareVersions(tc.v1, tc.v2)
		if res != tc.expected {
			t.Errorf("CompareVersions(%s, %s): expected %d, got %d", tc.v1, tc.v2, tc.expected, res)
		}
	}
}

func TestCheckUpdateAndPackageVerification(t *testing.T) {
	tempDir, err := os.MkdirTemp("", "gth_update_test_*")
	if err != nil {
		t.Fatalf("failed to create temp dir: %v", err)
	}
	defer os.RemoveAll(tempDir)

	mockPayload := []byte("MZ_MOCK_INSTALLER_PAYLOAD_V1.1.0")
	hash := sha256.Sum256(mockPayload)
	expectedHash := hex.EncodeToString(hash[:])
	expectedSize := int64(len(mockPayload))

	mockServer := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		_ = json.NewEncoder(w).Encode(ReleaseInfo{
			Version:      "1.1.0",
			DownloadURL:  "https://example.com/setup-1.1.0.exe",
			SHA256:       expectedHash,
			FileSize:     expectedSize,
			ReleaseNotes: "Added new search features and security updates",
			Mandatory:    false,
			PublishedAt:  time.Now(),
		})
	}))
	defer mockServer.Close()

	svc := NewUpdateService("1.0.0", mockServer.URL, mockServer.Client())

	ctx := context.Background()
	release, hasNew, err := svc.CheckUpdate(ctx)
	if err != nil {
		t.Fatalf("CheckUpdate failed: %v", err)
	}

	if !hasNew {
		t.Fatal("expected hasNew to be true for 1.1.0 vs 1.0.0")
	}
	if release.Version != "1.1.0" {
		t.Errorf("expected version 1.1.0, got %s", release.Version)
	}

	// 1. 测试合法安装包校验
	pkgFile := filepath.Join(tempDir, "setup.exe")
	if err := os.WriteFile(pkgFile, mockPayload, 0644); err != nil {
		t.Fatalf("failed to write mock installer: %v", err)
	}

	valid, err := svc.VerifyUpdatePackage(pkgFile, expectedHash, expectedSize)
	if err != nil || !valid {
		t.Fatalf("expected valid package verification, got %v, err: %v", valid, err)
	}

	// 2. 测试哈希篡改包校验失败
	corruptedHash := "0000000000000000000000000000000000000000000000000000000000000000"
	validCorrupt, errCorrupt := svc.VerifyUpdatePackage(pkgFile, corruptedHash, expectedSize)
	if validCorrupt || errCorrupt == nil {
		t.Fatal("expected failure on hash mismatch, got valid")
	}

	// 3. 测试文件体积不一致校验失败
	validSize, errSize := svc.VerifyUpdatePackage(pkgFile, expectedHash, expectedSize+100)
	if validSize || errSize == nil {
		t.Fatal("expected failure on file size mismatch, got valid")
	}
}

