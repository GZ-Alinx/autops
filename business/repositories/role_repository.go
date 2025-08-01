package repositories

import (
	"github.com/GZ-Alinx/autops/business/models"
	"github.com/GZ-Alinx/autops/internal/database"
	"gorm.io/gorm"
)

// RoleRepository 角色仓库接口
type RoleRepository interface {
	// GetByNameIn 根据角色名称列表获取角色
	GetByNameIn(names []string) ([]models.Role, error)
	// Create 创建新角色
	Create(role *models.Role) error
	// GetAll 获取所有角色
	GetAll() ([]models.Role, error)
	// GetByID 根据ID获取角色
	GetByID(id uint) (*models.Role, error)
	// Update 更新角色信息
	Update(role *models.Role) error
	// Delete 删除角色
	Delete(id uint) error
}

// roleRepository 角色仓库GORM实现
type roleRepository struct {
	db *gorm.DB
}

// NewRoleRepository 创建角色仓库实例
func NewRoleRepository() RoleRepository {
	return &roleRepository{
		db: database.DB,
	}
}

// GetByNameIn 根据角色名称列表获取角色
func (r *roleRepository) GetByNameIn(names []string) ([]models.Role, error) {
	var roles []models.Role
	if err := r.db.Where("name IN ?", names).Find(&roles).Error; err != nil {
		return nil, err
	}
	return roles, nil
}

// Create 创建新角色
func (r *roleRepository) Create(role *models.Role) error {
	return r.db.Create(role).Error
}

// GetAll 获取所有角色
func (r *roleRepository) GetAll() ([]models.Role, error) {
	var roles []models.Role
	if err := r.db.Find(&roles).Error; err != nil {
		return nil, err
	}
	return roles, nil
}

// GetByID 根据ID获取角色
func (r *roleRepository) GetByID(id uint) (*models.Role, error) {
	var role models.Role
	if err := r.db.First(&role, id).Error; err != nil {
		return nil, err
	}
	return &role, nil
}

// Update 更新角色信息
func (r *roleRepository) Update(role *models.Role) error {
	return r.db.Save(role).Error
}

// Delete 删除角色
func (r *roleRepository) Delete(id uint) error {
	// 先删除用户角色关联
	if err := r.db.Table("user_roles").Where("role_id = ?", id).Delete(nil).Error; err != nil {
		return err
	}
	// 再删除角色
	return r.db.Delete(&models.Role{}, id).Error
}
