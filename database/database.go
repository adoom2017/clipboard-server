package database

import (
	"fmt"
	"os"
	"path/filepath"

	"clipboard-server/models"

	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
	"gorm.io/gorm/logger"
)

var DB *gorm.DB

func Initialize() error {
	var err error

	dbPath := os.Getenv("DB_PATH")
	if dbPath == "" {
		dbPath = "data/clipboard.db"
	}

	// 从数据库路径中提取目录路径并创建目录
	dbDir := filepath.Dir(dbPath)
	if err := os.MkdirAll(dbDir, 0755); err != nil {
		return fmt.Errorf("failed to create database directory %s: %v", dbDir, err)
	}

	fmt.Printf("Created/verified database directory: %s\n", dbDir)

	logLevel := logger.Silent
	if os.Getenv("DB_DEBUG") == "true" {
		logLevel = logger.Info
	}

	DB, err = gorm.Open(sqlite.Open(dbPath), &gorm.Config{
		Logger: logger.Default.LogMode(logLevel),
	})
	if err != nil {
		return fmt.Errorf("failed to connect to database: %v", err)
	}

	DB.Exec("PRAGMA journal_mode = WAL;")
	DB.Exec("PRAGMA synchronous = NORMAL;")
	DB.Exec("PRAGMA cache_size = 1000;")
	DB.Exec("PRAGMA foreign_keys = ON;")
	DB.Exec("PRAGMA temp_store = memory;")

	if err := autoMigrate(); err != nil {
		return fmt.Errorf("failed to migrate database: %v", err)
	}

	fmt.Printf("Database initialized successfully at: %s\n", dbPath)
	return nil
}

func autoMigrate() error {
	return DB.AutoMigrate(
		&models.User{},
		&models.ClipboardItem{},
	)
}

func GetDB() *gorm.DB {
	return DB
}

func Close() error {
	if DB != nil {
		sqlDB, err := DB.DB()
		if err != nil {
			return err
		}
		return sqlDB.Close()
	}
	return nil
}

func HealthCheck() error {
	if DB == nil {
		return fmt.Errorf("database connection is nil")
	}

	sqlDB, err := DB.DB()
	if err != nil {
		return fmt.Errorf("failed to get underlying sql.DB: %v", err)
	}

	if err := sqlDB.Ping(); err != nil {
		return fmt.Errorf("failed to ping database: %v", err)
	}

	return nil
}

func GetStats() map[string]interface{} {
	stats := make(map[string]interface{})

	if DB == nil {
		stats["status"] = "disconnected"
		return stats
	}

	sqlDB, err := DB.DB()
	if err != nil {
		stats["status"] = "error"
		stats["error"] = err.Error()
		return stats
	}

	dbStats := sqlDB.Stats()
	stats["status"] = "connected"
	stats["open_connections"] = dbStats.OpenConnections
	stats["in_use"] = dbStats.InUse
	stats["idle"] = dbStats.Idle

	var userCount, clipboardCount int64
	DB.Model(&models.User{}).Count(&userCount)
	DB.Model(&models.ClipboardItem{}).Count(&clipboardCount)

	stats["user_count"] = userCount
	stats["clipboard_item_count"] = clipboardCount

	return stats
}

func CreateIndexes() error {
	indexes := []string{
		"CREATE INDEX IF NOT EXISTS idx_clipboard_items_user_timestamp ON clipboard_items(user_id, timestamp DESC);",
		"CREATE INDEX IF NOT EXISTS idx_clipboard_items_type ON clipboard_items(type);",
		"CREATE INDEX IF NOT EXISTS idx_clipboard_items_content ON clipboard_items(content);",
		"CREATE INDEX IF NOT EXISTS idx_users_username ON users(username);",
		"CREATE INDEX IF NOT EXISTS idx_users_email ON users(email);",
		"CREATE INDEX IF NOT EXISTS idx_users_active ON users(is_active);",
	}

	for _, index := range indexes {
		if err := DB.Exec(index).Error; err != nil {
			return fmt.Errorf("failed to create index: %v", err)
		}
	}

	fmt.Println("Database indexes created successfully")
	return nil
}

func Cleanup(daysOld int) error {
	if daysOld <= 0 {
		return fmt.Errorf("daysOld must be greater than 0")
	}

	result := DB.Where("created_at < datetime('now', '-' || ? || ' days')",
		daysOld).Delete(&models.ClipboardItem{})

	if result.Error != nil {
		return fmt.Errorf("failed to cleanup old clipboard items: %v", result.Error)
	}

	fmt.Printf("Cleaned up %d old clipboard items\n", result.RowsAffected)
	return nil
}

func Vacuum() error {
	if err := DB.Exec("VACUUM").Error; err != nil {
		return fmt.Errorf("failed to vacuum database: %v", err)
	}

	fmt.Println("Database vacuum completed")
	return nil
}
