package controllers

import (
	"net/http"
	"strconv"

	"errors"

	"github.com/GZ-Alinx/autops/business/models"
	"github.com/GZ-Alinx/autops/business/services"

	"github.com/GZ-Alinx/autops/internal/logger"
	"github.com/GZ-Alinx/autops/internal/middleware"
	"github.com/GZ-Alinx/autops/internal/response"
	"github.com/gin-gonic/gin"
	"go.uber.org/zap"
)

// UserController 用户控制器
// PasswordUpdateRequest 密码修改请求结构体
// @Description 用户密码修改请求参数
type PasswordUpdateRequest struct {
	OldPassword string `json:"old_password" binding:"required"`
	NewPassword string `json:"new_password" binding:"required,min=6"`
}

// LoginRequest 登录请求结构体
// @Description 用户登录请求参数
type LoginRequest struct {
	Username string `json:"username" binding:"required"`
	Password string `json:"password" binding:"required"`
}

// RegisterRequest 注册请求结构体
// @Description 用户注册请求参数
type RegisterRequest struct {
	Username string  `json:"username" binding:"required"`
	Password string  `json:"password" binding:"required,min=6"`
	Email    string  `json:"email" binding:"required,email"`
	Phone    *string `json:"phone"`
	Nickname string  `json:"nickname"`
}

// LoginResponse 登录响应结构体
// @Description 用户登录响应数据
type LoginResponse struct {
	Token    string      `json:"token"`
	UserInfo models.User `json:"user_info"`
}

// ListResponse 列表响应结构体
// @Description 分页列表响应数据
type ListResponse struct {
	Total int64         `json:"total"`
	Items []models.User `json:"items"`
}

// UpdateRequest 更新请求结构体
// @Description 用户信息更新请求参数
type UpdateRequest struct {
	Email    string  `json:"email" binding:"omitempty,email"`
	Phone    *string `json:"phone"`
	Nickname string  `json:"nickname"`
	Avatar   string  `json:"avatar"`
	Status   int     `json:"status" binding:"omitempty,oneof=0 1"`
}

type UserController struct {
	userService services.UserService
}

// NewUserController 创建用户控制器实例
func NewUserController(userService services.UserService) *UserController {
	return &UserController{
		userService: userService,
	}
}

// @Summary 用户登录
// @Description 用户登录获取JWT令牌
// @Tags 用户管理
// @Accept json
// @Produce json
// @Param login body LoginRequest true "登录信息"
// @Success 200 {object} response.Response{data=LoginResponse}
// @Failure 400 {object} response.Response
// @Failure 401 {object} response.Response
// @Failure 500 {object} response.Response
// @Router user/login [post]
// Login 用户登录
func (uc *UserController) Login(ctx *gin.Context) {
	var req struct {
		Username string `json:"username" binding:"required"`
		Password string `json:"password" binding:"required"`
	}

	logger.Logger.Info("开始用户登录操作")

	if err := ctx.ShouldBindJSON(&req); err != nil {
		logger.Logger.Warn("用户登录失败: 请求参数验证失败", zap.Error(err))
		response.Fail(ctx, http.StatusBadRequest, err)
		return
	}

	logger.Logger.Info("用户登录参数验证通过", zap.String("username", req.Username))

	user, err := uc.userService.GetUserByUsername(req.Username)
	if err != nil {
		logger.Logger.Error("获取用户失败", zap.Error(err))
		response.Fail(ctx, http.StatusUnauthorized, errors.New("用户名或密码错误"))
		return
	}

	if !uc.userService.VerifyPassword(user, req.Password) {
		response.Fail(ctx, http.StatusUnauthorized, errors.New("用户名或密码错误"))
		return
	}

	// 生成JWT令牌
	token, err := middleware.GenerateToken(strconv.Itoa(int(user.ID)), user.Username)
	if err != nil {
		logger.Logger.Error("生成令牌失败", zap.Error(err))
		response.Fail(ctx, http.StatusInternalServerError, errors.New("生成令牌失败"))
		return
	}

	logger.Logger.Info("用户登录成功", zap.Uint("userID", user.ID), zap.String("username", user.Username))

	response.Success(ctx, gin.H{
		"token": token,
		"user":  user,
	})
}

// @Summary 用户添加
// @Description 创建新用户账号
// @Tags 用户管理
// @Accept json
// @Produce json
// @Param user body RegisterRequest true "用户添加信息"
// @Success 200 {object} response.Response{data=models.User}
// @Failure 400 {object} response.Response
// @Failure 500 {object} response.Response
// @Router users/register [post]
// Register 用户添加
func (uc *UserController) Register(ctx *gin.Context) {
	var req struct {
		Username string  `json:"username" binding:"required"`
		Password string  `json:"password" binding:"required,min=6"`
		Email    string  `json:"email" binding:"required,email"`
		Phone    *string `json:"phone"`
	}

	logger.Logger.Info("开始用户注册操作")

	if err := ctx.ShouldBindJSON(&req); err != nil {
		logger.Logger.Warn("用户注册失败: 请求参数验证失败", zap.Error(err))
		response.Fail(ctx, http.StatusBadRequest, err)
		return
	}

	logger.Logger.Info("用户注册参数验证通过", zap.String("username", req.Username), zap.String("email", req.Email))

	// 处理空手机号，转换为空指针
	var phone *string
	if req.Phone != nil && *req.Phone == "" {
		phone = nil
	} else {
		phone = req.Phone
	}

	user, err := uc.userService.CreateUser(req.Username, req.Password, req.Email, phone)
	if err != nil {
		logger.Logger.Error("创建用户失败", zap.Error(err))
		response.Fail(ctx, http.StatusInternalServerError, err)
		return
	}

	logger.Logger.Info("用户注册成功", zap.Uint("userID", user.ID), zap.String("username", user.Username))

	logger.Logger.Info("更新用户成功", zap.Uint("userID", user.ID), zap.String("username", user.Username))

	response.Success(ctx, user)
}

// @Summary 获取用户信息
// @Description 根据用户ID获取用户详情
// @Tags 用户管理
// @Accept json
// @Produce json
// @Param id path int true "用户ID"
// @Success 200 {object} response.Response{data=models.User}
// @Failure 400 {object} response.Response
// @Failure 401 {object} response.Response
// @Failure 404 {object} response.Response
// @Failure 500 {object} response.Response
// @Security BearerAuth
// @Router /users/{id} [get]
// GetUser 获取用户信息
func (uc *UserController) GetUser(ctx *gin.Context) {
	logger.Logger.Info("开始获取用户信息操作")

	idStr := ctx.Param("id")
	id, err := strconv.ParseUint(idStr, 10, 64)
	if err != nil {
		logger.Logger.Warn("获取用户信息失败: 无效的用户ID", zap.String("idStr", idStr), zap.Error(err))
		response.Fail(ctx, http.StatusBadRequest, err)
		return
	}

	logger.Logger.Info("用户ID参数验证通过", zap.Uint64("id", id))

	user, err := uc.userService.GetUserByID(uint(id))
	if err != nil {
		logger.Logger.Error("获取用户失败", zap.Error(err))
		response.Fail(ctx, http.StatusInternalServerError, err)
		return
	}

	logger.Logger.Info("获取用户信息成功", zap.Uint("userID", user.ID), zap.String("username", user.Username))

	response.Success(ctx, user)
}

// @Summary 更新用户
// @Description 根据ID更新用户信息
// @Tags 用户管理
// @Accept json
// @Produce json
// @Param id path int true "用户ID"
// @Param user body UpdateRequest true "用户更新信息"
// @Success 200 {object} response.Response{data=models.User}
// @Failure 400 {object} response.Response
// @Failure 401 {object} response.Response
// @Failure 403 {object} response.Response
// @Failure 404 {object} response.Response
// @Failure 500 {object} response.Response
// @Security BearerAuth
// @Router /users/{id} [put]
// UpdateUser 更新用户
func (uc *UserController) UpdateUser(ctx *gin.Context) {
	logger.Logger.Info("开始更新用户操作")

	idStr := ctx.Param("id")
	id, err := strconv.ParseUint(idStr, 10, 64)
	if err != nil {
		logger.Logger.Warn("更新用户失败: 无效的用户ID", zap.String("idStr", idStr), zap.Error(err))
		response.Fail(ctx, http.StatusBadRequest, err)
		return
	}

	logger.Logger.Info("用户ID参数验证通过", zap.Uint64("id", id))

	var req UpdateRequest
	if err := ctx.ShouldBindJSON(&req); err != nil {
		logger.Logger.Warn("更新用户失败: 请求参数验证失败", zap.Error(err))
		response.Fail(ctx, http.StatusBadRequest, err)
		return
	}

	logger.Logger.Info("更新参数验证通过")

	user, err := uc.userService.GetUserByID(uint(id))
	if err != nil {
		logger.Logger.Error("获取用户失败", zap.Error(err))
		response.Fail(ctx, http.StatusInternalServerError, err)
		return
	}

	// 更新用户信息
	if req.Nickname != "" {
		user.Nickname = req.Nickname
	}
	if req.Email != "" {
		user.Email = req.Email
	}
	if req.Phone != nil && *req.Phone != "" {
		user.Phone = req.Phone
	} else {
		user.Phone = nil
	}
	if req.Avatar != "" {
		user.Avatar = req.Avatar
	}
	if req.Status != 0 {
		user.Status = req.Status
	}

	if err := uc.userService.UpdateUser(user); err != nil {
		logger.Logger.Error("更新用户失败", zap.Error(err))
		response.Fail(ctx, http.StatusInternalServerError, err)
		return
	}

	response.Success(ctx, user)
}

// @Summary 删除用户
// @Description 根据ID删除用户
// @Tags 用户管理
// @Accept json
// @Produce json
// @Param id path int true "用户ID"
// @Success 200 {object} response.Response
// @Failure 400 {object} response.Response
// @Failure 401 {object} response.Response
// @Failure 403 {object} response.Response
// @Failure 404 {object} response.Response
// @Failure 500 {object} response.Response
// @Security BearerAuth
// @Router /users/{id} [delete]
// DeleteUser 删除用户
func (uc *UserController) DeleteUser(ctx *gin.Context) {
	logger.Logger.Info("开始删除用户操作")

	idStr := ctx.Param("id")
	id, err := strconv.ParseUint(idStr, 10, 64)
	if err != nil {
		logger.Logger.Warn("删除用户失败: 无效的用户ID", zap.String("idStr", idStr), zap.Error(err))
		response.Fail(ctx, http.StatusBadRequest, err)
		return
	}

	logger.Logger.Info("用户ID参数验证通过", zap.Uint64("id", id))

	if err := uc.userService.DeleteUser(uint(id)); err != nil {
		logger.Logger.Error("删除用户失败", zap.Error(err))
		response.Fail(ctx, http.StatusInternalServerError, err)
		return
	}

	logger.Logger.Info("删除用户成功", zap.Uint64("id", id))
	response.Success(ctx, "密码修改成功")
}

// @Summary 用户列表
// @Description 分页获取用户列表
// @Tags 用户管理
// @Accept json
// @Produce json
// @Param page query int false "页码(默认1)"
// @Param pageSize query int false "每页条数(默认10)"
// @Success 200 {object} response.Response{data=ListResponse{items=models.User}}
// @Failure 400 {object} response.Response
// @Failure 401 {object} response.Response
// @Failure 500 {object} response.Response
// @Security BearerAuth
// @Router /users [get]

// ListUsers 获取用户列表
func (c *UserController) ListUsers(ctx *gin.Context) {
	logger.Logger.Info("开始获取用户列表操作")

	pageStr := ctx.Query("page")
	pageSizeStr := ctx.Query("page_size")

	page, _ := strconv.Atoi(pageStr)
	pageSize, _ := strconv.Atoi(pageSizeStr)

	if page <= 0 {
		page = 1
		logger.Logger.Debug("分页参数重置: 页码设置为1")
	}
	if pageSize <= 0 || pageSize > 100 {
		pageSize = 10
		logger.Logger.Debug("分页参数重置: 每页数量设置为10")
	}

	logger.Logger.Info("分页参数验证通过", zap.Int("page", page), zap.Int("pageSize", pageSize))

	users, total, err := c.userService.ListUsers(page, pageSize)
	if err != nil {
		logger.Logger.Error("获取用户列表失败", zap.Error(err))
		response.Fail(ctx, http.StatusInternalServerError, err)
		return
	}

	logger.Logger.Info("获取用户列表成功", zap.Int("count", len(users)), zap.Int64("total", total))

	response.Success(ctx, gin.H{
		"list":  users,
		"total": total,
		"page":  page,
		"size":  pageSize,
	})
}

// UpdatePassword 修改密码
// @Summary 修改密码
// @Description 修改用户密码
// @Tags 用户管理
// @Accept json
// @Produce json
// @Param id path uint true "用户ID"
// @Param body body PasswordUpdateRequest true "密码修改请求"
// @Success 200 {object} response.Response{data=string}
// @Failure 400 {object} response.Response{data=string}
// @Failure 401 {object} response.Response{data=string}
// @Failure 500 {object} response.Response{data=string}
// @Security BearerAuth
// @Router /users/{id}/password [put]
func (uc *UserController) UpdatePassword(ctx *gin.Context) {
	logger.Logger.Info("开始修改用户密码操作")

	idStr := ctx.Param("id")
	id, err := strconv.ParseUint(idStr, 10, 64)
	if err != nil {
		logger.Logger.Warn("修改密码失败: 无效的用户ID", zap.String("idStr", idStr), zap.Error(err))
		response.Fail(ctx, http.StatusBadRequest, err)
		return
	}

	logger.Logger.Info("用户ID参数验证通过", zap.Uint64("id", id))

	var req PasswordUpdateRequest
	if err := ctx.ShouldBindJSON(&req); err != nil {
		logger.Logger.Warn("修改密码失败: 请求参数验证失败", zap.Error(err))
		response.Fail(ctx, http.StatusBadRequest, err)
		return
	}

	logger.Logger.Info("密码修改参数验证通过")

	user, err := uc.userService.GetUserByID(uint(id))
	if err != nil {
		logger.Logger.Error("获取用户失败", zap.Error(err))
		response.Fail(ctx, http.StatusInternalServerError, err)
		return
	}

	// 验证旧密码
	logger.Logger.Info("开始验证旧密码")
	if !uc.userService.VerifyPassword(user, req.OldPassword) {
		logger.Logger.Warn("修改密码失败: 旧密码错误", zap.Uint("userID", user.ID))
		response.Fail(ctx, http.StatusUnauthorized, errors.New("旧密码错误"))
		return
	}

	logger.Logger.Info("旧密码验证通过", zap.Uint("userID", user.ID))

	// 更新密码
	logger.Logger.Info("开始更新密码")
	if err := uc.userService.UpdatePassword(user, req.NewPassword); err != nil {
		logger.Logger.Error("修改密码失败: 更新密码出错", zap.Uint("userID", user.ID), zap.Error(err))
		response.Fail(ctx, http.StatusInternalServerError, err)
		return
	}

	logger.Logger.Info("密码更新成功", zap.Uint("userID", user.ID))
	response.Success(ctx, "密码修改成功")
}
