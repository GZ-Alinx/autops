# Autops项目开发文档

## 1. 项目概述
Autops是一个基于Go语言开发的权限管理系统，采用Gin框架和RBAC权限模型，提供用户管理、角色分配和API权限控制功能。系统集成Swagger文档，支持RESTful API设计规范，具备完善的错误处理和统一响应格式。

## 2. 技术栈
- **后端框架**: Gin v1.9.1
- **ORM**: GORM v2.0
- **权限控制**: Casbin v2.0
- **API文档**: Swagger/OpenAPI
- **日志**: Zap
- **配置管理**: Viper
- **数据库**: MySQL
- **认证**: JWT

## 3. 项目结构
```
├── business/        # 业务逻辑层
│   ├── controllers/ # 控制器
│   ├── models/      # 数据模型
│   ├── repositories/ # 数据访问层
│   ├── routes/      # 路由定义
│   └── services/    # 服务层
├── config.yaml      # 配置文件
├── configs/         # 配置文件目录
│   └── casbin_model.conf # Casbin权限模型配置
├── docs/            # Swagger文档
├── internal/        # 内部模块
│   ├── config/      # 配置管理
│   ├── database/    # 数据库连接
│   ├── global/      # 全局变量
│   ├── logger/      # 日志
│   ├── middleware/  # 中间件
│   └── response/    # 响应处理
├── main.go          # 入口文件
└── postman_collection.json # Postman测试集合
```

## 4. 中间件说明
### 4.1 JWT认证中间件 (jwt.go)
**功能**: 负责JWT令牌的验证和解析

**使用方式**: 在需要认证的路由组中使用
```go
router.Use(middleware.JWTMiddleware())
```

**工作流程**: 
1. 从请求头中获取Authorization字段
2. 验证令牌格式是否正确
3. 解析并验证令牌有效性
4. 将用户信息存入上下文

**核心函数**: 
- `JWTMiddleware()`: 返回JWT认证中间件处理函数
- `GenerateToken(userID, username string) (string, error)`: 生成JWT令牌

### 4.2 Casbin权限中间件 (casbin.go)
**功能**: 基于Casbin实现API访问权限控制

**使用方式**: 在需要权限控制的路由组中使用
```go
router.Use(middleware.CasbinMiddleware())
```

**工作流程**: 
1. 获取当前请求的用户和角色信息
2. 获取请求路径和方法
3. 使用Casbin检查权限
4. 根据检查结果允许或拒绝请求

### 4.3 CORS中间件 (cors.go)
**功能**: 处理跨域请求

**使用方式**: 全局使用
```go
router.Use(middleware.CorsMiddleware())
```

**配置项**: 
- `AllowOrigins`: 允许的源
- `AllowMethods`: 允许的HTTP方法
- `AllowHeaders`: 允许的请求头
- `AllowCredentials`: 是否允许凭证
- `MaxAge`: 预检请求的缓存时间

## 5. API文档
### 5.1 基础信息
- 基础URL: `/api/v1`
- 认证方式: JWT (Authorization: Bearer Token)
- 响应格式: 统一JSON格式

### 5.2 用户管理API
#### 5.2.1 用户登录
- **路径**: `/user/login`
- **方法**: `POST`
- **认证**: 无需认证
- **请求参数**:
  ```json
  {
    "username": "string", // 用户名 (必填)
    "password": "string"  // 密码 (必填)
  }
  ```
- **响应示例**:
  ```json
  {
    "code": 200,
    "message": "success",
    "data": {
      "token": "string",
      "user": {
        "id": 1,
        "username": "admin",
        "email": "admin@example.com",
        "phone": "13800138000",
        "created_at": "2023-01-01T00:00:00Z"
      }
    }
  }
  ```

#### 5.2.2 用户添加
- **路径**: `/users/register`
- **方法**: `POST`
- **认证**: 需要JWT
- **请求参数**:
  ```json
  {
    "username": "string", // 用户名 (必填)
    "password": "string", // 密码 (必填，最小长度6)
    "email": "string",    // 邮箱 (必填，需符合邮箱格式)
    "phone": "string",    // 电话 (可选)
    "nickname": "string"  // 昵称 (可选)
  }
  ```
- **响应示例**:
  ```json
  {
    "code": 200,
    "message": "success",
    "data": {
      "id": 2,
      "username": "testuser",
      "email": "test@example.com",
      "phone": "13800138000",
      "created_at": "2023-01-01T00:00:00Z"
    }
  }
  ```

#### 5.2.3 获取用户详情
- **路径**: `/users/{id}`
- **方法**: `GET`
- **认证**: 需要JWT
- **权限**: `users:read`
- **路径参数**:
  - `id`: 用户ID
- **响应示例**:
  ```json
  {
    "code": 200,
    "message": "success",
    "data": {
      "id": 1,
      "username": "admin",
      "email": "admin@example.com",
      "phone": "13800138000",
      "created_at": "2023-01-01T00:00:00Z",
      "roles": ["admin"]
    }
  }
  ```

#### 5.2.4 更新用户信息
- **路径**: `/users/{id}`
- **方法**: `PUT`
- **认证**: 需要JWT
- **权限**: `users:update`
- **路径参数**:
  - `id`: 用户ID
- **请求参数**:
  ```json
  {
    "email": "string",    // 邮箱 (可选，需符合邮箱格式)
    "phone": "string",    // 电话 (可选)
    "nickname": "string", // 昵称 (可选)
    "avatar": "string",   // 头像URL (可选)
    "status": 1            // 状态 (可选，只能是0或1)
  }
  ```
- **响应示例**:
  ```json
  {
    "code": 200,
    "message": "success",
    "data": {
      "id": 1,
      "username": "updatedadmin",
      "email": "updated@example.com",
      "phone": "13800138000",
      "updated_at": "2023-01-02T00:00:00Z"
    }
  }
  ```

#### 5.2.5 删除用户
- **路径**: `/users/{id}`
- **方法**: `DELETE`
- **认证**: 需要JWT
- **权限**: `users:delete`
- **路径参数**:
  - `id`: 用户ID
- **响应示例**:
  ```json
  {
    "code": 200,
    "message": "success",
    "data": null
  }
  ```

#### 5.2.6 密码修改
- **路径**: `/users/{id}/password`
- **方法**: `PUT`
- **认证**: 需要JWT
- **权限**: `users:update`
- **路径参数**:
  - `id`: 用户ID
- **请求参数**:
  ```json
  {
    "old_password": "string", // 旧密码 (必填)
    "new_password": "string"  // 新密码 (必填，最小长度6)
  }
  ```
- **响应示例**:
  ```json
  {
    "code": 200,
    "message": "密码修改成功",
    "data": null
  }
  ```

#### 5.2.7 获取用户列表
- **路径**: `/users`
- **方法**: `GET`
- **认证**: 需要JWT
- **权限**: `users:list`
- **查询参数**:
  - `page`: 页码 (默认: 1)
  - `pageSize`: 每页数量 (默认: 10)
- **响应示例**:
  ```json
  {
    "code": 200,
    "message": "success",
    "data": {
      "total": 100,
      "list": [
        {
          "id": 1,
          "username": "admin",
          "email": "admin@example.com",
          "phone": "13800138000",
          "created_at": "2023-01-01T00:00:00Z"
        },
        // ...更多用户
      ],
      "page": 1,
      "pageSize": 10
    }
  }
  ```

### 5.3 角色管理API
#### 5.3.1 创建角色
- **路径**: `/roles`
- **方法**: `POST`
- **认证**: 需要JWT
- **权限**: `roles:create`
- **请求参数**:
  ```json
  {
    "name": "string",        // 角色名称 (必填，长度2-50)
    "description": "string"  // 角色描述 (可选，最大长度255)
  }
  ```
- **响应示例**:
  ```json
  {
    "code": 200,
    "message": "success",
    "data": {
      "id": 3,
      "name": "editor",
      "description": "编辑角色",
      "created_at": "2023-01-01T00:00:00Z"
    }
  }
  ```

#### 5.3.2 获取所有角色
- **路径**: `/roles`
- **方法**: `GET`
- **认证**: 需要JWT
- **权限**: `roles:list`
- **响应示例**:
  ```json
  {
    "code": 200,
    "message": "success",
    "data": [
      {
        "id": 1,
        "name": "admin",
        "description": "管理员角色",
        "created_at": "2023-01-01T00:00:00Z"
      },
      {
        "id": 2,
        "name": "user",
        "description": "普通用户角色",
        "created_at": "2023-01-01T00:00:00Z"
      }
    ]
  }
  ```

#### 5.3.3 获取角色详情
- **路径**: `/roles/{id}`
- **方法**: `GET`
- **认证**: 需要JWT
- **权限**: `roles:read`
- **路径参数**:
  - `id`: 角色ID
- **响应示例**:
  ```json
  {
    "code": 200,
    "message": "success",
    "data": {
      "id": 1,
      "name": "admin",
      "description": "管理员角色",
      "created_at": "2023-01-01T00:00:00Z",
      "permissions": [
        {
          "id": 1,
          "resource": "/users",
          "action": "GET",
          "description": "查看用户列表权限"
        },
        // ...更多权限
      ]
    }
  }
  ```

#### 5.3.4 更新角色
- **路径**: `/roles/{id}`
- **方法**: `PUT`
- **认证**: 需要JWT
- **权限**: `roles:update`
- **路径参数**:
  - `id`: 角色ID
- **请求参数**:
  ```json
  {
    "id": 1,                  // 角色ID (必填)
    "name": "string",        // 角色名称 (可选，长度2-50)
    "description": "string"  // 角色描述 (可选，最大长度255)
  }
  ```
- **响应示例**:
  ```json
  {
    "code": 200,
    "message": "success",
    "data": {
      "id": 3,
      "name": "editor",
      "description": "更新后的编辑角色",
      "updated_at": "2023-01-02T00:00:00Z"
    }
  }
  ```

#### 5.3.5 删除角色
- **路径**: `/roles/{id}`
- **方法**: `DELETE`
- **认证**: 需要JWT
- **权限**: `roles:delete`
- **路径参数**:
  - `id`: 角色ID
- **响应示例**:
  ```json
  {
    "code": 200,
    "message": "success",
    "data": null
  }
  ```

### 5.4 权限管理API
#### 5.4.1 更新用户角色权限关联
- **路径**: `/permissions/user-role`
- **方法**: `PUT`
- **认证**: 需要JWT
- **权限**: `permissions:update`
- **请求参数**: 
  ```json
  {
    "user_id": 1,           // 用户ID (必填)
    "roles": ["admin", "user"]  // 角色名称列表 (必填，至少一个角色)
  }
  ```
- **响应示例**: 
  ```json
  {
    "code": 200,
    "message": "success",
    "data": "成功为用户分配 2 个角色"
  }
  ```

#### 5.4.2 添加权限策略
- **路径**: `/permissions/policy`
- **方法**: `POST`
- **认证**: 需要JWT
- **权限**: `permissions:create`
- **请求参数**:
  ```json
  {
    "role": "string",   // 角色名称 (必填)
    "path": "string",   // 资源路径 (必填)
    "method": "string"  // HTTP方法 (必填，只能是GET, POST, PUT, DELETE, PATCH之一)
  }
  ```
- **响应示例**:
  ```json
  {
    "code": 200,
    "message": "添加权限策略成功",
    "data": null
  }
  ```

#### 5.4.2 删除权限策略
- **路径**: `/permissions/policy`
- **方法**: `DELETE`
- **认证**: 需要JWT
- **权限**: `permissions:delete`
- **请求参数**:
  ```json
  {
    "role": "string",   // 角色名称 (必填)
    "path": "string",   // 资源路径 (必填)
    "method": "string"  // HTTP方法 (必填)
  }
  ```
- **响应示例**:
  ```json
  {
    "code": 200,
    "message": "删除权限策略成功",
    "data": null
  }
  ```

#### 5.4.3 获取所有权限策略
- **路径**: `/permissions/policies`
- **方法**: `GET`
- **认证**: 需要JWT
- **权限**: `permissions:list`
- **响应示例**:
  ```json
  {
    "code": 200,
    "message": "success",
    "data": [
      "admin,/users,GET",
      "admin,/users,POST",
      "user,/users,GET",
      // ...更多策略
    ]
  }
  ```

### 5.5 测试API
#### 5.5.1 健康检查
- **路径**: `/health`
- **方法**: `GET`
- **认证**: 无需认证
- **响应示例**:
  ```json
  {
    "code": 200,
    "message": "success",
    "data": "OK"
  }
  ```

#### 5.5.2 测试权限
- **路径**: `/api/v1/test`
- **方法**: `GET`
- **认证**: 需要JWT
- **响应示例**:
  ```json
  {
    "code": 200,
    "message": "success",
    "data": {
      "has_permission": true
    }
  }
  ```

## 6. 权限模型
系统使用Casbin实现RBAC权限模型，支持路径通配符匹配，权限定义在`configs/casbin_model.conf`文件中：

```
[request_definition]
r = sub, obj, act

[policy_definition]
p = sub, obj, act

[role_definition]
g = _, _

[policy_effect]
e = some(where (p.eft == allow))

[matcher]
m = g(r.sub, p.sub) && keyMatch(r.obj, p.obj) && r.act == p.act
```

> 注意：`keyMatch`函数支持路径中的`*`通配符匹配，例如`/api/v1/users/*`可以匹配`/api/v1/users/1`、`/api/v1/users/2`等具体用户路径。

- `sub`: 主体 (用户或角色)
- `obj`: 客体 (资源路径)
- `act`: 动作 (HTTP方法)
- `g`: 角色继承关系
- `p`: 权限策略

## 7. 响应格式
系统采用统一的JSON响应格式：

```json
{
  "code": 200,          // 状态码
  "message": "success", // 响应消息
  "data": {}            // 响应数据 (可选)
}
```

## 8. 错误处理
系统定义了多种错误响应函数，对应不同的HTTP状态码：
- `ParamError`: 参数错误 (400)
- `Unauthorized`: 未授权 (401)
- `Forbidden`: 禁止访问 (403)
- `NotFound`: 资源不存在 (404)
- `ServerError`: 服务器错误 (500)

## 9. 快速开始
1. 克隆仓库
2. 安装依赖: `go mod download`
3. 配置数据库: 编辑`config.yaml`
4. 生成Swagger文档: `swag init -g main.go --output docs`
5. 启动服务: `go run main.go`
6. 访问API文档: http://localhost:8081/swagger/index.html

## 10. 开发建议
1. 遵循RESTful API设计规范
2. 使用Swagger注解为API添加文档
3. 优先使用系统提供的响应函数返回统一格式
4. 新增API时，同时添加相应的权限控制
5. 使用Postman测试集合进行API测试