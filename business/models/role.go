package models

import (
	"time"

	"gorm.io/gorm"
)

// Role 角色模型
type Role struct {
	ID          uint      `gorm:"primarykey" json:"id"`
	Name        string    `gorm:"size:50;uniqueIndex;not null" json:"name"` // 角色名称，如admin, editor
	Description string    `gorm:"size:255" json:"description"`              // 角色描述
	CreatedAt   time.Time `json:"created_at"`
	UpdatedAt   time.Time `json:"updated_at"`
	// DeletedAt 软删除字段
	// @SerializedName deleted_at
	// @Nullable
	// DeletedAt 软删除字段
	// @Schema(type="string", format="date-time")
	DeletedAt   gorm.DeletedAt `gorm:"index,softDelete:deleted_at" json:"deleted_at,omitempty"`
	Users       []User         `gorm:"many2many:user_roles;foreignKey:ID;joinForeignKey:RoleID;References:ID;joinReferences:UserID" json:"users,omitempty"`                   // 多对多关联用户
	Permissions []Permission   `gorm:"many2many:role_permissions;foreignKey:ID;joinForeignKey:RoleID;References:ID;joinReferences:PermissionID" json:"permissions,omitempty"` // 多对多关联权限
}

// UserRole 用户角色关联表
type UserRole struct {
	UserID    uint           `gorm:"primarykey" json:"user_id"`
	RoleID    uint           `gorm:"primarykey" json:"role_id"`
	User      User           `gorm:"foreignKey:UserID" json:"-"`
	Role      Role           `gorm:"foreignKey:RoleID" json:"-"`
	CreatedAt time.Time      `json:"created_at"`
	UpdatedAt time.Time      `json:"updated_at"`
	DeletedAt gorm.DeletedAt `gorm:"index" json:"deleted_at,omitempty"`
}
