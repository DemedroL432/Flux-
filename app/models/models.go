package models

import (
	"strings"
	"time"
)

// DefaultLocalInboundUser — имя пользователя локального mixed inbound (SOCKS/HTTP); пароль хранится отдельно.
const DefaultLocalInboundUser = "pulse"

type VlessServer struct {
	ID                    string            `json:"id" yaml:"id"`
	Name                  string            `json:"name" yaml:"name"`
	RawVLESS              string            `json:"rawVless" yaml:"rawVless"`
	Host                  string            `json:"host" yaml:"host"`
	Port                  string            `json:"port" yaml:"port"`
	UUID                  string            `json:"uuid" yaml:"uuid"`
	Params                map[string]string `json:"params,omitempty" yaml:"params,omitempty"`
	CreatedAt             time.Time         `json:"createdAt" yaml:"createdAt"`
	Source                string            `json:"source,omitempty" yaml:"source,omitempty"`
	FromSubscription      bool              `json:"fromSubscription,omitempty" yaml:"fromSubscription,omitempty"`
	LastCheckOK           bool              `json:"lastCheckOk,omitempty" yaml:"lastCheckOk,omitempty"`
	LastCheckLatencyMs    int               `json:"lastCheckLatencyMs,omitempty" yaml:"lastCheckLatencyMs,omitempty"`
	LastCheckDownloadMbps float64           `json:"lastCheckDownloadMbps,omitempty" yaml:"lastCheckDownloadMbps,omitempty"`
	LastCheckedAt         time.Time         `json:"lastCheckedAt,omitempty" yaml:"lastCheckedAt,omitempty"`
	Recommended           bool              `json:"recommended,omitempty" yaml:"recommended,omitempty"`
}

type Server = VlessServer

type ProxyMode string

const (
	ProxyModeSystem ProxyMode = "system_proxy"
	ProxyModeTun    ProxyMode = "tun"
	ProxyModeDirect ProxyMode = "direct_with_rules"
)

type RouteAction string

const (
	RouteProxy     RouteAction = "proxy"
	RouteDPIBypass RouteAction = "dpi-bypass"
	RouteDirect    RouteAction = "direct"
	RouteBlock     RouteAction = "block"
)

type RouteRule struct {
	Name           string      `json:"name" yaml:"name"`
	Domains        []string    `json:"domains,omitempty" yaml:"domains,omitempty"`
	DomainSuffixes []string    `json:"domainSuffixes,omitempty" yaml:"domainSuffixes,omitempty"`
	CIDRs          []string    `json:"cidrs,omitempty" yaml:"cidrs,omitempty"`
	Processes      []string    `json:"processes,omitempty" yaml:"processes,omitempty"`
	Action         RouteAction `json:"action" yaml:"action"`
}

type RoutingProfile struct {
	Name  string      `json:"name" yaml:"name"`
	Rules []RouteRule `json:"rules" yaml:"rules"`
	// Final: пусто — sing-box route.final как раньше (от ProxyMode); "direct" | "proxy" — явный финал.
	Final string `json:"final,omitempty" yaml:"final,omitempty"`
}

type DPIProfile struct {
	Name       string   `json:"name" yaml:"name"`
	Tool       string   `json:"tool" yaml:"tool"`             // e.g. "goodbyedpi", "zapret"
	ConfigPath string   `json:"configPath" yaml:"configPath"` // passed as-is to tool args if needed
	Args       []string `json:"args" yaml:"args"`             // optional extra args; may include "{config}"
}

type AppConfig struct {
	Servers         []Server      `json:"servers" yaml:"servers"`
	VlessServers    []VlessServer `json:"vlessServers,omitempty" yaml:"vlessServers,omitempty"`
	SubscriptionURL string        `json:"subscriptionUrl,omitempty" yaml:"subscriptionUrl,omitempty"`
	// SubscriptionExpiresUnix — время окончания подписки (Unix), из заголовка Subscription-Userinfo; 0 = неизвестно.
	SubscriptionExpiresUnix int64 `json:"subscriptionExpiresUnix,omitempty" yaml:"subscriptionExpiresUnix,omitempty"`
	// SubscriptionProviderLabel — подпись провайдера (хост URL или имя из ответа).
	SubscriptionProviderLabel string           `json:"subscriptionProviderLabel,omitempty" yaml:"subscriptionProviderLabel,omitempty"`
	ActiveServerID            string           `json:"activeServerId" yaml:"activeServerId"`
	ProxyMode                 ProxyMode        `json:"proxyMode" yaml:"proxyMode"`
	ProxyEnabled              bool             `json:"proxyEnabled" yaml:"proxyEnabled"`
	ProxyPort                 int              `json:"proxyPort" yaml:"proxyPort"` // mixed inbound 127.0.0.1:ProxyPort
	DPIEnabled                bool             `json:"dpiEnabled" yaml:"dpiEnabled"`
	ActiveDPIProfile          string           `json:"activeDpiProfile" yaml:"activeDpiProfile"`
	RoutingProfile            string           `json:"routingProfile" yaml:"routingProfile"`
	RoutingProfiles           []RoutingProfile `json:"routingProfiles" yaml:"routingProfiles"`
	// RoutingBuiltinRev — версия встроенных профилей; при увеличении в коде профили перезаписываются (см. config.ApplyRoutingBuiltinUpgrade).
	RoutingBuiltinRev int          `json:"routingBuiltinRev,omitempty" yaml:"routingBuiltinRev,omitempty"`
	DPIProfiles       []DPIProfile `json:"dpiProfiles" yaml:"dpiProfiles"`
	// DnsPreset: dhcp | xbox | custom (xbox = 111.88.96.50/51 см. https://xbox-dns.ru/ )
	DnsPreset    string `json:"dnsPreset" yaml:"dnsPreset"`
	DnsPrimary   string `json:"dnsPrimary" yaml:"dnsPrimary"`
	DnsSecondary string `json:"dnsSecondary" yaml:"dnsSecondary"`
	// ProcessProxyMap — пользовательские/встроенные process-based правила proxy-only. Ключ: имя процесса, значение: true.
	ProcessProxyMap map[string]bool `json:"processProxyMap,omitempty" yaml:"processProxyMap,omitempty"`
	// Ссылка на Telegram канал/чат проекта (информация, статус)
	TelegramInfoURL string `json:"telegramInfoUrl" yaml:"telegramInfoUrl"`
	// Подпись для интерфейса (например имя канала); при клике открывается TelegramInfoURL
	TelegramChannelName string `json:"telegramChannelName" yaml:"telegramChannelName"`
	// LicenseTier: free | premium | supporter (авторская подписка; сторонние VLESS не затрагиваются)
	LicenseTier string `json:"licenseTier,omitempty" yaml:"licenseTier,omitempty"`
	// LicenseExpiresUnix — 0 = без срока или не активировано
	LicenseExpiresUnix int64 `json:"licenseExpiresUnix,omitempty" yaml:"licenseExpiresUnix,omitempty"`
	// LicenseToken — опаковый токен с сервера активации (опционально)
	LicenseToken string `json:"licenseToken,omitempty" yaml:"licenseToken,omitempty"`
	// LicenseAPIBase — базовый URL сервера активации; пусто = только офлайн free
	LicenseAPIBase string `json:"licenseApiBase,omitempty" yaml:"licenseApiBase,omitempty"`
	// PremiumServerNameSubstrings — если имя или хост сервера содержит подстроку (без регистра), узел считается «Premium» от авторов
	PremiumServerNameSubstrings []string `json:"premiumServerNameSubstrings,omitempty" yaml:"premiumServerNameSubstrings,omitempty"`
	// DpiListEnabled — явное состояние списков Zapret lists/*.txt (ключ — имя без .txt). Отсутствие ключа = включено.
	DpiListEnabled map[string]bool `json:"dpiListEnabled,omitempty" yaml:"dpiListEnabled,omitempty"`
	// UserRoutingRules — пользовательские правила маршрутизации (приоритет выше встроенной библиотеки и профиля — см. MergeSingBoxRoutingRules).
	UserRoutingRules []RouteRule `json:"userRoutingRules,omitempty" yaml:"userRoutingRules,omitempty"`
	// StrictRoutingMode — «чистый гейминг»: не подмешивать DefaultRoutingLibraryRules (только процессы + пользователь + активный профиль).
	StrictRoutingMode bool `json:"strictRoutingMode,omitempty" yaml:"strictRoutingMode,omitempty"`
	// LocalInboundUser — логин для SOCKS/HTTP на 127.0.0.1 (mixed inbound sing-box); пусто → pulse.
	LocalInboundUser string `json:"localInboundUser,omitempty" yaml:"localInboundUser,omitempty"`
	// LocalInboundPassword — пароль для локального mixed inbound и секрет Clash API ядра; генерируется при первом запуске.
	LocalInboundPassword string `json:"localInboundPassword,omitempty" yaml:"localInboundPassword,omitempty"`
}

// DpiListIsEnabled возвращает true, если список zapret lists/<name>.txt не помечен как выключенный.
func DpiListIsEnabled(c *AppConfig, base string) bool {
	if c == nil || c.DpiListEnabled == nil || len(c.DpiListEnabled) == 0 {
		return true
	}
	if v, ok := c.DpiListEnabled[base]; ok {
		return v
	}
	return true
}

func (c AppConfig) EffectiveServers() []VlessServer {
	seen := make(map[string]struct{})
	out := make([]VlessServer, 0, len(c.VlessServers)+len(c.Servers))
	for _, s := range c.VlessServers {
		if s.ID == "" {
			continue
		}
		if _, ok := seen[s.ID]; ok {
			continue
		}
		seen[s.ID] = struct{}{}
		out = append(out, s)
	}
	for _, s := range c.Servers {
		if s.ID == "" {
			continue
		}
		if _, ok := seen[s.ID]; ok {
			continue
		}
		seen[s.ID] = struct{}{}
		out = append(out, VlessServer(s))
	}
	if len(out) == 0 {
		return nil
	}
	return out
}

func (c AppConfig) FindActiveServer() (VlessServer, bool) {
	if c.ActiveServerID == "" {
		return VlessServer{}, false
	}
	for _, s := range c.EffectiveServers() {
		if s.ID == c.ActiveServerID {
			return s, true
		}
	}
	return VlessServer{}, false
}

// ListenPort returns mixed proxy port (default 8080).
func (c AppConfig) ListenPort() int {
	if c.ProxyPort >= 1 && c.ProxyPort <= 65535 {
		return c.ProxyPort
	}
	return 8080
}

// LocalProxyAuth возвращает учётные данные локального mixed inbound / Clash API.
func (c AppConfig) LocalProxyAuth() (user, password string) {
	user = strings.TrimSpace(c.LocalInboundUser)
	if user == "" {
		user = DefaultLocalInboundUser
	}
	password = strings.TrimSpace(c.LocalInboundPassword)
	return user, password
}

// LicenseActive — авторская подписка premium/supporter и не истёк срок (ExpiresUnix>now).
func (c AppConfig) LicenseActive() bool {
	t := strings.ToLower(strings.TrimSpace(c.LicenseTier))
	if t != "premium" && t != "supporter" {
		return false
	}
	if c.LicenseExpiresUnix <= 0 {
		return true
	}
	return time.Now().Unix() < c.LicenseExpiresUnix
}

// Имена двух режимов маршрутизации (дублируют config-константы,
// чтобы models не зависел от config).
const (
	RoutingModeFullProxy = "full-proxy"
	RoutingModeManual    = "manual"
)

// ResolveRoutingProfileAlias приводит сохранённое имя профиля к одному из двух режимов.
// Старые имена (default, all-proxy, rf-smart-split, selective-proxy, gaming-friendly,
// all-direct, bypass-russia) остались в конфигах прошлых версий и должны продолжать работать.
func ResolveRoutingProfileAlias(name string) string {
	switch strings.ToLower(strings.TrimSpace(name)) {
	case "", "default", "all-proxy", "full-proxy":
		return RoutingModeFullProxy
	case "all-direct", "bypass-russia", "bypass-rf", "rf-smart-split",
		"selective-proxy", "gaming-friendly", "manual":
		return RoutingModeManual
	default:
		return name
	}
}

// ActiveRoutingProfile возвращает выбранный профиль маршрутизации из RoutingProfiles
// с автоматическим разрешением старых имён на два актуальных режима.
func (c AppConfig) ActiveRoutingProfile() (RoutingProfile, bool) {
	name := strings.TrimSpace(c.RoutingProfile)
	if name == "" && len(c.RoutingProfiles) == 0 {
		return RoutingProfile{}, false
	}

	resolved := ResolveRoutingProfileAlias(name)

	// 1. Поиск по разрешённому имени.
	for _, rp := range c.RoutingProfiles {
		if strings.EqualFold(rp.Name, resolved) {
			return rp, true
		}
	}

	// 2. Поиск по прямому имени (профиль сохранён с точным именем).
	if name != "" && !strings.EqualFold(name, resolved) {
		for _, rp := range c.RoutingProfiles {
			if strings.EqualFold(rp.Name, name) {
				return rp, true
			}
		}
	}

	// 3. Fallback: синтетический профиль, если в списке его нет.
	switch resolved {
	case RoutingModeManual:
		return RoutingProfile{
			Name:  RoutingModeManual,
			Final: "direct",
			Rules: []RouteRule{
				{Name: "local-direct", Domains: []string{"*.local", "localhost", "*.lan"}, Action: RouteDirect},
				{Name: "private-lan-direct", CIDRs: []string{"10.0.0.0/8", "172.16.0.0/12", "192.168.0.0/16", "127.0.0.0/8"}, Action: RouteDirect},
			},
		}, true
	case RoutingModeFullProxy:
		return RoutingProfile{
			Name:  RoutingModeFullProxy,
			Final: "proxy",
			Rules: []RouteRule{
				{Name: "local-direct", Domains: []string{"*.local", "localhost", "*.lan"}, Action: RouteDirect},
				{Name: "private-lan-direct", CIDRs: []string{"10.0.0.0/8", "172.16.0.0/12", "192.168.0.0/16", "127.0.0.0/8"}, Action: RouteDirect},
			},
		}, true
	}

	if len(c.RoutingProfiles) > 0 {
		return c.RoutingProfiles[0], true
	}

	return RoutingProfile{}, false
}

// ServerMatchesAuthorPremium — узел помечен как авторский Premium по подстрокам имени/хоста.
func ServerMatchesAuthorPremium(s VlessServer, markers []string) bool {
	for _, m := range markers {
		m = strings.TrimSpace(strings.ToLower(m))
		if m == "" {
			continue
		}
		n := strings.ToLower(s.Name)
		h := strings.ToLower(s.Host)
		if strings.Contains(n, m) || strings.Contains(h, m) {
			return true
		}
	}
	return false
}
