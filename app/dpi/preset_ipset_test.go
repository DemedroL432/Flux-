package dpi

import (
	"strings"
	"testing"

	"pulseshield/app/models"
)

func TestFilterPreset_NoConfig_ReturnsUnchanged(t *testing.T) {
	in := "--filter-tcp=80,443\n--ipset=lists/youtube.txt\n--lua-desync=multisplit\n--new\n"
	if got := FilterPresetLinesByEnabledLists(in, nil); got != in {
		t.Fatalf("nil cfg must pass through unchanged")
	}
	cfg := &models.AppConfig{}
	if got := FilterPresetLinesByEnabledLists(in, cfg); got != in {
		t.Fatalf("empty DpiListEnabled must pass through unchanged")
	}
}

func TestFilterPreset_DropsOrphanedDesyncSection(t *testing.T) {
	// Секция таргетируется по двум ipset; оба выключены и desync остаётся → секцию выкинуть.
	in := strings.Join([]string{
		"--filter-tcp=80,443",
		"--ipset=lists/ipset-cloudflare.txt",
		"--ipset=lists/ipset-amazon.txt",
		"--lua-desync=syndata:blob=tls_google",
		"--new",
		"--filter-tcp=443",
		"--hostlist=lists/youtube.txt",
		"--lua-desync=multisplit",
		"--new",
	}, "\n")

	cfg := &models.AppConfig{DpiListEnabled: map[string]bool{
		"ipset-cloudflare": false,
		"ipset-amazon":     false,
	}}

	out := FilterPresetLinesByEnabledLists(in, cfg)
	if strings.Contains(out, "syndata") {
		t.Fatalf("orphaned catch-all desync section must be removed, got:\n%s", out)
	}
	if !strings.Contains(out, "lists/youtube.txt") {
		t.Fatalf("untouched youtube section must remain, got:\n%s", out)
	}
}

func TestFilterPreset_KeepsSectionWhenOneListStillEnabled(t *testing.T) {
	in := strings.Join([]string{
		"--filter-tcp=80,443",
		"--ipset=lists/ipset-cloudflare.txt",
		"--ipset=lists/ipset-amazon.txt",
		"--lua-desync=syndata:blob=tls_google",
		"--new",
	}, "\n")

	cfg := &models.AppConfig{DpiListEnabled: map[string]bool{
		"ipset-cloudflare": false, // выключен
		"ipset-amazon":     true,  // включён — секция остаётся
	}}

	out := FilterPresetLinesByEnabledLists(in, cfg)
	if !strings.Contains(out, "ipset-amazon.txt") {
		t.Fatalf("enabled list line must stay")
	}
	if strings.Contains(out, "ipset-cloudflare.txt") {
		t.Fatalf("disabled list line must be removed")
	}
	if !strings.Contains(out, "syndata") {
		t.Fatalf("section with a remaining target must keep its desync")
	}
}

func TestFilterPreset_KeepsPortOnlySection(t *testing.T) {
	// Секция изначально без таргетинг-фильтров (чистый порт) — намеренная, не трогаем.
	in := strings.Join([]string{
		"--filter-udp=443",
		"--lua-desync=fake:blob=quic_google",
		"--new",
	}, "\n")
	cfg := &models.AppConfig{DpiListEnabled: map[string]bool{"youtube": false}}
	out := FilterPresetLinesByEnabledLists(in, cfg)
	if !strings.Contains(out, "fake:blob=quic_google") {
		t.Fatalf("intentional port-only section must be preserved")
	}
}
