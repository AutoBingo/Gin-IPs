// 自定义参数验证器需要完善名称、字典、函数，并在 parameter.go 中完善提示
package route_request

import (
	"Gin-IPs/src/configure"
)

// 自定义参数验证器名称
const (
	ValNameCheckOid string = "check_oid"
)

// 自定义参数验证器字典
var validatorMapper = map[string]func(field validator.FieldLevel) bool{
	ValNameCheckOid: CheckOid,
}

// 自定义参数验证器函数
func CheckOid(field validator.FieldLevel) bool {
	oid := configure.Oid(field.Field().String())
	for _, id := range configure.OidArray {
		if oid == id {
			return true
		}
	}
	return false
}
