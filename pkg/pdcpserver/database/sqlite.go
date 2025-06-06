package database

import (
	"fmt"
	"pdcplet/pkg/pdcpserver/model"

	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
)

var DB *gorm.DB

func InitSQLite(configPath string) error {
	db, err := gorm.Open(sqlite.Open(configPath+"?_txlock=immediate"), &gorm.Config{})
	if err != nil {
		return fmt.Errorf("数据库连接失败: %v", err)
	}

	db.Exec("PRAGMA journal_mode=WAL;")        // 启用WAL
	db.Exec("PRAGMA synchronous=NORMAL;")      // 平衡性能与数据安全
	db.Exec("PRAGMA busy_timeout=5000;")       // 设置5秒锁等待超时
	db.Exec("PRAGMA cache_size = 1000000000;") // 增加SQLite缓存
	db.Exec("PRAGMA foreign_keys = true;")     // 执行外键
	db.Exec("PRAGMA busy_timeout = 5000;")     // 设置一个更大的busy_timeout有助于防止SQLITE_BUSY错误

	// 自动迁移表结构
	if err := db.AutoMigrate(&model.VirtualMachineRecord{}); err != nil {
		return fmt.Errorf("表迁移失败: %v", err)
	}

	DB = db
	return nil
}
