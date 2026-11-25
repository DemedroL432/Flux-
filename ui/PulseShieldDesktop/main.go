package main

import (
	"context"
	"embed"
	"fmt"
	"os"
	"os/exec"
	"strings"
	"syscall"
	"time"

	"github.com/wailsapp/wails/v2"
	"github.com/wailsapp/wails/v2/pkg/options"
	"github.com/wailsapp/wails/v2/pkg/options/assetserver"
	winopt "github.com/wailsapp/wails/v2/pkg/options/windows"
	"github.com/wailsapp/wails/v2/pkg/runtime"
	"golang.org/x/sys/windows"
)

// Version is set at link time: -ldflags "-X main.Version=1.2.3" (release). Default matches desktop line 0.7.x.
var Version = "0.7.8"

//go:embed all:frontend/dist
var assets embed.FS

func main() {
	ok, err := ensureElevatedOnStartup()
	if err != nil {
		println("Error:", err.Error())
		return
	}
	if !ok {
		return
	}

	initWinToast()

	app := NewApp(Version)
	startSessionEndNotifier(app)

	runErr := wails.Run(&options.App{
		Title:         "Flux",
		Width:         1280,
		Height:        720,
		MinWidth:      960,
		MinHeight:     640,
		DisableResize: false,
		Frameless:     true,
		StartHidden:   hasStartHiddenArg(os.Args[1:]),
		AssetServer: &assetserver.Options{
			Assets: assets,
		},
		BackgroundColour: &options.RGBA{R: 14, G: 16, B: 19, A: 1},
		Windows: &winopt.Options{
			Theme:        winopt.Dark,
			BackdropType: winopt.Mica,
		},
		OnStartup: app.startup,
		// Нельзя вызывать Hide синхронно из этого колбэка — WebView2 зависает/падает.
		// Откладываем скрытие окна после возврата из обработчика закрытия.
		// При forceQuit (крестик / «Выход» в трее) возвращаем false — это настоящий выход.
		OnBeforeClose: func(ctx context.Context) (prevent bool) {
			if app.ShouldExitOnClose() {
				return false
			}
			go func(c context.Context) {
				time.Sleep(80 * time.Millisecond)
				runtime.WindowHide(c)
				notifyTrayHidden()
			}(ctx)
			return true
		},
		OnShutdown: func(ctx context.Context) {
			app.ShutdownCleanup()
		},
		Bind: []interface{}{
			app,
		},
	})

	if runErr != nil {
		println("Error:", runErr.Error())
	}
}

func ensureElevatedOnStartup() (bool, error) {
	var token windows.Token
	if err := windows.OpenProcessToken(windows.CurrentProcess(), windows.TOKEN_QUERY, &token); err != nil {
		return false, err
	}
	isElevated := token.IsElevated()
	_ = token.Close()
	if isElevated {
		return true, nil
	}

	exePath, err := os.Executable()
	if err != nil {
		return false, err
	}
	cmd := fmt.Sprintf("Start-Process -FilePath '%s' -Verb RunAs", psQuote(exePath))
	if len(os.Args) > 1 {
		cmd += " -ArgumentList @(" + strings.Join(quotePSArgs(os.Args[1:]), ", ") + ")"
	}
	ps := exec.Command("powershell.exe", "-NoProfile", "-ExecutionPolicy", "Bypass", "-WindowStyle", "Hidden", "-Command", cmd)
	ps.SysProcAttr = &syscall.SysProcAttr{
		HideWindow:    true,
		CreationFlags: 0x08000000, // CREATE_NO_WINDOW
	}
	if err := ps.Run(); err != nil {
		return false, err
	}
	return false, nil
}

func quotePSArgs(args []string) []string {
	out := make([]string, 0, len(args))
	for _, a := range args {
		out = append(out, "'"+psQuote(a)+"'")
	}
	return out
}

func psQuote(s string) string {
	return strings.ReplaceAll(s, "'", "''")
}

func hasStartHiddenArg(args []string) bool {
	for _, a := range args {
		switch strings.ToLower(strings.TrimSpace(a)) {
		case "--autostart", "--minimized":
			return true
		}
	}
	return false
}
