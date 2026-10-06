package main

import (
	"strings"
	"sync"
	"testing"
)

func forceNsExecMode(nsenter bool) {
	nsExecOnce = sync.Once{}
	nsExecOnce.Do(func() {})
	nsExecNsenter = nsenter
}

func TestNsCmdUsesIpNetnsExecByDefault(t *testing.T) {
	forceNsExecMode(false)
	cmd := nsCmd("fo1", "true")
	joined := strings.Join(cmd.Args, " ")
	if !strings.Contains(joined, "netns exec fo1") {
		t.Fatalf("普通系统应走 ip netns exec，实际 %v", cmd.Args)
	}
	if strings.Contains(joined, "nsenter") {
		t.Fatalf("普通系统不应走 nsenter，实际 %v", cmd.Args)
	}
}

func TestNsCmdFallsBackToNsenterWhenRestricted(t *testing.T) {
	forceNsExecMode(true)
	cmd := nsCmd("fo1", "true")
	joined := strings.Join(cmd.Args, " ")
	if !strings.Contains(joined, "nsenter") {
		t.Fatalf("受限环境应走 nsenter，实际 %v", cmd.Args)
	}
	if strings.Contains(joined, "netns exec") {
		t.Fatalf("受限环境不应再走 ip netns exec，实际 %v", cmd.Args)
	}
}
