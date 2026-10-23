//go:build windows

package core

import (
	"fmt"
	"os/exec"
	"strconv"
	"strings"
	"syscall"
)

// KillStraySingBoxOnListenPort завершает процессы sing-box, которые всё ещё слушают
// mixed-порт (например после сбоя или потери дочернего процесса).
func KillStraySingBoxOnListenPort(port int) {
	if port < 1 || port > 65535 {
		return
	}
	killStraySingBoxViaNetTCP(port)
	killStraySingBoxViaNetstat(port)
}

func killStraySingBoxViaNetTCP(port int) {
	ps := fmt.Sprintf(
		"$ErrorActionPreference='SilentlyContinue'; "+
			"Get-NetTCPConnection -LocalPort %d -State Listen | ForEach-Object { "+
			"$proc = Get-Process -Id $_.OwningProcess -ErrorAction SilentlyContinue; "+
			"if ($proc -and $proc.ProcessName -ieq 'sing-box') { Stop-Process -Id $proc.Id -Force } }",
		port,
	)
	cmd := exec.Command("powershell.exe", "-NoProfile", "-ExecutionPolicy", "Bypass", "-WindowStyle", "Hidden", "-Command", ps)
	cmd.SysProcAttr = &syscall.SysProcAttr{
		HideWindow:    true,
		CreationFlags: 0x08000000, // CREATE_NO_WINDOW
	}
	_ = cmd.Run()
}

// Fallback, если Get-NetTCPConnection недоступен: netstat -ano + taskkill.
func killStraySingBoxViaNetstat(port int) {
	out, err := exec.Command("netstat", "-ano").Output()
	if err != nil {
		return
	}
	portSuffix := fmt.Sprintf(":%d", port)
	for _, line := range strings.Split(string(out), "\n") {
		line = strings.TrimSpace(line)
		if line == "" {
			continue
		}
		low := strings.ToLower(line)
		// "listening" покрывает EN: LISTENING; "ожидание" — RU-локаль Windows.
		// Оба варианта проверяем по ToLower (low), чтобы избежать смешанного сравнения.
		if !strings.Contains(low, "listening") && !strings.Contains(low, "ожидание") {
			continue
		}
		fields := strings.Fields(line)
		if len(fields) < 4 {
			continue
		}
		local := fields[1]
		if !strings.HasSuffix(local, portSuffix) {
			continue
		}
		pidStr := fields[len(fields)-1]
		pid, err := strconv.Atoi(pidStr)
		if err != nil || pid <= 0 {
			continue
		}
		if !tasklistLineLooksLikeSingBox(pid) {
			continue
		}
		kill := exec.Command("taskkill.exe", "/PID", pidStr, "/F", "/T")
		kill.SysProcAttr = &syscall.SysProcAttr{
			HideWindow:    true,
			CreationFlags: 0x08000000,
		}
		_ = kill.Run()
	}
}

func tasklistLineLooksLikeSingBox(pid int) bool {
	out, err := exec.Command("tasklist.exe", "/FI", fmt.Sprintf("PID eq %d", pid), "/FO", "CSV", "/NH").Output()
	if err != nil {
		return false
	}
	return strings.Contains(strings.ToLower(string(out)), "sing-box")
}
