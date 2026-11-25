package config

import (
	"fmt"
	"net"
	"strings"

	"pulseshield/app/models"
)

// GetUserRoutingRules возвращает только пользовательские правила (без профиля и библиотеки).
func GetUserRoutingRules(cfg models.AppConfig) []models.RouteRule {
	if len(cfg.UserRoutingRules) == 0 {
		return nil
	}
	out := make([]models.RouteRule, len(cfg.UserRoutingRules))
	copy(out, cfg.UserRoutingRules)
	return out
}

// AddUserRoutingRule добавляет правило: target "proxy" | "direct" | "dpi-bypass" | "block".
// Для target=dpi-bypass домен дополнительно попадает в zapret/lists/custom-dpi.txt,
// чтобы winws2 действительно обходил его (правило маршрутизации само по себе
// влияет только на sing-box и до DPI не доходит).
func AddUserRoutingRule(cfg *models.AppConfig, pattern string, target string) error {
	if cfg == nil {
		return fmt.Errorf("config is nil")
	}
	pattern = strings.TrimSpace(pattern)
	if pattern == "" {
		return fmt.Errorf("пустой домен или IP")
	}
	action := models.RouteAction(strings.ToLower(strings.TrimSpace(target)))
	switch action {
	case models.RouteProxy, models.RouteDirect, models.RouteDPIBypass, models.RouteBlock:
	default:
		return fmt.Errorf(`target должен быть proxy, direct, dpi-bypass или block`)
	}
	rule, ok := buildUserRuleFromPattern(pattern, action)
	if !ok {
		return fmt.Errorf("некорректный шаблон домена, IP или процесса: %s", pattern)
	}
	cfg.UserRoutingRules = append(cfg.UserRoutingRules, rule)
	return nil
}

// CustomDpiEntriesForRules возвращает домены из пользовательских правил с действием
// dpi-bypass. Используется для полной пересборки custom-dpi.txt, чтобы список
// всегда соответствовал текущему набору правил.
func CustomDpiEntriesForRules(cfg *models.AppConfig) []string {
	if cfg == nil {
		return nil
	}
	out := make([]string, 0, len(cfg.UserRoutingRules))
	for _, r := range cfg.UserRoutingRules {
		if r.Action != models.RouteDPIBypass {
			continue
		}
		out = append(out, r.Domains...)
		for _, s := range r.DomainSuffixes {
			out = append(out, s)
		}
	}
	return out
}

// normalizeUserRuleToken нормализует пользовательский ввод правила.
// Возвращает значение и тип: "exact" | "suffix" | "cidr" | "process" | "".
func normalizeUserRuleToken(raw string) (string, string) {
	q := strings.TrimSpace(raw)
	if q == "" {
		return "", ""
	}
	low := strings.ToLower(q)
	exact := false
	if strings.HasPrefix(low, "full:") {
		exact = true
		q = strings.TrimSpace(q[len("full:"):])
		low = strings.ToLower(q)
	}

	for _, pref := range []string{"http://", "https://", "//"} {
		if strings.HasPrefix(low, pref) {
			q = q[len(pref):]
			low = strings.ToLower(q)
			break
		}
	}
	q = strings.TrimSpace(q)
	low = strings.ToLower(q)
	if q == "" {
		return "", ""
	}

	// IPv6 в квадратных скобках: "[::1]:443", "[::1]"
	if strings.HasPrefix(q, "[") {
		if end := strings.IndexByte(q, ']'); end > 0 {
			host := strings.TrimSpace(q[1:end])
			if ip := net.ParseIP(host); ip != nil {
				return (&net.IPNet{IP: ip, Mask: net.CIDRMask(128, 128)}).String(), "cidr"
			}
		}
		return "", ""
	}

	// CIDR проверяем ДО отсечения пути, иначе "10.0.0.0/8" потеряет маску.
	if strings.Contains(q, "/") {
		if _, ipNet, err := net.ParseCIDR(q); err == nil {
			return ipNet.String(), "cidr"
		}
	}

	// Отсекаем путь, query и фрагмент URL.
	if i := strings.IndexAny(q, "/?#"); i >= 0 {
		q = q[:i]
	}
	// Отсекаем userinfo (user:pass@host).
	if i := strings.LastIndex(q, "@"); i >= 0 {
		q = q[i+1:]
	}
	// Отсекаем порт (host:port), не ломая голый IPv6 без скобок.
	if h, _, err := net.SplitHostPort(q); err == nil {
		q = h
	}
	q = strings.ToLower(strings.TrimSpace(q))
	if q == "" {
		return "", ""
	}

	if ip := net.ParseIP(q); ip != nil {
		bits := 32
		if ip.To4() == nil {
			bits = 128
		}
		return (&net.IPNet{IP: ip, Mask: net.CIDRMask(bits, bits)}).String(), "cidr"
	}

	if strings.HasSuffix(q, ".exe") && len(q) > len(".exe") {
		// Если ввели путь Windows C:\...\app.exe — берём только имя exe.
		if idx := strings.LastIndexAny(q, `\/`); idx >= 0 {
			q = q[idx+1:]
		}
		return q, "process"
	}

	for strings.HasPrefix(q, "*.") || strings.HasPrefix(q, ".") {
		if strings.HasPrefix(q, "*.") {
			q = strings.TrimPrefix(q, "*.")
		} else {
			q = strings.TrimPrefix(q, ".")
		}
	}
	q = strings.Trim(q, ".*")
	if q == "" {
		return "", ""
	}
	if exact {
		return q, "exact"
	}
	return q, "suffix"
}

func buildUserRuleFromPattern(pattern string, action models.RouteAction) (models.RouteRule, bool) {
	val, kind := normalizeUserRuleToken(pattern)
	if val == "" {
		return models.RouteRule{}, false
	}
	name := "user:" + strings.TrimSpace(pattern)
	switch kind {
	case "cidr":
		return models.RouteRule{Name: name, CIDRs: []string{val}, Action: action}, true
	case "process":
		return models.RouteRule{Name: name, Processes: []string{val}, Action: action}, true
	case "exact":
		return models.RouteRule{Name: name, Domains: []string{val}, Action: action}, true
	case "suffix":
		return models.RouteRule{Name: name, DomainSuffixes: []string{val}, Action: action}, true
	default:
		return models.RouteRule{}, false
	}
}

func userRuleFromPattern(pattern string, action models.RouteAction) models.RouteRule {
	if r, ok := buildUserRuleFromPattern(pattern, action); ok {
		return r
	}
	p := strings.ToLower(strings.TrimSpace(pattern))
	return models.RouteRule{Name: "user:" + strings.TrimSpace(pattern), DomainSuffixes: []string{p}, Action: action}
}

// NormalizeUserRoutingRules мигрирует старые/некорректные пользовательские правила:
// переносит wildcard-маски "*.domain.com" из Domains в DomainSuffixes, очищает URL/порты/регистр
// и выделяет ".exe" в Processes. Возвращает true, если конфиг изменился.
func NormalizeUserRoutingRules(cfg *models.AppConfig) bool {
	if cfg == nil || len(cfg.UserRoutingRules) == 0 {
		return false
	}
	changed := false
	for i := range cfg.UserRoutingRules {
		r := cfg.UserRoutingRules[i]
		origDomains := append([]string(nil), r.Domains...)
		origSuffixes := append([]string(nil), r.DomainSuffixes...)
		origCIDRs := append([]string(nil), r.CIDRs...)
		origProcs := append([]string(nil), r.Processes...)

		var newDomains, newSuffixes, newCIDRs, newProcs []string
		seenDom := map[string]bool{}
		seenSuf := map[string]bool{}
		seenCIDR := map[string]bool{}
		seenProc := map[string]bool{}

		addUnique := func(dst *[]string, seen map[string]bool, val string) {
			if val == "" || seen[val] {
				return
			}
			seen[val] = true
			*dst = append(*dst, val)
		}

		// Элементы из r.Domains: всё с "*." / "." / "://" / "/" / ":" / ".exe" прогоняем через нормализацию.
		// Чистые exact-домены без "*" (например, созданные с full:) остаются в Domains в нижнем регистре.
		for _, raw := range r.Domains {
			t := strings.TrimSpace(raw)
			if t == "" {
				changed = true
				continue
			}
			hasWildcardOrProto := strings.HasPrefix(t, "*.") || strings.HasPrefix(t, ".") ||
				strings.Contains(t, "://") || strings.ContainsAny(t, "/?#") ||
				strings.HasSuffix(strings.ToLower(t), ".exe")
			if !hasWildcardOrProto {
				val, kind := normalizeUserRuleToken("full:" + t)
				switch kind {
				case "cidr":
					addUnique(&newCIDRs, seenCIDR, val)
				case "process":
					addUnique(&newProcs, seenProc, val)
				case "exact":
					addUnique(&newDomains, seenDom, val)
				}
				continue
			}
			val, kind := normalizeUserRuleToken(t)
			switch kind {
			case "cidr":
				addUnique(&newCIDRs, seenCIDR, val)
			case "process":
				addUnique(&newProcs, seenProc, val)
			case "suffix", "exact":
				addUnique(&newSuffixes, seenSuf, val)
			}
		}

		for _, raw := range r.DomainSuffixes {
			val, kind := normalizeUserRuleToken(raw)
			switch kind {
			case "cidr":
				addUnique(&newCIDRs, seenCIDR, val)
			case "process":
				addUnique(&newProcs, seenProc, val)
			case "suffix", "exact":
				addUnique(&newSuffixes, seenSuf, val)
			}
		}

		for _, raw := range r.CIDRs {
			val, kind := normalizeUserRuleToken(raw)
			if kind == "cidr" {
				addUnique(&newCIDRs, seenCIDR, val)
			}
		}

		for _, raw := range r.Processes {
			p := strings.ToLower(strings.TrimSpace(raw))
			addUnique(&newProcs, seenProc, p)
		}

		if !equalStringSlices(origDomains, newDomains) ||
			!equalStringSlices(origSuffixes, newSuffixes) ||
			!equalStringSlices(origCIDRs, newCIDRs) ||
			!equalStringSlices(origProcs, newProcs) {
			cfg.UserRoutingRules[i].Domains = newDomains
			cfg.UserRoutingRules[i].DomainSuffixes = newSuffixes
			cfg.UserRoutingRules[i].CIDRs = newCIDRs
			cfg.UserRoutingRules[i].Processes = newProcs
			changed = true
		}
	}
	return changed
}

func equalStringSlices(a, b []string) bool {
	if len(a) != len(b) {
		return false
	}
	for i := range a {
		if a[i] != b[i] {
			return false
		}
	}
	return true
}

// RemoveUserRoutingRule удаляет первое совпадение по имени правила user:<pattern> или по совпадению паттерна.
func RemoveUserRoutingRule(cfg *models.AppConfig, pattern string) error {
	if cfg == nil {
		return fmt.Errorf("config is nil")
	}
	pattern = strings.TrimSpace(pattern)
	if pattern == "" {
		return fmt.Errorf("пустой ключ")
	}
	targetName := "user:" + pattern
	rules := cfg.UserRoutingRules[:0]
	for _, r := range cfg.UserRoutingRules {
		if r.Name == targetName || userRuleSignature(r) == pattern {
			continue
		}
		rules = append(rules, r)
	}
	if len(rules) == len(cfg.UserRoutingRules) {
		return fmt.Errorf("правило не найдено")
	}
	cfg.UserRoutingRules = rules
	return nil
}

func userRuleSignature(r models.RouteRule) string {
	if len(r.CIDRs) > 0 {
		return r.CIDRs[0]
	}
	if len(r.Domains) > 0 {
		return r.Domains[0]
	}
	if len(r.DomainSuffixes) > 0 {
		return "." + r.DomainSuffixes[0]
	}
	if len(r.Processes) > 0 {
		return r.Processes[0]
	}
	return ""
}

// RemoveUserRoutingRuleAt удаляет правило по индексу в cfg.UserRoutingRules.
//
// Это единственный надёжный способ, когда на один домен заведено несколько правил
// с разными действиями: удаление по шаблону снесло бы их все, а по действию —
// можно перепутать порядок. Индекс приходит из того же снимка, что показан в UI.
func RemoveUserRoutingRuleAt(cfg *models.AppConfig, index int) error {
	if cfg == nil {
		return fmt.Errorf("config is nil")
	}
	if index < 0 || index >= len(cfg.UserRoutingRules) {
		return fmt.Errorf("правило не найдено: индекс %d вне диапазона", index)
	}
	cfg.UserRoutingRules = append(cfg.UserRoutingRules[:index], cfg.UserRoutingRules[index+1:]...)
	return nil
}

// FindUserRoutingRuleIndex находит правило по шаблону и действию.
// Возвращает -1, если такого нет.
func FindUserRoutingRuleIndex(cfg *models.AppConfig, pattern, action string) int {
	if cfg == nil {
		return -1
	}
	pattern = strings.TrimSpace(pattern)
	action = strings.TrimSpace(action)
	for i, r := range cfg.UserRoutingRules {
		if userRuleSignature(r) != pattern {
			continue
		}
		if action != "" && string(r.Action) != action {
			continue
		}
		return i
	}
	return -1
}
