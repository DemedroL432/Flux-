//go:build windows

package main

import (
	"os"
	"sync"

	toast "git.sr.ht/~jackmordaunt/go-toast/v2"
)

var toastSetup sync.Once

func initWinToast() {
	toastSetup.Do(func() {
		exe, _ := os.Executable()
		_ = toast.SetAppData(toast.AppData{
			AppID:         "Flux",
			ActivationExe: exe,
		})
	})
}

func notifyTrayHidden() {
	initWinToast()
	n := toast.Notification{
		AppID: "Flux",
		Title: "Flux",
		Body:  "Окно скрыто — приложение работает в трее. Откройте снова через меню иконки у часов.",
		Audio: toast.Default,
	}
	_ = n.Push()
}
