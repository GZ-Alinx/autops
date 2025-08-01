package middleware

import (
	"time"

	"gorm.io/gorm"
)

// @title 权限管理系统 API
// @version 1.0
// @description 基于Gin和Casbin的权限管理系统
// @termsOfService http://swagger.io/terms/
// @contact.name API Support
// @contact.email admin@example.com
// @license.name Apache 2.0
// @license.url http://www.apache.org/licenses/LICENSE-2.0.html
// @host localhost:8888
// @BasePath /api/v1
func init() {
	// 这里可以添加初始化代码
}

// Swagger文档中gorm.DeletedAt类型的映射
// @Schema(type="string", format="date-time")
type DeletedAt gorm.DeletedAt

// 为DeletedAt类型添加MarshalJSON方法，确保在Swagger中正确显示
func (d DeletedAt) MarshalJSON() ([]byte, error) {
	if d.Time.IsZero() {
		return []byte("null"), nil
	}
	return d.Time.MarshalJSON()
}

// 为DeletedAt类型添加UnmarshalJSON方法
func (d *DeletedAt) UnmarshalJSON(data []byte) error {
	if string(data) == "null" {
		d.Time = time.Time{}
		d.Valid = false
		return nil
	}
	err := d.Time.UnmarshalJSON(data)
	if err == nil {
		d.Valid = true
	}
	return err
}
