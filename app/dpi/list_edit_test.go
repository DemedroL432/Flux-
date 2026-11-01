package dpi

import (
	"testing"
)

// TestValidListNameRejectsTraversal проверяет защиту от выхода за пределы каталога списков.
func TestValidListNameRejectsTraversal(t *testing.T) {
	bad := []string{"", "   ", "../secret", "a/b", `a\b`, "..", "x..y", "a:b", "a|b", "a*b", "a?b"}
	for _, n := range bad {
		if _, err := validListName(n); err == nil {
			t.Errorf("validListName(%q) must fail", n)
		}
	}
	good := []string{"youtube", "custom-dpi", "ipset-youtube", "a_b-c.1"}
	for _, n := range good {
		if _, err := validListName(n); err != nil {
			t.Errorf("validListName(%q) = %v, want ok", n, err)
		}
	}
}

// TestValidListNameStripsTxt проверяет, что ".txt" отбрасывается.
func TestValidListNameStripsTxt(t *testing.T) {
	got, err := validListName("youtube.txt")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if got != "youtube" {
		t.Fatalf("got %q, want youtube", got)
	}
}

// TestWriteDpiListContentDedup проверяет нормализацию, дедупликацию и многострочный ввод.
func TestWriteDpiListContentDedup(t *testing.T) {
	// Проверяем чистую логику подготовки записей без файловой системы:
	// повторяем нормализацию из WriteDpiListContent.
	entries := []string{"https://A.com/x", "a.com", "  b.org  ", "b.org\nc.net"}
	seen := map[string]struct{}{}
	var clean []string
	for _, e := range entries {
		for _, line := range splitLines(e) {
			n := NormalizeHostEntry(line)
			if n == "" {
				continue
			}
			if _, ok := seen[n]; ok {
				continue
			}
			seen[n] = struct{}{}
			clean = append(clean, n)
		}
	}
	if len(clean) != 3 {
		t.Fatalf("expected 3 unique entries, got %v", clean)
	}
	want := []string{"a.com", "b.org", "c.net"}
	for i, w := range want {
		if clean[i] != w {
			t.Fatalf("entry %d = %q, want %q (all=%v)", i, clean[i], w, clean)
		}
	}
}

func splitLines(s string) []string {
	var out []string
	cur := ""
	for _, r := range s {
		if r == '\n' {
			out = append(out, cur)
			cur = ""
			continue
		}
		if r == '\r' {
			continue
		}
		cur += string(r)
	}
	out = append(out, cur)
	return out
}