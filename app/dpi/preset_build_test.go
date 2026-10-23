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
