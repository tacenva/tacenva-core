package accessgroup

import (
	"errors"

	"github.com/tacenva/tacenva-core/entity"
	"gorm.io/gorm"
)

func (s *Service) Destroy(tx *gorm.DB, sotID string) error {
	if sotID == "" {
		return errors.New("sot ID is required")
	}

	return s.repository.DestroyBySotID(tx, sotID)
}

func (r *repository) DestroyBySotID(tx *gorm.DB, sotID string) error {
	return tx.
		Where("sot_id = ?", sotID).
		Delete(&entity.AccessGroup{}).Error
}
