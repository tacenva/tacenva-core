package contract

import "github.com/tacenva/tacenva-core/entity"

type EnrollRequest struct {
	Hostname  string `json:"hostname"`
	PublicKey string `json:"public_key"`
}

type EnrollResponse struct {
	SoTHostname     string             `json:"sot_hostname"`
	Status          entity.UserStatus  `json:"status"`
	AccessLevel     entity.AccessLevel `json:"access_level"`
	AuthToken       string             `json:"auth_token"`
	GroupAssignment *GroupAssignment   `json:"group_assignment,omitempty"`
}

type GroupAssignment struct {
	ID    string               `json:"group_id,omitempty"`
	Items []entity.AccessGroup `json:"items,omitempty"`
}

type AuthData struct {
	Status      entity.UserStatus  `json:"status"`
	AccessLevel entity.AccessLevel `json:"access_level"`

	GroupAssignment *GroupAssignment `json:"group_assignment,omitempty"`
}
