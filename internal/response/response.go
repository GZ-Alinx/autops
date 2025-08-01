package response

import (
	"net/http"

	"github.com/gin-gonic/gin"
)

// Response 统一响应格式
// @Description API统一响应结构体
type Response struct {
	Code    int         `json:"code" swaggertype:"integer" example:"200" description:"状态码"`
	Message string      `json:"message" example:"操作成功" description:"响应消息"`
	Data    interface{} `json:"data,omitempty" description:"响应数据"`
}

// Success 成功响应
func Success(c *gin.Context, data interface{}) {
	c.JSON(http.StatusOK, Response{
		Code:    http.StatusOK,
		Message: "success",
		Data:    data,
	})
}

// OkWithData 成功响应（带数据）
func OkWithData(c *gin.Context, data interface{}) {
	Success(c, data)
}

// Fail 失败响应
func Fail(c *gin.Context, code int, err error) {
	message := "failed"
	if err != nil {
		message = err.Error()
	}

	c.JSON(code, Response{
		Code:    code,
		Message: message,
		Data:    nil,
	})
}

// ParamError 参数错误响应
func ParamError(c *gin.Context, err error) {
	Fail(c, http.StatusBadRequest, err)
}

// BadRequest 请求参数错误响应
func BadRequest(c *gin.Context, err error) {
	Fail(c, http.StatusBadRequest, err)
}

// Unauthorized 未授权响应
func Unauthorized(c *gin.Context, err error) {
	Fail(c, http.StatusUnauthorized, err)
}

// Forbidden 禁止访问响应
func Forbidden(c *gin.Context, err error) {
	Fail(c, http.StatusForbidden, err)
}

// NotFound 资源不存在响应
func NotFound(c *gin.Context, err error) {
	Fail(c, http.StatusNotFound, err)
}

// ServerError 服务器错误响应
func ServerError(c *gin.Context, err error) {
	Fail(c, http.StatusInternalServerError, err)
}

// InternalServerError 服务器内部错误响应
func InternalServerError(c *gin.Context, err error) {
	Fail(c, http.StatusInternalServerError, err)
}
