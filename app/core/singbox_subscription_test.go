package core

import (
	"encoding/json"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"testing"

	"pulseshield/app/models"
	"pulseshield/app/multiserver"
)

// TestSubscriptionToSingBoxConfig — сквозная проверка: текст подписки в любом
// поддерживаемом формате → разбор в VlessServer → реальная валидация ядром sing-box.
// Ловит расхождения между парсером ссылок и генератором конфига.
func TestSubscriptionToSingBoxConfig(t *testing.T) {
	singBox := filepath.Join("..", "..", "third_party", "core", "sing-box.exe")
	haveBinary := true
	if _, err := os.Stat(singBox); err != nil {
		haveBinary = false
	}

	bodies := map[string]string{
		"uri-lines": strings.Join([]string{
			"vless://11111111-2222-3333-4444-555555555555@a.example.com:443?security=tls&sni=a.example.com&type=ws&path=%2Fws&host=a.example.com#A",
			"trojan://pw@b.example.com:8443?security=tls&sni=b.example.com&type=grpc&serviceName=svc#B",
			"hysteria2://pw@c.example.com:8443?sni=c.example.com&obfs=salamander&obfs-password=ob#C",
			"ss://YWVzLTI1Ni1nY206cGFzcw==@d.example.com:8388#D",
			"tuic://11111111-2222-3333-4444-555555555555:pw@e.example.com:443?sni=e.example.com#E",
			"anytls://pw@f.example.com:8443?sni=f.example.com#F",
			"socks5://user:pw@g.example.com:1080#G",
		}, "\n"),
		"clash-block": `proxies:
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
  - name: "FI hy2"
    type: hysteria2
    server: fi.example.com
    port: 8443
    password: hypass
    sni: fi.example.com
  - name: "SS"
    type: ss
    server: hk.example.com
    port: 8388
    cipher: chacha20-ietf-poly1305
    password: sspass
`,
		"clash-flow": `proxies:
  - {name: A, type: vmess, server: a.example.com, port: 443, uuid: 11111111-2222-3333-4444-555555555555, alterId: 0, cipher: auto, network: ws, ws-opts: {path: /v, headers: {Host: a.example.com}}, tls: true}
  - {name: C, type: tuic, server: c.example.com, port: 443, uuid: 11111111-2222-3333-4444-555555555555, password: p, sni: c.example.com}
  - {name: D, type: anytls, server: d.example.com, port: 443, password: p, sni: d.example.com}
`,
		"scheme-aliases": strings.Join([]string{
			"ss2://YWVzLTI1Ni1nY206cGFzcw==@h.example.com:8388#H",
			"socks5h://user:pw@i.example.com:1080#I",
			"hy://u:p@j.example.com:443?sni=j.example.com#J",
		}, "\n"),
	}

	for name, body := range bodies {
		t.Run(name, func(t *testing.T) {
			servers, err := multiserver.ParseSubscriptionBody(body)
			if err != nil {
				t.Fatalf("ParseSubscriptionBody: %v", err)
			}
			if len(servers) == 0 {
				t.Fatal("no servers parsed")
			}
			for i := range servers {
				srv := *servers[i]
				t.Run(srv.ID+"-"+srv.Name, func(t *testing.T) {
					cfg := models.AppConfig{
						VlessServers:   []models.VlessServer{srv},
						ActiveServerID: srv.ID,
						ProxyMode:      models.ProxyModeSystem,
						ProxyPort:      2080,
					}
					raw, err := RenderSingBoxConfig(cfg, t.TempDir())
					if err != nil {
						t.Fatalf("RenderSingBoxConfig(%s): %v", srv.Name, err)
					}
					if !json.Valid(raw) {
						t.Fatal("rendered config is not valid JSON")
					}
					if !haveBinary {
						return
					}
					path := filepath.Join(t.TempDir(), "sing-box.json")
					if err := os.WriteFile(path, raw, 0o644); err != nil {
						t.Fatalf("write: %v", err)
					}
					if out, err := exec.Command(singBox, "check", "-c", path).CombinedOutput(); err != nil {
						t.Fatalf("sing-box check failed for %q: %v\n%s\n---config---\n%s", srv.Name, err, out, raw)
					}
				})
			}
		})
	}
}
