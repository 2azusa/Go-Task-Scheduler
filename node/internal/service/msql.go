package service

import (
	"crony/common/models"

	"gorm.io/gorm"
)

func RegisterTables(db *gorm.DB) {
	_ = db.AutoMigrate(
		models.User{},
		models.Node{},
		models.Job{},
		models.JobLog{},
	)
}
