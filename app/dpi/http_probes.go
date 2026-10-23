package dpi

import (
	"context"
	"net/http"
	"strings"
	"time"
)

// DefaultProbeURLs — HTTPS-эндпоинты для проверки доступности «типичных» сервисов (могут быть недоступны без обхода).
func DefaultProbeURLs() []string {
	return []string{
		"https://www.youtube.com/",
		"https://discord.com/",
		"https://www.instagram.com/",
		"https://www.google.com/generate_204",
	}
}

// RunHTTPProbes выполняет GET по каждому URL; считает успехом ответ с кодом < 500 и без ошибки соединения.
func RunHTTPProbes(urls []string, perURL time.Duration) (hits int, total int, lines []string) {
	total = len(urls)
	if total == 0 {
		return 0, 0, nil
	}
	client := &http.Client{
		Timeout: perURL,
		CheckRedirect: func(_ *http.Request, via []*http.Request) error {
			if len(via) >= 4 {
				return http.ErrUseLastResponse
			}
			return nil
		},
	}
	for _, raw := range urls {
		u := strings.TrimSpace(raw)
		if u == "" {
			continue
		}
		ctx, cancel := context.WithTimeout(context.Background(), perURL)
		req, err := http.NewRequestWithContext(ctx, http.MethodGet, u, nil)
		if err != nil {
			cancel()
			lines = append(lines, shortErr(u, err))
			continue
		}
		req.Header.Set("User-Agent", "PulseShieldDpiProbe/1.0")
		resp, err := client.Do(req)
		cancel()
		if err != nil {
			lines = append(lines, shortErr(u, err))
			continue
		}
		_ = resp.Body.Close()
		if resp.StatusCode < 500 {
			hits++
			lines = append(lines, u+": ok ("+resp.Status+")")
		} else {
			lines = append(lines, u+": "+resp.Status)
		}
	}
	return hits, total, lines
}

func shortErr(u string, err error) string {
	s := err.Error()
	if len(s) > 120 {
		s = s[:117] + "…"
	}
	return u + ": " + s
}
