package multiserver

import (
	"strings"
	"testing"
)

// TestFullChain_AliasesInMixedSubscription — алиасы схем должны проходить весь путь:
// текст подписки -> ParseSubscriptionBody -> серверы с корректным протоколом.
func TestFullChain_AliasesInMixedSubscription(t *testing.T) {
	body := strings.Join([]string{
		"vless://11111111-2222-3333-4444-555555555555@a.example.com:443?security=tls&sni=a.example.com#A",
		"ss2://YWVzLTI1Ni1nY206cGFzcw==@b.example.com:8388#B",
		"socks5h://user:pw@c.example.com:1080#C",
		"hy://u:p@d.example.com:443#D",
		"wg://e.example.com:51820?private_key=P&public_key=Q#E",
		"trojan://pw@f.example.com:443#F",
	}, "\n")

	servers, err := ParseSubscriptionBody(body)
	if err != nil {
		t.Fatalf("ParseSubscriptionBody: %v", err)
	}
	if len(servers) != 6 {
		t.Fatalf("expected 6 servers, got %d", len(servers))
	}

	want := map[string]int{"vless": 1, "shadowsocks": 1, "socks": 1, "hysteria": 1, "wireguard": 1, "trojan": 1}
	got := map[string]int{}
	for _, s := range servers {
		got[ProtocolOf(s)]++
	}
	for proto, n := range want {
		if got[proto] != n {
			t.Errorf("protocol %q: got %d, want %d (all=%v)", proto, got[proto], n, got)
		}
	}
}