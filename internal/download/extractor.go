package download

import (
	"archive/zip"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"strings"
)

// ExtractTrainerZip 解压 ZIP 修改器包，返回主可执行文件 (.exe) 的路径
func ExtractTrainerZip(zipPath string, destDir string) (string, error) {
	reader, err := zip.OpenReader(zipPath)
	if err != nil {
		return "", fmt.Errorf("failed to open zip file: %w", err)
	}
	defer reader.Close()

	if err := os.MkdirAll(destDir, 0755); err != nil {
		return "", fmt.Errorf("failed to create destination dir: %w", err)
	}

	var mainExePath string

	for _, file := range reader.File {
		// 防止 Zip Slip 漏洞 (路径穿越检查)
		cleanPath := filepath.Clean(file.Name)
		if strings.HasPrefix(cleanPath, "..") || strings.Contains(cleanPath, ":") {
			continue
		}

		targetPath := filepath.Join(destDir, cleanPath)

		if file.FileInfo().IsDir() {
			_ = os.MkdirAll(targetPath, file.Mode())
			continue
		}

		if err := os.MkdirAll(filepath.Dir(targetPath), 0755); err != nil {
			return "", err
		}

		srcFile, err := file.Open()
		if err != nil {
			return "", err
		}

		destFile, err := os.OpenFile(targetPath, os.O_WRONLY|os.O_CREATE|os.O_TRUNC, file.Mode())
		if err != nil {
			srcFile.Close()
			return "", err
		}

		_, copyErr := io.Copy(destFile, srcFile)
		srcFile.Close()
		destFile.Close()
		if copyErr != nil {
			return "", copyErr
		}

		// 记录主可执行文件
		if strings.HasSuffix(strings.ToLower(targetPath), ".exe") && mainExePath == "" {
			mainExePath = targetPath
		}
	}

	if mainExePath == "" {
		// 如果包内没有 exe，返回解压目录路径
		return destDir, nil
	}
	return mainExePath, nil
}

