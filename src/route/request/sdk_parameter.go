package route_request

import "Gin-IPs/src/configure"

type ReqGetParaSearchIp struct {
	Parameter
	Ip  string        `form:"ip" validate:"required"`
	Oid configure.Oid `form:"oid" validate:"check_oid"`
}

type ReqPostParaSearchIp struct {
	Parameter
	Ip  string        `json:"ip" validate:"required"`
	Oid configure.Oid `json:"oid" validate:"check_oid"`
}
