package user

import (
	"github.com/tacenva/tacenva-core/genericerror"
	"gorm.io/gorm"
)

func (s *Service) Destroy(tx *gorm.DB, sotID string) error {
	if sotID == "" {
		return genericerror.ErrRequired("sot id")
	}

	return s.repository.DestroyBySotID(tx, sotID)
}

func (r *repository) DestroyBySotID(tx *gorm.DB, sotID string) error {
	return tx.Exec(`
		DELETE FROM users
		WHERE access_unit_id IN (
			SELECT id FROM access_units WHERE sot_id = ?
		)
	`, sotID).Error
}
