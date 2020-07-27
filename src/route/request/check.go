package route_request

import (
	"Gin-IPs/src/utils/network"
	"errors"
	"fmt"
)

func CheckIp(ipArr []string, low, high int) error {
	if low > len(ipArr) || len(ipArr) > high {
		return errors.New(fmt.Sprintf("请求IP数量超过限制"))
	}
	for _, ip := range ipArr {
		if !network.MatchIpPattern(ip) {
			return errors.New(fmt.Sprintf("错误的IP格式:%s", ip))
		}
		if network.ErrorIpPattern(ip) {
			return errors.New(fmt.Sprintf("不支持的IP:%s", ip))
		}
	}
	return nil
}
