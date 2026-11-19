package diagnostics

import (
	"context"
	"fmt"
	"net"
	"net/http"
	"os"
	"path/filepath"
	"runtime"
	"strings"
	"time"

	"GameTrainerHub/internal/config"
	"GameTrainerHub/internal/database"
	"GameTrainerHub/internal/download"
	"GameTrainerHub/internal/logger"
	"GameTrainerHub/internal/model"
)

// RunNetworkDiagnosis 测试对资源源与关键域名的连通性及延迟
func RunNetworkDiagnosis(targetHost string) *model.NetworkDiagnosis {
	if targetHost == "" {
		targetHost = "flingtrainer.com"
	}

	diag := &model.NetworkDiagnosis{
		TargetHost: targetHost,
	}

	// 1. DNS 解析
	ips, err := net.LookupIP(targetHost)
	if err != nil {
		diag.DNSResolved = false
		diag.ErrorMessage = fmt.Sprintf("DNS 解析失败: %v", err)
		return diag
	}
	diag.DNSResolved = true
	var ipStrs []string
	for _, ip := range ips {
		ipStrs = append(ipStrs, ip.String())
	}
	diag.IPAddresses = strings.Join(ipStrs, ", ")

	// 2. HTTP 探测连通性与延迟
	start := time.Now()
	client := &http.Client{
		Timeout: 5 * time.Second,
	}
	url := fmt.Sprintf("https://%s", targetHost)
	resp, err := client.Get(url)
	latency := time.Since(start).Milliseconds()
	diag.LatencyMs = latency

	if err != nil {
		diag.IsReachable = false
		diag.ErrorMessage = fmt.Sprintf("HTTP 连接失败: %v (耗时 %dms)", err, latency)
		return diag
	}
	defer resp.Body.Close()

	diag.HTTPStatus = resp.StatusCode
	diag.IsReachable = (resp.StatusCode >= 200 && resp.StatusCode < 500)
	return diag
}

// DiagnoseEnvironment 执行系统全方位运行与环境检测
func DiagnoseEnvironment(cfg *config.AppConfig, db *database.DB, activeDlCount, localCount, favCount int) (*model.DiagnosticReport, error) {
	rep := &model.DiagnosticReport{
		GeneratedAt:        time.Now(),
		AppVersion:         "1.0.0",
		OS:                 runtime.GOOS,
		Arch:               runtime.GOARCH,
		DownloadDir:        cfg.DownloadDir,
		ActiveDownloads:    activeDlCount,
		LocalTrainersCount: localCount,
		FavoritesCount:     favCount,
		DBPath:             cfg.GetDBPath(),
	}

	// 1. 下载目录可写性检查
	testFile := filepath.Join(cfg.DownloadDir, fmt.Sprintf(".diag_write_%d.tmp", time.Now().UnixNano()))
	if err := os.WriteFile(testFile, []byte("diag"), 0644); err == nil {
		rep.DownloadDirWritable = true
		_ = os.Remove(testFile)
	} else {
		rep.DownloadDirWritable = false
	}

	// 2. 磁盘空间检查
	if err := download.CheckDiskSpace(cfg.DownloadDir, 1024*1024); err == nil {
		rep.AvailableDiskBytes = 1024 * 1024 * 1024 // 充足空间
	}

	// 3. 数据库完整性检查 (PRAGMA integrity_check)
	var intCheck string
	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()
	if db != nil {
		err := db.QueryRowContext(ctx, "PRAGMA integrity_check;").Scan(&intCheck)
		rep.DBIntegrityOK = (err == nil && strings.ToLower(intCheck) == "ok")
	}

	// 4. 网络诊断
	rep.Network = RunNetworkDiagnosis("flingtrainer.com")

	return rep, nil
}

// ExportDiagnosticsReport 生成并导出诊断报告文件
func ExportDiagnosticsReport(report *model.DiagnosticReport, logs []model.LogEntry, exportDir string) (string, error) {
	if err := os.MkdirAll(exportDir, 0755); err != nil {
		return "", err
	}

	nowStr := time.Now().Format("20060102_150405")
	fileName := fmt.Sprintf("diagnostics_%s.txt", nowStr)
	filePath := filepath.Join(exportDir, fileName)

	var sb strings.Builder
	sb.WriteString("==================================================================\n")
	sb.WriteString("                 Game Trainer Hub 诊断与运行报告                  \n")
	sb.WriteString("==================================================================\n\n")

	sb.WriteString(fmt.Sprintf("生成时间: %s\n", report.GeneratedAt.Format("2006-01-02 15:04:05")))
	sb.WriteString(fmt.Sprintf("客户端版本: %s\n", report.AppVersion))
	sb.WriteString(fmt.Sprintf("操作系统环境: %s/%s\n", report.OS, report.Arch))
	sb.WriteString(fmt.Sprintf("默认下载路径: %s (可写: %v)\n", report.DownloadDir, report.DownloadDirWritable))
	sb.WriteString(fmt.Sprintf("SQLite数据库: %s (完整性检查: %v)\n", report.DBPath, report.DBIntegrityOK))
	sb.WriteString(fmt.Sprintf("活跃任务数: %d | 本地修改器: %d | 收藏游戏: %d\n\n",
		report.ActiveDownloads, report.LocalTrainersCount, report.FavoritesCount))

	sb.WriteString("---------------------- 网络环境连通性检测 ----------------------\n")
	if report.Network != nil {
		sb.WriteString(fmt.Sprintf("目标服务器: %s\n", report.Network.TargetHost))
		sb.WriteString(fmt.Sprintf("DNS解析状态: %v (IP: %s)\n", report.Network.DNSResolved, report.Network.IPAddresses))
		sb.WriteString(fmt.Sprintf("HTTP状态码: %d | 响应延迟: %dms | 可连通: %v\n",
			report.Network.HTTPStatus, report.Network.LatencyMs, report.Network.IsReachable))
		if report.Network.ErrorMessage != "" {
			sb.WriteString(fmt.Sprintf("异常信息: %s\n", report.Network.ErrorMessage))
		}
	} else {
		sb.WriteString("无网络诊断信息\n")
	}
	sb.WriteString("\n")

	sb.WriteString("---------------------- 最近运行日志摘要 (已脱敏) ----------------------\n")
	if len(logs) == 0 {
		sb.WriteString("暂无近期日志条目\n")
	} else {
		for _, l := range logs {
			maskedMsg := logger.MaskSensitiveData(l.Message)
			maskedDetails := logger.MaskSensitiveData(l.Details)
			sb.WriteString(fmt.Sprintf("[%s] [%s] [%s] %s %s\n",
				l.Timestamp.Format("15:04:05.000"), l.Level, l.Category, maskedMsg, maskedDetails))
		}
	}
	sb.WriteString("\n==================================================================\n")

	err := os.WriteFile(filePath, []byte(sb.String()), 0644)
	if err != nil {
		return "", err
	}
	return filePath, nil
}
