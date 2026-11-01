package dpi

import (
	"testing"
	"time"
)

// TestRunProbeTargetsEmpty проверяет, что пустой список не падает.
func TestProbeSummaryZero(t *testing.T) {
	s := RunProbeTargets(nil, time.Second)
	if s.Total != 0 || s.OK != 0 || len(s.Results) != 0 {
		t.Fatalf("expected empty summary, got %+v", s)
	}
}

// TestProbeSummaryScore проверяет, что Score отражает число успехов и среднюю задержку.
func TestProbeSummaryScore(t *testing.T) {
	s := ProbeSummary{OK: 3, Total: 6, AvgMs: 120}
	hits, avg := s.Score()
	if hits != 3 || avg != 120 {
		t.Fatalf("Score() = %d,%d want 3,120", hits, avg)
	}
}

// TestDefaultProbeTargetsIncludesDiscordGateway проверяет, что Discord проверяется
// и по gateway: голосовой шлюз блокируется отдельно от сайта.
func TestDefaultProbeTargetsIncludesDiscordGateway(t *testing.T) {
	targets := DefaultProbeTargets()
	if len(targets) == 0 {
		t.Fatal("no probe targets")
	}
	var hasGateway, hasDiscord bool
	for _, tg := range targets {
		if tg.Service == "Discord Gateway" {
			hasGateway = true
		}
		if tg.Service == "Discord" {
			hasDiscord = true
		}
		if tg.URL == "" || tg.Service == "" {
			t.Fatalf("incomplete target: %+v", tg)
		}
	}
	if !hasGateway || !hasDiscord {
		t.Fatalf("expected Discord and Discord Gateway probes, got %+v", targets)
	}
}

// TestShortProbeErr проверяет сжатие текста ошибки.
func TestShortProbeErr(t *testing.T) {
	long := make([]byte, 300)
	for i := range long {
		long[i] = 'a'
	}
	got := shortProbeErr(errString(string(long)))
	if len(got) > 121 {
		t.Fatalf("error text not truncated: len=%d", len(got))
	}
}

type errString string

func (e errString) Error() string { return string(e) }