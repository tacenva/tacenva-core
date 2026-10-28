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
	var units []contract.AccessUnitWithUserCount

	err := r.db.
		Table("access_units AS au").
		Select(`
			au.*,
			COUNT(u.access_unit_id) AS user_count
		`).
		Joins(`
			LEFT JOIN users u
				ON u.access_unit_id = au.id
		`).
		Where(
			"au.group_id = ? AND au.sot_id = ?",
			groupID,
			sotID,
		).
		Group("au.id").
		Find(&units).
		Error
	if err != nil {
		return nil, err
	}

	return units, nil
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
