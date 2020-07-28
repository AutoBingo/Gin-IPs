package route_request

import (
	"Gin-IPs/src/configure"
	"encoding/json"
	"fmt"
)

// base
type Parameter struct {
}

// 参数验证之错误提示
func (r *Parameter) ParseError(errs error) (configure.Code, string) {
	if errs == nil {
		return configure.RequestSuccess, ""
	}
	switch errs.(type) {
	case *json.UnmarshalTypeError: // 不能使用 jsoniter
		err := errs.(*json.UnmarshalTypeError)
		msg := fmt.Sprintf("请求参数 %s 类型错误", err.Field)
		//msg := fmt.Sprintf("参数 %s 类型错误，期望类型是 %s", err.Field, err.Type.String())
		return configure.RequestParameterTypeError, msg
	case *validator.InvalidValidationError:
		return configure.RequestOtherError, errs.Error()
	case validator.ValidationErrors: // 也可以用    msg := v.Translate(errs.(validator.ValidationErrors)）
		for _, err := range errs.(validator.ValidationErrors) {
			switch err.Tag() {
			case ValNameCheckOid: // 自定义
				return configure.RequestParameterRangeError, fmt.Sprintf("请求参数 %s 需要传入 object id", err.Field())
			case "required":
				return configure.RequestParameterMiss, fmt.Sprintf("请求参数缺少 %s", err.Field())
			case "min", "max":
				return configure.RequestParameterRangeError, fmt.Sprintf("请求参数 %s 的值范围错误", err.Field())
			}
		}
	}
	return configure.RequestOtherError, "未知错误"
}
