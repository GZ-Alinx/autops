package models

import (
	"time"

	"gorm.io/gorm"
)

// User 用户模型
type User struct {
	ID        uint           `gorm:"primarykey" json:"id"`
	Username  string         `gorm:"size:50;uniqueIndex;not null" json:"username"`
	Password  string         `gorm:"size:100;not null" json:"-"` // 密码不返回给前端
	Email     string         `gorm:"size:100;uniqueIndex" json:"email"`
	Phone     *string        `gorm:"size:20;uniqueIndex:idx_users_phone,uniqueWhere:phone IS NOT NULL" json:"phone,omitempty"`
	Nickname  string         `gorm:"size:50" json:"nickname"`
	Avatar    string         `gorm:"size:255" json:"avatar"`
	Status    int            `gorm:"default:1" json:"status"` // 1:正常, 0:禁用
	CreatedAt time.Time      `json:"created_at"`
	UpdatedAt time.Time      `json:"updated_at"`
	DeletedAt gorm.DeletedAt `gorm:"index" json:"-"`
	Roles     []Role         `gorm:"many2many:user_roles;foreignKey:ID;joinForeignKey:UserID;References:ID;joinReferences:RoleID" json:"roles,omitempty"` // 多对多关联角色
}
