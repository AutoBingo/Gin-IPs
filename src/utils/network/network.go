package network

import (
	"fmt"
	"net"
	"strings"
)

// 通过向外网发一个UDP请求得到本地默认路由对应的IP地址，排除局域网，UDP本身是无连接的
func GetLoginIp() (string, error) {
	if udpAddr, err := net.ResolveUDPAddr("udp4", "8.8.8.8:8"); err != nil {
		return "", err
	} else {
		if udpConn, err := net.DialUDP("udp", nil, udpAddr); err != nil {
			return "", err
		} else {
			defer udpConn.Close()
			ipPort := strings.Split(udpConn.LocalAddr().String(), ":")
			fmt.Println(ipPort)
			if len(ipPort) >= 2 {
				return ipPort[0], nil
			}
			return "", nil
		}
	}
}

// IP地址格式匹配  "010.99.32.88" 属于正常IP
func MatchIpPattern(ip string) bool {
	//pattern := `^((2[0-4]\d|25[0-5]|[01]?\d\d?)\.){3}(2[0-4]\d|25[0-5]|[01]?\d\d?)$`
	//reg := regexp.MustCompile(pattern)
	//return reg.MatchString(ip)
	if net.ParseIP(ip) == nil {
		return false
	}
	return true
}

// 排查错误的IP
func ErrorIpPattern(ip string) bool {
	errorIpMapper := map[string]bool{
		"192.168.122.1": true,
		"192.168.250.1": true,
		"192.168.255.1": true,
		"192.168.99.1":  true,
		"192.168.56.1":  true,
		"10.10.10.1":    true,
	}
	errorIpPrefixPattern := []string{"127.0.0.", "169.254.", "11.1.", "10.176."}
	errorIpSuffixPattern := []string{".0.1"}
	if _, ok := errorIpMapper[ip]; ok {
		return true
	}
	for _, p := range errorIpPrefixPattern {
		if strings.HasPrefix(ip, p) {
			return true
		}
	}
	for _, p := range errorIpSuffixPattern {
		if strings.HasSuffix(ip, p) {
			return true
		}
	}
	return false
}
