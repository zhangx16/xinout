package main

import (
	"fmt"
	"math/rand"
	"net"
	"strconv"
	"strings"
)

// 随机端口的取值范围，落在 IANA 动态端口区间内，避开常见服务。
const (
	randPortMin = 20000
	randPortMax = 60000
)

// freeRandomPort 随机挑一个当前空闲的 TCP 端口。
//
// taken 里的端口会被跳过，用于避开本进程已经分配但还没真正监听的端口。
// 实际可用性以能否 bind 为准，这样不会和系统上其他进程抢。
func freeRandomPort(taken map[int]bool) (int, error) {
	for i := 0; i < 200; i++ {
		port := randPortMin + rand.Intn(randPortMax-randPortMin)
		if taken[port] {
			continue
		}
		if portAvailable(port) {
			return port, nil
		}
	}
	return 0, fmt.Errorf("找不到可用端口（已尝试 200 次）")
}

// portAvailable 通过尝试监听来判断端口是否真的空闲。
func portAvailable(port int) bool {
	ln, err := net.Listen("tcp", fmt.Sprintf("0.0.0.0:%d", port))
	if err != nil {
		return false
	}
	ln.Close()
	return true
}

// parseOptionalPort 解析用户填的端口。空字符串表示随机，返回 0。
func parseOptionalPort(s string) (int, error) {
	s = strings.TrimSpace(s)
	if s == "" {
		return 0, nil
	}
	n, err := strconv.Atoi(s)
	if err != nil || n < 1 || n > 65535 {
		return 0, fmt.Errorf("端口不合法，需要 1-65535")
	}
	return n, nil
}

// consecutivePorts 分配 n 个端口。start 为 0 时全部随机；否则用 start, start+1, …。
// taken 里的端口视为已占用。成功时不会改传入的 taken。
func consecutivePorts(start, n int, taken map[int]bool) ([]int, error) {
	if n < 1 {
		return nil, fmt.Errorf("数量至少为 1")
	}
	local := map[int]bool{}
	for p, v := range taken {
		if v {
			local[p] = true
		}
	}
	out := make([]int, 0, n)
	if start == 0 {
		for i := 0; i < n; i++ {
			p, err := freeRandomPort(local)
			if err != nil {
				return nil, err
			}
			local[p] = true
			out = append(out, p)
		}
		return out, nil
	}
	if start < 1 || start > 65535 {
		return nil, fmt.Errorf("端口 %d 不合法", start)
	}
	if start+n-1 > 65535 {
		return nil, fmt.Errorf("从 %d 起连续 %d 个端口超出 65535", start, n)
	}
	for i := 0; i < n; i++ {
		p := start + i
		if local[p] {
			return nil, fmt.Errorf("端口 %d 已被占用", p)
		}
		// 用户指定的端口不再先 Listen 再关掉做探测：探测会把端口打进
		// TIME_WAIT，随后真正监听失败，NAT 映射的端口就对不上。
		local[p] = true
		out = append(out, p)
	}
	return out, nil
}

// normalizeProvisionPorts 处理 NAT 场景：用户只填了一个对外端口且选了节点模板，
// 这个端口必须给节点链接（客户端连它），SOCKS5 走本机随机，不必占面板放行的口。
func normalizeProvisionPorts(req *ProvisionRequest) {
	if req.TemplateID > 0 && req.InboundPort == 0 && req.SocksPort != 0 {
		req.InboundPort = req.SocksPort
		req.SocksPort = 0
	}
}

// portsOverlap 判断两段闭区间是否相交。start 为 0 表示该段未指定，不相交。
func portsOverlap(a, an, b, bn int) bool {
	if a == 0 || b == 0 || an < 1 || bn < 1 {
		return false
	}
	a2, b2 := a+an-1, b+bn-1
	return a <= b2 && b <= a2
}
