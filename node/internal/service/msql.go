package service

import (
	"pulse/common/models"

	"gorm.io/gorm"
)

// RegisterTables 用于向数据库注册并自动迁移表结构。
func RegisterTables(db *gorm.DB) {
	_ = db.AutoMigrate(
		models.User{},   // 用户表
		models.Node{},   // 节点表
		models.Job{},    // 任务表
		models.JobLog{}, // 任务日志表
	)
}
