package multiserver

import (
	"encoding/base64"
	"testing"
	"time"

	"pulseshield/app/models"
)

func TestDecodeSubscriptionPayload_Base64(t *testing.T) {
	raw := "vless://uuid@example.com:443?security=tls#Example"
	enc := base64.StdEncoding.EncodeToString([]byte(raw))
	decoded, err := decodeSubscriptionPayload(enc)
	if err != nil {
		t.Fatalf("decode failed: %v", err)
	}
	if decoded != raw {
		t.Fatalf("decoded payload mismatch: %q", decoded)
	}
}

func TestParseVlessString(t *testing.T) {
	raw := "vless://123e4567-e89b-12d3-a456-426614174000@example.com:8443?security=reality&fp=chrome&sni=example.com&type=ws&path=%2Fws#NodeA"
	srv, err := ParseVlessString(raw)
	if err != nil {
		t.Fatalf("ParseVlessString failed: %v", err)
	}
	if srv.Host != "example.com" || srv.Port != "8443" || srv.Name != "NodeA" {
		t.Fatalf("unexpected parsed server: %+v", srv)
	}
	if srv.Params["security"] != "reality" || srv.Params["fp"] != "chrome" {
		t.Fatalf("unexpected params: %+v", srv.Params)
	}
	srv2, err := ParseVlessString(raw)
	if err != nil {
		t.Fatalf("ParseVlessString second parse failed: %v", err)
	}
	if srv.ID != srv2.ID {
		t.Fatalf("expected stable ID for same VLESS link, got %q vs %q", srv.ID, srv2.ID)
	}
}

func TestValidateAndTestServersFiltersInvalidHost(t *testing.T) {
	servers := []*models.VlessServer{
		{Host: "", Port: "443"},
		{Host: "127.0.0.1", Port: "1"},
	}
	valid := ValidateAndTestServers(servers, 50*time.Millisecond)
	if len(valid) != 0 {
		t.Fatalf("expected no valid servers, got %d", len(valid))
	}
}

func TestParseTrojanAndHysteria2(t *testing.T) {
	tr, err := ParseVlessString("trojan://secretpass@nl.example.com:8443?security=tls&sni=nl.example.com&type=grpc&serviceName=%2F2096%2FxN6QgI1c#NL-Trojan")
	if err != nil {
		t.Fatalf("ParseVlessString trojan failed: %v", err)
	}
	if tr.Params["proto"] != "trojan" || tr.UUID != "secretpass" || tr.Host != "nl.example.com" || tr.Port != "8443" {
		t.Fatalf("unexpected trojan server: %+v", tr)
	}

	hy, err := ParseVlessString("hysteria2://hypass@de.example.com:32767?sni=www.bing.com&obfs=salamander&obfs-password=obfspass#DE-Hy2")
	if err != nil {
		t.Fatalf("ParseVlessString hysteria2 failed: %v", err)
	}
	if hy.Params["proto"] != "hysteria2" || hy.UUID != "hypass" || hy.Host != "de.example.com" || hy.Port != "32767" || hy.Params["obfs"] != "salamander" {
		t.Fatalf("unexpected hysteria2 server: %+v", hy)
	}

	if _, err := ParseVlessString("vless://00000000-0000-0000-0000-000000000000@127.0.0.1:1?type=tcp#InfoDummy"); err == nil {
		t.Fatalf("expected dummy 127.0.0.1:1 server to be rejected")
	}
}

