package dpi

import (
	"os/exec"
	"sync"
	"syscall"
	"testing"
)

func TestRunner_StopWhenCmdIsNil(t *testing.T) {
	r := &Runner{}
	if r.IsRunning() {
		t.Fatal("expected IsRunning to be false for new Runner")
	}
	if r.ProcessAlive() {
		t.Fatal("expected ProcessAlive to be false for new Runner")
	}

	// Calling Stop() when r.cmd is nil must not panic and must succeed (calling CleanupStrayProcesses).
	if err := r.Stop(); err != nil {
		t.Fatalf("Stop() with nil cmd returned error: %v", err)
	}

	if r.activeTool != "" {
		t.Fatalf("expected activeTool to be empty, got %q", r.activeTool)
	}
}

func TestRunner_ConcurrentStop(t *testing.T) {
	r := &Runner{}

	var wg sync.WaitGroup
	for i := 0; i < 5; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			if err := r.Stop(); err != nil {
				t.Errorf("concurrent Stop() failed: %v", err)
			}
		}()
	}
	wg.Wait()
}

func TestRunner_StopActiveProcess(t *testing.T) {
	cmd := exec.Command("cmd.exe", "/c", "pause")
	cmd.SysProcAttr = &syscall.SysProcAttr{
		HideWindow:    true,
		CreationFlags: 0x08000000,
	}
	if err := cmd.Start(); err != nil {
		t.Fatalf("failed to start dummy process: %v", err)
	}

	r := &Runner{
		cmd:        cmd,
		activeTool: "test_tool",
	}

	if !r.IsRunning() {
		t.Fatal("expected IsRunning to be true")
	}

	if err := r.Stop(); err != nil {
		t.Fatalf("Stop() failed: %v", err)
	}

	if r.IsRunning() {
		t.Fatal("expected IsRunning to be false after Stop()")
	}
	if r.cmd != nil {
		t.Fatal("expected r.cmd to be nil after Stop()")
	}
	if r.activeTool != "" {
		t.Fatalf("expected activeTool to be empty, got %q", r.activeTool)
	}

	// Double Stop() on now-nil cmd should succeed without issue
	if err := r.Stop(); err != nil {
		t.Fatalf("second Stop() failed: %v", err)
	}
}

func TestCleanupStrayProcesses_Direct(t *testing.T) {
	// Directly verify CleanupStrayProcesses executes without panic
	CleanupStrayProcesses()
}
