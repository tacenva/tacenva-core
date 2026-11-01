package entity

import (
	"github.com/oklog/ulid/v2"
	"gorm.io/gorm"
)

type AccessLevel string

const (
	LevelAdmin AccessLevel = "admin"
	LevelWrite AccessLevel = "write"
	LevelRead  AccessLevel = "read"
)

type AccessUnit struct {
	ID          string      `json:"id" gorm:"type:varchar(26);primaryKey"`
	Name        string      `json:"name" gorm:"type:varchar(100);not null;uniqueIndex:idx_unit_name"`
	GroupID     string      `json:"group_id" gorm:"type:varchar(26);not null;uniqueIndex:idx_unit_name"`
	SotID       string      `json:"-" gorm:"type:varchar(26);not null;uniqueIndex:idx_unit_name"`
	Level       AccessLevel `json:"level" gorm:"type:varchar(10);not null"`
	PublicKey   string      `json:"public_key" gorm:"type:text;not null"`
	Revoked     bool        `json:"revoked" gorm:"not null;default:false"`
	AutoApprove bool        `json:"auto_approve" gorm:"not null;default:false"`

	Group AccessGroup `json:"group,omitempty" gorm:"foreignKey:GroupID,SotID;references:ID,SotID;constraint:OnUpdate:CASCADE,OnDelete:RESTRICT"`
	Users []User      `json:"users,omitempty" gorm:"foreignKey:AccessUnitID;references:ID;constraint:OnDelete:CASCADE,OnUpdate:CASCADE"`
}

func (l AccessLevel) IsValid() bool {
	switch l {
	case LevelAdmin, LevelWrite, LevelRead:
		return true
	default:
		return false
	}
}

func (l AccessLevel) AllowsWrite() bool {
	return l == LevelAdmin || l == LevelWrite
}

func (l AccessLevel) AllowsAdmin() bool {
	return l == LevelAdmin
}

func (au *AccessUnit) HasWriteAccess() bool {
	return au.Level.AllowsWrite()
}

func (au *AccessUnit) HasAdminAccess() bool {
	return au.Level.AllowsAdmin()
}

func (au *AccessUnit) IsAllowed() bool {
	return !au.Revoked
}

func (au *AccessUnit) BeforeCreate(tx *gorm.DB) error {
	if au.ID == "" {
		au.ID = ulid.Make().String()
	}
	return nil
}
