package contract

import "github.com/tacenva/tacenva-core/entity"

type AccessGroupTree struct {
	Group            entity.AccessGroup              `json:"group"`
	AccessUnitCount  int64                           `json:"access_unit_count"`
	UserCount        int64                           `json:"user_count"`
	PendingUserCount int64                           `json:"pending_user_count"`
	Children         []AccessGroupHierarchyWithCount `json:"children"`
}

type AccessGroupHierarchyWithCount struct {
	entity.AccessGroupHierarchy
	AccessUnitCount  int64 `json:"access_unit_count"`
	UserCount        int64 `json:"user_count"`
	PendingUserCount int64 `json:"pending_user_count"`
}

type PublicKeyOfGroupsReq struct {
	GroupIDs []string `json:"group_ids"`
}

type AccessUnitWithUserCount struct {
	entity.AccessUnit
	UserCount     int64 `json:"user_count" gorm:"-"`
	PendingCount  int64 `json:"pending_count" gorm:"-"`
	ApprovedCount int64 `json:"approved_count" gorm:"-"`
	RevokedCount  int64 `json:"revoked_count" gorm:"-"`
}
