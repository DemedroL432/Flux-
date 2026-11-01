package dpi

import (
	"bytes"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
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
	var outBuf bytes.Buffer
	cmd.Stdout = &outBuf
	cmd.Stderr = &outBuf
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
	errText := strings.TrimSpace(outBuf.String())
	if errText != "" {
		lines := strings.Split(errText, "\n")
		var lastLine string
		for i := len(lines) - 1; i >= 0; i-- {
			l := strings.TrimSpace(lines[i])
			if l != "" {
				lastLine = l
				break
			}
		}
		if len(lastLine) > 120 {
			lastLine = lastLine[:120] + "..."
		}
		return false, fmt.Sprintf("процесс завершился: %s", lastLine), nil
	}
	return false, "процесс завершился — смотрите лог DPI, антивирус или права администратора", nil
}
