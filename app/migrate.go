package app

import (
	"github.com/tacenva/tacenva-core/entity"
	"gorm.io/gorm"
)

func Migrate(db *gorm.DB) error {
	if err := db.Exec("PRAGMA foreign_keys = ON").Error; err != nil {
		return err
	}

	return db.AutoMigrate(
		// &entity.Permission{},
		// &entity.PermissionGroup{},
		// &entity.PermissionGroupHierarchy{},
		&entity.User{},
		&entity.SyncChange{},
		&entity.SyncState{},
	)
}
