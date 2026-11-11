package accessgroup

import (
	"github.com/tacenva/tacenva-core/entity"
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
	return tx.
		Where("sot_id = ?", sotID).
		Delete(&entity.AccessGroup{}).Error
}
