package util

import (
	"fmt"
	"math/rand"
	"net"
	"runtime"
	"strings"
)

// 获取本机内网IP
func GetLocalEthernetIP() (string, error) {
	// 获取所有网络接口
	interfaces, err := net.Interfaces()
	if err != nil {
		return "", fmt.Errorf("failed to get network interfaces: %v", err)
	}

	// 根据操作系统选择适当的接口名称前缀
	var targetPrefixes []string
	switch runtime.GOOS {
	case "windows":
		targetPrefixes = []string{"WLAN", "以太网", "Ethernet", "eth", "en"}
	case "darwin": // macOS
		targetPrefixes = []string{"en"}
	case "linux":
		targetPrefixes = []string{"eth", "enp", "ens"}
	default:
		targetPrefixes = []string{"eth", "en"}
	}

	for _, iface := range interfaces {
		// 跳过回环接口和未启用的接口
		if iface.Flags&net.FlagLoopback != 0 || iface.Flags&net.FlagUp == 0 {
			continue
		}

		// 检查接口名称是否匹配目标前缀
		var match bool
		for _, prefix := range targetPrefixes {
			if strings.HasPrefix(iface.Name, prefix) {
				match = true
				break
			}
		}
		if !match {
			continue
		}

		// 获取接口的地址
		addrs, err := iface.Addrs()
		if err != nil {
			continue
		}

		// 遍历地址，寻找IPv4地址
		for _, addr := range addrs {
			var ip net.IP
			switch v := addr.(type) {
			case *net.IPNet:
				ip = v.IP
			case *net.IPAddr:
				ip = v.IP
			}

			// 跳过IPv6和非全局单播地址
			if ip == nil || ip.IsLoopback() || !ip.IsGlobalUnicast() {
				continue
			}

			// 返回第一个找到的IPv4地址
			if ip.To4() != nil {
				return ip.String(), nil
			}
		}
	}

	return "", fmt.Errorf("no Ethernet IPv4 address found")
}

// 在指定范围内随机查找可用端口
func GetAvailablePort(minPort, maxPort int) (int, error) {
	// 验证端口范围
	if minPort < 1 || maxPort > 65535 || minPort > maxPort {
		return 0, fmt.Errorf("invalid port range: %d-%d", minPort, maxPort)
	}

	// 最大尝试次数（避免无限循环）
	maxAttempts := maxPort - minPort
	for i := 0; i < maxAttempts; i++ {
		// 随机选择一个端口
		port := minPort + rand.Intn(maxPort-minPort+1)

		// 检查端口是否可用
		addr := fmt.Sprintf(":%d", port)
		listener, err := net.Listen("tcp", addr)
		if err == nil {
			listener.Close()
			return port, nil
		}
	}

	return 0, fmt.Errorf("no available port found in range %d-%d after %d attempts", minPort, maxPort, maxAttempts)
}
