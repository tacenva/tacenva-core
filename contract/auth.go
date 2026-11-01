package contract

import "github.com/tacenva/tacenva-core/entity"

type EnrollRequest struct {
	Hostname  string `json:"hostname"`
	PublicKey string `json:"public_key"`
}

type EnrollResponse struct {
	SotHostname     string             `json:"sot_hostname"`
	Status          entity.UserStatus  `json:"status"`
	AccessLevel     entity.AccessLevel `json:"access_level"`
	AuthToken       string             `json:"auth_token"`
	GroupAssignment *GroupAssignment   `json:"group_assignment,omitempty"`
}

type GroupAssignment struct {
	ID             string               `json:"group_id,omitempty"`
	LastSyncChange entity.SyncChange    `json:"last_sync_change,omitempty"`
	GroupKeyData   GroupKeyData         `json:"group_key_data,omitempty"`
	Items          []entity.AccessGroup `json:"items,omitempty"`
}

type GroupKeyData struct {
	WrappedKey string `gorm:"column:wrapped_key"`
	PublicKey  string `gorm:"column:public_key"`
}

type AuthDataResp struct {
	Status      entity.UserStatus  `json:"status"`
	AccessLevel entity.AccessLevel `json:"access_level"`

	GroupAssignment *GroupAssignment `json:"group_assignment,omitempty"`
}

type AuthDataReq struct {
	AccessGroupLastChangeID string `json:"last_change_id"`
}
