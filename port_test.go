package main

import (
	"net"
	"testing"
)

func TestParseOptionalPort(t *testing.T) {
	if n, err := parseOptionalPort(""); err != nil || n != 0 {
		t.Fatalf("空值应为 0: %d %v", n, err)
	}
	if n, err := parseOptionalPort(" 443 "); err != nil || n != 443 {
		t.Fatalf("443: %d %v", n, err)
	}
	if _, err := parseOptionalPort("0"); err == nil {
		t.Fatal("0 应非法")
	}
	if _, err := parseOptionalPort("65536"); err == nil {
		t.Fatal("65536 应非法")
	}
	if _, err := parseOptionalPort("abc"); err == nil {
		t.Fatal("非数字应非法")
	}
}

func TestConsecutivePortsRandom(t *testing.T) {
	got, err := consecutivePorts(0, 3, nil)
	if err != nil {
		t.Fatal(err)
	}
	if len(got) != 3 {
		t.Fatalf("想要 3 个，实际 %v", got)
	}
	seen := map[int]bool{}
	for _, p := range got {
		if p < randPortMin || p >= randPortMax {
			t.Fatalf("随机端口越界: %d", p)
		}
		if seen[p] {
			t.Fatalf("重复端口 %d", p)
		}
		seen[p] = true
	}
}

func TestConsecutivePortsExact(t *testing.T) {
	ln, err := net.Listen("tcp", "127.0.0.1:0")
	if err != nil {
		t.Fatal(err)
	}
	defer ln.Close()
	base := ln.Addr().(*net.TCPAddr).Port
	ln.Close()

	got, err := consecutivePorts(base, 2, nil)
	if err != nil {
		t.Fatal(err)
	}
	if len(got) != 2 || got[0] != base || got[1] != base+1 {
		t.Fatalf("应为 %d,%d 实际 %v", base, base+1, got)
	}
}

func TestNormalizeProvisionPorts(t *testing.T) {
	req := ProvisionRequest{TemplateID: 7, SocksPort: 21683}
	normalizeProvisionPorts(&req)
	if req.InboundPort != 21683 || req.SocksPort != 0 {
		t.Fatalf("有模板时单端口应给节点: %+v", req)
	}
	req = ProvisionRequest{SocksPort: 21683}
	normalizeProvisionPorts(&req)
	if req.SocksPort != 21683 || req.InboundPort != 0 {
		t.Fatalf("无模板时应保持 SOCKS 端口: %+v", req)
	}
}

func TestConsecutivePortsTakenMap(t *testing.T) {
	if _, err := consecutivePorts(30000, 1, map[int]bool{30000: true}); err == nil {
		t.Fatal("taken 中的端口应当报错")
	}
}

func TestPortsOverlap(t *testing.T) {
	if !portsOverlap(100, 3, 102, 2) {
		t.Fatal("100-102 与 102-103 应相交")
	}
	if portsOverlap(100, 3, 103, 2) {
		t.Fatal("100-102 与 103-104 不应相交")
	}
	if portsOverlap(0, 3, 100, 2) {
		t.Fatal("未指定起点不应相交")
	}
}
