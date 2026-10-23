//go:build windows

package uiwin

import (
	"context"
	"math"
	"net/http"
	"net/url"
	"sync"
	"time"

	"github.com/lxn/walk"
)

type LatencyGraph struct {
	mu     sync.Mutex
	values []float64
	maxN   int

	widget *walk.CustomWidget
	cancel context.CancelFunc
	currentProxy string
}

func NewLatencyGraph(maxN int) *LatencyGraph {
	if maxN <= 0 {
		maxN = 40
	}
	return &LatencyGraph{maxN: maxN}
}

func (g *LatencyGraph) Attach(w *walk.CustomWidget) {
	g.widget = w
}

func (g *LatencyGraph) Stop() {
	if g.cancel != nil {
		g.cancel()
		g.cancel = nil
	}
}

func (g *LatencyGraph) EnsureRunning(proxyURL string) {
	if proxyURL == g.currentProxy && g.cancel != nil {
		return
	}
	g.Stop()
	g.currentProxy = proxyURL
	ctx, cancel := context.WithCancel(context.Background())
	g.cancel = cancel

	go func() {
		t := time.NewTicker(3 * time.Second)
		defer t.Stop()
		for {
			select {
			case <-ctx.Done():
				return
			case <-t.C:
				ms := measureLatency(g.currentProxy)
				g.push(ms)
				if g.widget != nil {
					g.widget.Invalidate()
				}
			}
		}
	}()
}

func (g *LatencyGraph) push(ms float64) {
	g.mu.Lock()
	defer g.mu.Unlock()
	if math.IsNaN(ms) || math.IsInf(ms, 0) {
		return
	}
	g.values = append(g.values, ms)
	if len(g.values) > g.maxN {
		g.values = g.values[len(g.values)-g.maxN:]
	}
}

func (g *LatencyGraph) snapshot() []float64 {
	g.mu.Lock()
	defer g.mu.Unlock()
	out := make([]float64, len(g.values))
	copy(out, g.values)
	return out
}

func (g *LatencyGraph) Latest() float64 {
	g.mu.Lock()
	defer g.mu.Unlock()
	if len(g.values) == 0 {
		return 0
	}
	return g.values[len(g.values)-1]
}

func measureLatency(proxyURL string) float64 {
	transport := &http.Transport{}
	if proxyURL != "" {
		if u, err := url.Parse(proxyURL); err == nil {
			transport.Proxy = http.ProxyURL(u)
		}
	}
	client := &http.Client{
		Timeout:   5 * time.Second,
		Transport: transport,
	}
	start := time.Now()
	resp, err := client.Get("https://example.com/")
	if err != nil {
		return 0
	}
	_ = resp.Body.Close()
	return float64(time.Since(start).Milliseconds())
}

func (g *LatencyGraph) Paint(canvas *walk.Canvas, bounds walk.Rectangle) error {
	// Background
	bg, _ := walk.NewSolidColorBrush(rgbCardAlt())
	defer bg.Dispose()
	canvas.FillRectangle(bg, bounds)

	values := g.snapshot()
	if len(values) < 2 {
		return nil
	}

	// Find max (ignore zeros)
	max := 1.0
	for _, v := range values {
		if v > max {
			max = v
		}
	}
	if max < 1 {
		max = 1
	}

	padding := 10
	w := float64(bounds.Width - 2*padding)
	h := float64(bounds.Height - 2*padding)
	if w <= 1 || h <= 1 {
		return nil
	}

	pen, _ := walk.NewCosmeticPen(walk.PenSolid, rgbAccent())
	defer pen.Dispose()

	// Draw polyline
	n := len(values)
	prevX := float64(bounds.X + padding)
	prevY := float64(bounds.Y + padding + int(h))

	for i, v := range values {
		x := float64(bounds.X+padding) + (float64(i) * w / float64(n-1))
		yy := h
		if v > 0 {
			yy = h * (1.0 - (v / max))
		}
		y := float64(bounds.Y+padding) + yy

		if i > 0 {
			canvas.DrawLine(pen, walk.Point{X: int(prevX), Y: int(prevY)}, walk.Point{X: int(x), Y: int(y)})
		}
		prevX, prevY = x, y
	}

	return nil
}

