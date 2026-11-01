package multiserver

import (
	"context"
	"crypto/sha1"
	"crypto/tls"
	"encoding/hex"
	"fmt"
	"io"
	"math"
	"net"
	"net/http"
	"net/url"
	"regexp"
	"sort"
	"strconv"
	"strings"
	"sync"
	"time"

	"pulseshield/app/models"
)

const userAgent = "PulseShield/0.6.0"

// SubscriptionMeta — метаданные из ответа провайдера (часто в заголовке Subscription-Userinfo).
type SubscriptionMeta struct {
	ExpiresUnix   int64
	ProviderLabel string
	UploadBytes   int64
	DownloadBytes int64
	TotalBytes    int64
	// Format — распознанный формат тела подписки (plain-uri / base64 / clash-yaml / sing-box-json / v2ray-json).
	Format string
	// Name — имя профиля из заголовка Profile-Title (если провайдер его отдаёт).
	Name string
}

var (
	reUserinfoExpire   = regexp.MustCompile(`(?i)\bexpire\s*=\s*(\d+)`)
	reUserinfoUpload   = regexp.MustCompile(`(?i)\bupload\s*=\s*(\d+)`)
	reUserinfoDownload = regexp.MustCompile(`(?i)\bdownload\s*=\s*(\d+)`)
	reUserinfoTotal    = regexp.MustCompile(`(?i)\btotal\s*=\s*(\d+)`)
)

func parseSubscriptionUserinfoHeader(h string) (meta SubscriptionMeta) {
	h = strings.TrimSpace(h)
	if h == "" {
		return meta
	}
	parseInto := func(re *regexp.Regexp, dst *int64) {
		if m := re.FindStringSubmatch(h); len(m) == 2 {
			if v, err := strconv.ParseInt(m[1], 10, 64); err == nil {
				*dst = v
			}
		}
	}
	parseInto(reUserinfoExpire, &meta.ExpiresUnix)
	parseInto(reUserinfoUpload, &meta.UploadBytes)
	parseInto(reUserinfoDownload, &meta.DownloadBytes)
	parseInto(reUserinfoTotal, &meta.TotalBytes)
	return meta
}

// decodeProfileTitle разбирает заголовок Profile-Title: "base64:...." или обычная строка.
func decodeProfileTitle(h string) string {
	h = strings.TrimSpace(h)
	if h == "" {
		return ""
	}
	if strings.HasPrefix(strings.ToLower(h), "base64:") {
		if dec, ok := decodeBase64Loose(h[len("base64:"):]); ok {
			return strings.TrimSpace(dec)
		}
	}
	return h
}

func providerLabelFromURL(raw string) string {
	u, err := url.Parse(strings.TrimSpace(raw))
	if err != nil || u.Host == "" {
		return ""
	}
	host := u.Hostname()
	if host == "" {
		return ""
	}
	return host
}

// FetchSubscription загружает список прокси-ссылок (любой поддерживаемый формат) без метаданных.
func FetchSubscription(rawURL string) ([]string, error) {
	links, _, err := FetchSubscriptionWithMeta(rawURL)
	return links, err
}

// FetchSubscriptionWithMeta загружает подписку напрямую или через локальный прокси 127.0.0.1:8080.
func FetchSubscriptionWithMeta(rawURL string) ([]string, SubscriptionMeta, error) {
	return FetchSubscriptionWithProxy(rawURL, 8080)
}

// normalizeSubscriptionURL приводит пользовательский ввод к пригодному для запроса URL:
// добавляет https:// если схема не указана и отбрасывает мусор вокруг адреса.
func normalizeSubscriptionURL(raw string) string {
	s := strings.TrimSpace(raw)
	s = strings.Trim(s, "\"'`<>")
	s = strings.ReplaceAll(s, "\r", "")
	s = strings.ReplaceAll(s, "\n", "")
	if i := strings.IndexAny(s, " \t"); i > 0 {
		s = s[:i]
	}
	if s == "" {
		return ""
	}
	low := strings.ToLower(s)
	if strings.HasPrefix(low, "http://") || strings.HasPrefix(low, "https://") {
		return s
	}
	// Пользователь мог вставить "sub.example.com/api/v1/client/subscribe?token=..."
	if !strings.Contains(s, "://") {
		return "https://" + strings.TrimPrefix(s, "//")
	}
	return s
}

// FetchSubscriptionWithProxy пытается скачать подписку, обходя блокировки ТСПУ:
// 1. Сначала пробует через локальный прокси 127.0.0.1:<proxyPort> (если туннель уже поднят).
// 2. Пробует как оригинальный URL, так и HTTPS-апгрейд (на случай блокировки plaintext HTTP по DPI).
// 3. Fallback — прямой запрос мимо прокси.
//
// Формат тела определяется автоматически (plain / base64 / Clash YAML / sing-box JSON / Xray JSON).
func FetchSubscriptionWithProxy(rawURL string, localProxyPort int) ([]string, SubscriptionMeta, error) {
	var meta SubscriptionMeta
	subscriptionURL := normalizeSubscriptionURL(rawURL)
	if subscriptionURL == "" {
		return nil, meta, fmt.Errorf("subscription URL is empty")
	}
	meta.ProviderLabel = providerLabelFromURL(subscriptionURL)

	candidates := []string{subscriptionURL}
	if strings.HasPrefix(strings.ToLower(subscriptionURL), "http://") {
		httpsURL := "https://" + subscriptionURL[len("http://"):]
		candidates = append(candidates, httpsURL)
	}

	var clients []*http.Client
	if localProxyPort >= 1 && localProxyPort <= 65535 {
		proxyAddr := fmt.Sprintf("127.0.0.1:%d", localProxyPort)
		if conn, err := net.DialTimeout("tcp", proxyAddr, 350*time.Millisecond); err == nil {
			_ = conn.Close()
			if pu, err := url.Parse("http://" + proxyAddr); err == nil {
				tr := &http.Transport{
					Proxy:           http.ProxyURL(pu),
					TLSClientConfig: &tls.Config{InsecureSkipVerify: true},
				}
				clients = append(clients, &http.Client{Timeout: 10 * time.Second, Transport: tr})
			}
		}
	}
	directTr := &http.Transport{
		Proxy:           nil, // всегда напрямую, игнорируя битый системный прокси если ядро выключено
		TLSClientConfig: &tls.Config{InsecureSkipVerify: true},
	}
	clients = append(clients, &http.Client{Timeout: 8 * time.Second, Transport: directTr})

	var lastErr error
	for _, client := range clients {
		for _, targetURL := range candidates {
			req, err := http.NewRequest(http.MethodGet, targetURL, nil)
			if err != nil {
				lastErr = err
				continue
			}
			req.Header.Set("User-Agent", userAgent)
			// Клиенты подписок (Happ / v2rayN / Clash) шлют эти заголовки — многие панели
			// без них отдают 404 или HTML-страницу вместо конфига.
			req.Header.Set("Accept", "*/*")
			req.Header.Set("Accept-Language", "ru-RU,ru;q=0.9,en;q=0.8")
			resp, err := client.Do(req)
			if err != nil {
				lastErr = err
				continue
			}
			body, readErr := io.ReadAll(io.LimitReader(resp.Body, 8<<20))
			hUserinfo := resp.Header.Get("Subscription-Userinfo")
			hProfileTitle := resp.Header.Get("Profile-Title")
			_ = resp.Body.Close()
			if resp.StatusCode < 200 || resp.StatusCode >= 300 {
				lastErr = fmt.Errorf("subscription request failed: %s", resp.Status)
				continue
			}
			if readErr != nil {
				lastErr = readErr
				continue
			}
			contentType := strings.ToLower(resp.Header.Get("Content-Type"))
			if strings.Contains(contentType, "text/html") && !looksLikeClashYAML(string(body)) {
				// Панель отдала HTML-заглушку (страница входа/ошибки) — это не конфиг.
				lastErr = fmt.Errorf("провайдер вернул HTML-страницу вместо подписки (проверьте ссылку)")
				continue
			}
			links, fmtName, decErr := parseSubscriptionLinksWithFormat(string(body))
			if decErr != nil {
				lastErr = decErr
				continue
			}
			if len(links) == 0 {
				lastErr = fmt.Errorf("empty subscription payload")
				continue
			}
			meta.Format = fmtName
			if hUserinfo != "" {
				m := parseSubscriptionUserinfoHeader(hUserinfo)
				if m.ExpiresUnix != 0 {
					meta.ExpiresUnix = m.ExpiresUnix
				}
				meta.UploadBytes = m.UploadBytes
				meta.DownloadBytes = m.DownloadBytes
				meta.TotalBytes = m.TotalBytes
			}
			if name := decodeProfileTitle(hProfileTitle); name != "" {
				meta.Name = name
				meta.ProviderLabel = name
			}
			return links, meta, nil
		}
	}
	if lastErr == nil {
		lastErr = fmt.Errorf("failed to fetch subscription")
	}
	return nil, meta, lastErr
}

// ParseVlessString разбирает прокси-ссылку. Историческое имя сохранено ради совместимости,
// но функция понимает все поддерживаемые протоколы (vless/vmess/trojan/ss/hysteria2/tuic/anytls/socks/http).
func ParseVlessString(vlessURL string) (*models.VlessServer, error) {
	return ParseProxyLink(vlessURL)
}

var probeHostIPCache sync.Map

func resolveProbeHostIP(host string) string {
	h := strings.TrimSpace(host)
	if h == "" || net.ParseIP(h) != nil {
		return h
	}
	lh := strings.ToLower(h)
	if cached, ok := probeHostIPCache.Load(lh); ok {
		return cached.(string)
	}
	pickIPv4 := func(ips []net.IP) string {
		for _, ip := range ips {
			if v4 := ip.To4(); v4 != nil && !v4.IsLoopback() && !v4.IsUnspecified() {
				return v4.String()
			}
		}
		if len(ips) > 0 {
			return ips[0].String()
		}
		return ""
	}
	ctxSys, cancelSys := context.WithTimeout(context.Background(), 1200*time.Millisecond)
	ips, err := net.DefaultResolver.LookupIP(ctxSys, "ip4", h)
	cancelSys()
	if err == nil {
		if resolved := pickIPv4(ips); resolved != "" {
			probeHostIPCache.Store(lh, resolved)
			return resolved
		}
	}
	for _, dnsAddr := range []string{"8.8.8.8:53", "1.1.1.1:53", "77.88.8.8:53"} {
		target := dnsAddr
		r := &net.Resolver{
			PreferGo: true,
			Dial: func(ctx context.Context, network, address string) (net.Conn, error) {
				d := net.Dialer{Timeout: 900 * time.Millisecond}
				return d.DialContext(ctx, "udp", target)
			},
		}
		ctx, cancel := context.WithTimeout(context.Background(), 1000*time.Millisecond)
		ips, err := r.LookupIP(ctx, "ip4", h)
		cancel()
		if err == nil {
			if resolved := pickIPv4(ips); resolved != "" {
				probeHostIPCache.Store(lh, resolved)
				return resolved
			}
		}
	}
	return h
}

// ProbeServerMetrics проверяет TCP-рукопожатие и пинг к узлу БЕЗ отправки открытого TLS SNI srv.Host,
// чтобы ТСПУ провайдера (например «Подряд») не блокировал IP узла по незамаскированному SNI.
func ProbeServerMetrics(srv *models.VlessServer, timeout time.Duration) {
	if srv == nil || strings.TrimSpace(srv.Host) == "" {
		return
	}
	host := strings.TrimSpace(srv.Host)
	if host == "127.0.0.1" || host == "0.0.0.0" || srv.Port == "1" || srv.Port == "0" {
		srv.LastCheckOK = false
		return
	}
	netType := strings.ToLower(strings.TrimSpace(srv.Params["type"]))
	if netType == "xhttp" || netType == "splithttp" {
		srv.LastCheckOK = false
		return
	}
	if timeout <= 0 {
		timeout = 2500 * time.Millisecond
	}
	dialHost := resolveProbeHostIP(host)
	dialPort := strings.TrimSpace(srv.Port)
	if dialPort == "" {
		dialPort = "443"
	}
	// Hysteria2 работает по UDP (например порт 36712), поэтому для оценки доступности хоста и RTT
	// проверяем TCP 443 на том же IP-адресе сервера.
	if srv.Params["_proto"] == "hysteria2" {
		dialPort = "443"
	}

	var conn net.Conn
	var err error
	var latMs int
	for attempt := 0; attempt < 2; attempt++ {
		start := time.Now()
		conn, err = net.DialTimeout("tcp", net.JoinHostPort(dialHost, dialPort), timeout)
		if err == nil {
			latMs = int(time.Since(start).Milliseconds())
			_ = conn.Close()
			break
		}
	}
	srv.LastCheckedAt = time.Now()
	if err != nil {
		srv.LastCheckOK = false
		return
	}
	if latMs <= 0 {
		latMs = 1
	}
	srv.LastCheckOK = true
	srv.LastCheckLatencyMs = latMs
	srv.LastCheckDownloadMbps = roundMbps(latencyAsMbpsScore(latMs))
}

func ValidateAndTestServers(servers []*models.VlessServer, timeout time.Duration) []*models.VlessServer {
	if timeout <= 0 {
		timeout = 2500 * time.Millisecond
	}
	valid := make([]*models.VlessServer, 0, len(servers))
	var mu sync.Mutex
	var wg sync.WaitGroup
	sem := make(chan struct{}, 24)
	for i := range servers {
		srv := servers[i]
		if srv == nil || strings.TrimSpace(srv.Host) == "" {
			continue
		}
		if srv.Host == "127.0.0.1" || srv.Host == "0.0.0.0" || srv.Port == "1" || srv.Port == "0" {
			continue
		}
		netType := strings.ToLower(strings.TrimSpace(srv.Params["type"]))
		if netType == "xhttp" || netType == "splithttp" {
			continue
		}
		wg.Add(1)
		go func(s *models.VlessServer) {
			defer wg.Done()
			sem <- struct{}{}
			defer func() { <-sem }()
			ProbeServerMetrics(s, timeout)
			if !s.LastCheckOK {
				return
			}
			mu.Lock()
			valid = append(valid, s)
			mu.Unlock()
		}(srv)
	}
	wg.Wait()
	markRecommended(valid)
	return valid
}

func latencyAsMbpsScore(latMs int) float64 {
	if latMs <= 0 {
		return 0
	}
	ms := float64(latMs)
	return math.Min(350, 10000/(ms+8))
}

func roundMbps(v float64) float64 {
	if v <= 0 {
		return 0
	}
	return math.Round(v*10) / 10
}

func splitVlessLines(payload string) []string {
	return splitProxyLines(payload)
}

// parseSubscriptionLinksWithFormat определяет формат тела подписки и возвращает ссылки + имя формата.
func parseSubscriptionLinksWithFormat(body string) ([]string, string, error) {
	links, err := ExtractSubscriptionLinks(body)
	if err != nil {
		return nil, "", err
	}
	trimmed := strings.TrimSpace(body)
	format := "plain-uri"
	switch {
	case strings.HasPrefix(trimmed, "{") || strings.HasPrefix(trimmed, "["):
		format = "sing-box-json"
	case looksLikeClashYAML(trimmed):
		format = "clash-yaml"
	default:
		if len(splitProxyLines(trimmed)) == 0 {
			format = "base64"
		}
	}
	return links, format, nil
}

// decodeSubscriptionPayload оставлен для обратной совместимости (используется в тестах):
// снимает base64-обёртку или возвращает исходный plain-текст.
func decodeSubscriptionPayload(payload string) (string, error) {
	trimmed := strings.TrimSpace(payload)
	if trimmed == "" {
		return "", fmt.Errorf("subscription payload is empty")
	}
	if len(splitProxyLines(trimmed)) > 0 {
		return trimmed, nil
	}
	if dec, ok := decodeSubscriptionBody(trimmed); ok {
		return dec, nil
	}
	return "", fmt.Errorf("subscription payload is not valid base64")
}

func serverID(raw, host, port string) string {
	sum := sha1.Sum([]byte(strings.TrimSpace(strings.ToLower(raw))))
	return fmt.Sprintf("%s_%s_%s", hex.EncodeToString(sum[:8]), strings.ReplaceAll(host, ":", "_"), port)
}

func serverTransportTier(s *models.VlessServer) int {
	if s == nil {
		return 9
	}
	proto := strings.ToLower(strings.TrimSpace(s.Params["_proto"]))
	sec := strings.ToLower(strings.TrimSpace(s.Params["security"]))
	netType := strings.ToLower(strings.TrimSpace(s.Params["type"]))
	hostLow := strings.ToLower(strings.TrimSpace(s.Host))
	nameLow := strings.ToLower(strings.TrimSpace(s.Name))
	isRU := strings.HasPrefix(hostLow, "ru") || strings.Contains(nameLow, "россия") || strings.Contains(nameLow, "russia")
	if netType == "ws" {
		if isRU {
			return 6
		}
		return 4 // WebSocket чаще даёт HTTP 404 на CDN и легче режется ТСПУ
	}
	if sec == "reality" || proto == "hysteria2" {
		if isRU {
			return 3
		}
		return 1 // Зарубежные Reality и Hysteria2 Salamander наиболее устойчивы к ТСПУ и открывают все сервисы
	}
	if isRU {
		return 5
	}
	return 2
}

func markRecommended(servers []*models.VlessServer) {
	for _, srv := range servers {
		srv.Recommended = false
	}
	sort.SliceStable(servers, func(i, j int) bool {
		if servers[i].LastCheckOK != servers[j].LastCheckOK {
			return servers[i].LastCheckOK
		}
		ti, tj := serverTransportTier(servers[i]), serverTransportTier(servers[j])
		if ti != tj {
			return ti < tj
		}
		if servers[i].LastCheckLatencyMs == servers[j].LastCheckLatencyMs {
			return servers[i].LastCheckDownloadMbps > servers[j].LastCheckDownloadMbps
		}
		if servers[i].LastCheckLatencyMs == 0 {
			return false
		}
		if servers[j].LastCheckLatencyMs == 0 {
			return true
		}
		return servers[i].LastCheckLatencyMs < servers[j].LastCheckLatencyMs
	})
	for i := 0; i < len(servers) && i < 3; i++ {
		if servers[i].LastCheckOK {
			servers[i].Recommended = true
		}
	}
}
