package config

import "pulseshield/app/models"

// РРјРµРЅР° РґРІСѓС… СЂРµР¶РёРјРѕРІ РјР°СЂС€СЂСѓС‚РёР·Р°С†РёРё. Р”СЂСѓРіРёС… РІСЃС‚СЂРѕРµРЅРЅС‹С… РїСЂРѕС„РёР»РµР№ РЅРµС‚:
// РІСЃС‘, С‡С‚Рѕ СЂР°РЅСЊС€Рµ РЅР°Р·С‹РІР°Р»РѕСЃСЊ default/selective-proxy/gaming-friendly/rf-smart-split/all-direct,
// СЃРІРѕСЂР°С‡РёРІР°РµС‚СЃСЏ РІ РѕРґРёРЅ РёР· СЌС‚РёС… РґРІСѓС… (СЃРј. MigrateRoutingProfileNames).
const (
	// RoutingModeGlobal — «Глобальный (Скрытие GEO)»: весь внешний трафик через прокси (только LAN напрямую).
	RoutingModeGlobal = "global"
	// RoutingModeFullProxy вЂ” В«РџРѕР»РЅС‹Р№ РїСЂРѕРєСЃРёВ»: РІСЃС‘, РєСЂРѕРјРµ Р»РѕРєР°Р»СЊРЅРѕРіРѕ Рё Р Р¤, РёРґС‘С‚ С‡РµСЂРµР· С‚СѓРЅРЅРµР»СЊ.
	RoutingModeFullProxy = "full-proxy"
	// RoutingModeManual вЂ” В«РџСЂРѕРєСЃРё СЃ СЂСѓС‡РЅС‹РјРё РїСЂР°РІРёР»Р°РјРёВ»: С‡РµСЂРµР· С‚СѓРЅРЅРµР»СЊ РёРґС‘С‚ С‚РѕР»СЊРєРѕ С‚Рѕ,
	// С‡С‚Рѕ РїРѕР»СЊР·РѕРІР°С‚РµР»СЊ РґРѕР±Р°РІРёР» СЃР°Рј (РїР»СЋСЃ СЃРїРёСЃРѕРє РјРµРґРёР°-СЃРµСЂРІРёСЃРѕРІ); РѕСЃС‚Р°Р»СЊРЅРѕРµ РЅР°РїСЂСЏРјСѓСЋ.
	RoutingModeManual = "manual"
)

// RoutingModeNames вЂ” РїРѕСЂСЏРґРѕРє РѕС‚РѕР±СЂР°Р¶РµРЅРёСЏ РІ UI.
func RoutingModeNames() []string {
	return []string{RoutingModeGlobal, RoutingModeFullProxy, RoutingModeManual}
}

// IsKnownRoutingMode СЃРѕРѕР±С‰Р°РµС‚, СЏРІР»СЏРµС‚СЃСЏ Р»Рё РёРјСЏ РѕРґРЅРёРј РёР· РїРѕРґРґРµСЂР¶РёРІР°РµРјС‹С… СЂРµР¶РёРјРѕРІ.
func IsKnownRoutingMode(name string) bool {
	return name == RoutingModeGlobal || name == RoutingModeFullProxy || name == RoutingModeManual
}

// legacyProfileToMode вЂ” СЃРѕРѕС‚РІРµС‚СЃС‚РІРёРµ СЃС‚Р°СЂС‹С… РІСЃС‚СЂРѕРµРЅРЅС‹С… РїСЂРѕС„РёР»РµР№ РЅРѕРІС‹Рј СЂРµР¶РёРјР°Рј.
// РџСЂРѕС„РёР»Рё, РєРѕС‚РѕСЂС‹С… Р·РґРµСЃСЊ РЅРµС‚, СЃС‡РёС‚Р°СЋС‚СЃСЏ РїРѕР»СЊР·РѕРІР°С‚РµР»СЊСЃРєРёРјРё Рё РЅРµ РїРµСЂРµРёРјРµРЅРѕРІС‹РІР°СЋС‚СЃСЏ.
var legacyProfileToMode = map[string]string{
	"default":            RoutingModeFullProxy,
	"all-direct":         RoutingModeManual,
	"selective-proxy":    RoutingModeManual,
	"gaming-friendly":    RoutingModeManual,
	"rf-smart-split":     RoutingModeManual,
	"bypass-russia":      RoutingModeManual,
	"bypass-rf":          RoutingModeManual,
	"general-dpi-bypass": RoutingModeManual,
}

// MigrateRoutingProfileNames РїРµСЂРµРІРѕРґРёС‚ СЃРѕС…СЂР°РЅС‘РЅРЅС‹Рµ РёРјРµРЅР° СЃС‚Р°СЂС‹С… РІСЃС‚СЂРѕРµРЅРЅС‹С… РїСЂРѕС„РёР»РµР№
// РЅР° РґРІР° Р°РєС‚СѓР°Р»СЊРЅС‹С… СЂРµР¶РёРјР°. Р’РѕР·РІСЂР°С‰Р°РµС‚ true, РµСЃР»Рё РёРјСЏ Р°РєС‚РёРІРЅРѕРіРѕ РїСЂРѕС„РёР»СЏ РёР·РјРµРЅРёР»РѕСЃСЊ,
// Рё СЃРїРёСЃРѕРє РёРјС‘РЅ, РєРѕС‚РѕСЂС‹Рµ Р±С‹Р»Рё РїРµСЂРµРёРјРµРЅРѕРІР°РЅС‹ (РґР»СЏ РѕС‡РёСЃС‚РєРё РґСѓР±Р»РёРєР°С‚РѕРІ).
func MigrateRoutingProfileNames(cfg *models.AppConfig) (activeChanged bool, renamed []string) {
	if cfg == nil {
		return false, nil
	}
	for i := range cfg.RoutingProfiles {
		old := cfg.RoutingProfiles[i].Name
		if mapped, ok := legacyProfileToMode[old]; ok {
			if cfg.RoutingProfiles[i].Name != mapped {
				cfg.RoutingProfiles[i].Name = mapped
				renamed = append(renamed, old+"в†’"+mapped)
			}
		}
	}
	if mapped, ok := legacyProfileToMode[cfg.RoutingProfile]; ok {
		if cfg.RoutingProfile != mapped {
			cfg.RoutingProfile = mapped
			activeChanged = true
		}
	}
	return activeChanged, renamed
}

// builtinRoutingPresets РІРѕР·РІСЂР°С‰Р°РµС‚ РґРІР° СЂРµР¶РёРјР° РјР°СЂС€СЂСѓС‚РёР·Р°С†РёРё.
//
//	РџРѕР»РЅС‹Р№ РїСЂРѕРєСЃРё (final: proxy) вЂ” РІ С‚СѓРЅРЅРµР»СЊ СѓС…РѕРґРёС‚ РІСЃС‘, РєСЂРѕРјРµ Р»РѕРєР°Р»СЊРЅС‹С… Р°РґСЂРµСЃРѕРІ,
//	РїСЂРёРІР°С‚РЅС‹С… СЃРµС‚РµР№ Рё СЂРѕСЃСЃРёР№СЃРєРёС… СЃРµСЂРІРёСЃРѕРІ (РѕРЅРё РёРґСѓС‚ РЅР°РїСЂСЏРјСѓСЋ, С‡С‚РѕР±С‹ РЅРµ РіРѕРЅСЏС‚СЊ
//	Р»РёС€РЅРёР№ С‚СЂР°С„РёРє С‡РµСЂРµР· VLESS Рё РЅРµ Р»РѕРјР°С‚СЊ РґРѕСЃС‚СѓРї Рє Р Р¤-СЂРµСЃСѓСЂСЃР°Рј).
//
//	Р СѓС‡РЅС‹Рµ РїСЂР°РІРёР»Р° (final: direct) вЂ” РІ С‚СѓРЅРЅРµР»СЊ РёРґС‘С‚ С‚РѕР»СЊРєРѕ РјРµРґРёР°-СЃРїРёСЃРѕРє Рё С‚Рѕ,
//	С‡С‚Рѕ РїРѕР»СЊР·РѕРІР°С‚РµР»СЊ РґРѕР±Р°РІРёР» СЃР°Рј; РѕСЃС‚Р°Р»СЊРЅРѕРµ РЅР°РїСЂСЏРјСѓСЋ.
//
// РџРѕР»СЊР·РѕРІР°С‚РµР»СЊСЃРєРёРµ РїСЂР°РІРёР»Р° (UserRoutingRules) РёРјРµСЋС‚ РїСЂРёРѕСЂРёС‚РµС‚ РЅР°Рґ РѕР±РѕРёРјРё СЂРµР¶РёРјР°РјРё,
// РїРѕСЌС‚РѕРјСѓ РґРѕР±Р°РІР»СЏС‚СЊ СЃРІРѕРё РёСЃРєР»СЋС‡РµРЅРёСЏ РјРѕР¶РЅРѕ РІ Р»СЋР±РѕРј РёР· РЅРёС….
func builtinRoutingPresets() []models.RoutingProfile {
	localDomains := []string{"*.local", "localhost", "*.lan"}

	localDirect := []models.RouteRule{
		{Name: "local-direct", Domains: localDomains, Action: models.RouteDirect},
	}
	lanCIDR := models.RouteRule{
		Name:   "private-lan-direct",
		CIDRs:  []string{"10.0.0.0/8", "172.16.0.0/12", "192.168.0.0/16", "127.0.0.0/8"},
		Action: models.RouteDirect,
	}
	ruDirect := models.RouteRule{
		Name:           "ru-tld-direct",
		DomainSuffixes: RuDirectDomainSuffixes(),
		Action:         models.RouteDirect,
	}
	ruServicesDirect := models.RouteRule{
		Name:           "ru-services-direct",
		Domains:        RuDirectDomains(),
		DomainSuffixes: RuDirectExtraSuffixes(),
		Action:         models.RouteDirect,
	}

	globalRules := append(append([]models.RouteRule{}, localDirect...), lanCIDR)

	base := append(append([]models.RouteRule{}, localDirect...), lanCIDR, ruDirect, ruServicesDirect)

	fullProxyRules := append([]models.RouteRule{}, base...)

	manualRules := append([]models.RouteRule{}, base...)

	return []models.RoutingProfile{
		{
			Name:  RoutingModeGlobal,
			Final: "proxy",
			Rules: globalRules,
		},
		{
			Name:  RoutingModeFullProxy,
			Final: "proxy",
			Rules: fullProxyRules,
		},
		{
			Name:  RoutingModeManual,
			Final: "direct",
			Rules: manualRules,
		},
	}
}

// dedupeRoutingProfiles РѕСЃС‚Р°РІР»СЏРµС‚ РїРµСЂРІС‹Р№ РїСЂРѕС„РёР»СЊ СЃ РґР°РЅРЅС‹Рј РёРјРµРЅРµРј (С‡РёРЅРёС‚ РґСѓР±Р»РёРєР°С‚С‹ РІ СЃРѕС…СЂР°РЅС‘РЅРЅРѕРј JSON).
func dedupeRoutingProfiles(cfg *models.AppConfig) bool {
	if cfg == nil || len(cfg.RoutingProfiles) == 0 {
		return false
	}
	before := len(cfg.RoutingProfiles)
	seen := make(map[string]struct{}, len(cfg.RoutingProfiles))
	out := cfg.RoutingProfiles[:0]
	for _, p := range cfg.RoutingProfiles {
		name := p.Name
		if name == "" {
			continue
		}
		if _, ok := seen[name]; ok {
			continue
		}
		seen[name] = struct{}{}
		out = append(out, p)
	}
	cfg.RoutingProfiles = out
	return len(cfg.RoutingProfiles) < before
}

// MergeBuiltinRoutingProfiles РїСЂРёРІРѕРґРёС‚ РїСЂРѕС„РёР»Рё Рє РґРІСѓРј СЂРµР¶РёРјР°Рј:
// РјРёРіСЂРёСЂСѓРµС‚ СЃС‚Р°СЂС‹Рµ РІСЃС‚СЂРѕРµРЅРЅС‹Рµ РёРјРµРЅР°, СѓРґР°Р»СЏРµС‚ РїСЂРѕС„РёР»Рё, РєРѕС‚РѕСЂС‹С… Р±РѕР»СЊС€Рµ РЅРµС‚,
// Рё РіР°СЂР°РЅС‚РёСЂСѓРµС‚ РЅР°Р»РёС‡РёРµ РѕР±РѕРёС… СЂРµР¶РёРјРѕРІ.
//
// РџРѕР»СЊР·РѕРІР°С‚РµР»СЊСЃРєРёРµ РїСЂРѕС„РёР»Рё СЃ СЃРѕР±СЃС‚РІРµРЅРЅС‹РјРё РёРјРµРЅР°РјРё СЃРѕС…СЂР°РЅСЏСЋС‚СЃСЏ: РјС‹ РЅРµ СѓРґР°Р»СЏРµРј
// С‚Рѕ, С‡С‚Рѕ С‡РµР»РѕРІРµРє СЃРѕР·РґР°Р» СЃР°Рј, вЂ” РёРЅР°С‡Рµ РµРіРѕ РїСЂР°РІРёР»Р° РјРѕР»С‡Р° РїСЂРѕРїР°РґСѓС‚.
func MergeBuiltinRoutingProfiles(cfg *models.AppConfig) bool {
	if cfg == nil {
		return false
	}
	changed := dedupeRoutingProfiles(cfg)

	// РЁР°Рі 1: СЃС‚Р°СЂС‹Рµ РІСЃС‚СЂРѕРµРЅРЅС‹Рµ РёРјРµРЅР° в†’ РЅРѕРІС‹Рµ СЂРµР¶РёРјС‹.
	if activeChanged, _ := MigrateRoutingProfileNames(cfg); activeChanged {
		changed = true
	}
	// РџРѕСЃР»Рµ РїРµСЂРµРёРјРµРЅРѕРІР°РЅРёСЏ РјРѕРіР»Рё РїРѕСЏРІРёС‚СЊСЃСЏ РґСѓР±Р»РёРєР°С‚С‹ (РЅР°РїСЂРёРјРµСЂ, Р±С‹Р»Рѕ Рё default, Рё rf-smart-split).
	if dedupeRoutingProfiles(cfg) {
		changed = true
	}

	// РЁР°Рі 2: РєР°РЅРѕРЅРёС‡РµСЃРєРёРµ РѕРїСЂРµРґРµР»РµРЅРёСЏ РґРІСѓС… СЂРµР¶РёРјРѕРІ.
	// РџРµСЂРµР·Р°РїРёСЃС‹РІР°РµРј РїСЂР°РІРёР»Р° Сѓ РїСЂРѕС„РёР»РµР№ СЃ СЌС‚РёРјРё РёРјРµРЅР°РјРё, С‡С‚РѕР±С‹ РїСЂР°РІРєРё РІСЃС‚СЂРѕРµРЅРЅРѕР№
	// Р»РѕРіРёРєРё РґРѕРµР·Р¶Р°Р»Рё РґРѕ СЃСѓС‰РµСЃС‚РІСѓСЋС‰РёС… СѓСЃС‚Р°РЅРѕРІРѕРє.
	builtins := builtinRoutingPresets()
	byName := make(map[string]models.RoutingProfile, len(builtins))
	for _, p := range builtins {
		byName[p.Name] = p
	}
	kept := cfg.RoutingProfiles[:0]
	for _, p := range cfg.RoutingProfiles {
		if bp, ok := byName[p.Name]; ok {
			kept = append(kept, bp)
			delete(byName, p.Name)
			continue
		}
		kept = append(kept, p)
	}
	cfg.RoutingProfiles = kept
	for _, name := range RoutingModeNames() {
		if p, ok := byName[name]; ok {
			cfg.RoutingProfiles = append(cfg.RoutingProfiles, p)
			changed = true
		}
	}

	// РЁР°Рі 3: Р°РєС‚РёРІРЅС‹Р№ РїСЂРѕС„РёР»СЊ РґРѕР»Р¶РµРЅ СЃСѓС‰РµСЃС‚РІРѕРІР°С‚СЊ.
	if !profileExists(cfg, cfg.RoutingProfile) {
		cfg.RoutingProfile = RoutingModeManual
		changed = true
	}
	return changed
}

// profileExists СЃРѕРѕР±С‰Р°РµС‚, РµСЃС‚СЊ Р»Рё РїСЂРѕС„РёР»СЊ СЃ С‚Р°РєРёРј РёРјРµРЅРµРј.
func profileExists(cfg *models.AppConfig, name string) bool {
	if cfg == nil || name == "" {
		return false
	}
	for _, p := range cfg.RoutingProfiles {
		if p.Name == name {
			return true
		}
	}
	return false
}
