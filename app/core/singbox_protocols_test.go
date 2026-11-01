package core

import (
	"encoding/json"
	"os"
	"os/exec"
	"path/filepath"
	"testing"

	"pulseshield/app/models"
)

// TestRenderSingBoxConfig_AllProtocolsValid РїСЂРѕРІРµСЂСЏРµС‚, С‡С‚Рѕ СЃРіРµРЅРµСЂРёСЂРѕРІР°РЅРЅС‹Р№ РєРѕРЅС„РёРі
// РґР»СЏ РєР°Р¶РґРѕРіРѕ РїРѕРґРґРµСЂР¶РёРІР°РµРјРѕРіРѕ РїСЂРѕС‚РѕРєРѕР»Р° РїСЂРёРЅРёРјР°РµС‚СЃСЏ РЅР°СЃС‚РѕСЏС‰РёРј СЏРґСЂРѕРј sing-box (sing-box check).
func TestRenderSingBoxConfig_AllProtocolsValid(t *testing.T) {
	singBox := filepath.Join("..", "..", "third_party", "core", "sing-box.exe")
	if _, err := os.Stat(singBox); err != nil {
		t.Skip("sing-box.exe not found, skipping binary validation")
	}

	cases := []models.VlessServer{
		{
			ID: "vless-reality", Name: "VLESS Reality", Host: "1.1.1.1", Port: "443",
			UUID:   "11111111-2222-3333-4444-555555555555",
			Params: map[string]string{"_proto": "vless", "security": "reality", "sni": "www.microsoft.com", "fp": "chrome", "pbk": "tL5CxgEQ7T5kw9Y6z9uqZywQA9fQjU4PBwEGXAXDTH4", "sid": "00", "type": "tcp", "flow": "xtls-rprx-vision"},
		},
		{
			ID: "vless-ws", Name: "VLESS WS", Host: "1.1.1.1", Port: "443",
			UUID:   "11111111-2222-3333-4444-555555555555",
			Params: map[string]string{"_proto": "vless", "security": "tls", "sni": "example.com", "type": "ws", "path": "/ws", "host": "example.com"},
		},
		{
			ID: "trojan-grpc", Name: "Trojan", Host: "1.1.1.1", Port: "443",
			UUID:   "trojanpass",
			Params: map[string]string{"_proto": "trojan", "proto": "trojan", "security": "tls", "sni": "example.com", "type": "grpc", "serviceName": "svc"},
		},
		{
			ID: "hy2", Name: "Hysteria2", Host: "1.1.1.1", Port: "8443",
			UUID:   "hypass",
			Params: map[string]string{"_proto": "hysteria2", "proto": "hysteria2", "sni": "example.com", "obfs": "salamander", "obfs-password": "obfspass"},
		},
		{
			ID: "vmess-ws", Name: "VMess WS", Host: "1.1.1.1", Port: "443",
			UUID: "11111111-2222-3333-4444-555555555555",
			Params: map[string]string{"_proto": "vmess", "type": "ws", "path": "/ray", "host": "example.com",
				"security": "tls", "sni": "example.com", "method": "auto"},
		},
		{
			ID: "ss", Name: "Shadowsocks", Host: "1.1.1.1", Port: "8388",
			UUID:   "sspass",
			Params: map[string]string{"_proto": "shadowsocks", "method": "aes-256-gcm"},
		},
		{
			ID: "tuic", Name: "TUIC", Host: "1.1.1.1", Port: "443",
			UUID:   "tuicuuid",
			Params: map[string]string{"_proto": "tuic", "tuic_uuid": "11111111-2222-3333-4444-555555555555", "tuic_password": "tuicpass", "sni": "example.com", "congestion_control": "bbr"},
		},
		{
			ID: "anytls", Name: "AnyTLS", Host: "1.1.1.1", Port: "8443",
			UUID:   "atpass",
			Params: map[string]string{"_proto": "anytls", "sni": "example.com"},
		},
		{
			ID: "hysteria-v1", Name: "Hysteria v1", Host: "1.1.1.1", Port: "36712",
			UUID:   "h1pass",
			Params: map[string]string{"_proto": "hysteria", "peer": "example.com", "upmbps": "100", "downmbps": "200"},
		},
		{
			ID: "socks5", Name: "SOCKS5", Host: "1.1.1.1", Port: "1080",
			UUID:   "pass",
			Params: map[string]string{"_proto": "socks", "username": "user"},
		},
	}

	for _, srv := range cases {
		t.Run(srv.ID, func(t *testing.T) {
			cfg := models.AppConfig{
				VlessServers:   []models.VlessServer{srv},
				ActiveServerID: srv.ID,
				ProxyMode:      models.ProxyModeSystem,
				ProxyPort:      2080,
			}
			raw, err := RenderSingBoxConfig(cfg, t.TempDir())
			if err != nil {
				t.Fatalf("render failed: %v", err)
			}
			if !json.Valid(raw) {
				t.Fatalf("rendered config is not valid JSON")
			}
			path := filepath.Join(t.TempDir(), "sing-box.json")
			if err := os.WriteFile(path, raw, 0o644); err != nil {
				t.Fatalf("write failed: %v", err)
			}
			out, err := exec.Command(singBox, "check", "-c", path).CombinedOutput()
			if err != nil {
				t.Fatalf("sing-box check failed for %s: %v\n%s\n---config---\n%s", srv.ID, err, out, raw)
			}
		})
	}
}
