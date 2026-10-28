package entity

import (
	"github.com/oklog/ulid/v2"
	"github.com/tacenva/tacenva-core/util/credential"
	"gorm.io/gorm"
)

type UserStatus string

const (
	UserStatusPending  UserStatus = "pending"
	UserStatusApproved UserStatus = "approved"
	UserStatusRevoked  UserStatus = "revoked"
)

var TokenSize = 32

type User struct {
	ID           string     `json:"id,omitempty" gorm:"type:varchar(26);primaryKey"`
	AccessUnitID string     `json:"access_unit_id" gorm:"type:varchar(26);not null;index"`
	Hostname     string     `json:"hostname,omitempty" gorm:"type:varchar(50);not null"`
	Token        string     `json:"token,omitempty" gorm:"type:varchar(128);not null"`
	Status       UserStatus `json:"status,omitempty" gorm:"type:varchar(20);not null"`

	AccessUnit AccessUnit `json:"access_unit,omitempty" gorm:"foreignKey:AccessUnitID;references:ID;constraint:OnDelete:CASCADE,OnUpdate:CASCADE"`
}

func (u *User) HasWriteAccess() bool {
	switch u.AccessUnit.Level {
	case LevelAdmin, LevelWrite:
		return true
	default:
		return false
	}
}

func (u *User) HasAdminAccess() bool {
	return u.AccessUnit.Level == LevelAdmin
}

func (u *User) IsApproved() bool {
	return u.Status == UserStatusApproved
}

func (u *User) BeforeCreate(tx *gorm.DB) error {
	if u.ID == "" {
		u.ID = ulid.Make().String()
	}

	token, err := credential.Generate(uint(TokenSize))
	if err != nil {
		return err
	}
	u.Token = token

	return nil
}
