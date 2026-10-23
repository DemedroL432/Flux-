package multiserver

import (
	"context"
	"crypto/sha1"
	"crypto/tls"
	"encoding/base64"
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
	ExpiresUnix     int64
	ProviderLabel   string
	UploadBytes     int64
	DownloadBytes   int64
	TotalBytes      int64
}

var reUserinfoExpire = regexp.MustCompile(`(?i)\bexpire\s*=\s*(\d+)`)

func parseSubscriptionUserinfoHeader(h string) (meta SubscriptionMeta) {
	h = strings.TrimSpace(h)
	if h == "" {
		return meta
	}
	if m := reUserinfoExpire.FindStringSubmatch(h); len(m) == 2 {
		if v, err := strconv.ParseInt(m[1], 10, 64); err == nil {
			meta.ExpiresUnix = v
		}
	}
	return meta
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

// FetchSubscription загружает список vless:// ссылок (без метаданных).
func FetchSubscription(rawURL string) ([]string, error) {
	links, _, err := FetchSubscriptionWithMeta(rawURL)
	return links, err
}

// FetchSubscriptionWithMeta загружает подписку напрямую или через локальный прокси 127.0.0.1:8080.
func FetchSubscriptionWithMeta(rawURL string) ([]string, SubscriptionMeta, error) {
	return FetchSubscriptionWithProxy(rawURL, 8080)
}

// FetchSubscriptionWithProxy пытается скачать подписку, обходя блокировки ТСПУ:
// 1. Сначала пробует через локальный прокси 127.0.0.1:<proxyPort> (если туннель уже поднят).
// 2. Пробует как оригинальный URL, так и HTTPS-апгрейд (на случай блокировки plaintext HTTP по DPI).
// 3. Fallback — прямой запрос мимо прокси.
func FetchSubscriptionWithProxy(rawURL string, localProxyPort int) ([]string, SubscriptionMeta, error) {
	var meta SubscriptionMeta
	subscriptionURL := strings.TrimSpace(rawURL)
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
			resp, err := client.Do(req)
			if err != nil {
				lastErr = err
				continue
			}
			body, readErr := io.ReadAll(io.LimitReader(resp.Body, 8<<20))
			hUserinfo := resp.Header.Get("Subscription-Userinfo")
			_ = resp.Body.Close()
			if resp.StatusCode < 200 || resp.StatusCode >= 300 {
				lastErr = fmt.Errorf("subscription request failed: %s", resp.Status)
				continue
			}
			if readErr != nil {
				lastErr = readErr
				continue
			}
			decoded, decErr := decodeSubscriptionPayload(string(body))
			if decErr != nil {
				lastErr = decErr
				continue
			}
			lines := splitVlessLines(decoded)
			if len(lines) == 0 {
				lastErr = fmt.Errorf("empty subscription payload")
				continue
			}
			if hUserinfo != "" {
				m := parseSubscriptionUserinfoHeader(hUserinfo)
				if m.ExpiresUnix != 0 {
					meta.ExpiresUnix = m.ExpiresUnix
				}
			}
			return lines, meta, nil
		}
	}
	if lastErr == nil {
		lastErr = fmt.Errorf("failed to fetch subscription")
	}
	return nil, meta, lastErr
}

func ParseVlessString(vlessURL string) (*models.VlessServer, error) {
	raw := strings.TrimSpace(vlessURL)
	low := strings.ToLower(raw)
	proto := ""
	switch {
	case strings.HasPrefix(low, "vless://"):
		proto = "vless"
	case strings.HasPrefix(low, "trojan://"):
		proto = "trojan"
	case strings.HasPrefix(low, "hysteria2://"), strings.HasPrefix(low, "hy2://"):
		proto = "hysteria2"
	default:
		return nil, fmt.Errorf("text is not a supported proxy link (vless/trojan/hysteria2)")
	}
	u, err := url.Parse(raw)
	if err != nil {
		return nil, err
	}
	host := strings.TrimSpace(u.Hostname())
	if host == "" || host == "127.0.0.1" || host == "0.0.0.0" || host == "localhost" {
		return nil, fmt.Errorf("proxy host is empty or dummy loopback")
	}
	port := u.Port()
	if port == "" {
		port = "443"
	}
	if port == "1" || port == "0" {
		return nil, fmt.Errorf("dummy info node port: %s", port)
	}
	uuid := u.User.Username()
	if uuid == "00000000-0000-0000-0000-000000000000" {
		return nil, fmt.Errorf("dummy info node uuid")
	}
	params := map[string]string{}
	for k, vs := range u.Query() {
		if len(vs) > 0 {
			params[k] = vs[0]
		}
	}
	params["_proto"] = proto
	if proto != "vless" && params["proto"] == "" {
		params["proto"] = proto
	}
	netType := strings.ToLower(strings.TrimSpace(params["type"]))
	if netType == "xhttp" || netType == "splithttp" {
		return nil, fmt.Errorf("unsupported Xray transport: %s", netType)
	}
	name := strings.TrimSpace(u.Fragment)
	if unesc, err := url.QueryUnescape(name); err == nil && unesc != "" {
		name = strings.TrimSpace(unesc)
	}
	if name == "" {
		name = host
	}
	srv := &models.VlessServer{
		ID:        serverID(raw, host, port),
		Name:      name,
		RawVLESS:  raw,
		Host:      host,
		Port:      port,
		UUID:      uuid,
		Params:    params,
		CreatedAt: time.Now(),
	}
	return srv, nil
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
	lines := strings.FieldsFunc(payload, func(r rune) bool {
		return r == '\n' || r == '\r'
	})
	out := make([]string, 0, len(lines))
	for _, line := range lines {
		line = strings.TrimSpace(line)
		low := strings.ToLower(line)
		if strings.HasPrefix(low, "vless://") ||
			strings.HasPrefix(low, "trojan://") ||
			strings.HasPrefix(low, "hysteria2://") ||
			strings.HasPrefix(low, "hy2://") {
			out = append(out, line)
		}
	}
	return out
}

func decodeSubscriptionPayload(payload string) (string, error) {
	trimmed := strings.TrimSpace(payload)
	low := strings.ToLower(trimmed)
	if strings.Contains(low, "vless://") || strings.Contains(low, "trojan://") || strings.Contains(low, "hysteria2://") {
		return trimmed, nil
	}
	raw := strings.NewReplacer("\n", "", "\r", "", "\t", "", " ", "").Replace(trimmed)
	if raw == "" {
		return "", fmt.Errorf("subscription payload is empty")
	}
	decoded, err := base64.StdEncoding.DecodeString(raw)
	if err != nil {
		decoded, err = base64.RawStdEncoding.DecodeString(raw)
	}
	if err != nil {
		decoded, err = base64.URLEncoding.DecodeString(raw)
	}
	if err != nil {
		decoded, err = base64.RawURLEncoding.DecodeString(raw)
	}
	if err != nil {
		return "", fmt.Errorf("subscription payload is not valid base64")
	}
	return string(decoded), nil
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
