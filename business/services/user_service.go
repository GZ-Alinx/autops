package services

import (
	"errors"

	"golang.org/x/crypto/bcrypt"

	"github.com/GZ-Alinx/autops/business/models"
	"github.com/GZ-Alinx/autops/business/repositories"
	"github.com/GZ-Alinx/autops/internal/logger"
	"go.uber.org/zap"
)

// UserService 用户服务接口
type UserService interface {
	CreateUser(username, password, email string, phone *string) (*models.User, error)
	GetUserByID(id uint) (*models.User, error)
	GetUserByUsername(username string) (*models.User, error)
	UpdateUser(user *models.User) error
	DeleteUser(id uint) error
	ListUsers(page, pageSize int) ([]*models.User, int64, error)
	VerifyPassword(user *models.User, password string) bool
	UpdatePassword(user *models.User, newPassword string) error
}

// userService 服务实现
type userService struct {
	repo repositories.UserRepository
}

// NewUserService 创建用户服务实例
func NewUserService(repo repositories.UserRepository) UserService {
	return &userService{
		repo: repo,
	}
}

// CreateUser 创建用户并加密密码，同时分配默认角色
func (s *userService) CreateUser(username, password, email string, phone *string) (*models.User, error) {
	// 检查用户是否已存在
	existingUser, _ := s.repo.GetByUsername(username)
	if existingUser.ID > 0 {
		return nil, errors.New("用户名已存在")
	}

	// 密码加密
	hashedPassword, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
	if err != nil {
		return nil, err
	}

	// 创建用户
	user := &models.User{
		Username: username,
		Password: string(hashedPassword),
		Email:    email,
		Phone:    phone,
	}

	if err := s.repo.Create(user); err != nil {
		return nil, err
	}

	// 分配默认角色 "user"
	roleRepo := repositories.NewRoleRepository()
	roles, err := roleRepo.GetByNameIn([]string{"user"})
	if err != nil {
		logger.Logger.Error("获取默认角色失败", zap.Error(err))
		// 即使角色分配失败，也不影响用户创建
	} else if len(roles) > 0 {
		// 关联用户和角色
		if err := s.repo.AssignRole(user.ID, roles[0].ID); err != nil {
			logger.Logger.Error("分配默认角色失败", zap.Error(err))
		} else {
			logger.Logger.Info("成功分配默认角色给新用户", zap.Uint("userID", user.ID), zap.Uint("roleID", roles[0].ID))
		}
	}

	return user, nil
}

// GetUserByID 根据ID获取用户
func (s *userService) GetUserByID(id uint) (*models.User, error) {
	return s.repo.GetByID(id)
}

// GetUserByUsername 根据用户名获取用户
func (s *userService) GetUserByUsername(username string) (*models.User, error) {
	return s.repo.GetByUsername(username)
}

// UpdateUser 更新用户
func (s *userService) UpdateUser(user *models.User) error {
	_, err := s.repo.Update(user)
	return err
}

// DeleteUser 删除用户
func (s *userService) DeleteUser(id uint) error {
	return s.repo.Delete(id)
}

// ListUsers 分页获取用户列表
func (s *userService) ListUsers(page, pageSize int) ([]*models.User, int64, error) {
	return s.repo.List(page, pageSize)
}

// VerifyPassword 验证密码
func (s *userService) VerifyPassword(user *models.User, password string) bool {
	err := bcrypt.CompareHashAndPassword([]byte(user.Password), []byte(password))
	return err == nil
}

// UpdatePassword 更新用户密码
func (s *userService) UpdatePassword(user *models.User, newPassword string) error {
	// 加密新密码
	hashedPassword, err := bcrypt.GenerateFromPassword([]byte(newPassword), bcrypt.DefaultCost)
	if err != nil {
		return err
	}

	// 更新用户密码
	user.Password = string(hashedPassword)
	_, err = s.repo.Update(user)
	return err
}

// InitAdminUser 初始化管理员用户
func InitAdminUser() error {
	logger.Logger.Info("开始初始化管理员用户")

	// 创建用户服务实例
	userRepo := repositories.NewUserRepository()
	userService := NewUserService(userRepo)

	// 创建admin用户，密码123456
	phone := ""
	user, err := userService.CreateUser("admin", "123456", "admin@example.com", &phone)
	if err != nil {
		logger.Logger.Error("创建管理员用户失败", zap.Error(err))
		return err
	}

	// 为管理员用户分配admin角色
	roleRepo := repositories.NewRoleRepository()
	roles, err := roleRepo.GetByNameIn([]string{"admin"})
	if err != nil {
		logger.Logger.Error("获取admin角色失败", zap.Error(err))
	} else if len(roles) > 0 {
		if err := userRepo.AssignRole(user.ID, roles[0].ID); err != nil {
			logger.Logger.Error("为管理员用户分配角色失败", zap.Error(err))
		} else {
			logger.Logger.Info("成功为管理员用户分配角色", zap.Uint("userID", user.ID), zap.Uint("roleID", roles[0].ID))
		}
	}

	logger.Logger.Info("管理员用户创建成功", zap.Uint("userID", user.ID))
	return nil
}
