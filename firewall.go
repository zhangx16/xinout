package main

import (
	"os/exec"
	"strconv"
	"strings"
)

// ensureTCPPortOpen 尽量在本机防火墙放行 TCP 端口。
//
// NAT 小鸡常见情况：云面板只做公网 DNAT，虚拟机里 INPUT 仍是 DROP，
// 进程已经在听，外面照样连不上。这里对本机 iptables / firewalld / ufw
// 尽力插入放行，失败忽略，不阻断主流程。
func ensureTCPPortOpen(port int) {
	if port < 1 || port > 65535 {
		return
	}
	p := strconv.Itoa(port)

	if _, err := exec.LookPath("iptables"); err == nil {
		if cmdRun(exec.Command("iptables", "-C", "INPUT", "-p", "tcp", "--dport", p, "-j", "ACCEPT")) != nil {
			_ = cmdRun(exec.Command("iptables", "-I", "INPUT", "1", "-p", "tcp", "--dport", p, "-j", "ACCEPT"))
		}
	}

	if _, err := exec.LookPath("firewall-cmd"); err == nil {
		if cmdRun(exec.Command("firewall-cmd", "--state")) == nil {
			_ = cmdRun(exec.Command("firewall-cmd", "--add-port="+p+"/tcp"))
		}
	}

	if _, err := exec.LookPath("ufw"); err == nil {
		out, err := cmdOutput(exec.Command("ufw", "status"))
		if err == nil && strings.Contains(strings.ToLower(string(out)), "status: active") {
			_ = cmdRun(exec.Command("ufw", "allow", p+"/tcp"))
		}
	}
}
