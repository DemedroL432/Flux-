package dpi

import (
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"syscall"
	"time"
)

// TestWinws2Preset запускает winws2 с @configRel (путь относительно каталога zapret, напр. presets/Gamingnew.txt),
// ждёт observe и проверяет, что процесс не завершился сразу (косвенный признак валидного пресета и окружения).
func TestWinws2Preset(configRel string, observe time.Duration) (ok bool, detail string, err error) {
	bin, dir, err := ResolveZapretWinws2()
	if err != nil {
		return false, "", err
	}
	relSlash := filepath.ToSlash(filepath.Clean(configRel))
	full := filepath.Join(dir, filepath.FromSlash(relSlash))
	if _, statErr := os.Stat(full); statErr != nil {
		return false, fmt.Sprintf("нет файла %s", relSlash), nil
	}
	cmd := exec.Command(bin, "@"+relSlash)
	cmd.Dir = dir
	cmd.SysProcAttr = &syscall.SysProcAttr{
		HideWindow:    true,
		CreationFlags: 0x08000000,
	}
	if err := cmd.Start(); err != nil {
		return false, fmt.Sprintf("не удалось запустить: %v", err), nil
	}
	pid := cmd.Process.Pid
	time.Sleep(observe)
	alive := verifyProcessAlive(pid)
	_ = cmd.Process.Kill()
	_, _ = cmd.Process.Wait()
	runHidden("taskkill.exe", "/F", "/IM", "winws2.exe", "/T")
	runHidden("taskkill.exe", "/F", "/IM", "winws.exe", "/T")
	if alive {
		return true, fmt.Sprintf("процесс удерживается ~%.0f с — пресет запускается", observe.Seconds()), nil
	}
	return false, "процесс завершился — смотрите лог DPI, антивирус или права администратора", nil
}
