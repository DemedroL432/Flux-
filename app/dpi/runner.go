package dpi

import (
	"errors"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"sync"
	"syscall"
	"time"

	"pulseshield/app/core"
	"pulseshield/app/models"
)

type Runner struct {
	mu         sync.Mutex
	cmd        *exec.Cmd
	lastLog    string
	activeTool string
}

func (r *Runner) Start(profile models.DPIProfile, cfg *models.AppConfig) error {
	r.mu.Lock()
	defer r.mu.Unlock()
	if r.cmd != nil && r.cmd.Process != nil {
		return nil
	}
	var (
		bin  string
		dir  string
		args []string
		err  error
	)

	switch profile.Tool {
	case "zapret_winws2":
		bin, dir, err = ResolveZapretWinws2()
		if err != nil {
			return err
		}
		atArg, err := PrepareWinwsPresetArg(dir, profile, cfg)
		if err != nil {
			return err
		}
		// winws2 accepts @<file> with args inside; run from zapret base dir
		// so relative paths like lua/, lists/, windivert.filter/ work.
		args = []string{atArg}
	default:
		bin, err = resolveToolBinary(profile.Tool)
		if err != nil {
			return err
		}
		args = make([]string, 0, len(profile.Args))
		for _, a := range profile.Args {
			args = append(args, strings.ReplaceAll(a, "{config}", profile.ConfigPath))
		}
	}

	logFile, logPath, err := core.OpenLogFile("dpi")
	if err != nil {
		return err
	}
	r.lastLog = logPath

	cmd := exec.Command(bin, args...)
	cmd.SysProcAttr = &syscall.SysProcAttr{
		HideWindow:    true,
		CreationFlags: 0x08000000, // CREATE_NO_WINDOW
	}
	if dir != "" {
		cmd.Dir = dir
	}
	cmd.Stdout = logFile
	cmd.Stderr = logFile
	if err := cmd.Start(); err != nil {
		return fmt.Errorf("запуск DPI не удался (права администратора, антивирус или нет файла): %w", err)
	}
	time.Sleep(280 * time.Millisecond)
	if cmd.Process != nil && !verifyProcessAlive(cmd.Process.Pid) {
		_ = cmd.Process.Kill()
		_, _ = cmd.Process.Wait()
		return fmt.Errorf("процесс DPI сразу завершился — откройте лог: %s", logPath)
	}
	r.cmd = cmd
	r.activeTool = profile.Tool
	go r.reap(cmd)
	return nil
}

func (r *Runner) reap(cmd *exec.Cmd) {
	err := cmd.Wait()
	r.mu.Lock()
	defer r.mu.Unlock()
	if r.cmd == cmd {
		r.cmd = nil
		r.activeTool = ""
	}
	_ = err
}

// CleanupStrayProcesses kills stray winws.exe and winws2.exe processes,
// and stops/deletes WinDivert driver services (Monkey, WinDivert, WinDivert14).
func CleanupStrayProcesses() {
	runHidden("taskkill.exe", "/F", "/IM", "winws.exe", "/T")
	runHidden("taskkill.exe", "/F", "/IM", "winws2.exe", "/T")
	for _, svc := range []string{"Monkey", "WinDivert", "WinDivert14"} {
		runHidden("sc.exe", "stop", svc)
		runHidden("sc.exe", "delete", svc)
	}
}

// CleanupStrayProcesses delegates to package-level CleanupStrayProcesses.
func (r *Runner) CleanupStrayProcesses() {
	CleanupStrayProcesses()
}

func (r *Runner) Stop() error {
	r.mu.Lock()
	defer r.mu.Unlock()

	if r.cmd != nil && r.cmd.Process != nil {
		_ = r.cmd.Process.Kill()
		_, _ = r.cmd.Process.Wait()
		r.cmd = nil
	}
	r.activeTool = ""

	CleanupStrayProcesses()
	return nil
}

func runHidden(name string, args ...string) {
	cmd := exec.Command(name, args...)
	cmd.SysProcAttr = &syscall.SysProcAttr{
		HideWindow:    true,
		CreationFlags: 0x08000000, // CREATE_NO_WINDOW
	}
	_ = cmd.Run()
}

func (r *Runner) IsRunning() bool {
	r.mu.Lock()
	defer r.mu.Unlock()
	return r.cmd != nil && r.cmd.Process != nil
}

// ProcessAlive returns whether the DPI child process PID is still running (Windows).
func (r *Runner) ProcessAlive() bool {
	r.mu.Lock()
	defer r.mu.Unlock()
	if r.cmd == nil || r.cmd.Process == nil {
		return false
	}
	return verifyProcessAlive(r.cmd.Process.Pid)
}

func (r *Runner) LastLogPath() string {
	r.mu.Lock()
	defer r.mu.Unlock()
	return r.lastLog
}

func resolveToolBinary(tool string) (string, error) {
	exe, err := os.Executable()
	if err != nil {
		return "", err
	}
	baseDir := filepath.Dir(exe)
	candidates := []string{
		filepath.Join(baseDir, "dpi", tool+".exe"),
		filepath.Join(baseDir, tool+".exe"),
	}
	candidates = append(candidates, ancestorCandidates(baseDir, "third_party", "dpi", tool+".exe")...)
	for _, c := range candidates {
		info, err := os.Stat(c)
		if err == nil && !info.IsDir() {
			return c, nil
		}
	}
	return "", errors.New("dpi tool binary not found near app executable")
}
