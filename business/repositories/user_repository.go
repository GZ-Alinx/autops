package repositories

import (
	"github.com/GZ-Alinx/autops/business/models"
	"github.com/GZ-Alinx/autops/internal/database"
)

// UserRepository 用户仓库接口
type UserRepository interface {
	GetByID(id uint) (*models.User, error)
	GetByUsername(username string) (*models.User, error)
	Create(user *models.User) error
	Update(user *models.User) (int64, error)
	Delete(id uint) error
	List(page, pageSize int) ([]*models.User, int64, error)
	AssignRole(userID, roleID uint) error
}

// userRepository GORM实现
type userRepository struct{}

// NewUserRepository 创建用户仓库实例
func NewUserRepository() UserRepository {
	return &userRepository{}
}

// GetByID 根据ID获取用户（包含角色）
func (r *userRepository) GetByID(id uint) (*models.User, error) {
	var user models.User
	result := database.DB.Preload("Roles").First(&user, id)
	return &user, result.Error
}

// GetByUsername 根据用户名获取用户（包含角色）
func (r *userRepository) GetByUsername(username string) (*models.User, error) {
	var user models.User
	result := database.DB.Where("username = ?", username).Preload("Roles").First(&user)
	return &user, result.Error
}

// Create 创建用户
func (r *userRepository) Create(user *models.User) error {
	return database.DB.Create(user).Error
}

// Update 更新用户
func (r *userRepository) Update(user *models.User) (int64, error) {
	result := database.DB.Save(user)
	return result.RowsAffected, result.Error
}

// Delete 删除用户
func (r *userRepository) Delete(id uint) error {
	return database.DB.Delete(&models.User{}, id).Error
}

// List 分页获取用户列表（包含角色）
func (r *userRepository) List(page, pageSize int) ([]*models.User, int64, error) {
	var users []*models.User
	var total int64

	// 获取总数
	database.DB.Model(&models.User{}).Count(&total)

	// 分页查询并预加载角色
	offset := (page - 1) * pageSize
	result := database.DB.Preload("Roles").Offset(offset).Limit(pageSize).Find(&users)

	return users, total, result.Error
}

// AssignRole 为用户分配角色
func (r *userRepository) AssignRole(userID, roleID uint) error {
	// 创建用户角色关联记录
	result := database.DB.Exec("INSERT INTO user_roles (user_id, role_id) VALUES (?, ?)", userID, roleID)
	return result.Error
}
