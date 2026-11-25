package config

import "pulseshield/app/models"

// CurrentRoutingBuiltinRev увеличивайте при изменении builtinRoutingPresets, чтобы подтянуть профили у существующих установок.
const CurrentRoutingBuiltinRev = 11

// XboxDnsFriendlyDomainSuffixes — сервисы, для которых на https://xbox-dns.ru/ заявлен Smart DNS без VPN
// (Xbox Live, игры в т.ч. Supercell/Bungie, ChatGPT/Claude/Gemini/Copilot, JetBrains/GitHub/Notion/Framer, Twitch/Spotify/xCloud).
// Правило direct (dpi-bypass) стоит до VLESS — не дублировать суффиксы в proxy-only и DPI.
func XboxDnsFriendlyDomainSuffixes() []string {
	return []string{
		// Xbox Live, Store, xCloud
		"xbox.com",
		"xboxlive.com",
		"xboxservices.com",
		"xboxgamepass.com",
		"microsoftstore.com",
		"msftconnecttest.com",
		"copilot.microsoft.com",
		// Игры (сайт: Destiny 2, Marathon, Stumble Guys, Supercell — Brawl Stars, Clash of Clans и др.)
		"bungie.net",
		"supercell.com",
		"clashofclans.com",
		"brawlstars.com",
		"clashroyale.com",
		"haydaygame.com",
		"boombeach.com",
		"stumbleguys.com",
		"kitkagames.com",
		// AI: ChatGPT, Claude, Gemini, Copilot (+ Cursor как IDE в той же «dev» ветке маршрута)
		"openai.com",
		"chatgpt.com",
		"oaistatic.com",
		"oaiusercontent.com",
		"anthropic.com",
		"claude.ai",
		"gemini.google.com",
		"ai.google.dev",
		// cursor.sh / cursor.com: браузер без VPN; сам Cursor IDE — см. CursorDesktopProxyRule (раньше этого списка).
		"cursor.sh",
		"cursor.com",
		// JetBrains, GitHub / Copilot, Notion, Framer
		"jetbrains.com",
		"jetbrains.space",
		"github.com",
		"githubusercontent.com",
		"githubcopilot.com",
		"notion.so",
		"notion.com",
		"framer.com",
		"framer.website",
		// Twitch, Spotify
		"spotify.com",
		"scdn.co",
		"spotifycdn.com",
		"twitch.tv",
		"ttvnw.net",
		"jtvnw.net",
	}
}

// RuDirectDomainSuffixes — всё, что должно идти мимо VLESS для российского сегмента по TLD/кириллице.
func RuDirectDomainSuffixes() []string {
	return []string{
		"ru",
		"xn--p1ai",
		"su",
		"xn--h1alffa9f",
		"xn--80adxhks",
	}
}

// RuDirectDomains — отдельные российские сервисы и госдомены без невалидных wildcard-масок.
func RuDirectDomains() []string {
	return []string{
		"gov.ru",
		"kremlin.ru",
		"nalog.ru",
		"gosuslugi.ru",
		"yandex.ru",
		"yandex.net",
		"ya.ru",
		"vk.com",
		"vk.ru",
		"vk-cdn.net",
		"mail.ru",
		"ok.ru",
		"dzen.ru",
		"rutube.ru",
		"kinopoisk.ru",
		"avito.ru",
		"sberbank.ru",
		"sber.ru",
		"tbank.ru",
		"ozon.ru",
		"wildberries.ru",
	}
}

// RuDirectExtraSuffixes — суффиксы российских сервисов вне зоны .ru для domain_suffix в sing-box.
func RuDirectExtraSuffixes() []string {
	return []string{
		"vk.com",
		"vk-cdn.net",
		"yandex.net",
	}
}

// AIServiceDomainSuffixes — полный список доменов популярных нейросетей (ChatGPT, Claude, Gemini, Grok и др.),
// которым необходим зарубежный IP-адрес через прокси.
func AIServiceDomainSuffixes() []string {
	return []string{
		"openai.com",
		"chatgpt.com",
		"oaistatic.com",
		"oaiusercontent.com",
		"sora.com",
		"anthropic.com",
		"claude.ai",
		"claudeusercontent.com",
		"gemini.google.com",
		"aistudio.google.com",
		"makersuite.google.com",
		"ai.google.dev",
		"generativelanguage.googleapis.com",
		"notebooklm.google.com",
		"deepmind.google",
		"labs.google",
		"alkalimakersuite-pa.clients6.google.com",
		"grok.com",
		"x.ai",
		"copilot.microsoft.com",
		"perplexity.ai",
		"mistral.ai",
		"replicate.com",
		"huggingface.co",
		"openrouter.ai",
		"poe.com",
		"midjourney.com",
		"cursor.sh",
		"cursor.com",
	}
}

// ProxyOnlyMediaDomainSuffixes — мессенджеры, которых нет на xbox-dns.ru и которым нужен полный VLESS (MTProto, DC по IP).
// Telegram / WhatsApp / Signal / Messenger сюда намеренно: Smart DNS их не «чинит» в правилах маршрутизации — трафик идёт в outbound proxy, не в xbox-dns-direct.
func ProxyOnlyMediaDomainSuffixes() []string {
	return []string{
		"telegram.org",
		"t.me",
		"telegram.me",
		"telesco.pe",
		"telegram-cdn.org",
		"whatsapp.com",
		"whatsapp.net",
		"signal.org",
		"messenger.com",
	}
}

// DPIBypassMediaDomainSuffixes — веб/медиа: DPI + direct, без VLESS; не дублировать Xbox и proxy-only.
func DPIBypassMediaDomainSuffixes() []string {
	return []string{
		"youtube.com",
		"youtu.be",
		"ytimg.com",
		"ggpht.com",
		"google.com",
		"gstatic.com",
		"googleapis.com",
		"gvt1.com",
		"googleusercontent.com",
		"googlevideo.com",
		"twitter.com",
		"twimg.com",
		"x.com",
		"instagram.com",
		"facebook.com",
		"fbcdn.net",
		"reddit.com",
		"redd.it",
		"wikipedia.org",
		"medium.com",
		"discord.com",
		"discord.gg",
		"discordapp.com",
		"discord.media",
		"discord.co",
		"discordstatus.com",
		"slack.com",
		"slack-edge.com",
		"slack-msgs.com",
		"line.me",
		"line-apps.com",
		"viber.com",
		"snapchat.com",
		"sc-cdn.net",
		"zoom.us",
		"zoom.com",
		"zoomgov.com",
		"webex.com",
		"webexcontent.com",
		"teams.microsoft.com",
		"teams.live.com",
		"skype.com",
		"skypeassets.com",
		"element.io",
		"matrix.org",
		"kick.com",
		"proton.me",
		"pm.me",
	}
}

// SelectiveMediaDomainSuffixes — объединение proxy+DPI для selective / gaming (Xbox — отдельным правилом).
func SelectiveMediaDomainSuffixes() []string {
	return mergeUniqueStringSlices(ProxyOnlyMediaDomainSuffixes(), DPIBypassMediaDomainSuffixes())
}

func ProcessProxyRules(cfg *models.AppConfig) []models.RouteRule {
	processes := make([]string, 0)
	if cfg != nil {
		for name, enabled := range cfg.ProcessProxyMap {
			if enabled && name != "" {
				processes = append(processes, name)
			}
		}
	}
	if len(processes) == 0 {
		processes = []string{
			"Telegram.exe", "telegram.exe",
			"Cursor.exe", "cursor.exe",
			"Discord.exe", "discord.exe",
			"steam.exe", "steamwebhelper.exe",
		}
	}
	return []models.RouteRule{
		{
			Name:      "process-proxy-map",
			Processes: processes,
			Action:    models.RouteProxy,
		},
	}
}

// TelegramDesktopProxyRule — MTProto/DC по IP не попадают в domain_suffix.
func TelegramDesktopProxyRule() models.RouteRule {
	return models.RouteRule{
		Name:      "telegram-desktop-proxy",
		Processes: []string{"Telegram.exe", "telegram.exe"},
		Action:    models.RouteProxy,
	}
}

// CursorDesktopProxyRule — трафик без подходящего SNI (остальное по доменам — в xbox-dns-direct).
func CursorDesktopProxyRule() models.RouteRule {
	return models.RouteRule{
		Name:      "cursor-desktop-proxy",
		Processes: []string{"Cursor.exe", "cursor.exe"},
		Action:    models.RouteProxy,
	}
}

// XboxDnsDirectRule — direct до proxy-only (экономия VLESS при пресете Xbox DNS).
func XboxDnsDirectRule() models.RouteRule {
	return models.RouteRule{
		Name:           "xbox-dns-direct",
		DomainSuffixes: XboxDnsFriendlyDomainSuffixes(),
		Action:         models.RouteDPIBypass,
	}
}

func mergeUniqueStringSlices(a, b []string) []string {
	seen := make(map[string]struct{}, len(a)+len(b))
	out := make([]string, 0, len(a)+len(b))
	for _, s := range append(append([]string{}, a...), b...) {
		if s == "" {
			continue
		}
		if _, ok := seen[s]; ok {
			continue
		}
		seen[s] = struct{}{}
		out = append(out, s)
	}
	return out
}

// ApplyRoutingBuiltinUpgrade подменяет встроенные профили из builtinRoutingPresets() при устаревшей RoutingBuiltinRev.
func ApplyRoutingBuiltinUpgrade(cfg *models.AppConfig) bool {
	if cfg == nil || cfg.RoutingBuiltinRev >= CurrentRoutingBuiltinRev {
		return false
	}
	if cfg.ProcessProxyMap != nil {
		// Убираем навязанное автопроксирование Steam из прошлых ревизий (rev 8): Steam/Discord/YouTube обрабатываются через DPI.
		delete(cfg.ProcessProxyMap, "steam.exe")
		delete(cfg.ProcessProxyMap, "Steam.exe")
		delete(cfg.ProcessProxyMap, "steamwebhelper.exe")
	}
	byName := map[string]models.RoutingProfile{}
	for _, p := range builtinRoutingPresets() {
		byName[p.Name] = p
	}
	// Синхронизируем правила у профилей двух актуальных режимов.
	for i := range cfg.RoutingProfiles {
		if np, ok := byName[cfg.RoutingProfiles[i].Name]; ok {
			cfg.RoutingProfiles[i] = np
		}
	}
	cfg.RoutingBuiltinRev = CurrentRoutingBuiltinRev
	return true
}
