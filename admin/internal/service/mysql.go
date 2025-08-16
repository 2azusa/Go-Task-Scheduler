package service

import (
	"fmt"
	"os"
	"pulse/common/models"
	"pulse/common/pkg/logger"

	"errors"

	"gorm.io/gorm"
)

// RegisterTables 函数用于在数据库中注册数据表，并初始化种子数据
func RegisterTables(db *gorm.DB) error {
	// 调用 AutoMigrate 方法，自动迁移表结构
	err := db.AutoMigrate(
		models.User{},
		models.Node{},
		models.Job{},
		models.JobLog{},
		models.Script{},
	)
	if err != nil {
		// 将错误记录日志
		logger.GetLogger().Error(fmt.Sprintf("register table failed, error: %s", err.Error()))
		os.Exit(0)
	}
	// 定义一个 User 类型的切片，用于初始化种子数据
	entities := []models.User{
		{UserName: "root", Password: "7szho", Role: models.RoleAdmin, Email: "123321@qq.com"},
	}
	// 调用 checkDataExist 函数检查数据是否存在，避免重复插入
	if exist := checkDataExist(db); !exist {
		// 如果数据不存在，则向用户表中插入初始数据
		if err := db.Table(models.PulseUserTableName).Create(&entities).Error; err != nil {
			return fmt.Errorf("failed to initialize table data: %w", err)
		}
	}
	logger.GetLogger().Info("register table success")
	return nil
}

// checkDataExist 函数用于检查数据库是否存入 root 用户
func checkDataExist(db *gorm.DB) bool {
	// 调用 `errors.Is` 函数来判断返回的错误是否就是 gorm.ErrRecordNotFound
	if errors.Is(db.Table(models.PulseUserTableName).Where("username = ?", "root").First(&models.User{}).Error, gorm.ErrRecordNotFound) { // 判断是否存在数据
		return false
	}
	return true
}
