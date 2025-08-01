package controllers

import (
	"errors"
	"fmt"
	"strconv"
	"strings"

	"github.com/gin-gonic/gin"
	"go.uber.org/zap"
	"gorm.io/gorm"

	"github.com/GZ-Alinx/autops/business/models"
	"github.com/GZ-Alinx/autops/business/repositories"
	"github.com/GZ-Alinx/autops/internal/database"
	"github.com/GZ-Alinx/autops/internal/global"
	"github.com/GZ-Alinx/autops/internal/logger"
	"github.com/GZ-Alinx/autops/internal/response"
)

// PermissionController 权限管理控制器
type PermissionController struct{}

// NewPermissionController 创建权限控制器实例
func NewPermissionController() *PermissionController {
	return &PermissionController{}
}

// @Summary 添加权限策略
// @Description 为角色添加资源访问权限
// @Tags 权限管理
// @Accept application/json
// @Produce application/json
// @Param data body PolicyRequest true "权限策略信息"
// @Success 200 {object} response.Response{data=string}
// @Failure 400 {object} response.Response{msg=string}
// @Failure 500 {object} response.Response{msg=string}
// @Security ApiKeyAuth
// @Router /permissions/policy [post]
func (pc *PermissionController) AddPolicy(c *gin.Context) {
	var req PolicyRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		logger.Logger.Warn("添加权限策略失败: 请求参数验证失败", zap.Error(err))
		response.BadRequest(c, fmt.Errorf("请求参数验证失败: %v", err))
		return
	}

	// 添加权限策略
	validMethods := map[string]bool{"GET": true, "POST": true, "PUT": true, "DELETE": true, "PATCH": true}
	if req.Role == "" || req.Path == "" || req.Method == "" || !validMethods[req.Method] {
		logger.Logger.Warn("添加权限策略失败: 无效的请求参数", zap.String("role", req.Role), zap.String("path", req.Path), zap.String("method", req.Method))
		response.BadRequest(c, fmt.Errorf("无效的请求参数: 角色、路径不能为空且方法必须为GET/POST/PUT/DELETE/PATCH之一"))
		return
	}

	logger.Logger.Info("开始添加权限策略", zap.String("role", req.Role), zap.String("path", req.Path), zap.String("method", req.Method))

	if global.Enforcer == nil {
		logger.Logger.Error("添加权限策略失败: 权限管理器未初始化")
		response.InternalServerError(c, fmt.Errorf("权限管理器未初始化"))
		return
	}

	// 检查权限是否已存在
	var permission models.Permission
	result := database.DB.Where("resource = ? AND action = ?", req.Path, req.Method).First(&permission)
	if result.Error != nil {
		if !errors.Is(result.Error, gorm.ErrRecordNotFound) {
			logger.Logger.Error("查询权限失败", zap.String("path", req.Path), zap.String("method", req.Method), zap.Error(result.Error))
			response.InternalServerError(c, fmt.Errorf("查询权限失败: %v", result.Error))
			return
		}
		logger.Logger.Info("权限不存在，将创建新权限", zap.String("path", req.Path), zap.String("method", req.Method))
	} else {
		logger.Logger.Info("权限已存在", zap.Int("permissionID", int(permission.ID)), zap.String("path", req.Path), zap.String("method", req.Method))
	}

	// 如果权限不存在则创建
	if errors.Is(result.Error, gorm.ErrRecordNotFound) {
		permission = models.Permission{
			Resource:    req.Path,
			Action:      req.Method,
			Description: fmt.Sprintf("%s %s权限", req.Method, req.Path),
		}
		if err := database.DB.Create(&permission).Error; err != nil {
			logger.Logger.Error("创建权限失败", zap.String("path", req.Path), zap.String("method", req.Method), zap.Error(err))
			response.InternalServerError(c, fmt.Errorf("创建权限失败: %v", err))
			return
		}
		logger.Logger.Info("创建权限成功", zap.Int("permissionID", int(permission.ID)), zap.String("path", req.Path), zap.String("method", req.Method))
	}

	// 检查角色是否存在
	var role models.Role
	if err := database.DB.Where("name = ?", req.Role).First(&role).Error; err != nil {
		logger.Logger.Error("查询角色失败", zap.String("role", req.Role), zap.Error(err))
		response.InternalServerError(c, fmt.Errorf("查询角色失败: %v", err))
		return
	}
	logger.Logger.Info("查询角色成功", zap.Int("roleID", int(role.ID)), zap.String("roleName", role.Name))

	// 检查角色权限关联是否已存在
	var rolePermission models.RolePermission
	result = database.DB.Where("role_id = ? AND permission_id = ?", role.ID, permission.ID).First(&rolePermission)
	if result.Error != nil {
		if !errors.Is(result.Error, gorm.ErrRecordNotFound) {
			logger.Logger.Error("查询角色权限关联失败", zap.Int("roleID", int(role.ID)), zap.Int("permissionID", int(permission.ID)), zap.Error(result.Error))
			response.InternalServerError(c, fmt.Errorf("查询角色权限关联失败: %v", result.Error))
			return
		}
		logger.Logger.Info("角色权限关联不存在，将创建新关联", zap.Int("roleID", int(role.ID)), zap.Int("permissionID", int(permission.ID)))
	} else {
		logger.Logger.Warn("角色权限关联已存在", zap.Int("roleID", int(role.ID)), zap.Int("permissionID", int(permission.ID)))
	}

	if !errors.Is(result.Error, gorm.ErrRecordNotFound) {
		logger.Logger.Warn("权限策略已存在", zap.String("role", req.Role), zap.String("path", req.Path), zap.String("method", req.Method))
		response.BadRequest(c, fmt.Errorf("权限策略已存在"))
		return
	}

	// 创建角色权限关联
	rolePermission = models.RolePermission{
		RoleID:       role.ID,
		PermissionID: permission.ID,
	}
	if err := database.DB.Create(&rolePermission).Error; err != nil {
		logger.Logger.Error("创建角色权限关联失败", zap.Int("roleID", int(role.ID)), zap.Int("permissionID", int(permission.ID)), zap.Error(err))
		response.InternalServerError(c, fmt.Errorf("创建角色权限关联失败: %v", err))
		return
	}
	logger.Logger.Info("创建角色权限关联成功", zap.Int("roleID", int(role.ID)), zap.Int("permissionID", int(permission.ID)))

	// 同步到casbin_rule
	logger.Logger.Info("开始同步权限策略到casbin", zap.String("role", req.Role), zap.String("path", req.Path), zap.String("method", req.Method))
	ok, err := global.Enforcer.AddPolicy(req.Role, req.Path, req.Method)
	if err != nil {
		logger.Logger.Error("添加权限策略失败", zap.String("role", req.Role), zap.String("path", req.Path), zap.String("method", req.Method), zap.Error(err))
		response.InternalServerError(c, fmt.Errorf("添加权限策略失败: %v", err))
		return
	}
	if !ok {
		logger.Logger.Warn("权限策略已存在于casbin", zap.String("role", req.Role), zap.String("path", req.Path), zap.String("method", req.Method))
		response.BadRequest(c, fmt.Errorf("权限策略已存在"))
		return
	}
	logger.Logger.Info("添加权限策略到casbin成功", zap.String("role", req.Role), zap.String("path", req.Path), zap.String("method", req.Method))

	// 保存策略变更
	if err := global.Enforcer.SavePolicy(); err != nil {
		logger.Logger.Error("保存权限策略失败", zap.Error(err))
		response.InternalServerError(c, fmt.Errorf("保存权限策略失败: %v", err))
		return
	}
	logger.Logger.Info("保存权限策略成功")

	logger.Logger.Info("添加权限策略操作完成", zap.String("role", req.Role), zap.String("path", req.Path), zap.String("method", req.Method))
	response.OkWithData(c, "添加权限策略成功")
}

// @Summary 删除权限策略
// @Description 移除角色的资源访问权限
// @Tags 权限管理
// @Accept application/json
// @Produce application/json
// @Param data body PolicyRequest true "权限策略信息"
// @Success 200 {object} response.Response{data=string}
// @Failure 400 {object} response.Response{msg=string}
// @Failure 500 {object} response.Response{msg=string}
// @Security ApiKeyAuth
// @Router /permissions/policy [delete]
func (pc *PermissionController) RemovePolicy(c *gin.Context) {
	var req PolicyRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		logger.Logger.Warn("删除权限策略失败: 请求参数验证失败", zap.Error(err))
		response.BadRequest(c, err)
		return
	}

	// 删除权限策略
	if req.Role == "" || req.Path == "" || req.Method == "" {
		logger.Logger.Warn("删除权限策略失败: 角色、路径和方法不能为空", zap.String("role", req.Role), zap.String("path", req.Path), zap.String("method", req.Method))
		response.BadRequest(c, fmt.Errorf("角色、路径和方法不能为空"))
		return
	}

	logger.Logger.Info("开始删除权限策略", zap.String("role", req.Role), zap.String("path", req.Path), zap.String("method", req.Method))

	if global.Enforcer == nil {
		logger.Logger.Error("删除权限策略失败: 权限管理器未初始化")
		response.InternalServerError(c, fmt.Errorf("权限管理器未初始化"))
		return
	} // 查询权限
	var permission models.Permission
	if err := database.DB.Where("resource = ? AND action = ?", req.Path, req.Method).First(&permission).Error; err != nil {
		logger.Logger.Error("查询权限失败", zap.String("path", req.Path), zap.String("method", req.Method), zap.Error(err))
		response.InternalServerError(c, fmt.Errorf("查询权限失败: %v", err))
		return
	}
	logger.Logger.Info("查询权限成功", zap.Int("permissionID", int(permission.ID)), zap.String("path", req.Path), zap.String("method", req.Method))

	// 查询角色
	var role models.Role
	if err := database.DB.Where("name = ?", req.Role).First(&role).Error; err != nil {
		logger.Logger.Error("查询角色失败", zap.String("role", req.Role), zap.Error(err))
		response.InternalServerError(c, fmt.Errorf("查询角色失败: %v", err))
		return
	}
	logger.Logger.Info("查询角色成功", zap.Int("roleID", int(role.ID)), zap.String("roleName", role.Name))

	// 删除角色权限关联
	if err := database.DB.Where("role_id = ? AND permission_id = ?", role.ID, permission.ID).Delete(&models.RolePermission{}).Error; err != nil {
		logger.Logger.Error("删除角色权限关联失败", zap.Int("roleID", int(role.ID)), zap.Int("permissionID", int(permission.ID)), zap.Error(err))
		response.InternalServerError(c, fmt.Errorf("删除角色权限关联失败: %v", err))
		return
	}
	logger.Logger.Info("删除角色权限关联成功", zap.Int("roleID", int(role.ID)), zap.Int("permissionID", int(permission.ID)))

	// 如果权限无关联角色，删除权限
	var count int64
	database.DB.Model(&models.RolePermission{}).Where("permission_id = ?", permission.ID).Count(&count)
	logger.Logger.Info("检查权限关联角色数量", zap.Int64("count", count), zap.Int("permissionID", int(permission.ID)))
	if count == 0 {
		logger.Logger.Info("权限无关联角色，将删除权限", zap.Int("permissionID", int(permission.ID)))
		if err := database.DB.Delete(&permission).Error; err != nil {
			logger.Logger.Error("删除权限失败", zap.Int("permissionID", int(permission.ID)), zap.Error(err))
			response.InternalServerError(c, fmt.Errorf("删除权限失败: %v", err))
			return
		}
		logger.Logger.Info("删除权限成功", zap.Int("permissionID", int(permission.ID)))
	} else {
		logger.Logger.Info("权限仍有关联角色，不删除权限", zap.Int("permissionID", int(permission.ID)), zap.Int64("关联角色数", count))
	}

	// 从casbin_rule删除策略
	logger.Logger.Info("开始从casbin删除权限策略", zap.String("role", req.Role), zap.String("path", req.Path), zap.String("method", req.Method))
	ok, err := global.Enforcer.RemovePolicy(req.Role, req.Path, req.Method)
	if err != nil {
		logger.Logger.Error("删除权限策略失败", zap.String("role", req.Role), zap.String("path", req.Path), zap.String("method", req.Method), zap.Error(err))
		response.InternalServerError(c, fmt.Errorf("删除权限策略失败: %v", err))
		return
	}
	if !ok {
		logger.Logger.Warn("权限策略不存在于casbin", zap.String("role", req.Role), zap.String("path", req.Path), zap.String("method", req.Method))
		response.BadRequest(c, fmt.Errorf("权限策略不存在"))
		return
	}
	logger.Logger.Info("从casbin删除权限策略成功", zap.String("role", req.Role), zap.String("path", req.Path), zap.String("method", req.Method))

	// 保存策略变更
	if err := global.Enforcer.SavePolicy(); err != nil {
		logger.Logger.Error("保存权限策略失败", zap.Error(err))
		response.InternalServerError(c, fmt.Errorf("保存权限策略失败: %v", err))
		return
	}
	logger.Logger.Info("保存权限策略成功")

	logger.Logger.Info("删除权限策略操作完成", zap.String("role", req.Role), zap.String("path", req.Path), zap.String("method", req.Method))
	response.OkWithData(c, "删除权限策略成功")

	// 保存策略变更
	if err := global.Enforcer.SavePolicy(); err != nil {
		response.InternalServerError(c, fmt.Errorf("保存权限策略失败: %v", err))
		return
	}

	response.OkWithData(c, "删除权限策略成功")
}

// @Summary 获取所有权限策略
// @Description 获取系统中所有的RBAC权限策略
// @Tags 权限管理
// @Produce application/json
// @Success 200 {object} response.Response{data=[]string}
// @Failure 500 {object} response.Response{msg=string}
// @Security ApiKeyAuth
// @Router /permissions/policies [get]
func (pc *PermissionController) GetPolicies(c *gin.Context) {
	// 获取所有策略
	logger.Logger.Info("正在获取所有权限策略")
	policies, err := global.Enforcer.GetPolicy()
	if err != nil {
		logger.Logger.Error("获取权限策略失败: " + err.Error())
		response.InternalServerError(c, fmt.Errorf("获取权限策略失败: %v", err))
		return
	}
	// 转换策略为字符串格式
	var policyStrings []string
	for _, p := range policies {
		policyStrings = append(policyStrings, strings.Join(p, ","))
	}
	response.OkWithData(c, policyStrings)
}

// @Summary 创建角色
// @Description 创建新的角色
// @Tags 角色管理
// @Accept json
// @Produce json
// @Param data body CreateRoleRequest true "角色信息"
// @Security ApiKeyAuth
// @Success 200 {object} response.Response{data=models.Role}
// @Failure 400 {object} response.Response{msg=string}
// @Failure 500 {object} response.Response{msg=string}
// @Router /roles [post]
func (pc *PermissionController) CreateRole(c *gin.Context) {
	var req CreateRoleRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.BadRequest(c, fmt.Errorf("请求参数验证失败: %v", err))
		return
	}

	// 检查角色是否已存在
	roleRepo := repositories.NewRoleRepository()
	roles, err := roleRepo.GetByNameIn([]string{req.Name})
	if err != nil {
		response.InternalServerError(c, fmt.Errorf("查询角色失败: %v", err))
		return
	}
	if len(roles) > 0 {
		response.BadRequest(c, fmt.Errorf("角色已存在"))
		return
	}

	// 创建新角色
	newRole := &models.Role{
		Name:        req.Name,
		Description: req.Description,
	}
	if err := roleRepo.Create(newRole); err != nil {
		response.InternalServerError(c, fmt.Errorf("创建角色失败: %v", err))
		return
	}

	// 同步Casbin策略
	if err := database.SyncCasbinPolicy(); err != nil {
		logger.Logger.Error("同步Casbin策略失败", zap.Error(err))
	}

	response.OkWithData(c, newRole)
}

// @Summary 获取所有角色
// @Description 获取系统中所有角色
// @Tags 角色管理
// @Produce json
// @Security ApiKeyAuth
// @Success 200 {object} response.Response{data=[]models.Role}
// @Failure 500 {object} response.Response{msg=string}
// @Router /roles [get]
func (pc *PermissionController) GetAllRoles(c *gin.Context) {
	roleRepo := repositories.NewRoleRepository()
	roles, err := roleRepo.GetAll()
	if err != nil {
		response.InternalServerError(c, fmt.Errorf("获取角色列表失败: %v", err))
		return
	}
	response.OkWithData(c, roles)
}

// @Summary 获取角色详情
// @Description 根据ID获取角色详情
// @Tags 角色管理
// @Produce json
// @Param id path int true "角色ID"
// @Security ApiKeyAuth
// @Success 200 {object} response.Response{data=models.Role}
// @Failure 400 {object} response.Response{msg=string}
// @Failure 404 {object} response.Response{msg=string}
// @Failure 500 {object} response.Response{msg=string}
// @Router /roles/{id} [get]
func (pc *PermissionController) GetRoleByID(c *gin.Context) {
	idStr := c.Param("id")
	roleID, err := strconv.ParseUint(idStr, 10, 64)
	if err != nil {
		response.BadRequest(c, fmt.Errorf("无效的角色ID: %v", err))
		return
	}

	roleRepo := repositories.NewRoleRepository()
	// 需要实现GetByID方法
	role, err := roleRepo.GetByID(uint(roleID))
	if err != nil {
		response.InternalServerError(c, fmt.Errorf("查询角色失败: %v", err))
		return
	}
	if role.ID == 0 {
		response.NotFound(c, fmt.Errorf("角色不存在"))
		return
	}

	// 同步Casbin策略
	if err := database.SyncCasbinPolicy(); err != nil {
		logger.Logger.Error("同步Casbin策略失败", zap.Error(err))
	}

	response.OkWithData(c, role)
}

// @Summary 更新角色
// @Description 更新角色信息
// @Tags 角色管理
// @Accept json
// @Produce json
// @Param id path int true "角色ID"
// @Param data body UpdateRoleRequest true "角色信息"
// @Security ApiKeyAuth
// @Success 200 {object} response.Response{data=models.Role}
// @Failure 400 {object} response.Response{msg=string}
// @Failure 404 {object} response.Response{msg=string}
// @Failure 500 {object} response.Response{msg=string}
// @Router /roles/{id} [put]
func (pc *PermissionController) UpdateRole(c *gin.Context) {
	var req UpdateRoleRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.BadRequest(c, fmt.Errorf("请求参数验证失败: %v", err))
		return
	}

	roleID := req.ID
	roleRepo := repositories.NewRoleRepository()
	// 需要实现GetByID方法
	role, err := roleRepo.GetByID(uint(roleID))
	if err != nil {
		response.InternalServerError(c, fmt.Errorf("查询角色失败: %v", err))
		return
	}
	if role.ID == 0 {
		response.NotFound(c, fmt.Errorf("角色不存在"))
		return
	}

	// 检查名称是否已被其他角色使用
	if req.Name != role.Name {
		roles, err := roleRepo.GetByNameIn([]string{req.Name})
		if err != nil {
			response.InternalServerError(c, fmt.Errorf("查询角色失败: %v", err))
			return
		}
		if len(roles) > 0 {
			response.BadRequest(c, fmt.Errorf("角色名称已存在"))
			return
		}
	}

	// 更新角色信息
	role.Name = req.Name
	role.Description = req.Description
	// 需要实现Update方法
	if err := roleRepo.Update(role); err != nil {
		response.InternalServerError(c, fmt.Errorf("更新角色失败: %v", err))
		return
	}

	response.OkWithData(c, role)
}

// @Summary 删除角色
// @Description 根据ID删除角色
// @Tags 角色管理
// @Produce json
// @Param id path int true "角色ID"
// @Security ApiKeyAuth
// @Success 200 {object} response.Response{data=string}
// @Failure 400 {object} response.Response{msg=string}
// @Failure 404 {object} response.Response{msg=string}
// @Failure 500 {object} response.Response{msg=string}
// @Router /roles/{id} [delete]
func (pc *PermissionController) DeleteRole(c *gin.Context) {
	idStr := c.Param("id")
	roleID, err := strconv.ParseUint(idStr, 10, 64)
	if err != nil {
		response.BadRequest(c, fmt.Errorf("无效的角色ID: %v", err))
		return
	}

	roleRepo := repositories.NewRoleRepository()
	// 需要实现GetByID方法
	role, err := roleRepo.GetByID(uint(roleID))
	if err != nil {
		response.InternalServerError(c, fmt.Errorf("查询角色失败: %v", err))
		return
	}
	if role.ID == 0 {
		response.NotFound(c, fmt.Errorf("角色不存在"))
		return
	}

	// 需要实现Delete方法
	if err := roleRepo.Delete(uint(roleID)); err != nil {
		response.InternalServerError(c, fmt.Errorf("删除角色失败: %v", err))
		return
	}

	// 同步Casbin策略
	if err := database.SyncCasbinPolicy(); err != nil {
		logger.Logger.Error("同步Casbin策略失败", zap.Error(err))
	}

	response.OkWithData(c, "角色删除成功")
}

// @Summary 更新用户角色
// @Description 更新指定用户的角色列表（会替换现有角色）
// @Tags 权限管理
// @Accept json
// @Produce json
// @Param data body UpdateUserRoleRequest true "用户ID和角色列表"
// @Security ApiKeyAuth
// @Success 200 {object} response.Response{data=string}
// @Failure 400 {object} response.Response{msg=string}
// @Failure 404 {object} response.Response{msg=string}
// @Failure 500 {object} response.Response{msg=string}
// @Router /permissions/user-role [put]
func (pc *PermissionController) UpdateUserRole(c *gin.Context) {
	var req UpdateUserRoleRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.BadRequest(c, fmt.Errorf("请求参数验证失败: %v", err))
		return
	}

	// 查询用户
	userRepo := repositories.NewUserRepository()
	user, err := userRepo.GetByID(req.UserID)
	if err != nil {
		response.InternalServerError(c, fmt.Errorf("查询用户失败: %v", err))
		return
	}
	if user.ID == 0 {
		response.NotFound(c, fmt.Errorf("用户不存在"))
		return
	}

	// 查询角色是否存在
	roleRepo := repositories.NewRoleRepository()
	roles, err := roleRepo.GetByNameIn(req.Roles)
	if err != nil {
		response.InternalServerError(c, fmt.Errorf("查询角色失败: %v", err))
		return
	}
	if len(roles) != len(req.Roles) {
		response.BadRequest(c, fmt.Errorf("部分角色不存在"))
		return
	}
	// 更新用户角色关联
	if err := database.DB.Model(&user).Association("Roles").Replace(roles); err != nil {
		response.InternalServerError(c, fmt.Errorf("更新用户角色失败: %v", err))
		return
	}

	// 同步Casbin策略
	if err := database.SyncCasbinPolicy(); err != nil {
		logger.Logger.Error("同步Casbin策略失败", zap.Error(err))
	}

	response.OkWithData(c, fmt.Sprintf("成功为用户分配 %d 个角色", len(roles)))
}

// PolicyRequest 权限策略请求结构
// @Description 权限策略请求参数，包含角色名称、资源路径和HTTP方法
type PolicyRequest struct {
	Role   string `json:"role" binding:"required"`   // 角色名称
	Path   string `json:"path" binding:"required"`   // 资源路径
	Method string `json:"method" binding:"required"` // HTTP方法
}

// CreateRoleRequest 创建角色请求结构
// @Description 创建新角色的请求参数
// @Param Name body string true "角色名称"
// @Param Description body string false "角色描述"
type CreateRoleRequest struct {
	Name        string `json:"name" binding:"required,min=2,max=50"` // 角色名称
	Description string `json:"description" binding:"max=255"`        // 角色描述
}

// UpdateRoleRequest 更新角色请求结构
// @Description 更新角色信息的请求参数
// @Param ID path int true "角色ID"
// @Param Name body string true "角色名称"
// @Param Description body string false "角色描述"
type UpdateRoleRequest struct {
	ID          uint   `json:"id" binding:"required"`
	Name        string `json:"name" binding:"required,min=2,max=50"` // 角色名称
	Description string `json:"description" binding:"max=255"`        // 角色描述
}

// UpdateUserRoleRequest 更新用户角色请求结构
// @Description 更新用户角色的请求参数，支持多角色分配
// @Param UserID body int true "用户ID"
// @Param Roles body []string true "角色名称列表"
type UpdateUserRoleRequest struct {
	UserID uint     `json:"user_id" binding:"required"`
	Roles  []string `json:"roles" binding:"required,min=1"` // 至少分配一个角色
}
