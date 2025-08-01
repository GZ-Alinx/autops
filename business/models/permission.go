package models

import (
	"time"

	"gorm.io/gorm"
)

// Permission 权限模型
type Permission struct {
	ID          uint           `gorm:"primarykey" json:"id"`
	Resource    string         `gorm:"size:255;not null" json:"resource"` // API资源路径
	Action      string         `gorm:"size:50;not null" json:"action"`    // HTTP方法
	Description string         `gorm:"size:255" json:"description"`       // 权限描述
	CreatedAt   time.Time      `json:"created_at"`
	UpdatedAt   time.Time      `json:"updated_at"`
	DeletedAt   gorm.DeletedAt `gorm:"index,softDelete:deleted_at" json:"deleted_at,omitempty"`
	Roles       []Role         `gorm:"many2many:role_permissions;foreignKey:ID;joinForeignKey:PermissionID;References:ID;joinReferences:RoleID" json:"roles,omitempty"`
}
