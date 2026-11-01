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
	rule := userRuleFromPattern(pattern, action)
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

func userRuleFromPattern(pattern string, action models.RouteAction) models.RouteRule {
	name := "user:" + pattern
	if strings.Contains(pattern, "/") {
		if _, _, err := net.ParseCIDR(pattern); err == nil {
			return models.RouteRule{Name: name, CIDRs: []string{pattern}, Action: action}
		}
	}
	if ip := net.ParseIP(pattern); ip != nil {
		bits := 32
		if ip.To4() == nil {
			bits = 128
		}
		cidr := (&net.IPNet{IP: ip, Mask: net.CIDRMask(bits, bits)}).String()
		return models.RouteRule{Name: name, CIDRs: []string{cidr}, Action: action}
	}
	if strings.HasPrefix(pattern, "*.") {
		return models.RouteRule{Name: name, Domains: []string{pattern}, Action: action}
	}
	if strings.HasPrefix(pattern, ".") {
		suf := strings.TrimPrefix(pattern, ".")
		return models.RouteRule{Name: name, DomainSuffixes: []string{suf}, Action: action}
	}
	return models.RouteRule{Name: name, DomainSuffixes: []string{pattern}, Action: action}
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
