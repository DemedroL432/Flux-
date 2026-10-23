package dpi

import (
	"fmt"
	"log"
	"os"
	"path/filepath"
	"sort"
	"strings"
	"unicode"

	"pulseshield/app/models"
)

// dpiListDescriptions — человекочитаемые подписи для известных lists/*.txt (ключ без .txt).
var dpiListDescriptions = map[string]string{
	"youtube":                  "YouTube (hostlist)",
	"youtube_v2":               "YouTube v2",
	"youtubeGV":                "YouTube Google Video",
	"youtubeQ":                 "YouTube качество / CDN",
	"russia-youtube":           "YouTube (RU)",
	"russia-youtubeQ":          "YouTube Q (RU)",
	"russia-youtube-rtmps":     "YouTube RTMPS (RU)",
	"ipset-youtube":            "YouTube IP (ipset)",
	"ipset-russia-youtube":     "YouTube IP (RU)",
	"ipset-googlevideo":        "Google Video / CDN",
	"discord":                  "Discord (hostlist)",
	"russia-discord":           "Discord (RU)",
	"russia-discord-ipset":     "Discord IP (RU)",
	"ipset-discord":            "Discord IP",
	"telegram":                 "Telegram (hostlist)",
	"ipset-telegram":           "Telegram IP",
	"instagram":                "Instagram (hostlist)",
	"ipset-instagram":          "Instagram IP",
	"facebook":                 "Facebook (hostlist)",
	"ipset-facebook":           "Facebook IP",
	"twitter":                  "X/Twitter (hostlist)",
	"ipset-twitter":            "X/Twitter IP",
	"twitch":                   "Twitch (hostlist)",
	"ipset-twitch":             "Twitch IP",
	"github":                   "GitHub (hostlist)",
	"ipset-github":             "GitHub IP",
	"google":                   "Google (hostlist)",
	"list-google":              "Google список",
	"cloudflare-ipset":         "Cloudflare IPv4",
	"cloudflare-ipset_v6":      "Cloudflare IPv6",
	"ipset-cloudflare":         "Cloudflare IP (набор)",
	"ipset-cloudflare1":        "Cloudflare IP (набор 2)",
	"steam":                    "Steam / сообщество (hostlist)",
	"ipset-steam":              "Steam IP",
	"roblox":                   "Roblox (hostlist)",
	"ipset-roblox":             "Roblox IP",
	"ipset-riot-games":         "Riot Games / Valorant",
	"ipset-lol-ru":             "League of Legends RU",
	"ipset-lol-euw":            "League EUW",
	"whatsapp":                 "WhatsApp (hostlist)",
	"ipset-whatsapp":           "WhatsApp IP",
	"amazon":                   "Amazon (hostlist)",
	"ipset-amazon":             "Amazon IP",
	"netrogat":                 "Исключения / «не трогать»",
	"other":                    "Прочие сервисы",
	"other2":                   "Прочие сервисы (2)",
	"list-general":             "Общий hostlist",
	"list-exclude":             "Исключения hostlist",
	"ipset-exclude":            "Исключения IP",
	"ipset-all":                "Большой общий ipset",
	"ipset-base":               "Базовый ipset",
	"ipset-ru":                 "Российские IP (исключения)",
	"ipset-dns":                "DNS / резолверы (исключение)",
	"russia-blacklist":         "RU blacklist hostlist",
	"soundcloud":               "SoundCloud",
	"ipset-soundcloud":         "SoundCloud IP",
	"rutracker":                "RuTracker",
	"ipset-rutracker":          "RuTracker IP",
	"rutor":                    "Rutor",
	"itch":                     "itch.io",
	"porn":                     "Adult (hostlist)",
	"ipset-porn":               "Adult IP",
	"speedtest":                "Speedtest",
	"ipset-anydesk":            "AnyDesk IP",
	"ipset-censorliber":        "Censor liber",
	"ipset-claude":             "Claude / Anthropic CDN",
	"claude":                   "Claude (hostlist)",
	"ipset-ovh":                "OVH",
	"ipset-tankix":             "Tanki X",
	"ipset-warp":               "WARP",
	"ipset-zapretkvn":          "Zapret KVN",
	"ntcparty":                 "Ntc Party",
	"allzone":                  "All zone",
	"my-ipset":                 "Пользовательский ipset",
	"myhostlist":               "Пользовательский hostlist",
	"mycdnlist":                "Пользовательский CDN",
	"service_config":           "Служебный конфиг списков",
}

func describeDpiListFile(base string) string {
	if d, ok := dpiListDescriptions[base]; ok {
		return d
	}
	return humanizeFileBase(base)
}

func humanizeFileBase(base string) string {
	base = strings.ReplaceAll(base, "_", " ")
	base = strings.ReplaceAll(base, "-", " ")
	r := []rune(strings.ToLower(base))
	for i := range r {
		if i == 0 || r[i-1] == ' ' {
			r[i] = unicode.ToUpper(r[i])
		}
	}
	return string(r)
}

// GetAvailableDpiLists сканирует zapret/lists/*.txt и возвращает состояние из cfg.DpiListEnabled.
func GetAvailableDpiLists(cfg models.AppConfig) ([]DpiList, error) {
	zdir, err := ResolveZapretDir()
	if err != nil {
		log.Printf("[DPI] GetAvailableDpiLists: zapret: %v", err)
		return []DpiList{}, fmt.Errorf("zapret: %w", err)
	}
	ldir := filepath.Join(zdir, "lists")
	entries, err := os.ReadDir(ldir)
	if err != nil {
		log.Printf("[DPI] GetAvailableDpiLists: lists ReadDir %s: %v", ldir, err)
		return []DpiList{}, fmt.Errorf("lists недоступна (%s): %w", ldir, err)
	}
	var names []string
	for _, e := range entries {
		if e.IsDir() {
			continue
		}
		n := e.Name()
		if strings.HasSuffix(strings.ToLower(n), ".txt") {
			names = append(names, strings.TrimSuffix(n, ".txt"))
		}
	}
	sort.Strings(names)
	if len(names) == 0 {
		log.Printf("[DPI] GetAvailableDpiLists: в %s нет файлов .txt", ldir)
	}
	out := make([]DpiList, 0, len(names))
	for _, base := range names {
		out = append(out, DpiList{
			Name:        base,
			Description: describeDpiListFile(base),
			Enabled:     models.DpiListIsEnabled(&cfg, base),
		})
	}
	return out, nil
}

// ToggleDpiList обновляет карту выключенных списков (включение удаляет ключ).
func ToggleDpiList(cfg *models.AppConfig, name string, enabled bool) error {
	name = strings.TrimSpace(name)
	if name == "" {
		return fmt.Errorf("пустое имя списка")
	}
	if strings.Contains(name, "..") || strings.ContainsAny(name, `/\`) {
		return fmt.Errorf("недопустимое имя списка")
	}
	if cfg == nil {
		return fmt.Errorf("config is nil")
	}
	if cfg.DpiListEnabled == nil {
		cfg.DpiListEnabled = map[string]bool{}
	}
	if enabled {
		delete(cfg.DpiListEnabled, name)
		if len(cfg.DpiListEnabled) == 0 {
			cfg.DpiListEnabled = nil
		}
	} else {
		cfg.DpiListEnabled[name] = false
	}
	return nil
}
