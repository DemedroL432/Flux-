package dpi

import (
	"context"
	"net/http"
	"strings"
	"sync"
	"time"
)

// ProbeTarget — эндпоинт проверки и человеческое имя сервиса.
type ProbeTarget struct {
	Service string
	URL     string
}

// ProbeResult — итог проверки одного эндпоинта.
type ProbeResult struct {
	Service string        `json:"service"`
	URL     string        `json:"url"`
	OK      bool          `json:"ok"`
	Status  string        `json:"status"`
	Latency time.Duration `json:"-"`
	LatencyMs int64       `json:"latencyMs"`
	Err     string        `json:"error,omitempty"`
}

// DefaultProbeTargets — сервисы, которые чаще всего ломаются под DPI.
// Discord проверяем и по веб-домену, и по gateway: голосовой шлюз живёт на
// отдельном хосте и нередко блокируется, даже когда сайт открывается.
func DefaultProbeTargets() []ProbeTarget {
	return []ProbeTarget{
		{Service: "YouTube", URL: "https://www.youtube.com/generate_204"},
		{Service: "Discord", URL: "https://discord.com/api/v9/gateway"},
		{Service: "Discord Gateway", URL: "https://gateway.discord.gg/"},
		{Service: "Twitch", URL: "https://www.twitch.tv/"},
		{Service: "Instagram", URL: "https://www.instagram.com/"},
		{Service: "Google", URL: "https://www.google.com/generate_204"},
	}
}

// DefaultProbeURLs оставлен для совместимости: список URL без имён сервисов.
func DefaultProbeURLs() []string {
	targets := DefaultProbeTargets()
	out := make([]string, 0, len(targets))
	for _, t := range targets {
		out = append(out, t.URL)
	}
	return out
}

// ProbeSummary — сводка по всем пробам одного пресета.
type ProbeSummary struct {
	OK        int           `json:"ok"`
	Total     int           `json:"total"`
	AvgMs     int64         `json:"avgMs"`
	MaxMs     int64         `json:"maxMs"`
	Results   []ProbeResult `json:"results"`
	FailedFor []string      `json:"failedFor"`
}

// Score возвращает вес результата: число успешных проб.
// Средняя задержка используется как тай-брейкер при равенстве.
func (s ProbeSummary) Score() (hits int, avgMs int64) {
	return s.OK, s.AvgMs
}

// RunProbeTargets выполняет пробы параллельно и возвращает сводку.
//
// Параллельность важна: последовательная проверка шести эндпоинтов с таймаутом
// по 6 секунд добавляла бы до 36 секунд на каждого кандидата автоподбора.
func RunProbeTargets(targets []ProbeTarget, perURL time.Duration) ProbeSummary {
	summary := ProbeSummary{Total: len(targets)}
	if len(targets) == 0 {
		return summary
	}

	client := &http.Client{
		Timeout: perURL,
		CheckRedirect: func(_ *http.Request, via []*http.Request) error {
			if len(via) >= 4 {
				return http.ErrUseLastResponse
			}
			return nil
		},
		Transport: &http.Transport{
			// Проба должна видеть реальное состояние канала, а не переиспользованное
			// соединение, поднятое до включения обхода.
			DisableKeepAlives:   true,
			MaxIdleConns:        1,
			TLSHandshakeTimeout: perURL,
		},
	}

	results := make([]ProbeResult, len(targets))
	var wg sync.WaitGroup
	for i, t := range targets {
		wg.Add(1)
		go func(idx int, target ProbeTarget) {
			defer wg.Done()
			results[idx] = runOneProbe(client, target, perURL)
		}(i, t)
	}
	wg.Wait()

	var sumMs int64
	var succeeded int
	for _, r := range results {
		if r.OK {
			succeeded++
			sumMs += r.LatencyMs
			if r.LatencyMs > summary.MaxMs {
				summary.MaxMs = r.LatencyMs
			}
		} else {
			summary.FailedFor = append(summary.FailedFor, r.Service)
		}
	}
	summary.OK = succeeded
	summary.Results = results
	if succeeded > 0 {
		summary.AvgMs = sumMs / int64(succeeded)
	}
	return summary
}

// runOneProbe проверяет один эндпоинт и замеряет время ответа.
func runOneProbe(client *http.Client, target ProbeTarget, perURL time.Duration) ProbeResult {
	res := ProbeResult{Service: target.Service, URL: target.URL}
	ctx, cancel := context.WithTimeout(context.Background(), perURL)
	defer cancel()

	req, err := http.NewRequestWithContext(ctx, http.MethodGet, target.URL, nil)
	if err != nil {
		res.Err = err.Error()
		res.Status = "bad request"
		return res
	}
	req.Header.Set("User-Agent", "Mozilla/5.0 (Windows NT 10.0; Win64; x64) FluxProbe/1.0")

	start := time.Now()
	resp, err := client.Do(req)
	elapsed := time.Since(start)
	res.Latency = elapsed
	res.LatencyMs = elapsed.Milliseconds()

	if err != nil {
		res.Err = shortProbeErr(err)
		res.Status = "нет ответа"
		return res
	}
	defer func() { _ = resp.Body.Close() }()

	// 2xx/3xx/4xx означают, что до сервиса достучались (4xx — уже его ответ).
	// 5xx и обрыв на TLS считаем провалом: признак вмешательства в канал.
	if resp.StatusCode < 500 {
		res.OK = true
		res.Status = resp.Status
	} else {
		res.Status = resp.Status
		res.Err = "сервер ответил ошибкой"
	}
	return res
}

// shortProbeErr сжимает текст ошибки до читаемого вида.
func shortProbeErr(err error) string {
	s := err.Error()
	// Убираем длинный префикс URL из ошибок net/http.
	if i := strings.LastIndex(s, ": "); i >= 0 && i < len(s)-2 {
		trimmed := strings.TrimSpace(s[i+2:])
		if trimmed != "" {
			s = trimmed
		}
	}
	if len(s) > 120 {
		s = s[:117] + "…"
	}
	return s
}

// RunHTTPProbes оставлен для совместимости со старым вызовом.
func RunHTTPProbes(urls []string, perURL time.Duration) (hits int, total int, lines []string) {
	targets := make([]ProbeTarget, 0, len(urls))
	for _, u := range urls {
		targets = append(targets, ProbeTarget{Service: u, URL: u})
	}
	s := RunProbeTargets(targets, perURL)
	for _, r := range s.Results {
		if r.OK {
			lines = append(lines, r.URL+": ok ("+r.Status+")")
		} else {
			lines = append(lines, r.URL+": "+r.Status)
		}
	}
	return s.OK, s.Total, lines
}

func shortErr(u string, err error) string {
	s := err.Error()
	if len(s) > 120 {
		s = s[:117] + "…"
	}
	return u + ": " + s
}
