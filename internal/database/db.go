package database

import (
	"context"
	"database/sql"
	"fmt"
	"os"
	"path/filepath"
	"sync"
	"time"

	"GameTrainerHub/internal/logger"

	_ "modernc.org/sqlite"
)

type DB struct {
	*sql.DB
	dsn string
	mu  sync.RWMutex
}

var (
	instance *DB
	once     sync.Once
)

// Open 打开 SQLite 数据库并配置连接池与 PRAGMA
func Open(dbPath string) (*DB, error) {
	dir := filepath.Dir(dbPath)
	if err := os.MkdirAll(dir, 0755); err != nil {
		return nil, fmt.Errorf("failed to create db directory: %w", err)
	}

	// SQLite 参数优化：WAL 模式、外键约束、Busy 超时
	dsn := fmt.Sprintf("%s?_pragma=journal_mode(WAL)&_pragma=synchronous(NORMAL)&_pragma=foreign_keys(1)&_pragma=busy_timeout(5000)", dbPath)
	
	sqlDB, err := sql.Open("sqlite", dsn)
	if err != nil {
		return nil, fmt.Errorf("failed to open sqlite database: %w", err)
	}

	// 设置连接池：由于是嵌入式 SQLite WAL 模式，读写并发适中
	sqlDB.SetMaxOpenConns(5)
	sqlDB.SetMaxIdleConns(2)
	sqlDB.SetConnMaxLifetime(1 * time.Hour)

	// 强制启用外键约束
	if _, err := sqlDB.Exec("PRAGMA foreign_keys = ON;"); err != nil {
		_ = sqlDB.Close()
		return nil, fmt.Errorf("failed to enable foreign_keys pragma: %w", err)
	}

	// 测试连接可用性
	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()
	if err := sqlDB.PingContext(ctx); err != nil {
		_ = sqlDB.Close()
		return nil, fmt.Errorf("failed to ping database: %w", err)
	}

	db := &DB{
		DB:  sqlDB,
		dsn: dsn,
	}

	// 执行自动迁移
	if err := db.Migrate(); err != nil {
		_ = sqlDB.Close()
		return nil, fmt.Errorf("database migration failed: %w", err)
	}

	logger.Info("Database initialized successfully", "path", dbPath)
	return db, nil
}

// InitGlobalDB 初始化全局单例数据库
func InitGlobalDB(dbPath string) (*DB, error) {
	var err error
	once.Do(func() {
		instance, err = Open(dbPath)
	})
	return instance, err
}

// GetDB 获取全局数据库实例
func GetDB() *DB {
	return instance
}

// WithTx 事务执行包装器
func (db *DB) WithTx(ctx context.Context, fn func(tx *sql.Tx) error) error {
	tx, err := db.BeginTx(ctx, nil)
	if err != nil {
		return fmt.Errorf("failed to begin transaction: %w", err)
	}

	defer func() {
		if p := recover(); p != nil {
			_ = tx.Rollback()
			panic(p)
		}
	}()

	if err := fn(tx); err != nil {
		if rbErr := tx.Rollback(); rbErr != nil {
			return fmt.Errorf("tx err: %v, rollback err: %w", err, rbErr)
		}
		return err
	}

	if err := tx.Commit(); err != nil {
		return fmt.Errorf("failed to commit transaction: %w", err)
	}
	return nil
}
