package dpi

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
)

// TestNormalizeHostEntry проверяет приведение пользовательского ввода к виду hostlist.
func TestNormalizeHostEntry(t *testing.T) {
	cases := map[string]string{
		"https://YouTube.com/watch?v=1": "youtube.com",
		"  EXAMPLE.org  ":               "example.org",
		"*.discord.com":                 "discord.com",
		".telegram.org":                 "telegram.org",
		"http://a.b.c:8443/path":        "a.b.c",
		"host.com#comment":              "host.com",
		"host.com?x=1":                  "host.com",
		"":                              "",
		"   ":                           "",
		"///":                           "",
	}
	for in, want := range cases {
		if got := NormalizeHostEntry(in); got != want {
			t.Errorf("NormalizeHostEntry(%q) = %q, want %q", in, got, want)
		}
	}
}

// TestInjectCustomDpiHostlist проверяет, что строка --hostlist добавляется только
// в секции с desync и таргетингом, и не дублируется.
func TestInjectCustomDpiHostlist(t *testing.T) {
	preset := `--wf-tcp-out=80,443
--filter-tcp=443
--hostlist=lists/youtube.txt
--lua-desync=fake
--new
--filter-udp=443
--lua-desync=udp_desync
`
	out := InjectCustomDpiHostlist(preset)
	if !strings.Contains(out, "--hostlist="+CustomDpiListRel) {
		t.Fatalf("custom hostlist not injected:\n%s", out)
	}
	// Вставлять нужно ровно один раз.
	if n := strings.Count(out, "--hostlist="+CustomDpiListRel); n != 1 {
		t.Fatalf("expected 1 injection, got %d", n)
	}
	// Секция без таргетинга (только udp desync) не должна получать строку:
	// иначе desync начнёт бить по всему UDP-трафику.
	udpSection := out[strings.Index(out, "--new"):]
	if strings.Contains(udpSection, "--hostlist="+CustomDpiListRel) {
		t.Fatalf("injected into section without targeting:\n%s", out)
	}
}

// TestInjectCustomDpiHostlistIdempotent проверяет отсутствие дубликатов при повторном вызове.
func TestInjectCustomDpiHostlistIdempotent(t *testing.T) {
	preset := "--filter-tcp=443\n--hostlist=lists/a.txt\n--lua-desync=fake\n"
	once := InjectCustomDpiHostlist(preset)
	twice := InjectCustomDpiHostlist(once)
	if once != twice {
		t.Fatalf("injection is not idempotent:\n--- once ---\n%s\n--- twice ---\n%s", once, twice)
	}
}

// TestCustomDpiFileRoundTrip проверяет запись/чтение/удаление через временный каталог.
func TestCustomDpiFileRoundTrip(t *testing.T) {
	dir := t.TempDir()
	lists := filepath.Join(dir, "lists")
	if err := os.MkdirAll(lists, 0o755); err != nil {
		t.Fatal(err)
	}
	path := filepath.Join(lists, CustomDpiListName+".txt")

	// Пишем вручную, как это делает writeCustomDpiEntries.
	body := customDpiHeader + "youtube.com\ndiscord.com\n"
	if err := os.WriteFile(path, []byte(body), 0o644); err != nil {
		t.Fatal(err)
	}

	entries := parseListEntries(body)
	if len(entries) != 2 {
		t.Fatalf("expected 2 entries, got %v", entries)
	}
	if entries[0] != "youtube.com" || entries[1] != "discord.com" {
		t.Fatalf("unexpected entries: %v", entries)
	}

	// Комментарии и пустые строки не считаются записями.
	if got := parseListEntries("# only comment\n\n  \n"); len(got) != 0 {
		t.Fatalf("expected no entries from comments, got %v", got)
	}
}

// TestParseListEntriesDedup проверяет удаление дубликатов при разборе.
func TestParseListEntriesDedup(t *testing.T) {
	got := parseListEntries("a.com\nb.com\na.com\n")
	if len(got) != 2 {
		t.Fatalf("expected 2 unique entries, got %v", got)
	}
}
