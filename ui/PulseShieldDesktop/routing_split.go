package main

import (
	"fmt"
	"path/filepath"

	"pulseshield/app/config"
	"pulseshield/app/models"
)

// RoutingSplitEntry — строка для UI «прокси / напрямую».
type RoutingSplitEntry struct {
	Pattern string `json:"pattern"`
	Target  string `json:"target"`
	System  bool   `json:"system"`
	Label   string `json:"label,omitempty"`
	Source  string `json:"source"` // user | profile | library
}

func flattenRouteRule(r models.RouteRule, source string) []RoutingSplitEntry {
	tgt := string(r.Action)
	label := r.Name
	system := source != "user"
	var out []RoutingSplitEntry
	for _, d := range r.Domains {
		out = append(out, RoutingSplitEntry{Pattern: d, Target: tgt, System: system, Label: label, Source: source})
	}
	for _, s := range r.DomainSuffixes {
		out = append(out, RoutingSplitEntry{Pattern: "." + s, Target: tgt, System: system, Label: label, Source: source})
	}
	for _, c := range r.CIDRs {
		out = append(out, RoutingSplitEntry{Pattern: c, Target: tgt, System: system, Label: label, Source: source})
	}
	for _, p := range r.Processes {
		out = append(out, RoutingSplitEntry{Pattern: p, Target: tgt, System: system, Label: label, Source: source})
	}
	return out
}

// GetRoutingSplitView объединяет пользовательские правила, профиль и библиотеку (если не StrictRoutingMode).
func (a *App) GetRoutingSplitView() ([]RoutingSplitEntry, error) {
	a.panelCacheMu.Lock()
	if a.routingSplitCacheReady {
		out := append([]RoutingSplitEntry(nil), a.routingSplitCached...)
		a.panelCacheMu.Unlock()
		return out, nil
	}
	a.panelCacheMu.Unlock()

	o, err := a.routingSplitViewLocked()
	if err != nil {
		return o, err
	}
	a.panelCacheMu.Lock()
	a.routingSplitCached = append([]RoutingSplitEntry(nil), o...)
	a.routingSplitCacheReady = true
	a.panelCacheMu.Unlock()
	return o, nil
}

func (a *App) routingSplitViewLocked() ([]RoutingSplitEntry, error) {
	a.mu.Lock()
	defer a.mu.Unlock()
	if a.st == nil {
		return nil, fmt.Errorf("app state is not initialized")
	}
	snap := a.st.Snapshot()

	zdir, ok := config.ResolveZapretDirNearExe()
	listsDir := ""
	if ok {
		listsDir = filepath.Join(zdir, "lists")
	}

	var out []RoutingSplitEntry
	for _, r := range snap.UserRoutingRules {
		out = append(out, flattenRouteRule(r, "user")...)
	}
	if rp, ok := snap.ActiveRoutingProfile(); ok {
		for _, r := range rp.Rules {
			out = append(out, flattenRouteRule(r, "profile")...)
		}
	}
	if !snap.StrictRoutingMode {
		builtins := config.DefaultRoutingLibraryRules(listsDir)
		for _, r := range builtins {
			out = append(out, flattenRouteRule(r, "library")...)
		}
	}
	return out, nil
}

// AddUserRoutingRule добавляет пользовательское правило (не системное).
func (a *App) AddUserRoutingRule(pattern string, target string) error {
	a.mu.Lock()
	defer a.mu.Unlock()
	if a.st == nil {
		return fmt.Errorf("app state is not initialized")
	}
	var verr error
	a.st.Update(func(cfg *models.AppConfig) {
		verr = config.AddUserRoutingRule(cfg, pattern, target)
	})
	if verr != nil {
		return verr
	}
	err := config.Save(a.st.Snapshot())
	if err == nil {
		a.invalidateRoutingDpiPanelCaches()
	}
	return err
}

// RemoveUserRoutingRule удаляет пользовательское правило по шаблону (как в колонке Pattern).
func (a *App) RemoveUserRoutingRule(pattern string) error {
	a.mu.Lock()
	defer a.mu.Unlock()
	if a.st == nil {
		return fmt.Errorf("app state is not initialized")
	}
	var verr error
	a.st.Update(func(cfg *models.AppConfig) {
		verr = config.RemoveUserRoutingRule(cfg, pattern)
	})
	if verr != nil {
		return verr
	}
	err := config.Save(a.st.Snapshot())
	if err == nil {
		a.invalidateRoutingDpiPanelCaches()
	}
	return err
}
