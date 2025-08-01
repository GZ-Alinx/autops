package global

import (
	casbin "github.com/casbin/casbin/v2"
)

var (
	// Enforcer Casbin权限控制实例
	Enforcer *casbin.Enforcer
)