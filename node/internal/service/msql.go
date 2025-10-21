package service

import (
	"pulse/common/models"

	"gorm.io/gorm"
)

// RegisterTables 执行数据库表的自动迁移
func RegisterTables(db *gorm.DB) {
	_ = db.AutoMigrate(
		models.User{},
		models.Node{},
		models.Job{},
		models.JobLog{},
	)
}
