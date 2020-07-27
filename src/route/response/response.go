package route_response

import "Gin-IPs/src/configure"

type ResponseData struct {
	Page     int64         `json:"page"`  // 分页显示，页码
	PageSize int64         `json:"page_size"`  // 分页显示，页面大小
	Size     int64         `json:"size"`  // 返回的元素总数
	Total    int64         `json:"total"`  // 筛选条件计算得到的总数，但可能不会全部返回
	List     []interface{} `json:"list"`
}

type Response struct {
	Code    configure.Code `json:"code"`  // 响应码
	Message string         `json:"message"` // 响应描述
	Data    ResponseData   `json:"data"`  // 最终返回数据
}
