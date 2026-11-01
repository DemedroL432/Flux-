package multiserver

import (
	"testing"
)

// TestParseProxyLink_SchemeAliases проверяет алиасы схем, которые встречаются у панелей
// (Happ, v2rayN, Marzban, sing-box-совместимые подписки): ss2://, socks5h://, socks4a://,
// hy://, wg://. Все они должны давать тот же протокол, что и каноническая схема.
func TestParseProxyLink_SchemeAliases(t *testing.T) {
	cases := []struct {
		name      string
		link      string
		wantHost  string
		wantPort  string
		wantProto string
	}{
		{
			name:      "ss2 alias",
			link:      "ss2://YWVzLTI1Ni1nY206cGFzcw==@hk.example.com:8388#HK",
			wantHost:  "hk.example.com",
			wantPort:  "8388",
			wantProto: "shadowsocks",
		},
		{
			name:      "socks5h alias",
			link:      "socks5h://user:pw@proxy.example.com:1080#SOCKS5H",
			wantHost:  "proxy.example.com",
			wantPort:  "1080",
			wantProto: "socks",
		},
		{
			name:      "socks4a alias",
			link:      "socks4a://user:pw@proxy.example.com:1080#SOCKS4A",
			wantHost:  "proxy.example.com",
			wantPort:  "1080",
			wantProto: "socks",
		},
		{
			name:      "hy alias",
			link:      "hy://hyuser:pass@fr.example.com:443?sni=fr.example.com#FR",
			wantHost:  "fr.example.com",
			wantPort:  "443",
			wantProto: "hysteria",
		},
		{
			name:      "wg alias",
			link:      "wg://wg.example.com:51820?private_key=PRIV&public_key=PUB&address=10.0.0.2%2F32#WG",
			wantHost:  "wg.example.com",
			wantPort:  "51820",
			wantProto: "wireguard",
		},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			if !IsProxyLink(tc.link) {
				t.Fatalf("IsProxyLink(%q) = false, want true", tc.link)
			}
			srv, err := ParseProxyLink(tc.link)
			if err != nil {
				t.Fatalf("ParseProxyLink(%q): %v", tc.link, err)
			}
			if srv.Host != tc.wantHost {
				t.Errorf("host = %q, want %q", srv.Host, tc.wantHost)
			}
			if srv.Port != tc.wantPort {
				t.Errorf("port = %q, want %q", srv.Port, tc.wantPort)
			}
			if got := ProtocolOf(srv); got != tc.wantProto {
				t.Errorf("protocol = %q, want %q", got, tc.wantProto)
			}
		})
	}
}

// TestParseProxyLink_DefaultPorts проверяет, что при отсутствии порта подставляется
// правильный дефолт для каждой схемы, включая алиасы.
func TestParseProxyLink_DefaultPorts(t *testing.T) {
	cases := []struct {
		link string
		port string
	}{
		{"ss2://YWVzLTI1Ni1nY206cGFzcw==@a.example.com#A", "8388"},
		{"socks5h://u:p@b.example.com#B", "1080"},
		{"socks4a://u:p@c.example.com#C", "1080"},
		{"hy://u:p@d.example.com#D", "36712"},
		{"wg://e.example.com?private_key=P&public_key=Q#E", "51820"},
	}
	for _, tc := range cases {
		srv, err := ParseProxyLink(tc.link)
		if err != nil {
			t.Fatalf("ParseProxyLink(%q): %v", tc.link, err)
		}
		if srv.Port != tc.port {
			t.Errorf("%q: port = %q, want %q", tc.link, srv.Port, tc.port)
		}
	}
}

// TestExtractSubscriptionLinks_HTMLWithAliases — алиасы должны находиться и при
// «выковыривании» ссылок из HTML-обёртки панели, где список схем раньше был захардкожен.
func TestExtractSubscriptionLinks_HTMLWithAliases(t *testing.T) {
	body := `<html><body><pre>ss2://YWVzLTI1Ni1nY206cGFzcw==@hk.example.com:8388#HK
socks5h://user:pw@proxy.example.com:1080#SOCKS5H
hy://u:p@fr.example.com:443#FR</pre></body></html>`

	links, err := ExtractSubscriptionLinks(body)
	if err != nil {
		t.Fatalf("ExtractSubscriptionLinks(html aliases): %v", err)
	}
	if len(links) != 3 {
		t.Fatalf("expected 3 links, got %d: %v", len(links), links)
	}
	got := map[string]bool{}
	for _, l := range links {
		srv, err := ParseProxyLink(l)
		if err != nil {
			t.Fatalf("parse %q: %v", l, err)
		}
		got[ProtocolOf(srv)] = true
	}
	for _, want := range []string{"shadowsocks", "socks", "hysteria"} {
		if !got[want] {
			t.Errorf("protocol %q not found in HTML scrape (got %v)", want, got)
		}
	}
}

// TestExtractSubscriptionLinks_SchemeAliases — алиасы схем должны находиться
// и внутри текста подписки, а не только при явном разборе одной ссылки.
func TestExtractSubscriptionLinks_SchemeAliases(t *testing.T) {
	body := "ss2://YWVzLTI1Ni1nY206cGFzcw==@hk.example.com:8388#HK\n" +
		"socks5h://user:pw@proxy.example.com:1080#SOCKS5H\n" +
		"hy://u:p@fr.example.com:443#FR\n"

	links, err := ExtractSubscriptionLinks(body)
	if err != nil {
		t.Fatalf("ExtractSubscriptionLinks: %v", err)
	}
	if len(links) != 3 {
		t.Fatalf("expected 3 links, got %d: %v", len(links), links)
	}

	got := map[string]bool{}
	for _, l := range links {
		srv, err := ParseProxyLink(l)
		if err != nil {
			t.Fatalf("parse %q: %v", l, err)
		}
		got[ProtocolOf(srv)] = true
	}
	for _, want := range []string{"shadowsocks", "socks", "hysteria"} {
		if !got[want] {
			t.Errorf("protocol %q not found (got %v)", want, got)
		}
	}
}
