package models

import (
	"time"

	"gorm.io/gorm"
)

// RolePermission 角色权限关联模型
type RolePermission struct {
	RoleID       uint           `gorm:"primarykey" json:"role_id"`
	PermissionID uint           `gorm:"primarykey" json:"permission_id"`
	CreatedAt    time.Time      `json:"created_at"`
	UpdatedAt    time.Time      `json:"updated_at"`
	DeletedAt    gorm.DeletedAt `gorm:"index" json:"deleted_at,omitempty"`
	Role         Role           `gorm:"foreignKey:RoleID" json:"role,omitempty"`
	Permission   Permission     `gorm:"foreignKey:PermissionID" json:"permission,omitempty"`
}
