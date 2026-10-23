package main

import (
	"fmt"

	"pulseshield/app/config"
	"pulseshield/app/dpi"
	"pulseshield/app/models"
)

// GetAvailableDpiLists возвращает все lists/*.txt рядом с Zapret и флаг включения из конфига.
func (a *App) GetAvailableDpiLists() ([]dpi.DpiList, error) {
	a.panelCacheMu.Lock()
	if a.dpiListsCacheReady {
		out := append([]dpi.DpiList(nil), a.dpiListsCached...)
		a.panelCacheMu.Unlock()
		return out, nil
	}
	a.panelCacheMu.Unlock()

	a.mu.Lock()
	if a.st == nil {
		a.mu.Unlock()
		return []dpi.DpiList{}, fmt.Errorf("app state is not initialized")
	}
	lists, err := dpi.GetAvailableDpiLists(a.st.Snapshot())
	a.mu.Unlock()

	if err != nil {
		return lists, err
	}

	a.panelCacheMu.Lock()
	a.dpiListsCached = append([]dpi.DpiList(nil), lists...)
	a.dpiListsCacheReady = true
	a.panelCacheMu.Unlock()

	return lists, nil
}

// GetZapretListsPathHint возвращает путь к каталогу lists для подсказки в UI (рядом с .exe).
func (a *App) GetZapretListsPathHint() string {
	return dpi.ZapretListsPathForHint()
}

// ToggleDpiList включает/выключает использование списка при патче пресета winws2.
// Если DPI сейчас работает, автоматически перезапускает его в фоне, чтобы обновить фильтры.
func (a *App) ToggleDpiList(name string, enabled bool) error {
	a.mu.Lock()
	if a.st == nil {
		a.mu.Unlock()
		return fmt.Errorf("app state is not initialized")
	}
	var terr error
	a.st.Update(func(cfg *models.AppConfig) {
		terr = dpi.ToggleDpiList(cfg, name, enabled)
	})
	if terr != nil {
		a.mu.Unlock()
		return terr
	}
	snap := a.st.Snapshot()
	err := config.Save(snap)
	if err == nil {
		a.invalidateRoutingDpiPanelCaches()
	}
	shouldRestartDPI := snap.DPIEnabled && a.dpiRunner != nil && a.dpiRunner.IsRunning()
	a.mu.Unlock()

	if shouldRestartDPI {
		go func() {
			_ = a.runApplySync(false, 0)
		}()
	}
	return err
}
