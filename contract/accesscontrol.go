package contract

import "github.com/tacenva/tacenva-core/entity"

type AccessGroupTree struct {
	Group           entity.AccessGroup              `json:"group"`
	AccessUnitCount int64                           `json:"access_unit_count"`
	UserCount       int64                           `json:"user_count"`
	Children        []AccessGroupHierarchyWithCount `json:"children"`
}

type AccessGroupHierarchyWithCount struct {
	entity.AccessGroupHierarchy
	AccessUnitCount int64 `json:"access_unit_count"`
	UserCount       int64 `json:"user_count"`
}

type PublicKeyOfGroupsReq struct {
	GroupIDs []string `json:"group_ids"`
}

type AccessUnitWithUserCount struct {
	entity.AccessUnit
	UserCount int64 `json:"user_count"`
}
