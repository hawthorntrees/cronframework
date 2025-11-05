package utils

import (
	"errors"
	"net"
	"strings"
)

var ErrNoValidIP = errors.New("未找到有效的非环回/非私有网络IP")

func GetLocalIP() (string, error) {
	addrs, err := net.InterfaceAddrs()
	if err != nil {
		return "", err
	}

	for _, addr := range addrs {
		ipNet, ok := addr.(*net.IPNet)
		if !ok || ipNet.IP.IsLoopback() {
			continue
		}

		ipV4 := ipNet.IP.To4()
		if ipV4 == nil {
			continue
		}

		ipStr := ipV4.String()
		if !isPrivateIP(ipStr) {
			return ipStr, nil
		}
	}

	return "", ErrNoValidIP
}

func isPrivateIP(ip string) bool {
	return strings.HasPrefix(ip, "10.") ||
		(strings.HasPrefix(ip, "172.") &&
			(strings.Split(ip, ".")[1] >= "16" && strings.Split(ip, ".")[1] <= "31")) ||
		strings.HasPrefix(ip, "192.168.")
}
