package transactionmanager

import (
	"fmt"

	"github.com/tacenva/database"
	"gorm.io/gorm"
)

func Transaction(
	gormDB *gorm.DB,
	tacenvaDbFile *database.DatabaseFile,
	fn func(
		gormTx *gorm.DB,
		fileTx *database.DatabaseFile,
	) error,
) error {
	if gormDB == nil {
		return fmt.Errorf("gorm database is nil")
	}

	if tacenvaDbFile == nil {
		return fmt.Errorf("tacenva database file is nil")
	}

	if fn == nil {
		return fmt.Errorf("transaction callback is nil")
	}

	// Begin GORM transaction.
	gormTx := gormDB.Begin()
	if gormTx.Error != nil {
		return gormTx.Error
	}
	defer gormTx.Rollback()

	// Begin Tacenva transaction.
	fileTx, err := tacenvaDbFile.Begin()
	if err != nil {
		return err
	}
	defer fileTx.Rollback()

	// Execute callback.
	if err := fn(gormTx, fileTx.DB()); err != nil {
		return err
	}

	// Commit GORM.
	if err := gormTx.Commit().Error; err != nil {
		return fmt.Errorf("gorm commit failed: %w", err)
	}

	// Commit Tacenva.
	if err := fileTx.Commit(); err != nil {
		return fmt.Errorf(
			"tacenva commit failed after gorm commit: %w",
			err,
		)
	}

	return nil
}
