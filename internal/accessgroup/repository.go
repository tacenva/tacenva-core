package accessgroup

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

func (r *repository) Transaction(
	fn func(tx *gorm.DB) error,
) error {
	return r.db.Transaction(fn)
}

func (r *repository) CreateGroup(
	tx *gorm.DB,
	accessGroup *entity.AccessGroup,
) error {
	return tx.Create(accessGroup).Error
}

func (r *repository) CreateGroupHierarchy(
	tx *gorm.DB,
	accessGroupHierarchy *entity.AccessGroupHierarchy,
) error {
	return tx.Create(accessGroupHierarchy).Error
}

func (r *repository) UpdateGroup(
	tx *gorm.DB,
	accessGroup *entity.AccessGroup,
) error {
	return tx.Model(&entity.AccessGroup{}).
		Where("id = ? AND sot_id = ?", accessGroup.ID, accessGroup.SotID).
		Select("Name").
		Updates(accessGroup).Error
}

func (r *repository) DeleteGroup(
	tx *gorm.DB,
	id string,
	sotID string,
) error {
	result := tx.
		Where("id = ? AND sot_id = ?", id, sotID).
		Delete(&entity.AccessGroup{})

	if result.Error != nil {
		return result.Error
	}

	if result.RowsAffected == 0 {
		return gorm.ErrRecordNotFound
	}

	return nil
}

func (r *repository) GetGroupByID(
	tx *gorm.DB,
	id string,
	sotID string,
) (*entity.AccessGroup, error) {
	var accessGroup entity.AccessGroup

	err := tx.
		Where("id = ? AND sot_id = ?", id, sotID).
		First(&accessGroup).
		Error
	if err != nil {
		return nil, err
	}

	return &accessGroup, nil
}

type childRow struct {
	SotID           string `gorm:"column:sot_id"`
	ParentGroupID   string `gorm:"column:parent_group_id"`
	ChildGroupID    string `gorm:"column:child_group_id"`
	WrappedChildKey string `gorm:"column:wrapped_child_key"`
	ChildName       string `gorm:"column:child_name"`
	ChildPublicKey  string `gorm:"column:child_public_key"`
	AccessUnitCount int64  `gorm:"column:access_unit_count"`
	UserCount       int64  `gorm:"column:user_count"`
}

type groupCountRow struct {
	GroupID         string `gorm:"column:group_id"`
	AccessUnitCount int64  `gorm:"column:access_unit_count"`
	UserCount       int64  `gorm:"column:user_count"`
}

func (r *repository) GetGroupWithChildren(
	tx *gorm.DB,
	id string,
	sotID string,
) ([]contract.AccessGroupTree, error) {
	var groups []entity.AccessGroup

	// Query 1: ambil group yang diminta atau semua root group.
	query := tx.Model(&entity.AccessGroup{}).
		Where("sot_id = ?", sotID)

	if id != "" {
		query = query.Where("id = ?", id)
	} else {
		query = query.Where(`
			NOT EXISTS (
				SELECT 1
				FROM access_group_hierarchies h
				WHERE h.sot_id = access_groups.sot_id
				  AND h.child_group_id = access_groups.id
			)
		`)
	}

	if err := query.Find(&groups).Error; err != nil {
		return nil, err
	}

	if len(groups) == 0 {
		return []contract.AccessGroupTree{}, nil
	}

	groupIDs := make([]string, 0, len(groups))
	for _, group := range groups {
		groupIDs = append(groupIDs, group.ID)
	}

	// Siapkan response parent.
	result := make([]contract.AccessGroupTree, 0, len(groups))
	groupIndex := make(map[string]int, len(groups))

	for _, group := range groups {
		groupIndex[group.ID] = len(result)

		result = append(result, contract.AccessGroupTree{
			Group:    group,
			Children: []contract.AccessGroupHierarchyWithCount{},
		})
	}

	// Query 2a: hitung AccessUnit/user untuk semua parent sekaligus.
	// Diasumsikan satu AccessUnit merepresentasikan satu user.
	var parentCounts []groupCountRow

	err := tx.
		Table("access_units").
		Select(`
			group_id,
			COUNT(*) AS access_unit_count,
			COUNT(*) AS user_count
		`).
		Where("sot_id = ? AND group_id IN ?", sotID, groupIDs).
		Group("group_id").
		Scan(&parentCounts).Error
	if err != nil {
		return nil, err
	}

	for _, count := range parentCounts {
		if index, ok := groupIndex[count.GroupID]; ok {
			result[index].AccessUnitCount = count.AccessUnitCount
			result[index].UserCount = count.UserCount
		}
	}

	// Query 2b: ambil semua child beserta jumlah AccessUnit/user.
	var children []childRow

	err = tx.
		Table("access_group_hierarchies AS h").
		Select(`
			h.sot_id,
			h.parent_group_id,
			h.child_group_id,
			h.wrapped_child_key,
			g.name AS child_name,
			g.public_key AS child_public_key,
			COUNT(au.group_id) AS access_unit_count,
			COUNT(au.group_id) AS user_count
		`).
		Joins(`
			JOIN access_groups g
				ON g.id = h.child_group_id
				AND g.sot_id = h.sot_id
		`).
		Joins(`
			LEFT JOIN access_units au
				ON au.group_id = h.child_group_id
				AND au.sot_id = h.sot_id
		`).
		Where(
			"h.sot_id = ? AND h.parent_group_id IN ?",
			sotID,
			groupIDs,
		).
		Group(`
			h.sot_id,
			h.parent_group_id,
			h.child_group_id,
			h.wrapped_child_key,
			g.name,
			g.public_key
		`).
		Scan(&children).Error
	if err != nil {
		return nil, err
	}

	// Masukkan child ke parent yang sesuai.
	for _, child := range children {
		index, ok := groupIndex[child.ParentGroupID]
		if !ok {
			continue
		}

		hierarchy := entity.AccessGroupHierarchy{
			SotID:           child.SotID,
			ParentGroupID:   child.ParentGroupID,
			ChildGroupID:    child.ChildGroupID,
			WrappedChildKey: child.WrappedChildKey,
			Child: entity.AccessGroup{
				ID:        child.ChildGroupID,
				SotID:     child.SotID,
				Name:      child.ChildName,
				PublicKey: child.ChildPublicKey,
			},
		}

		result[index].Children = append(
			result[index].Children,
			contract.AccessGroupHierarchyWithCount{
				AccessGroupHierarchy: hierarchy,
				AccessUnitCount:      child.AccessUnitCount,
				UserCount:            child.UserCount,
			},
		)
	}

	return result, nil
}

func (r *repository) FindAll(
	tx *gorm.DB,
	sotID string,
) ([]entity.AccessGroup, error) {
	var groups []entity.AccessGroup

	err := tx.Model(&entity.AccessGroup{}).
		Select("id", "name").
		Where("sot_id = ?", sotID).
		Order("name ASC").
		Find(&groups).Error
	if err != nil {
		return nil, err
	}

	return groups, nil
}

func (r *repository) FindKeysByIDs(
	tx *gorm.DB,
	sotID string,
	groupIDs []string,
) ([]entity.AccessGroup, error) {
	var groups []entity.AccessGroup

	if len(groupIDs) == 0 {
		return groups, nil
	}

	err := tx.Model(&entity.AccessGroup{}).
		Select("id", "public_key").
		Where("sot_id = ? AND id IN ?", sotID, groupIDs).
		Find(&groups).Error
	if err != nil {
		return nil, err
	}

	return groups, nil
}

func (r *repository) CreateGroupsBulk(
	tx *gorm.DB,
	sotID string,
	groups []entity.AccessGroup,
) error {
	if len(groups) == 0 {
		return nil
	}

	for i := range groups {
		groups[i].SotID = sotID
		groups[i].AccessUnits = nil
	}

	return tx.CreateInBatches(&groups, 100).Error
}

func (r *repository) UpdateGroupsBulk(
	tx *gorm.DB,
	sotID string,
	groups []entity.AccessGroup,
) error {
	for _, group := range groups {
		err := tx.Model(&entity.AccessGroup{}).
			Where("id = ? AND sot_id = ?", group.ID, sotID).
			Updates(map[string]interface{}{
				"name": group.Name,
				// "public_key": group.PublicKey,
			}).Error
		if err != nil {
			return err
		}
	}

	return nil
}

func (r *repository) DeleteGroupsBulk(
	tx *gorm.DB,
	sotID string,
	groupIDs []string,
) error {
	if len(groupIDs) == 0 {
		return nil
	}

	return tx.
		Where("sot_id = ? AND id IN ?", sotID, groupIDs).
		Delete(&entity.AccessGroup{}).Error
}
