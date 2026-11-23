package database

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"sync"
	"testing"
)

func TestOpenAndMigrations(t *testing.T) {
	tempDir, err := os.MkdirTemp("", "gth_db_test_*")
	if err != nil {
		t.Fatalf("failed to create temp dir: %v", err)
	}
	defer os.RemoveAll(tempDir)

	dbPath := filepath.Join(tempDir, "test.db")
	db, err := Open(dbPath)
	if err != nil {
		t.Fatalf("failed to open database: %v", err)
	}
	defer db.Close()

	ctx := context.Background()

	// 1. 验证迁移版本记录
	versions, err := db.GetAppliedMigrations(ctx)
	if err != nil {
		t.Fatalf("failed to get applied migrations: %v", err)
	}
	if len(versions) == 0 {
		t.Fatal("expected at least one applied migration")
	}
	if versions[0] != 1 {
		t.Errorf("expected migration version 1, got %d", versions[0])
	}

	// 2. 验证外键约束开启
	var fkEnabled int
	if err := db.QueryRowContext(ctx, "PRAGMA foreign_keys").Scan(&fkEnabled); err != nil {
		t.Fatalf("failed to check foreign_keys pragma: %v", err)
	}
	if fkEnabled != 1 {
		t.Errorf("expected PRAGMA foreign_keys=1, got %d", fkEnabled)
	}

	// 3. 验证默认 sources 数据插入
	var flingName string
	err = db.QueryRowContext(ctx, "SELECT name FROM sources WHERE id = 'fling'").Scan(&flingName)
	if err != nil {
		t.Fatalf("failed to query default source fling: %v", err)
	}
	if flingName != "FLiNG Trainer" {
		t.Errorf("expected 'FLiNG Trainer', got '%s'", flingName)
	}
}

func TestTransactionRollback(t *testing.T) {
	tempDir, err := os.MkdirTemp("", "gth_tx_test_*")
	if err != nil {
		t.Fatalf("failed to create temp dir: %v", err)
	}
	defer os.RemoveAll(tempDir)

	dbPath := filepath.Join(tempDir, "test.db")
	db, err := Open(dbPath)
	if err != nil {
		t.Fatalf("failed to open database: %v", err)
	}
	defer db.Close()

	ctx := context.Background()

	// 执行抛出错误的事务
	err = db.WithTx(ctx, func(tx *sql.Tx) error {
		_, execErr := tx.ExecContext(ctx, `
			INSERT INTO app_settings (key, value) VALUES ('tx_test', 'val1')
		`)
		if execErr != nil {
			return execErr
		}
		return errors.New("simulated error to trigger rollback")
	})

	if err == nil {
		t.Fatal("expected transaction error, got nil")
	}

	// 验证回滚后数据未被写入
	var count int
	_ = db.QueryRowContext(ctx, "SELECT COUNT(*) FROM app_settings WHERE key = 'tx_test'").Scan(&count)
	if count != 0 {
		t.Errorf("expected rollback to leave 0 rows, found %d", count)
	}
}

func TestTransactionCommit(t *testing.T) {
	tempDir, err := os.MkdirTemp("", "gth_tx_commit_*")
	if err != nil {
		t.Fatalf("failed to create temp dir: %v", err)
	}
	defer os.RemoveAll(tempDir)

	dbPath := filepath.Join(tempDir, "test.db")
	db, err := Open(dbPath)
	if err != nil {
		t.Fatalf("failed to open database: %v", err)
	}
	defer db.Close()

	ctx := context.Background()

	err = db.WithTx(ctx, func(tx *sql.Tx) error {
		_, execErr := tx.ExecContext(ctx, `
			INSERT INTO app_settings (key, value) VALUES ('tx_commit', 'success')
		`)
		return execErr
	})

	if err != nil {
		t.Fatalf("unexpected tx error: %v", err)
	}

	var val string
	err = db.QueryRowContext(ctx, "SELECT value FROM app_settings WHERE key = 'tx_commit'").Scan(&val)
	if err != nil {
		t.Fatalf("failed to query committed value: %v", err)
	}
	if val != "success" {
		t.Errorf("expected 'success', got '%s'", val)
	}
}

func TestIntegrityCheckAndBackup(t *testing.T) {
	tempDir, err := os.MkdirTemp("", "gth_backup_test_*")
	if err != nil {
		t.Fatalf("failed to create temp dir: %v", err)
	}
	defer os.RemoveAll(tempDir)

	dbPath := filepath.Join(tempDir, "primary.db")
	db, err := Open(dbPath)
	if err != nil {
		t.Fatalf("failed to open primary database: %v", err)
	}
	defer db.Close()

	ctx := context.Background()

	// 1. 完整性检查
	if err := db.IntegrityCheck(ctx); err != nil {
		t.Fatalf("initial integrity check failed: %v", err)
	}

	// 写入一些数据
	_, _ = db.ExecContext(ctx, "INSERT INTO app_settings (key, value) VALUES ('backup_key', 'backup_val');")

	// 2. 在线热备份
	backupPath := filepath.Join(tempDir, "backup", "snapshot.db")
	if err := db.Backup(ctx, backupPath); err != nil {
		t.Fatalf("database backup failed: %v", err)
	}

	// 3. 验证从备份库中可以读取到数据
	backupDB, err := Open(backupPath)
	if err != nil {
		t.Fatalf("failed to open backup database: %v", err)
	}
	defer backupDB.Close()

	var backupVal string
	err = backupDB.QueryRowContext(ctx, "SELECT value FROM app_settings WHERE key = 'backup_key'").Scan(&backupVal)
	if err != nil {
		t.Fatalf("failed to read from backup: %v", err)
	}
	if backupVal != "backup_val" {
		t.Errorf("expected 'backup_val', got '%s'", backupVal)
	}
}

func TestDatabaseConcurrentStress(t *testing.T) {
	tempDir, err := os.MkdirTemp("", "gth_stress_test_*")
	if err != nil {
		t.Fatalf("failed to create temp dir: %v", err)
	}
	defer os.RemoveAll(tempDir)

	dbPath := filepath.Join(tempDir, "stress.db")
	db, err := Open(dbPath)
	if err != nil {
		t.Fatalf("failed to open stress database: %v", err)
	}
	defer db.Close()

	ctx := context.Background()
	concurrency := 20
	iterations := 10

	var wg sync.WaitGroup
	errCh := make(chan error, concurrency*iterations)

	for i := 0; i < concurrency; i++ {
		wg.Add(1)
		workerID := i
		go func() {
			defer wg.Done()
			for j := 0; j < iterations; j++ {
				key := fmt.Sprintf("k_%d_%d", workerID, j)
				val := fmt.Sprintf("v_%d_%d", workerID, j)

				// 并发写入
				_, writeErr := db.ExecContext(ctx, "INSERT OR REPLACE INTO app_settings (key, value) VALUES (?, ?)", key, val)
				if writeErr != nil {
					errCh <- writeErr
					return
				}

				// 并发读取
				var readVal string
				readErr := db.QueryRowContext(ctx, "SELECT value FROM app_settings WHERE key = ?", key).Scan(&readVal)
				if readErr != nil {
					errCh <- readErr
					return
				}
				if readVal != val {
					errCh <- fmt.Errorf("data mismatch: expected %s, got %s", val, readVal)
					return
				}
			}
		}()
	}

	wg.Wait()
	close(errCh)

	for e := range errCh {
		t.Fatalf("concurrent stress test encountered error: %v", e)
	}

	// 最终完整性校验
	if err := db.IntegrityCheck(ctx); err != nil {
		t.Fatalf("post-stress integrity check failed: %v", err)
	}
}

func TestVerifySeededProductionDB(t *testing.T) {
	appData := os.Getenv("APPDATA")
	dbPath := filepath.Join(appData, "GameTrainerHub", "data", "trainers.db")
	if _, err := os.Stat(dbPath); os.IsNotExist(err) {
		t.Skip("production db not found")
	}

	db, err := Open(dbPath)
	if err != nil {
		t.Fatalf("Open production db failed: %v", err)
	}
	defer db.Close()

	var gameCount, trainerCount, ftsCount int
	_ = db.QueryRow("SELECT COUNT(id) FROM games").Scan(&gameCount)
	_ = db.QueryRow("SELECT COUNT(id) FROM trainers").Scan(&trainerCount)
	_ = db.QueryRow("SELECT COUNT(rowid) FROM game_search_fts").Scan(&ftsCount)

	t.Logf("=== PRODUCTION DB VERIFICATION ===")
	t.Logf("Total games in DB: %d", gameCount)
	t.Logf("Total trainers in DB: %d", trainerCount)
	t.Logf("Total FTS5 indexed: %d", ftsCount)

	if gameCount < 767 {
		t.Errorf("expected at least 767 games, got %d", gameCount)
	}
	if trainerCount < 767 {
		t.Errorf("expected at least 767 trainers, got %d", trainerCount)
	}
}
