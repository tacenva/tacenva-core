package accessunit

import (
	"errors"

	"github.com/tacenva/tacenva-core/entity"
	"gorm.io/gorm"
)

func (s *Service) Transaction(fn func(tx *gorm.DB) error) error {
	return s.repository.db.Transaction(fn)
}

func (s *Service) Destroy(tx *gorm.DB, sotID string) error {
	if sotID == "" {
		return errors.New("sot ID is required")
	}

	// 1. Delete access group keys
	if err := s.repository.DestroyKeyBySotID(tx, sotID); err != nil {
		return err
	}

	// 2. Delete access group hierarchy
	if err := s.repository.DestroyHierarchyBySotID(tx, sotID); err != nil {
		return err
	}

	// 3. Delete access units
	if err := s.repository.DestroyUnitBySotID(tx, sotID); err != nil {
		return err
	}

	return nil
}

func (r *repository) DestroyKeyBySotID(tx *gorm.DB, sotID string) error {
	return tx.
		Where("sot_id = ?", sotID).
		Delete(&entity.AccessGroupKey{}).Error
}

func (r *repository) DestroyHierarchyBySotID(tx *gorm.DB, sotID string) error {
	return tx.
		Where("sot_id = ?", sotID).
		Delete(&entity.AccessGroupHierarchy{}).Error
}

func (r *repository) DestroyUnitBySotID(tx *gorm.DB, sotID string) error {
	return tx.
		Where("sot_id = ?", sotID).
		Delete(&entity.AccessUnit{}).Error
}
