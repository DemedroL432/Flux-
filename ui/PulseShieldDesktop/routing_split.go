package main

import (
	"fmt"
	"net"
	"path/filepath"
	"strings"

	"pulseshield/app/config"
	"pulseshield/app/dpi"
	"pulseshield/app/models"
)

// RoutingSplitEntry — строка для UI «прокси / напрямую».
type RoutingSplitEntry struct {
	Pattern string `json:"pattern"`
	Target  string `json:"target"`
	System  bool   `json:"system"`
	Label   string `json:"label,omitempty"`
	Source  string `json:"source"` // user | profile | library
	// UserIndex — позиция в UserRoutingRules (только для source=user).
	// Нужна, чтобы удалять конкретное правило, когда на один домен их несколько.
	UserIndex int `json:"userIndex"`
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
	for i, r := range snap.UserRoutingRules {
		entries := flattenRouteRule(r, "user")
		for k := range entries {
			entries[k].UserIndex = i
		}
		out = append(out, entries...)
	}
	for _, r := range config.EffectiveProfileRules(snap) {
		out = append(out, flattenRouteRule(r, "profile")...)
	}
	resolved := models.ResolveRoutingProfileAlias(snap.RoutingProfile)
	isBuiltinMode := resolved == models.RoutingModeManual || resolved == models.RoutingModeFullProxy || resolved == models.RoutingModeGlobal
	if !snap.StrictRoutingMode && !isBuiltinMode {
		builtins := config.DefaultRoutingLibraryRules(listsDir)
		for _, r := range builtins {
			out = append(out, flattenRouteRule(r, "library")...)
		}
	}
	return out, nil
}

// AddUserRoutingRule добавляет пользовательское правило (не системное).
// Если цель — dpi-bypass, домен сразу пишется в custom-dpi.txt, иначе
// правило повлияет только на sing-box, а winws2 про него не узнает.
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
	snap := a.st.Snapshot()
	if err := config.Save(snap); err != nil {
		return err
	}
	a.invalidateRoutingDpiPanelCaches()

	// Список DPI пересобираем из всех правил — это снимает рассинхрон,
	// если файл правили вручную или версия обновилась.
	if err := dpi.SetCustomDpiEntries(config.CustomDpiEntriesForRules(&snap)); err != nil {
		return fmt.Errorf("правило добавлено, но список custom-dpi.txt не обновлён: %w", err)
	}
	a.invalidateRoutingDpiPanelCaches()
	return nil
}

// RemoveUserRoutingRuleAt удаляет пользовательское правило по индексу.
// Индекс берётся из GetRoutingSplitView (поле userIndex) — так корректно удаляется
// одно из нескольких правил на один и тот же домен.
func (a *App) RemoveUserRoutingRuleAt(index int) error {
	a.mu.Lock()
	defer a.mu.Unlock()
	if a.st == nil {
		return fmt.Errorf("app state is not initialized")
	}
	var verr error
	a.st.Update(func(cfg *models.AppConfig) {
		verr = config.RemoveUserRoutingRuleAt(cfg, index)
	})
	if verr != nil {
		return verr
	}
	snap := a.st.Snapshot()
	if err := config.Save(snap); err != nil {
		return err
	}
	a.invalidateRoutingDpiPanelCaches()

	if err := dpi.SetCustomDpiEntries(config.CustomDpiEntriesForRules(&snap)); err != nil {
		return fmt.Errorf("правило удалено, но список custom-dpi.txt не обновлён: %w", err)
	}
	a.invalidateRoutingDpiPanelCaches()
	return nil
}

// RemoveUserRoutingRule удаляет пользовательское правило по шаблону (как в колонке Pattern)
// и пересобирает custom-dpi.txt по оставшимся правилам.
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
	snap := a.st.Snapshot()
	if err := config.Save(snap); err != nil {
		return err
	}
	a.invalidateRoutingDpiPanelCaches()

	if err := dpi.SetCustomDpiEntries(config.CustomDpiEntriesForRules(&snap)); err != nil {
		return fmt.Errorf("правило удалено, но список custom-dpi.txt не обновлён: %w", err)
	}
	a.invalidateRoutingDpiPanelCaches()
	return nil
}

// RouteDecision — результат проверки «куда уйдёт трафик для адреса».
// Повторяет порядок buildRouteRules: процессы → пользователь → профиль → библиотека,
// затем route.final профиля (fallback на ProxyMode).
type RouteDecision struct {
	Query      string `json:"query"`
	Normalized string `json:"normalized"`
	Kind       string `json:"kind"`      // domain | ip | process
	Target     string `json:"target"`    // proxy | direct | dpi-bypass | block
	Source     string `json:"source"`    // process | user | profile | library | final
	RuleLabel  string `json:"ruleLabel"` // имя правила или outbound для final
	Final      bool   `json:"final"`     // true, если сработал final профиля
	Note       string `json:"note"`      // человеческая подсказка
	DPIOn      bool   `json:"dpiOn"`     // DPI включён и процесс запущен
	ProxyOn    bool   `json:"proxyOn"`   // прокси включён
}

func normalizeRouteQuery(raw string) (string, string) {
	q := strings.TrimSpace(strings.ToLower(raw))
	q = strings.TrimPrefix(q, "http://")
	q = strings.TrimPrefix(q, "https://")
	q = strings.TrimPrefix(q, "//")
	// IPv6 в квадратных скобках: "[::1]:443", "[::1]", "https://[v6]/path".
	// Скобки означают, что внутри гарантированно литерал адреса, а не CIDR.
	if strings.HasPrefix(q, "[") {
		if end := strings.IndexByte(q, ']'); end > 0 {
			host := q[1:end]
			if net.ParseIP(host) != nil {
				return host, "ip"
			}
		}
		return "", ""
	}
	// CIDR проверяем ДО отсечения пути: иначе "10.0.0.0/8" теряет маску.
	if i := strings.IndexByte(q, '/'); i > 0 {
		if _, _, err := net.ParseCIDR(q); err == nil {
			return q, "ip"
		}
	}
	if i := strings.IndexAny(q, "/?#"); i >= 0 {
		q = q[:i]
	}
	if i := strings.LastIndex(q, "@"); i >= 0 {
		q = q[i+1:]
	}
	if h, _, err := net.SplitHostPort(q); err == nil {
		q = h
	}
	q = strings.Trim(q, ".")
	if q == "" {
		return "", ""
	}
	if strings.Contains(q, "/") {
		if _, _, err := net.ParseCIDR(q); err == nil {
			return q, "ip"
		}
		return q, ""
	}
	if net.ParseIP(q) != nil {
		return q, "ip"
	}
	if strings.HasSuffix(strings.ToLower(q), ".exe") {
		return q, "process"
	}
	return q, "domain"
}

func domainMatchesExact(pattern, host string) bool {
	return strings.EqualFold(strings.TrimPrefix(pattern, "*."), host)
}

func domainMatchesSuffix(suffix, host string) bool {
	s := strings.TrimPrefix(strings.TrimPrefix(strings.ToLower(suffix), "."), "*.")
	if s == "" {
		return false
	}
	h := strings.ToLower(host)
	return h == s || strings.HasSuffix(h, "."+s)
}

func cidrMatches(cidr, host string) bool {
	ip := net.ParseIP(host)
	if ip == nil && host != "" {
		if ips, err := net.LookupHost(host); err == nil {
			for _, cand := range ips {
				if cidrMatches(cidr, cand) {
					return true
				}
			}
		}
		return false
	}
	if ip == nil {
		return false
	}
	_, network, err := net.ParseCIDR(cidr)
	if err != nil {
		return false
	}
	return network.Contains(ip)
}

// matchRule проверяет одно правило на совпадение с запросом.
func matchRule(r models.RouteRule, norm, kind string) bool {
	switch kind {
	case "process":
		for _, p := range r.Processes {
			if strings.EqualFold(p, norm) {
				return true
			}
		}
		return false
	case "ip":
		for _, c := range r.CIDRs {
			if c == norm || cidrMatches(c, norm) {
				return true
			}
		}
		// доменные правила для IP не применяются (в sing-box ip_cidr и domain — разные поля)
		return false
	default:
		for _, d := range r.Domains {
			if domainMatchesExact(d, norm) || domainMatchesSuffix(d, norm) {
				return true
			}
		}
		for _, s := range r.DomainSuffixes {
			if domainMatchesSuffix(s, norm) {
				return true
			}
		}
		return false
	}
}

func routeTargetLabel(target string) string {
	switch target {
	case string(models.RouteProxy):
		return "через прокси (VLESS)"
	case string(models.RouteDPIBypass):
		return "DPI-обход (напрямую, без VLESS)"
	case string(models.RouteDirect):
		return "напрямую, минуя прокси"
	case string(models.RouteBlock):
		return "заблокировано"
	default:
		return target
	}
}

// isPrivateRouteIP повторяет поведение ip_is_private из sing-box:
// приватные, loopback, link-local и ULA-адреса всегда уходят напрямую.
func isPrivateRouteIP(host string) bool {
	ip := net.ParseIP(strings.TrimSpace(host))
	if ip == nil {
		return false
	}
	return ip.IsPrivate() || ip.IsLoopback() || ip.IsLinkLocalUnicast() || ip.IsLinkLocalMulticast() || ip.IsUnspecified()
}

// VerifyRouteTarget показывает, какое правило сработает для домена, IP или exe.
// Только чтение состояния — ничего не меняет и не применяет.
func (a *App) VerifyRouteTarget(raw string) (*RouteDecision, error) {
	norm, kind := normalizeRouteQuery(raw)
	if norm == "" {
		return nil, fmt.Errorf("укажите домен, IP или имя процесса")
	}
	if kind == "" {
		return nil, fmt.Errorf("не похоже на домен, IP или .exe: %s", raw)
	}

	a.mu.Lock()
	if a.st == nil {
		a.mu.Unlock()
		return nil, fmt.Errorf("app state is not initialized")
	}
	snap := a.st.Snapshot()
	a.mu.Unlock()

	zdir, ok := config.ResolveZapretDirNearExe()
	listsDir := ""
	if ok {
		listsDir = filepath.Join(zdir, "lists")
	}

	type layer struct {
		source string
		rules  []models.RouteRule
	}
	resolved := models.ResolveRoutingProfileAlias(snap.RoutingProfile)
	isBuiltinMode := resolved == models.RoutingModeManual || resolved == models.RoutingModeFullProxy || resolved == models.RoutingModeGlobal
	layers := []layer{{source: "user", rules: snap.UserRoutingRules}}
	if !isBuiltinMode {
		layers = append(layers, layer{source: "process", rules: config.ProcessProxyRules(&snap)})
	}
	if kind != "process" {
		if profRules := config.EffectiveProfileRules(snap); len(profRules) > 0 {
			layers = append(layers, layer{source: "profile", rules: profRules})
		}
		if !snap.StrictRoutingMode && !isBuiltinMode {
			layers = append(layers, layer{source: "library", rules: config.DefaultRoutingLibraryRules(listsDir)})
		}
	}

	out := &RouteDecision{
		Query: raw, Normalized: norm, Kind: kind,
		DPIOn:   snap.DPIEnabled && snap.ActiveDPIProfile != "",
		ProxyOn: snap.ProxyEnabled,
	}

	// buildRouteRules ставит {"ip_is_private": true, "outbound": "direct"} первым правилом,
	// поэтому приватный адрес уходит напрямую раньше любых пользовательских правил.
	if kind == "ip" && isPrivateRouteIP(norm) {
		out.Target = string(models.RouteDirect)
		out.Source = "private"
		out.RuleLabel = "ip_is_private"
		out.Note = "приватный адрес — трафик идёт напрямую всегда, до всех правил маршрутизации"
		return out, nil
	}

	for _, l := range layers {
		for _, r := range l.rules {
			if !matchRule(r, norm, kind) {
				continue
			}
			out.Target = string(r.Action)
			out.Source = l.source
			out.RuleLabel = r.Name
			out.Note = routeTargetLabel(string(r.Action))
			switch l.source {
			case "process":
				out.Note += " · правило по процессу, оно выше доменов и профиля"
			case "user":
				out.Note += " · ваше правило, оно выше профиля"
			case "profile":
				out.Note += " · правило активного профиля"
			case "library":
				out.Note += " · встроенная библиотека Flux (нижний слой)"
			}
			return out, nil
		}
	}

	// Ничего не совпало — работает final профиля, иначе ProxyMode.
	// Повторяет routeFinalOutbound: если у профиля final пустой/неизвестный,
	// решение принимает proxyMode (direct_with_rules → direct).
	finalTarget := "proxy"
	if rp, ok := snap.ActiveRoutingProfile(); ok {
		switch strings.ToLower(strings.TrimSpace(rp.Final)) {
		case "direct":
			finalTarget = "direct"
		case "proxy":
			finalTarget = "proxy"
		default:
			if snap.ProxyMode == models.ProxyModeDirect {
				finalTarget = "direct"
			}
		}
	} else if snap.ProxyMode == models.ProxyModeDirect {
		finalTarget = "direct"
	}
	out.Target = finalTarget
	out.Source = "final"
	out.Final = true
	out.RuleLabel = "route.final"
	out.Note = "ни одно правило не совпало — трафик идёт по финальному правилу профиля: " + routeTargetLabel(finalTarget)
	if kind == "process" {
		out.Note = "процесс без правила — трафик идёт по финальному правилу профиля: " + routeTargetLabel(finalTarget)
	}
	return out, nil
}
