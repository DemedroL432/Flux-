package dpi

import (
	"strings"
	"testing"
)

func TestTrimLeadingPresetComments(t *testing.T) {
	raw := []byte("# a\n# b\n\n--wf-tcp-out=1\n")
	out := string(TrimLeadingPresetComments(raw))
	if strings.HasPrefix(strings.TrimSpace(out), "#") {
		t.Fatalf("expected comments stripped, got %q", out)
	}
	if !strings.Contains(out, "--wf-tcp-out") {
		t.Fatal(out)
	}
}

func TestReplaceWFPorts(t *testing.T) {
	in := []byte("--wf-tcp-out=1\n--wf-udp-out=2\n")
	out := string(ReplaceWFPorts(in, "80", "443-500"))
	if !strings.Contains(out, "--wf-tcp-out=80") || !strings.Contains(out, "--wf-udp-out=443-500") {
		t.Fatal(out)
	}
}

func TestDefaultGeneratedTuningCandidatesNonEmpty(t *testing.T) {
	c := DefaultGeneratedTuningCandidates()
	if len(c) < 2 {
		t.Fatal(len(c))
	}
	for _, x := range c {
		b := x.Build()
		if len(b) < 100 {
			t.Fatalf("%s: short body %d", x.ID, len(b))
		}
	}
}

func TestReplaceDiscordDesync(t *testing.T) {
	in := []byte(`--filter-tcp=443
--hostlist-domains=updates.discord.com,gateway.discord.gg
--out-range=-d8
--lua-desync=old_method:1

--new

--filter-tcp=80,443
--hostlist=lists/discord.txt
--out-range=-d8
--lua-desync=old_method:2
`)
	newDesync := "--lua-desync=new_test_method:123"
	out := string(ReplaceDiscordDesync(in, newDesync))
	if strings.Contains(out, "old_method") {
		t.Fatalf("expected old_method removed, got:\n%s", out)
	}
	if !strings.Contains(out, "new_test_method:123") {
		t.Fatalf("expected new_test_method present, got:\n%s", out)
	}
}
