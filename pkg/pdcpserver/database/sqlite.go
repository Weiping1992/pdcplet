package database

import (
	"fmt"
	"pdcplet/pkg/pdcpserver/model"

	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
)

var DB *gorm.DB

func InitSQLite(configPath string) error {
	db, err := gorm.Open(sqlite.Open(configPath), &gorm.Config{})
	if err != nil {
		return fmt.Errorf("数据库连接失败: %v", err)
	}

	// 自动迁移表结构
	if err := db.AutoMigrate(&model.VirtualMachineRecord{}); err != nil {
		return fmt.Errorf("表迁移失败: %v", err)
	}

	DB = db
	return nil
}
