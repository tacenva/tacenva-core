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
		&entity.AccessUnit{},
		&entity.AccessGroup{},
		&entity.AccessGroupHierarchy{},
		&entity.AccessGroupKey{},

		&entity.User{},
		&entity.SyncChange{},
		&entity.SyncState{},
	)
}
