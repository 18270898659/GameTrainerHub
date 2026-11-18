package database

import (
	"context"
	"database/sql"
	"errors"
	"os"
	"path/filepath"
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

	// 正常提交事务
	err = db.WithTx(ctx, func(tx *sql.Tx) error {
		_, execErr := tx.ExecContext(ctx, `
			INSERT INTO app_settings (key, value) VALUES ('tx_success', 'val_committed')
		`)
		return execErr
	})

	if err != nil {
		t.Fatalf("transaction commit failed: %v", err)
	}

	var val string
	err = db.QueryRowContext(ctx, "SELECT value FROM app_settings WHERE key = 'tx_success'").Scan(&val)
	if err != nil {
		t.Fatalf("failed to query committed row: %v", err)
	}
	if val != "val_committed" {
		t.Errorf("expected 'val_committed', got '%s'", val)
	}
}

