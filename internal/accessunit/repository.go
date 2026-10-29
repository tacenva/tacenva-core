package accessunit

import (
	"github.com/tacenva/tacenva-core/contract"
	"github.com/tacenva/tacenva-core/entity"
	"gorm.io/gorm"
)

type repository struct {
	db *gorm.DB
}

func NewRepository(db *gorm.DB) *repository {
	return &repository{
		db: db,
	}
}

func (r *repository) CreateUnit(
	tx *gorm.DB,
	accessUnit *entity.AccessUnit,
) error {
	return tx.Create(accessUnit).Error
}

func (r *repository) CreateGroupKey(
	tx *gorm.DB,
	accessGroupKey *entity.AccessGroupKey,
) error {
	return tx.Create(accessGroupKey).Error
}

func (r *repository) UpdateUnit(
	tx *gorm.DB,
	accessUnit *entity.AccessUnit,
) error {
	return tx.Model(&entity.AccessUnit{}).
		Where("id = ? AND sot_id = ?", accessUnit.ID, accessUnit.SotID).
		Select("Name", "Level", "Revoked", "AutoApprove").
		Updates(accessUnit).Error
}

func (r *repository) DeleteUnit(
	tx *gorm.DB,
	id string,
	sotID string,
) error {
	result := tx.
		Where("id = ? AND sot_id = ?", id, sotID).
		Delete(&entity.AccessUnit{})

	if result.Error != nil {
		return result.Error
	}

	if result.RowsAffected == 0 {
		return gorm.ErrRecordNotFound
	}

	return nil
}

func (r *repository) GetUnitByID(
	tx *gorm.DB,
	id string,
	sotID string,
) (*entity.AccessUnit, error) {
	var accessUnit entity.AccessUnit

	err := tx.
		Where("id = ? AND sot_id = ?", id, sotID).
		First(&accessUnit).
		Error
	if err != nil {
		return nil, err
	}

	return &accessUnit, nil
}

func (r *repository) GetUnitByPublicKey(
	sotID string,
	publicKey string,
) (*entity.AccessUnit, error) {
	var accessUnit entity.AccessUnit

	err := r.db.
		Where("public_key = ? AND sot_id = ?", publicKey, sotID).
		First(&accessUnit).
		Error
	if err != nil {
		return nil, err
	}

	return &accessUnit, nil
}

func (r *repository) GetUnitsByGroup(
	groupID string,
	sotID string,
) ([]contract.AccessUnitWithUserCount, error) {
	var accessUnits []entity.AccessUnit

	err := r.db.
		Preload("Users").
		Where(
			"group_id = ? AND sot_id = ?",
			groupID,
			sotID,
		).
		Find(&accessUnits).Error
	if err != nil {
		return nil, err
	}

	units := make(
		[]contract.AccessUnitWithUserCount,
		0,
		len(accessUnits),
	)

	for _, au := range accessUnits {
		unit := contract.AccessUnitWithUserCount{
			AccessUnit: au,
			UserCount:  int64(len(au.Users)),
		}

		for _, user := range au.Users {
			switch user.Status {
			case entity.UserStatusPending:
				unit.PendingCount++
			case entity.UserStatusApproved:
				unit.ApprovedCount++
			case entity.UserStatusRevoked:
				unit.RevokedCount++
			}
		}

		units = append(units, unit)
	}

	return units, nil
}

func (r *repository) FindAll(
	sotID string,
) ([]entity.AccessUnit, error) {
	var result []entity.AccessUnit

	err := r.db.
		Where("sot_id = ?", sotID).
		Find(&result).Error
	if err != nil {
		return nil, err
	}

	return result, nil
}

func (r *repository) GetUsersByUnit(
	unitID string,
	sotID string,
) ([]entity.User, error) {
	var users []entity.User

	err := r.db.
		Where(
			"access_unit_id = ? AND access_unit_id IN (?)",
			unitID,
			r.db.Model(&entity.AccessUnit{}).
				Select("id").
				Where("sot_id = ?", sotID),
		).
		Find(&users).
		Error
	if err != nil {
		return nil, err
	}

	return users, nil
}

func (r *repository) SetApprovalUser(
	userID string,
	isRevoked bool,
) error {
	status := entity.UserStatusApproved
	if isRevoked {
		status = entity.UserStatusRevoked
	}

	result := r.db.
		Model(&entity.User{}).
		Where("id = ?", userID).
		Update("status", status)

	if result.Error != nil {
		return result.Error
	}

	if result.RowsAffected == 0 {
		return gorm.ErrRecordNotFound
	}

	return nil
}

func (r *repository) DeleteUser(
	userID string,
) error {
	result := r.db.
		Where("id = ?", userID).
		Delete(&entity.User{})

	if result.Error != nil {
		return result.Error
	}

	if result.RowsAffected == 0 {
		return gorm.ErrRecordNotFound
	}

	return nil
}
