package multiserver

import (
	"encoding/base64"
	"strings"
	"testing"
)

// TestParseProxyLink_AllProtocols проверяет, что каждое семейство ссылок разбирается.
func TestParseProxyLink_AllProtocols(t *testing.T) {
	cases := []struct {
		name      string
		link      string
		wantHost  string
		wantPort  string
		wantProto string
		wantUUID  string
	}{
		{
			name:      "vless reality",
			link:      "vless://11111111-2222-3333-4444-555555555555@nl1.example.com:443?security=reality&sni=www.microsoft.com&fp=chrome&pbk=abc&sid=00&type=tcp&flow=xtls-rprx-vision#NL-Reality",
			wantHost:  "nl1.example.com",
			wantPort:  "443",
			wantProto: "vless",
			wantUUID:  "11111111-2222-3333-4444-555555555555",
		},
		{
			name:      "trojan grpc",
			link:      "trojan://passw0rd@de1.example.com:8443?security=tls&sni=de1.example.com&type=grpc&serviceName=svc#DE",
			wantHost:  "de1.example.com",
			wantPort:  "8443",
			wantProto: "trojan",
			wantUUID:  "passw0rd",
		},
		{
			name:      "hysteria2",
			link:      "hysteria2://hy2pass@fi1.example.com:8443?sni=fi1.example.com&obfs=salamander&obfs-password=secret#FI",
			wantHost:  "fi1.example.com",
			wantPort:  "8443",
			wantProto: "hysteria2",
			wantUUID:  "hy2pass",
		},
		{
			name:      "hy2 alias",
			link:      "hy2://hy2pass@se1.example.com:443?sni=se1.example.com#SE",
			wantHost:  "se1.example.com",
			wantPort:  "443",
			wantProto: "hysteria2",
			wantUUID:  "hy2pass",
		},
		{
			name:      "tuic v5",
			link:      "tuic://11111111-2222-3333-4444-555555555555:tuicpass@jp1.example.com:443?sni=jp1.example.com&alpn=h3&congestion_control=bbr#JP",
			wantHost:  "jp1.example.com",
			wantPort:  "443",
			wantProto: "tuic",
			wantUUID:  "tuicpass",
		},
		{
			name:      "anytls",
			link:      "anytls://atpass@sg1.example.com:8443?sni=sg1.example.com#SG",
			wantHost:  "sg1.example.com",
			wantPort:  "8443",
			wantProto: "anytls",
			wantUUID:  "atpass",
		},
		{
			name:      "socks5",
			link:      "socks5://user:pass@127.0.0.2:1080#SOCKS",
			wantHost:  "127.0.0.2",
			wantPort:  "1080",
			wantProto: "socks",
			wantUUID:  "pass",
		},
		{
			name:      "hysteria v1",
			link:      "hysteria://us1.example.com:36712?auth=h1pass&peer=us1.example.com&upmbps=100&downmbps=200&obfs=xplus#US",
			wantHost:  "us1.example.com",
			wantPort:  "36712",
			wantProto: "hysteria",
			wantUUID:  "h1pass",
		},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			srv, err := ParseProxyLink(tc.link)
			if err != nil {
				t.Fatalf("ParseProxyLink failed: %v", err)
			}
			if srv.Host != tc.wantHost || srv.Port != tc.wantPort {
				t.Fatalf("host/port = %s:%s, want %s:%s", srv.Host, srv.Port, tc.wantHost, tc.wantPort)
			}
			if got := ProtocolOf(srv); got != tc.wantProto {
				t.Fatalf("protocol = %q, want %q", got, tc.wantProto)
			}
			if srv.UUID != tc.wantUUID {
				t.Fatalf("secret = %q, want %q", srv.UUID, tc.wantUUID)
			}
			if srv.ID == "" || srv.Name == "" {
				t.Fatalf("empty ID or Name: %+v", srv)
			}
			if srv.RawVLESS != tc.link {
				t.Fatalf("RawVLESS lost")
			}
		})
	}
}

func TestParseProxyLink_Shadowsocks(t *testing.T) {
	// Слитный base64: ss://base64(method:password)@host:port#name
	cred := base64.StdEncoding.EncodeToString([]byte("aes-256-gcm:sspass"))
	link := "ss://" + cred + "@hk1.example.com:8388#HK"
	srv, err := ParseProxyLink(link)
	if err != nil {
		t.Fatalf("ss merged failed: %v", err)
	}
	if srv.Host != "hk1.example.com" || srv.Port != "8388" || srv.UUID != "sspass" {
		t.Fatalf("unexpected ss server: %+v", srv)
	}
	if srv.Params["method"] != "aes-256-gcm" || ProtocolOf(srv) != "shadowsocks" {
		t.Fatalf("unexpected ss params: %+v", srv.Params)
	}

	// Legacy SIP002: ss://base64(method:password@host:port)#name
	legacy := base64.StdEncoding.EncodeToString([]byte("chacha20-ietf-poly1305:pw2@tw1.example.com:443"))
	srv2, err := ParseProxyLink("ss://" + legacy + "#TW")
	if err != nil {
		t.Fatalf("ss legacy failed: %v", err)
	}
	if srv2.Host != "tw1.example.com" || srv2.Port != "443" || srv2.UUID != "pw2" {
		t.Fatalf("unexpected legacy ss server: %+v", srv2)
	}
}

func TestParseProxyLink_Vmess(t *testing.T) {
	payload := `{"v":"2","ps":"US-Node","add":"us1.example.com","port":"443","id":"11111111-2222-3333-4444-555555555555","aid":"0","scy":"auto","net":"ws","type":"none","host":"us1.example.com","path":"/ray","tls":"tls","sni":"us1.example.com","alpn":"h2,http/1.1","fp":"chrome"}`
	link := "vmess://" + base64.StdEncoding.EncodeToString([]byte(payload))
	srv, err := ParseProxyLink(link)
	if err != nil {
		t.Fatalf("vmess failed: %v", err)
	}
	if srv.Host != "us1.example.com" || srv.Port != "443" || srv.Name != "US-Node" {
		t.Fatalf("unexpected vmess server: %+v", srv)
	}
	if srv.UUID != "11111111-2222-3333-4444-555555555555" {
		t.Fatalf("unexpected vmess uuid: %q", srv.UUID)
	}
	if srv.Params["type"] != "ws" || srv.Params["path"] != "/ray" || srv.Params["security"] != "tls" {
		t.Fatalf("unexpected vmess params: %+v", srv.Params)
	}
	if ProtocolOf(srv) != "vmess" {
		t.Fatalf("protocol = %q", ProtocolOf(srv))
	}
}

func TestExtractSubscriptionLinks_Plain(t *testing.T) {
	body := "vless://a@x.example.com:443?security=tls#A\n\ntrojan://b@y.example.com:443#B\n# comment\nhy2://c@z.example.com:443#C\n"
	links, err := ExtractSubscriptionLinks(body)
	if err != nil {
		t.Fatalf("plain failed: %v", err)
	}
	if len(links) != 3 {
		t.Fatalf("expected 3 links, got %d: %v", len(links), links)
	}
}

func TestExtractSubscriptionLinks_Base64Wrapped(t *testing.T) {
	plain := "vless://a@x.example.com:443?security=tls#A\ntrojan://b@y.example.com:443#B\nhysteria2://c@z.example.com:443#C"
	enc := base64.StdEncoding.EncodeToString([]byte(plain))
	// добавляем переносы строк внутри, как это делают некоторые панели
	wrapped := enc[:20] + "\n" + enc[20:]
	links, err := ExtractSubscriptionLinks(wrapped)
	if err != nil {
		t.Fatalf("base64 failed: %v", err)
	}
	if len(links) != 3 {
		t.Fatalf("expected 3 links, got %d: %v", len(links), links)
	}
}

func TestExtractSubscriptionLinks_Base64URLSafe(t *testing.T) {
	plain := "vless://a@x.example.com:443?security=tls&pbk=??>>#A?" // содержит символы, дающие +/ в base64
	enc := base64.URLEncoding.EncodeToString([]byte(plain))
	links, err := ExtractSubscriptionLinks(enc)
	if err != nil {
		t.Fatalf("urlsafe base64 failed: %v", err)
	}
	if len(links) != 1 {
		t.Fatalf("expected 1 link, got %d", len(links))
	}
}

func TestExtractSubscriptionLinks_DoubleBase64(t *testing.T) {
	plain := "vless://a@x.example.com:443?security=tls#A\ntrojan://b@y.example.com:443#B"
	once := base64.StdEncoding.EncodeToString([]byte(plain))
	twice := base64.StdEncoding.EncodeToString([]byte(once))
	links, err := ExtractSubscriptionLinks(twice)
	if err != nil {
		t.Fatalf("double base64 failed: %v", err)
	}
	if len(links) != 2 {
		t.Fatalf("expected 2 links, got %d", len(links))
	}
}

func TestExtractSubscriptionLinks_ClashYAML(t *testing.T) {
	body := `port: 7890
socks-port: 7891
allow-lan: true
mode: rule
proxies:
  - name: "NL Reality"
    type: vless
    server: nl.example.com
    port: 443
    uuid: 11111111-2222-3333-4444-555555555555
    network: tcp
    tls: true
    servername: www.microsoft.com
    flow: xtls-rprx-vision
    client-fingerprint: chrome
    reality-opts:
      public-key: abc
      short-id: "00"
  - name: "DE Trojan"
    type: trojan
    server: de.example.com
    port: 8443
    password: trojanpass
    sni: de.example.com
    skip-cert-verify: true
  - name: "HK SS"
    type: ss
    server: hk.example.com
    port: 8388
    cipher: aes-256-gcm
    password: sspass
proxy-groups:
  - name: PROXY
    type: select
    proxies:
      - "NL Reality"
rules:
  - MATCH,PROXY
`
	links, err := ExtractSubscriptionLinks(body)
	if err != nil {
		t.Fatalf("clash yaml failed: %v", err)
	}
	if len(links) != 3 {
		t.Fatalf("expected 3 links, got %d: %v", len(links), links)
	}
	var protos []string
	for _, l := range links {
		srv, err := ParseProxyLink(l)
		if err != nil {
			t.Fatalf("failed to parse converted link %q: %v", l, err)
		}
		protos = append(protos, ProtocolOf(srv))
	}
	joined := strings.Join(protos, ",")
	if !strings.Contains(joined, "vless") || !strings.Contains(joined, "trojan") || !strings.Contains(joined, "shadowsocks") {
		t.Fatalf("unexpected protocols: %v", protos)
	}
}

func TestExtractSubscriptionLinks_SingBoxJSON(t *testing.T) {
	body := `{
  "outbounds": [
    {"type":"vless","tag":"NL","server":"nl.example.com","server_port":443,"uuid":"11111111-2222-3333-4444-555555555555","flow":"xtls-rprx-vision","tls":{"enabled":true,"server_name":"www.microsoft.com","utls":{"enabled":true,"fingerprint":"chrome"},"reality":{"enabled":true,"public_key":"abc","short_id":"00"}}},
    {"type":"hysteria2","tag":"FI","server":"fi.example.com","server_port":8443,"password":"hypass","tls":{"enabled":true,"server_name":"fi.example.com"}},
    {"type":"shadowsocks","tag":"HK","server":"hk.example.com","server_port":8388,"method":"aes-256-gcm","password":"sspass"},
    {"type":"direct","tag":"direct"},
    {"type":"block","tag":"block"}
  ]
}`
	links, err := ExtractSubscriptionLinks(body)
	if err != nil {
		t.Fatalf("sing-box json failed: %v", err)
	}
	if len(links) != 3 {
		t.Fatalf("expected 3 links (direct/block skipped), got %d: %v", len(links), links)
	}
}

func TestExtractSubscriptionLinks_XrayJSON(t *testing.T) {
	body := `{
  "outbounds": [
    {"protocol":"vless","tag":"X-NL","settings":{"vnext":[{"address":"nl.example.com","port":443,"users":[{"id":"11111111-2222-3333-4444-555555555555","flow":"xtls-rprx-vision"}]}]},"streamSettings":{"network":"tcp","security":"reality","realitySettings":{"publicKey":"abc","shortId":"00","serverName":"www.microsoft.com","fingerprint":"chrome"}}},
    {"protocol":"vmess","tag":"X-US","settings":{"vnext":[{"address":"us.example.com","port":443,"users":[{"id":"22222222-3333-4444-5555-666666666666"}]}]}}
  ]
}`
	links, err := ExtractSubscriptionLinks(body)
	if err != nil {
		t.Fatalf("xray json failed: %v", err)
	}
	if len(links) != 2 {
		t.Fatalf("expected 2 links, got %d: %v", len(links), links)
	}
}

func TestExtractSubscriptionLinks_HTMLWrapped(t *testing.T) {
	body := `<html><body><pre>vless://a@x.example.com:443?security=tls#A
trojan://b@y.example.com:443#B</pre></body></html>`
	links, err := ExtractSubscriptionLinks(body)
	if err != nil {
		t.Fatalf("html wrapped failed: %v", err)
	}
	if len(links) != 2 {
		t.Fatalf("expected 2 links, got %d: %v", len(links), links)
	}
}

func TestExtractSubscriptionLinks_EmptyIsError(t *testing.T) {
	if _, err := ExtractSubscriptionLinks("   "); err == nil {
		t.Fatal("expected error for empty body")
	}
	if _, err := ExtractSubscriptionLinks("not a subscription at all"); err == nil {
		t.Fatal("expected error for garbage body")
	}
}

func TestParseSubscriptionUserinfoHeader_Full(t *testing.T) {
	m := parseSubscriptionUserinfoHeader("upload=1024; download=2048; total=1073741824; expire=1893456000")
	if m.UploadBytes != 1024 || m.DownloadBytes != 2048 || m.TotalBytes != 1073741824 || m.ExpiresUnix != 1893456000 {
		t.Fatalf("unexpected meta: %+v", m)
	}
}

func TestDecodeProfileTitle(t *testing.T) {
	if got := decodeProfileTitle("base64:" + base64.StdEncoding.EncodeToString([]byte("Мой провайдер"))); got != "Мой провайдер" {
		t.Fatalf("base64 title = %q", got)
	}
	if got := decodeProfileTitle("Plain Title"); got != "Plain Title" {
		t.Fatalf("plain title = %q", got)
	}
	if got := decodeProfileTitle(""); got != "" {
		t.Fatalf("empty title = %q", got)
	}
}

func TestNormalizeSubscriptionURL(t *testing.T) {
	cases := map[string]string{
		"sub.example.com/api?token=1":      "https://sub.example.com/api?token=1",
		"https://sub.example.com/api":      "https://sub.example.com/api",
		"  'http://sub.example.com/api'  ": "http://sub.example.com/api",
		"//sub.example.com/api":            "https://sub.example.com/api",
	}
	for in, want := range cases {
		if got := normalizeSubscriptionURL(in); got != want {
			t.Fatalf("normalize(%q) = %q, want %q", in, got, want)
		}
	}
}

func TestIsProxyLink(t *testing.T) {
	yes := []string{
		"vless://a@x.example.com:443",
		"vmess://eyJ2IjoiMiJ9",
		"trojan://a@x.example.com:443",
		"ss://YWVzLTI1Ni1nY206cGFzcw==@x.example.com:8388",
		"hysteria2://a@x.example.com:443",
		"hy2://a@x.example.com:443",
		"tuic://a@x.example.com:443",
		"anytls://a@x.example.com:443",
		"socks5://a@x.example.com:1080",
	}
	for _, l := range yes {
		if !IsProxyLink(l) {
			t.Fatalf("expected link to be recognized: %s", l)
		}
	}
	no := []string{
		"https://sub.example.com/api/v1/client/subscribe?token=abc",
		"http://example.com/page",
		"hello world",
		"",
		"proxies:",
	}
	for _, l := range no {
		if IsProxyLink(l) {
			t.Fatalf("expected NOT a proxy link: %s", l)
		}
	}
}

func TestParseSubscriptionBody_Mixed(t *testing.T) {
	body := strings.Join([]string{
		"vless://11111111-2222-3333-4444-555555555555@nl.example.com:443?security=reality&sni=a.com&pbk=k&sid=00#NL",
		"hysteria2://pass@fi.example.com:8443?sni=fi.example.com#FI",
		"ssr://c29tZXRoaW5n", // не поддерживается — должен быть пропущен
	}, "\n")
	servers, err := ParseSubscriptionBody(body)
	if err != nil {
		t.Fatalf("mixed failed: %v", err)
	}
	if len(servers) != 2 {
		t.Fatalf("expected 2 supported servers, got %d", len(servers))
	}
}

// TestClashNestedOpts проверяет Clash.Meta-формат с вложенными блоками ws-opts / grpc-opts /
// reality-opts, записанными как подобъекты YAML.
func TestClashNestedOpts(t *testing.T) {
	body := `proxies:
  - name: "NL WS"
    type: vless
    server: nl.example.com
    port: 443
    uuid: 11111111-2222-3333-4444-555555555555
    network: ws
    tls: true
    servername: cdn.example.com
    ws-opts:
      path: /ws-route
      headers:
        Host: cdn.example.com
  - name: "DE gRPC"
    type: trojan
    server: de.example.com
    port: 8443
    password: pw
    network: grpc
    grpc-opts:
      grpc-service-name: TunService
  - name: "HK SS plugin"
    type: ss
    server: hk.example.com
    port: 8388
    cipher: chacha20-ietf-poly1305
    password: sspass
    plugin: obfs
    plugin-opts:
      mode: tls
      host: bing.com
`
	links, err := ExtractSubscriptionLinks(body)
	if err != nil {
		t.Fatalf("clash nested failed: %v", err)
	}
	if len(links) != 3 {
		t.Fatalf("expected 3 links, got %d: %v", len(links), links)
	}
	var wsPath, grpcSvc, ssPlugin string
	for _, l := range links {
		srv, err := ParseProxyLink(l)
		if err != nil {
			t.Fatalf("parse %q: %v", l, err)
		}
		switch ProtocolOf(srv) {
		case "vless":
			wsPath = srv.Params["path"]
			if got := srv.Params["host"]; got != "cdn.example.com" {
				t.Fatalf("ws host = %q, want cdn.example.com", got)
			}
		case "trojan":
			grpcSvc = srv.Params["serviceName"]
		case "shadowsocks":
			ssPlugin = srv.Params["plugin"]
		}
	}
	if wsPath != "/ws-route" {
		t.Fatalf("vless ws path = %q", wsPath)
	}
	if grpcSvc != "TunService" {
		t.Fatalf("trojan grpc service = %q", grpcSvc)
	}
	if !strings.Contains(ssPlugin, "mode=tls") {
		t.Fatalf("ss plugin = %q, want mode=tls", ssPlugin)
	}
}

// TestClashAllProtocols проверяет конвертацию каждого протокола Clash в рабочую ссылку.
func TestClashAllProtocols(t *testing.T) {
	body := `proxies:
  - {name: A, type: vmess, server: a.example.com, port: 443, uuid: 11111111-2222-3333-4444-555555555555, alterId: 0, cipher: auto, network: ws, ws-opts: {path: /v, headers: {Host: a.example.com}}, tls: true}
  - {name: B, type: hysteria2, server: b.example.com, port: 8443, password: pw, sni: b.example.com, obfs: salamander, obfs-password: ob}
  - {name: C, type: tuic, server: c.example.com, port: 443, uuid: 11111111-2222-3333-4444-555555555555, password: p, sni: c.example.com}
  - {name: D, type: anytls, server: d.example.com, port: 443, password: p, sni: d.example.com}
  - {name: E, type: socks5, server: e.example.com, port: 1080, username: u, password: p}
  - {name: F, type: wireguard, server: f.example.com, port: 51820, private-key: PRIV, public-key: PUB, ip: 10.0.0.2/32}
`
	links, err := ExtractSubscriptionLinks(body)
	if err != nil {
		t.Fatalf("clash all protocols failed: %v", err)
	}
	if len(links) != 6 {
		t.Fatalf("expected 6 links, got %d: %v", len(links), links)
	}
	want := map[string]bool{"vmess": false, "hysteria2": false, "tuic": false, "anytls": false, "socks": false, "wireguard": false}
	for _, l := range links {
		srv, err := ParseProxyLink(l)
		if err != nil {
			t.Fatalf("parse %q: %v", l, err)
		}
		p := ProtocolOf(srv)
		if _, ok := want[p]; ok {
			want[p] = true
		}
	}
	for p, ok := range want {
		if !ok {
			t.Fatalf("protocol %s was not produced (links=%v)", p, links)
		}
	}
}

// TestClashFlowStyleList проверяет YAML в «flow»-стиле (объекты в фигурных скобках).
func TestClashFlowStyleList(t *testing.T) {
	body := `proxies:
  - {name: "X", type: vless, server: x.example.com, port: 443, uuid: 11111111-2222-3333-4444-555555555555, tls: true, servername: x.example.com}
`
	links, err := ExtractSubscriptionLinks(body)
	if err != nil {
		t.Fatalf("flow style failed: %v", err)
	}
	if len(links) != 1 {
		t.Fatalf("expected 1 link, got %d", len(links))
	}
	srv, err := ParseProxyLink(links[0])
	if err != nil {
		t.Fatalf("parse: %v", err)
	}
	if srv.Host != "x.example.com" || srv.Params["security"] != "tls" {
		t.Fatalf("unexpected server: host=%q security=%q", srv.Host, srv.Params["security"])
	}
}

// TestBase64EncodedClashYAML — панель кодирует Clash-конфиг в base64 целиком.
func TestBase64EncodedClashYAML(t *testing.T) {
	yamlBody := "proxies:\n  - {name: Y, type: trojan, server: y.example.com, port: 443, password: pw}\n"
	enc := base64.StdEncoding.EncodeToString([]byte(yamlBody))
	links, err := ExtractSubscriptionLinks(enc)
	if err != nil {
		t.Fatalf("base64 clash failed: %v", err)
	}
	if len(links) != 1 {
		t.Fatalf("expected 1 link, got %d", len(links))
	}
	if !strings.HasPrefix(links[0], "trojan://") {
		t.Fatalf("unexpected link: %s", links[0])
	}
}

// TestParseSubscriptionBody_AllLinkTypes — сквозная проверка списка ссылок всех протоколов.
func TestParseSubscriptionBody_AllLinkTypes(t *testing.T) {
	vmessJSON := base64.StdEncoding.EncodeToString([]byte(`{"v":"2","ps":"V","add":"v.example.com","port":"443","id":"11111111-2222-3333-4444-555555555555","aid":"0","net":"ws","path":"/v","host":"v.example.com","tls":"tls"}`))
	ssB64 := base64.RawURLEncoding.EncodeToString([]byte("aes-256-gcm:sspass"))
	body := strings.Join([]string{
		"vless://11111111-2222-3333-4444-555555555555@a.example.com:443?security=tls&sni=a.example.com#A",
		"vmess://" + vmessJSON,
		"trojan://pw@c.example.com:443?sni=c.example.com#C",
		"ss://" + ssB64 + "@d.example.com:8388#D",
		"hysteria2://pw@e.example.com:8443?sni=e.example.com#E",
		"tuic://11111111-2222-3333-4444-555555555555:pw@f.example.com:443?sni=f.example.com#F",
		"anytls://pw@g.example.com:443?sni=g.example.com#G",
		"socks5://u:p@h.example.com:1080#H",
		"http://u:p@i.example.com:8080#I",
	}, "\n")
	servers, err := ParseSubscriptionBody(body)
	if err != nil {
		t.Fatalf("all link types failed: %v", err)
	}
	if len(servers) != 9 {
		t.Fatalf("expected 9 servers, got %d", len(servers))
	}
	got := map[string]int{}
	for _, s := range servers {
		got[ProtocolOf(s)]++
	}
	for _, p := range []string{"vless", "vmess", "trojan", "shadowsocks", "hysteria2", "tuic", "anytls", "socks", "http"} {
		if got[p] == 0 {
			t.Fatalf("protocol %s missing, got %v", p, got)
		}
	}
}
