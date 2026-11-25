package core

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"strings"
	"time"
)

// SingBoxClashAPIPort — локальный контроллер Clash API sing-box (только 127.0.0.1).
// Должен совпадать с experimental.clash_api.external_controller в sing-box.json.
const SingBoxClashAPIPort = 20991

// SingBoxMainOutboundTag — тег основного VLESS outbound в RenderSingBoxConfig (sing-box.json).
const SingBoxMainOutboundTag = "proxy"

// SingBoxClashGlobalProxyGroup — группа Clash API для выбора прокси по умолчанию после старта ядра.
const SingBoxClashGlobalProxyGroup = "GLOBAL"

// clashAPIHTTPBase — базовый URL без пути (/connections добавляется отдельно).
// В тестах может подменяться на httptest.Server.URL.
var clashAPIHTTPBase = fmt.Sprintf("http://127.0.0.1:%d", SingBoxClashAPIPort)

// ClashConnectionsURL адрес для GET /connections (то же, что слушает sing-box).
func ClashConnectionsURL() string {
	return clashAPIHTTPBase + "/connections"
}

func clashProxiesPutURL(groupName string) string {
	g := strings.TrimSpace(groupName)
	if g == "" {
		g = SingBoxClashGlobalProxyGroup
	}
	return clashAPIHTTPBase + "/proxies/" + url.PathEscape(g)
}

type clashSelectProxyBody struct {
	Name string `json:"name"`
}

// PutClashProxyGroupSelection вызывает Clash API PUT /proxies/{group} с телом {"name": outboundTag}.
// Хранение выбранного сервера — в AppConfig и перегенерации sing-box.json; вызов выравнивает Clash API
// (например GLOBAL → outbound с тегом "proxy").
func PutClashProxyGroupSelection(apiSecret, groupName, outboundTag string) error {
	secret := strings.TrimSpace(apiSecret)
	tag := strings.TrimSpace(outboundTag)
	if tag == "" {
		tag = SingBoxMainOutboundTag
	}
	urlStr := clashProxiesPutURL(groupName)
	bodyBytes, err := json.Marshal(clashSelectProxyBody{Name: tag})
	if err != nil {
		return err
	}
	var lastErr error
	for attempt := 1; attempt <= 4; attempt++ {
		if attempt > 1 {
			time.Sleep(time.Duration(90+70*attempt) * time.Millisecond)
		}
		status, err := putClashProxyGroupOnce(secret, urlStr, bodyBytes)
		if err == nil {
			return nil
		}
		lastErr = err
		if status == http.StatusUnauthorized || status == http.StatusForbidden {
			break
		}
		if status >= 400 && status < 500 && status != http.StatusTooManyRequests && status != 0 {
			break
		}
	}
	return lastErr
}

func putClashProxyGroupOnce(secret, urlStr string, body []byte) (int, error) {
	client := NewLoopbackAwareHTTPClient(5 * time.Second)
	req, err := http.NewRequest(http.MethodPut, urlStr, bytes.NewReader(body))
	if err != nil {
		return 0, err
	}
	req.Header.Set("Content-Type", "application/json")
	if secret != "" {
		req.Header.Set("Authorization", "Bearer "+secret)
	}
	resp, err := client.Do(req)
	if err != nil {
		return 0, err
	}
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusOK && resp.StatusCode != http.StatusNoContent {
		prefix, _ := io.ReadAll(io.LimitReader(resp.Body, 320))
		return resp.StatusCode, fmt.Errorf("clash api PUT %s -> %s body=%q", urlStr, resp.Status, strings.TrimSpace(string(prefix)))
	}
	return resp.StatusCode, nil
}

// CloseAllClashConnections закрывает все открытые соединения в sing-box через Clash API DELETE /connections.
// Это сбрасывает старые HTTP/2 и keep-alive потоки при смене сервера, маршрутов или режима захвата.
func CloseAllClashConnections(apiSecret string) error {
	secret := strings.TrimSpace(apiSecret)
	client := NewLoopbackAwareHTTPClient(2 * time.Second)
	req, err := http.NewRequest(http.MethodDelete, ClashConnectionsURL(), nil)
	if err != nil {
		return err
	}
	if secret != "" {
		req.Header.Set("Authorization", "Bearer "+secret)
	}
	resp, err := client.Do(req)
	if err != nil {
		return err
	}
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusOK && resp.StatusCode != http.StatusNoContent {
		return fmt.Errorf("clash api DELETE %s -> %s", ClashConnectionsURL(), resp.Status)
	}
	return nil
}

// ActiveConnection — одно соединение из Clash API sing-box.
type ActiveConnection struct {
	ID       string `json:"id"`
	Process  string `json:"process"`
	Remote   string `json:"remote"`
	Upload   int64  `json:"upload"`
	Download int64  `json:"download"`
}

type clashConnectionsEnvelope struct {
	Connections []json.RawMessage `json:"connections"`
}

// FetchClashConnections опрашивает sing-box Clash API (experimental.clash_api).
// apiSecret — ровно experimental.clash_api.secret из конфига sing-box (тот же, что LocalInboundPassword).
// Пустой секрет: заголовок Authorization не отправляется (режим без пароля на API).
func FetchClashConnections(apiSecret string) ([]ActiveConnection, error) {
	secret := strings.TrimSpace(apiSecret)
	urlStr := ClashConnectionsURL()
	var lastErr error
	for attempt := 1; attempt <= 5; attempt++ {
		if attempt > 1 {
			time.Sleep(time.Duration(100+80*attempt) * time.Millisecond)
		}
		rows, statusCode, err := fetchClashConnectionsOnce(secret)
		if err == nil {
			return rows, nil
		}
		lastErr = err
		if statusCode == http.StatusUnauthorized || statusCode == http.StatusForbidden {
			fmt.Printf("[PulseShield] clash_api: auth отклонён (%d) для %s — проверьте, что LocalInboundPassword совпадает с experimental.clash_api.secret в runtime/sing-box.json\n",
				statusCode, urlStr)
			break
		}
		// Не долбим сервер при явных клиентских ошибках.
		if statusCode >= 400 && statusCode < 500 && statusCode != http.StatusTooManyRequests {
			break
		}
		if attempt < 5 {
			fmt.Printf("[PulseShield] clash_api: попытка %d/5 не удалась (%v), status=%d url=%s\n", attempt, err, statusCode, urlStr)
		}
	}
	return nil, lastErr
}

func fetchClashConnectionsOnce(secret string) ([]ActiveConnection, int, error) {
	client := NewLoopbackAwareHTTPClient(5 * time.Second)
	req, err := http.NewRequest(http.MethodGet, ClashConnectionsURL(), nil)
	if err != nil {
		return nil, 0, err
	}
	if secret != "" {
		// sing-box: Authorization: Bearer ${secret}
		req.Header.Set("Authorization", "Bearer "+secret)
	}
	resp, err := client.Do(req)
	if err != nil {
		fmt.Printf("[PulseShield] clash_api: сеть GET %s err=%v (Authorization Bearer, len(secret)=%d)\n",
			req.URL.String(), err, len(secret))
		return nil, 0, err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		bodyPrefix, _ := io.ReadAll(io.LimitReader(resp.Body, 480))
		fmt.Printf("[PulseShield] clash_api: GET %s -> %s body_prefix=%q len(secret)=%d\n",
			req.URL.String(), resp.Status, strings.TrimSpace(string(bodyPrefix)), len(secret))
		return nil, resp.StatusCode, fmt.Errorf("clash api: %s", resp.Status)
	}

	var env clashConnectionsEnvelope
	if err := json.NewDecoder(resp.Body).Decode(&env); err != nil {
		fmt.Printf("[PulseShield] clash_api: decode JSON ответа %s err=%v\n", req.URL.String(), err)
		return nil, resp.StatusCode, err
	}
	out := make([]ActiveConnection, 0, len(env.Connections))
	for _, raw := range env.Connections {
		var wrap struct {
			ID       string `json:"id"`
			Upload   int64  `json:"upload"`
			Download int64  `json:"download"`
			Metadata struct {
				Process         string `json:"process"`
				Host            string `json:"host"`
				DestinationIP   string `json:"destinationIP"`
				DestinationPort string `json:"destinationPort"`
				Type            string `json:"type"`
				Network         string `json:"network"`
			} `json:"metadata"`
		}
		if err := json.Unmarshal(raw, &wrap); err != nil {
			continue
		}
		remote := clashRemoteLabel(wrap.Metadata.Host, wrap.Metadata.DestinationIP, wrap.Metadata.DestinationPort, wrap.Metadata.Type, wrap.Metadata.Network)
		proc := strings.TrimSpace(wrap.Metadata.Process)
		if proc == "" {
			proc = "—"
		}
		out = append(out, ActiveConnection{
			ID:       wrap.ID,
			Process:  proc,
			Remote:   remote,
			Upload:   wrap.Upload,
			Download: wrap.Download,
		})
	}
	return out, resp.StatusCode, nil
}

func clashRemoteLabel(host, dip, dport, typ, network string) string {
	host = strings.TrimSpace(host)
	dip = strings.TrimSpace(dip)
	dport = strings.TrimSpace(dport)
	network = strings.TrimSpace(network)
	typ = strings.TrimSpace(typ)
	if host != "" {
		if dport != "" {
			return host + ":" + dport
		}
		return host
	}
	if dip != "" {
		if dport != "" {
			return dip + ":" + dport
		}
		return dip
	}
	if typ != "" || network != "" {
		return strings.TrimSpace(typ + " " + network)
	}
	return "—"
}
