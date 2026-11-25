package main

import (
	"archive/zip"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"syscall"
	"time"

	"github.com/wailsapp/wails/v2/pkg/runtime"
	"pulseshield/app/bootlog"
)

const portableExeName = "Flux.exe"

const (
	updateProgressEvent  = "update-progress"
	updateAvailableEvent = "update-available"
	updateErrorEvent     = "update-error"
)

func allowedGitHubDownloadURL(raw string) error {
	u, err := url.Parse(raw)
	if err != nil || u.Scheme != "https" {
		return fmt.Errorf("нужен https URL")
	}
	h := strings.ToLower(strings.TrimSpace(u.Hostname()))
	if h == "" {
		return fmt.Errorf("пустой хост")
	}
	switch h {
	case "github.com", "objects.githubusercontent.com", "codeload.github.com":
		return nil
	default:
		if strings.HasSuffix(h, ".github.com") || strings.HasSuffix(h, ".githubusercontent.com") {
			return nil
		}
	}
	return fmt.Errorf("разрешены только загрузки с GitHub")
}

func (a *App) emitUpdateProgress(phase string, pct int) {
	if a == nil || a.ctx == nil {
		return
	}
	b, err := json.Marshal(map[string]interface{}{"phase": phase, "pct": pct})
	if err != nil {
		return
	}
	runtime.EventsEmit(a.ctx, updateProgressEvent, string(b))
}

func unzipSafe(zipPath, destDir string) error {
	r, err := zip.OpenReader(zipPath)
	if err != nil {
		return err
	}
	defer r.Close()

	destClean := filepath.Clean(destDir)
	for _, f := range r.File {
		p := filepath.Join(destDir, filepath.FromSlash(f.Name))
		clean := filepath.Clean(p)
		rel, err := filepath.Rel(destClean, clean)
		if err != nil || rel == ".." || strings.HasPrefix(rel, ".."+string(os.PathSeparator)) {
			return fmt.Errorf("некорректный путь в архиве: %s", f.Name)
		}
		if f.FileInfo().IsDir() {
			if err := os.MkdirAll(clean, 0o755); err != nil {
				return err
			}
			continue
		}
		if err := os.MkdirAll(filepath.Dir(clean), 0o755); err != nil {
			return err
		}
		rc, err := f.Open()
		if err != nil {
			return err
		}
		out, err := os.OpenFile(clean, os.O_WRONLY|os.O_CREATE|os.O_TRUNC, f.Mode())
		if err != nil {
			_ = rc.Close()
			return err
		}
		_, copyErr := io.Copy(out, rc)
		_ = out.Close()
		_ = rc.Close()
		if copyErr != nil {
			return copyErr
		}
	}
	return nil
}

func findPortablePayloadRoot(root string) (string, error) {
	candidates := []string{portableExeName, "PulseShield.exe"}
	for _, exeName := range candidates {
		if st, err := os.Stat(filepath.Join(root, exeName)); err == nil && !st.IsDir() {
			return root, nil
		}
	}
	entries, err := os.ReadDir(root)
	if err != nil {
		return "", err
	}
	for _, e := range entries {
		if !e.IsDir() {
			continue
		}
		sub := filepath.Join(root, e.Name())
		for _, exeName := range candidates {
			if st, err := os.Stat(filepath.Join(sub, exeName)); err == nil && !st.IsDir() {
				return sub, nil
			}
		}
	}
	return "", fmt.Errorf("в архиве нет %s (ожидается portable-сборка)", portableExeName)
}

func isNSISInstallerURL(raw string) bool {
	u, err := url.Parse(strings.TrimSpace(raw))
	if err != nil || u.Path == "" {
		return false
	}
	if !strings.HasSuffix(strings.ToLower(u.Path), ".exe") {
		return false
	}
	base := strings.ToLower(filepath.Base(u.Path))
	return strings.Contains(base, "setup") || strings.Contains(base, "installer")
}

func isInstalledProgramFilesPath(exePath string) bool {
	p := strings.ToLower(filepath.Clean(strings.TrimSpace(exePath)))
	if p == "" || p == "." {
		return false
	}
	for _, env := range []string{"ProgramFiles", "ProgramFiles(x86)"} {
		if base := strings.ToLower(filepath.Clean(strings.TrimSpace(os.Getenv(env)))); base != "" && base != "." {
			if p == base || strings.HasPrefix(p, base+string(os.PathSeparator)) {
				return true
			}
		}
	}
	return strings.Contains(p, `\program files\`) || strings.Contains(p, `\program files (x86)\`) ||
		strings.Contains(p, `/program files/`) || strings.Contains(p, `/program files (x86)/`)
}

func pickAutoUpdateURL(info UpdateInfo, exePath string) string {
	zipURL := strings.TrimSpace(info.DownloadURL)
	instURL := strings.TrimSpace(info.InstallerDownloadURL)
	if isInstalledProgramFilesPath(exePath) {
		if instURL != "" {
			return instURL
		}
		return zipURL
	}
	if zipURL != "" {
		return zipURL
	}
	return instURL
}

func cleanupOldExtractDirs(base string) {
	entries, err := os.ReadDir(base)
	if err != nil {
		return
	}
	for _, e := range entries {
		if e.IsDir() && strings.HasPrefix(e.Name(), "extract-") {
			_ = os.RemoveAll(filepath.Join(base, e.Name()))
		}
	}
}

// downloadToFileProgress скачивает файл; onProgress вызывается не чаще ~200 ms (pct -1 если размер неизвестен).
func downloadToFileProgress(ctx context.Context, rawURL, dest string, onProgress func(pct int)) error {
	if err := allowedGitHubDownloadURL(rawURL); err != nil {
		return err
	}
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, rawURL, nil)
	if err != nil {
		return err
	}
	req.Header.Set("User-Agent", "PulseShield-Update/1.0")
	req.Header.Set("Accept", "*/*")

	client := &http.Client{Transport: http.DefaultTransport}
	resp, err := client.Do(req)
	if err != nil {
		return err
	}
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusOK {
		b, _ := io.ReadAll(io.LimitReader(resp.Body, 512))
		return fmt.Errorf("скачивание: %s — %s", resp.Status, strings.TrimSpace(string(b)))
	}

	total := resp.ContentLength
	tmp := dest + ".part"
	f, err := os.Create(tmp)
	if err != nil {
		return err
	}

	buf := make([]byte, 256*1024)
	var written int64
	lastEmit := time.Time{}
	lastPct := -2

	flushProgress := func(force bool) {
		if onProgress == nil {
			return
		}
		pct := -1
		if total > 0 {
			pct = int(written * 100 / total)
			if pct > 99 {
				pct = 99
			}
		}
		now := time.Now()
		if !force && now.Sub(lastEmit) < 200*time.Millisecond {
			return
		}
		if !force && total > 0 && pct == lastPct && pct < 99 {
			return
		}
		lastEmit = now
		lastPct = pct
		onProgress(pct)
	}

	for {
		n, rerr := resp.Body.Read(buf)
		if n > 0 {
			if _, werr := f.Write(buf[:n]); werr != nil {
				_ = f.Close()
				_ = os.Remove(tmp)
				return werr
			}
			written += int64(n)
			flushProgress(false)
		}
		if rerr == io.EOF {
			break
		}
		if rerr != nil {
			_ = f.Close()
			_ = os.Remove(tmp)
			return rerr
		}
	}
	if err := f.Close(); err != nil {
		_ = os.Remove(tmp)
		return err
	}
	if onProgress != nil {
		onProgress(100)
	}
	if err := os.Rename(tmp, dest); err != nil {
		_ = os.Remove(tmp)
		return err
	}
	return nil
}

const psApplyUpdate = `param(
  [Parameter(Mandatory=$true)][string]$Source,
  [Parameter(Mandatory=$true)][string]$Target,
  [Parameter(Mandatory=$true)][string]$ExeName,
  [Parameter(Mandatory=$true)][string]$ScriptPath
)
$ErrorActionPreference = 'SilentlyContinue'
Start-Sleep -Seconds 3
Stop-Process -Name "Flux","PulseShield","PulseShieldDesktop","sing-box","winws2" -Force -ErrorAction SilentlyContinue
Start-Sleep -Seconds 1
& robocopy $Source $Target /E /R:5 /W:1 /NFL /NDL /NJH /NJS
$launch = Join-Path $Target $ExeName
if (!(Test-Path -LiteralPath $launch)) {
  $launch = Join-Path $Target "Flux.exe"
}
if (Test-Path -LiteralPath $launch) {
  Start-Process -FilePath $launch -WorkingDirectory $Target
}
Start-Sleep -Seconds 2
Remove-Item -LiteralPath $ScriptPath -Force -ErrorAction SilentlyContinue
`

const psApplyInstaller = `param(
  [Parameter(Mandatory=$true)][string]$InstallerPath,
  [Parameter(Mandatory=$true)][string]$TargetDir,
  [Parameter(Mandatory=$true)][string]$ExeName,
  [Parameter(Mandatory=$true)][string]$ScriptPath
)
$ErrorActionPreference = 'SilentlyContinue'
Start-Sleep -Seconds 2
Stop-Process -Name "Flux","PulseShield","PulseShieldDesktop","sing-box","winws2" -Force -ErrorAction SilentlyContinue
Start-Sleep -Seconds 1
Start-Process -FilePath $InstallerPath -ArgumentList "/S" -Wait
$candidates = @(
  (Join-Path "C:\Program Files\Flux" "Flux.exe"),
  (Join-Path $TargetDir $ExeName),
  (Join-Path $TargetDir "Flux.exe")
)
foreach ($c in $candidates) {
  if (Test-Path -LiteralPath $c) {
    Start-Process -FilePath $c -WorkingDirectory (Split-Path -Parent $c)
    break
  }
}
Start-Sleep -Seconds 2
Remove-Item -LiteralPath $ScriptPath -Force -ErrorAction SilentlyContinue
`

func (a *App) applyZipUpdate(ctx context.Context, raw string) error {
	appDir, restartName, base, err := a.resolveAppDirs()
	if err != nil {
		return err
	}
	zipPath := filepath.Join(base, "release-update.zip")
	extractDir := filepath.Join(base, "extract-"+time.Now().Format("20060102-150405"))

	if err := downloadToFileProgress(ctx, raw, zipPath, func(pct int) {
		a.emitUpdateProgress("download", pct)
	}); err != nil {
		return fmt.Errorf("скачивание: %w", err)
	}

	a.emitUpdateProgress("unpack", -1)
	cleanupOldExtractDirs(base)
	if err := os.MkdirAll(extractDir, 0o755); err != nil {
		return err
	}
	if err := unzipSafe(zipPath, extractDir); err != nil {
		return fmt.Errorf("распаковка: %w", err)
	}
	payloadRoot, err := findPortablePayloadRoot(extractDir)
	if err != nil {
		return err
	}

	scriptPath := filepath.Join(base, "apply-update.ps1")
	if err := os.WriteFile(scriptPath, []byte(psApplyUpdate), 0o644); err != nil {
		return err
	}

	if err := a.StopAll(); err != nil {
		return fmt.Errorf("остановка сервисов перед обновлением: %w", err)
	}

	psExe, err := exec.LookPath("powershell.exe")
	if err != nil {
		return fmt.Errorf("powershell.exe не найден")
	}
	cmd := exec.Command(psExe,
		"-NoProfile", "-ExecutionPolicy", "Bypass", "-WindowStyle", "Hidden",
		"-File", scriptPath,
		"-Source", payloadRoot,
		"-Target", appDir,
		"-ExeName", restartName,
		"-ScriptPath", scriptPath,
	)
	cmd.SysProcAttr = &syscall.SysProcAttr{
		HideWindow:    true,
		CreationFlags: 0x08000000, // CREATE_NO_WINDOW
	}
	if err := cmd.Start(); err != nil {
		return fmt.Errorf("запуск сценария обновления: %w", err)
	}

	a.quitForUpdate()
	return nil
}

func (a *App) resolveAppDirs() (appDir, restartName, updatesBase string, err error) {
	restartName = portableExeName
	exePath, err := os.Executable()
	if err != nil {
		return "", "", "", err
	}
	exePath, err = filepath.EvalSymlinks(exePath)
	if err != nil {
		return "", "", "", err
	}
	if baseExe := filepath.Base(exePath); strings.HasSuffix(strings.ToLower(baseExe), ".exe") {
		restartName = baseExe
	}
	appDir = filepath.Dir(exePath)
	base := filepath.Join(os.Getenv("LOCALAPPDATA"), "Flux", "updates")
	if err := os.MkdirAll(base, 0o755); err != nil {
		return "", "", "", err
	}
	return appDir, restartName, base, nil
}

func (a *App) applyInstallerUpdate(ctx context.Context, raw string) error {
	if !isNSISInstallerURL(raw) {
		return fmt.Errorf("URL не похож на NSIS-установщик (нужен .exe с «setup» или «installer» в имени)")
	}
	appDir, restartName, base, err := a.resolveAppDirs()
	if err != nil {
		return err
	}
	instPath := filepath.Join(base, "Flux-Update-Setup.exe")

	if err := downloadToFileProgress(ctx, raw, instPath, func(pct int) {
		a.emitUpdateProgress("download", pct)
	}); err != nil {
		return fmt.Errorf("скачивание: %w", err)
	}

	a.emitUpdateProgress("install", -1)
	if err := a.StopAll(); err != nil {
		return fmt.Errorf("остановка сервисов: %w", err)
	}

	scriptPath := filepath.Join(base, "apply-installer.ps1")
	if err := os.WriteFile(scriptPath, []byte(psApplyInstaller), 0o644); err != nil {
		return err
	}

	psExe, err := exec.LookPath("powershell.exe")
	if err != nil {
		return fmt.Errorf("powershell.exe не найден")
	}
	cmd := exec.Command(psExe,
		"-NoProfile", "-ExecutionPolicy", "Bypass", "-WindowStyle", "Hidden",
		"-File", scriptPath,
		"-InstallerPath", instPath,
		"-TargetDir", appDir,
		"-ExeName", restartName,
		"-ScriptPath", scriptPath,
	)
	cmd.SysProcAttr = &syscall.SysProcAttr{
		HideWindow:    true,
		CreationFlags: 0x08000000,
	}
	if err := cmd.Start(); err != nil {
		return fmt.Errorf("запуск установщика: %w", err)
	}

	a.quitForUpdate()
	return nil
}

func (a *App) quitForUpdate() {
	a.mu.Lock()
	a.forceQuit = true
	a.mu.Unlock()
	time.Sleep(150 * time.Millisecond)
	runtime.Quit(a.ctx)
}

// ApplyAutoUpdate: для .zip — portable (robocopy + перезапуск PulseShield.exe);
// для NSIS .exe (имя содержит setup/installer) — тихая установка /S, затем выход приложения.
func (a *App) ApplyAutoUpdate(downloadURL string) error {
	if !a.autoUpdateBusy.CompareAndSwap(false, true) {
		return fmt.Errorf("обновление уже выполняется")
	}
	defer a.autoUpdateBusy.Store(false)

	raw := strings.TrimSpace(downloadURL)
	if raw == "" {
		return fmt.Errorf("пустая ссылка на загрузку")
	}
	if err := allowedGitHubDownloadURL(raw); err != nil {
		return err
	}

	ctx, cancel := context.WithTimeout(context.Background(), 25*time.Minute)
	defer cancel()

	lu := strings.ToLower(raw)
	switch {
	case strings.HasSuffix(lu, ".zip"):
		return a.applyZipUpdate(ctx, raw)
	case strings.HasSuffix(lu, ".exe") && isNSISInstallerURL(raw):
		return a.applyInstallerUpdate(ctx, raw)
	default:
		return fmt.Errorf("неподдерживаемый тип: используйте .zip portable или NSIS Setup.exe с релиза")
	}
}

// ApplyLatestUpdate проверяет последний релиз на GitHub, выбирает подходящий пакет (portable .zip или NSIS Setup.exe) и запускает установку.
func (a *App) ApplyLatestUpdate() error {
	info := a.GetUpdateInfo()
	if !info.OK {
		return fmt.Errorf("%s", info.Message)
	}
	if info.UpToDate {
		return fmt.Errorf("установлена актуальная версия (%s)", info.CurrentVersion)
	}
	exePath, _ := os.Executable()
	if eval, err := filepath.EvalSymlinks(exePath); err == nil {
		exePath = eval
	}
	dlURL := pickAutoUpdateURL(info, exePath)
	if dlURL == "" {
		return fmt.Errorf("в релизе %s не найден файл обновления", info.LatestTag)
	}
	return a.ApplyAutoUpdate(dlURL)
}

func (a *App) runAutoUpdateLoop() {
	time.Sleep(10 * time.Second)
	var lastAttemptedTag string

	checkAndApplyOnce := func() {
		defer func() {
			if r := recover(); r != nil {
				bootlog.Append(fmt.Sprintf("auto-update panic: %v", r))
			}
		}()
		if a == nil || a.st == nil {
			return
		}
		if a.st.Snapshot().AutoUpdateDisabled || a.autoUpdateBusy.Load() {
			return
		}
		ver := strings.TrimSpace(a.GetAppVersion())
		if ver == "" || strings.EqualFold(ver, "dev") {
			return
		}
		info := a.GetUpdateInfo()
		if !info.OK || info.UpToDate || info.IsPrerelease || info.LatestTag == "" || info.LatestTag == lastAttemptedTag {
			return
		}
		exePath, _ := os.Executable()
		if eval, err := filepath.EvalSymlinks(exePath); err == nil {
			exePath = eval
		}
		dlURL := pickAutoUpdateURL(info, exePath)
		if dlURL == "" {
			return
		}
		lastAttemptedTag = info.LatestTag
		if a.ctx != nil {
			if b, err := json.Marshal(info); err == nil {
				runtime.EventsEmit(a.ctx, updateAvailableEvent, string(b))
			}
		}
		bootlog.Append(fmt.Sprintf("auto-update: starting %s -> %s (%s)", info.CurrentVersion, info.LatestTag, dlURL))
		if err := a.ApplyAutoUpdate(dlURL); err != nil {
			bootlog.Append(fmt.Sprintf("auto-update failed (%s): %v", info.LatestTag, err))
			if a.ctx != nil {
				runtime.EventsEmit(a.ctx, updateErrorEvent, err.Error())
			}
		}
	}

	checkAndApplyOnce()
	ticker := time.NewTicker(4 * time.Hour)
	defer ticker.Stop()
	for range ticker.C {
		checkAndApplyOnce()
	}
}
