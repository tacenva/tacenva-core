package entity

import (
	"github.com/oklog/ulid/v2"
	"gorm.io/gorm"
)

type AccessGroup struct {
	ID        string `json:"id" gorm:"type:varchar(26);primaryKey"`
	SotID     string `json:"-" gorm:"type:varchar(26);primaryKey"`
	Name      string `json:"name,omitempty" gorm:"type:varchar(100);not null"`
	PublicKey string `json:"public_key" gorm:"type:text;not null"`

	AccessUnits []AccessUnit `json:"access_units,omitempty" gorm:"foreignKey:GroupID,SotID;references:ID,SotID"`
}

type AccessGroupHierarchy struct {
	SotID           string `json:"-" gorm:"type:varchar(26);primaryKey"`
	ParentGroupID   string `json:"parent_group_id" gorm:"type:varchar(26);primaryKey;index"`
	ChildGroupID    string `json:"child_group_id" gorm:"type:varchar(26);primaryKey;index"`
	WrappedChildKey string `json:"wrapped_child_key" gorm:"type:text;default:null"`

	Parent AccessGroup `json:"parent,omitempty" gorm:"foreignKey:ParentGroupID,SotID;references:ID,SotID;constraint:OnUpdate:CASCADE,OnDelete:CASCADE"`
	Child  AccessGroup `json:"child,omitempty" gorm:"foreignKey:ChildGroupID,SotID;references:ID,SotID;constraint:OnUpdate:CASCADE,OnDelete:CASCADE"`
}

func (g *AccessGroup) BeforeCreate(tx *gorm.DB) error {
	if g.ID == "" {
		g.ID = ulid.Make().String()
	}
	return nil
}

type AccessGroupKey struct {
	AccessGroupID string `json:"access_group_id" gorm:"type:varchar(26);primaryKey"`
	AccessUnitID  string `json:"access_unit_id" gorm:"type:varchar(26);primaryKey"`
	SotID         string `json:"-" gorm:"type:varchar(26);primaryKey"`
	WrappedKey    string `json:"wrapped_key" gorm:"type:text;not null"`
}
