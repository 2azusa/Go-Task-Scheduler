package service

import (
	"pulse/common/models"

	"gorm.io/gorm"
)

// RegisterTables performs auto-migration for database tables.
func RegisterTables(db *gorm.DB) {
	_ = db.AutoMigrate(
		models.User{},
		models.Node{},
		models.Job{},
		models.JobLog{},
	)
}
