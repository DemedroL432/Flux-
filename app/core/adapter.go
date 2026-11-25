package core

import (
	"errors"
	"io"
	"os"
	"os/exec"
	"path/filepath"
	"sync"
	"syscall"
	"time"
)

type Adapter struct {
	mu     sync.Mutex
	cmd    *exec.Cmd
	binary string
	stdout io.Writer
	stderr io.Writer
}

func NewAdapter(binary string) *Adapter {
	return &Adapter{binary: binary}
}

func (a *Adapter) SetOutput(stdout, stderr io.Writer) {
	a.mu.Lock()
	defer a.mu.Unlock()
	a.stdout = stdout
	a.stderr = stderr
}

func (a *Adapter) Start(args ...string) error {
	a.mu.Lock()
	defer a.mu.Unlock()
	if a.cmd != nil && a.cmd.Process != nil {
		return nil
	}
	if a.binary == "" {
		return errors.New("core binary path is empty")
	}
	cmd := exec.Command(a.binary, args...)
	cmd.Dir = filepath.Dir(a.binary)
	// sing-box 1.12+: без domain_resolver — FATAL; временный режим совместимости при миграции конфига.
	cmd.Env = append(os.Environ(), "ENABLE_DEPRECATED_MISSING_DOMAIN_RESOLVER=true")
	cmd.SysProcAttr = &syscall.SysProcAttr{
		HideWindow:    true,
		CreationFlags: 0x08000000, // CREATE_NO_WINDOW
	}
	if a.stdout != nil {
		cmd.Stdout = a.stdout
	}
	if a.stderr != nil {
		cmd.Stderr = a.stderr
	}
	if err := cmd.Start(); err != nil {
		return err
	}
	a.cmd = cmd
	return nil
}

// Stop завершает процесс sing-box: SIGINT → ожидание до 3s → Kill.
// afterSignal вызывается сразу после отправки сигнала (до ожидания Wait),
// чтобы снять системный прокси, не дожидаясь завершения процесса.
func (a *Adapter) Stop(afterSignal func()) error {
	a.mu.Lock()
	if a.cmd == nil || a.cmd.Process == nil {
		a.mu.Unlock()
		return nil
	}
	cmd := a.cmd
	a.cmd = nil
	a.mu.Unlock()

	errInt := cmd.Process.Signal(os.Interrupt)
	var errTerm error
	if errInt != nil {
		// Fallback для процессов без поддержки SIGINT (Windows / no-console).
		errTerm = cmd.Process.Signal(syscall.SIGTERM)
	}

	if afterSignal != nil {
		afterSignal()
	}

	waitDone := make(chan struct{})
	go func() {
		// cmd.Wait() — а не cmd.Process.Wait() — корректно закрывает stdout/stderr pipes
		// и освобождает все ресурсы процесса. После Kill() она завершится в разумное время.
		_ = cmd.Wait()
		close(waitDone)
	}()

	if errInt != nil && errTerm != nil {
		// На Windows отправка POSIX-сигналов процессу без консоли не поддерживается — завершаем сразу без 3s ожидания.
		_ = cmd.Process.Kill()
		select {
		case <-waitDone:
		case <-time.After(2 * time.Second):
		}
		return nil
	}

	select {
	case <-waitDone:
		return nil
	case <-time.After(3 * time.Second):
	}

	_ = cmd.Process.Kill()
	// Ждём завершения горутины с cmd.Wait() — после Kill() она обязана завершиться.
	// Таймаут 2s — защита от зависания ядра ОС.
	select {
	case <-waitDone:
	case <-time.After(2 * time.Second):
		// Горутина зависла (аномалия ядра ОС) — не блокируем дальнейший shutdown.
	}
	return nil
}

func (a *Adapter) IsRunning() bool {
	a.mu.Lock()
	defer a.mu.Unlock()
	return a.cmd != nil && a.cmd.Process != nil
}
