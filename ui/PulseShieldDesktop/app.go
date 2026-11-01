package main

import (
	"bufio"
	"bytes"
	"context"
	_ "embed"
	"encoding/json"
	"fmt"
	"io"
	"net"
	"net/http"
	"os"
	"os/exec"
	"path/filepath"
	"sort"
	"strconv"
	"strings"
	"sync"
	"sync/atomic"
	"syscall"
	"time"

	"github.com/getlantern/systray"
	"github.com/getlantern/systray/example/icon"
	"github.com/wailsapp/wails/v2/pkg/runtime"
	"golang.org/x/sys/windows"
	"pulseshield/app/bootlog"
	"pulseshield/app/config"
	"pulseshield/app/core"
	"pulseshield/app/diagnostics"
	"pulseshield/app/dpi"
	"pulseshield/app/licensing"
	"pulseshield/app/models"
	"pulseshield/app/multiserver"
	"pulseshield/app/netdns"
	"pulseshield/app/state"
	"pulseshield/app/winproxy"
)

//go:embed guide_ru.txt
var programGuideRU string

const (
	runtimeApplyEvent    = "runtime-apply"
	runtimeWatchdogEvent = "runtime-watchdog"
	dpiConnectivityEvent = "dpi-connectivity"
	realtimeStatsEvent   = "realtime-stats"
)

// App struct
type App struct {
	ctx context.Context
	mu  sync.Mutex

	// applyMu сериализует тяжёлое применение ядра/DPI (без удержания mu на время I/O).
	applyMu sync.Mutex

	st          *state.AppState
	coreManager *core.Manager
	dpiRunner   *dpi.Runner

	appVersion string
	forceQuit  bool

	rtLastRecv uint64
	rtLastSent uint64
	rtLastAt   time.Time

	watchdogGen uint64 // увеличивается при новом apply — отменяет старые таймеры watchdog

	applyJobSeq atomic.Uint64 // монотонный id для событий runtime-apply (только Deferred)

	deviceOnce       sync.Once
	cachedDeviceHID  string // короткая подпись для компактного UI
	cachedDeviceFull string // полный SHA-256 hex HWID (на экране поддержки)

	shutdownOnce sync.Once

	// DPI connectivity probe (generate_204) — только когда DPI включён и процесс живой.
	dpiConnectivityWarn    bool
	dpiConnectivityDetail  string
	dpiConnectivityChecked time.Time

	rtSessionStart    time.Time
	rtSessionBaseRecv uint64
	rtSessionBaseSent uint64

	panelCacheMu           sync.Mutex
	dpiListsCached         []dpi.DpiList
	dpiListsCacheReady     bool
	routingSplitCached     []RoutingSplitEntry
	routingSplitCacheReady bool
}

type ImportedFile struct {
	Path    string `json:"path"`
	Content string `json:"content"`
}

type UIState struct {
	ProxyEnabled              bool                    `json:"proxyEnabled"`
	ProxyPort                 int                     `json:"proxyPort"`
	DPIEnabled                bool                    `json:"dpiEnabled"`
	Elevated                  bool                    `json:"elevated"`
	ProxyMode                 string                  `json:"proxyMode"`
	ActiveDPIProfile          string                  `json:"activeDpiProfile"`
	RoutingProfile            string                  `json:"routingProfile"`
	ActiveServerID            string                  `json:"activeServerId"`
	CoreStatus                string                  `json:"coreStatus"`
	DPIRunning                bool                    `json:"dpiRunning"`
	DnsPreset                 string                  `json:"dnsPreset"`
	DnsPrimary                string                  `json:"dnsPrimary"`
	DnsSecondary              string                  `json:"dnsSecondary"`
	SubscriptionURL           string                  `json:"subscriptionUrl"`
	SubscriptionExpiresUnix   int64                   `json:"subscriptionExpiresUnix"`
	SubscriptionProviderLabel string                  `json:"subscriptionProviderLabel"`
	TelegramInfoURL           string                  `json:"telegramInfoUrl"`
	TelegramChannelName       string                  `json:"telegramChannelName"`
	LicenseTier               string                  `json:"licenseTier"`
	LicenseExpiresUnix        int64                   `json:"licenseExpiresUnix"`
	LicenseActive             bool                    `json:"licenseActive"`
	DeviceIDShort             string                  `json:"deviceIdShort"`
	DeviceIDFull              string                  `json:"deviceIdFull"`
	LicenseAPIBase            string                  `json:"licenseApiBase"`
	PremiumServerMarkers      []string                `json:"premiumServerMarkers"`
	Servers                   []UIServer              `json:"servers"`
	DPIProfiles               []models.DPIProfile     `json:"dpiProfiles"`
	RoutingProfiles           []models.RoutingProfile `json:"routingProfiles"`
	StrictRoutingMode         bool                    `json:"strictRoutingMode"`
}

type UIServer struct {
	ID                    string  `json:"id"`
	Name                  string  `json:"name"`
	Host                  string  `json:"host"`
	Port                  string  `json:"port"`
	Protocol              string  `json:"protocol"`
	Source                string  `json:"source"`
	FromSubscription      bool    `json:"fromSubscription"`
	LastCheckOK           bool    `json:"lastCheckOk"`
	LastCheckLatencyMs    int     `json:"lastCheckLatencyMs"`
	LastCheckDownloadMbps float64 `json:"lastCheckDownloadMbps"`
	Recommended           bool    `json:"recommended"`
	Premium               bool    `json:"premium"`
}

type ProxyProcessDetail struct {
	Name         string  `json:"name"`
	PID          string  `json:"pid"`
	ConnCount    int     `json:"connCount"`
	DownloadBps  float64 `json:"downloadBps"`
	UploadBps    float64 `json:"uploadBps"`
	EstimateNote string  `json:"estimateNote"`
}

// ActiveConnectionSnapshot — строка монитора соединений sing-box (Clash API).
type ActiveConnectionSnapshot struct {
	ID       string `json:"id"`
	Process  string `json:"process"`
	Remote   string `json:"remote"`
	Upload   int64  `json:"upload"`
	Download int64  `json:"download"`
}

type RealtimeStats struct {
	Connected             bool                       `json:"connected"`
	Status                string                     `json:"status"`
	DownloadBps           float64                    `json:"downloadBps"`
	UploadBps             float64                    `json:"uploadBps"`
	ProxyProcessDetails   []ProxyProcessDetail       `json:"proxyProcessDetails"`
	ActiveConnections     []ActiveConnectionSnapshot `json:"activeConnections"`
	SessionDownloadBytes  int64                      `json:"sessionDownloadBytes"`
	SessionUploadBytes    int64                      `json:"sessionUploadBytes"`
	SessionElapsedSeconds int64                      `json:"sessionElapsedSeconds"`
}

type LogsState struct {
	CoreLogPath   string `json:"coreLogPath"`
	DPILogPath    string `json:"dpiLogPath"`
	CoreLogTail   string `json:"coreLogTail"`
	DPILogTail    string `json:"dpiLogTail"`
	CoreErrors    string `json:"coreErrors"`
	DPIErrors     string `json:"dpiErrors"`
	CoreConnLines string `json:"coreConnLines"`
	DPIConnLines  string `json:"dpiConnLines"`
}

// ServiceHealth — быстрая проверка: локальный mixed-прокси, TCP до VLESS-сервера, DPI-процесс.
type ServiceHealth struct {
	ProxyLocalOK      bool   `json:"proxyLocalOk"`
	ProxyListenPort   int    `json:"proxyListenPort"`
	VLESSRemote       string `json:"vlessRemote"`
	DPIRunning        bool   `json:"dpiRunning"`
	DPIEnabled        bool   `json:"dpiEnabled"`
	ProxyEnabled      bool   `json:"proxyEnabled"`
	ActiveServer      string `json:"activeServer"`
	ActiveDPIProfile  string `json:"activeDpiProfile"`
	DpiTool           string `json:"dpiTool"`
	DpiConfigPath     string `json:"dpiConfigPath"`
	ProxyPortOccupied bool   `json:"proxyPortOccupied"`
	// DpiConnectivityWarn — последняя проверка https://www.google.com/generate_204 не удалась (DPI включён, процесс запущен).
	DpiConnectivityWarn bool `json:"dpiConnectivityWarn"`
	// DpiConnectivityDetail — краткая причина (ошибка сети или неожиданный HTTP-код).
	DpiConnectivityDetail string `json:"dpiConnectivityDetail,omitempty"`
	// DpiConnectivityCheckedUnix — время последней проверки (Unix seconds), 0 если ещё не было.
	DpiConnectivityCheckedUnix int64 `json:"dpiConnectivityCheckedUnix"`
}

// DpiPresetTestResult — результат короткой проверки встроенного пресета winws2.
type DpiPresetTestResult struct {
	Label       string `json:"label"`
	SourceRel   string `json:"sourceRel"`
	ProfileName string `json:"profileName"`
	Ok          bool   `json:"ok"`
	Detail      string `json:"detail"`
}

// DpiTuneCandidateResult — этап подбора: живой пресет + счётчик HTTP-проб к сервисам.
type DpiTuneCandidateResult struct {
	Label       string `json:"label"`
	SourceRel   string `json:"sourceRel"`
	ProfileName string `json:"profileName"`
	PresetAlive bool   `json:"presetAlive"`
	ProbeHits   int    `json:"probeHits"`
	ProbeTotal  int    `json:"probeTotal"`
	Detail      string `json:"detail"`
}

// DpiTuneResult — итог автоподбора: таблица кандидатов и созданный профиль zapret:auto:…
type DpiTuneResult struct {
	Candidates     []DpiTuneCandidateResult `json:"candidates"`
	BestSourceRel  string                   `json:"bestSourceRel"`
	NewProfileName string                   `json:"newProfileName"`
	Message        string                   `json:"message"`
}

// zapretProfileNameFromConfigPath строит имя профиля как в config.SyncZapretPresetProfiles (presets/… → zapret:…).
func zapretProfileNameFromConfigPath(configPath string) string {
	s := filepath.ToSlash(strings.TrimPrefix(configPath, "presets/"))
	if s == "" {
		return ""
	}
	base := strings.TrimSuffix(s, filepath.Ext(s))
	id := strings.ReplaceAll(base, "/", ":")
	return "zapret:" + id
}

func findZapretProfileByConfigRel(cfg models.AppConfig, relSlash string) (models.DPIProfile, bool) {
	want := filepath.ToSlash(strings.TrimSpace(relSlash))
	for _, p := range cfg.DPIProfiles {
		if p.Tool != "zapret_winws2" {
			continue
		}
		if filepath.ToSlash(p.ConfigPath) == want {
			return p, true
		}
	}
	return models.DPIProfile{}, false
}

func (a *App) restoreDpiAfterTune(wasEnabled bool, prevProfile string) {
	a.mu.Lock()
	if a.st == nil {
		a.mu.Unlock()
		return
	}
	a.st.Update(func(cfg *models.AppConfig) {
		cfg.DPIEnabled = wasEnabled
		cfg.ActiveDPIProfile = prevProfile
	})
	_ = config.Save(a.st.Snapshot())
	a.mu.Unlock()
	_ = a.runApplySync(false, 0)
}

// NewApp creates a new App application struct
func NewApp(version string) *App {
	return &App{appVersion: strings.TrimSpace(version)}
}

// startup is called when the app starts. The context is saved
// so we can call the runtime methods
func (a *App) startup(ctx context.Context) {
	a.ctx = ctx
	cfg, err := config.LoadOrCreate()
	if err != nil {
		runtime.LogErrorf(ctx, "config load error: %v", err)
		cfg = config.DefaultConfig()
		config.SyncZapretPresetProfiles(&cfg)
		_ = config.MergeBuiltinRoutingProfiles(&cfg)
		config.EnsureLocalInboundSecrets(&cfg)
	}
	a.st = state.New(cfg)
	// Сразу синхронизируем домены из правил обхода DPI в custom-dpi.txt.
	initSnap := a.st.Snapshot()
	_ = dpi.SetCustomDpiEntries(config.CustomDpiEntriesForRules(&initSnap))
	a.invalidateRoutingDpiPanelCaches()

	coreBinary, _ := core.ResolveBinaryPath()
	a.coreManager = core.NewManager(core.NewAdapter(coreBinary))
	a.dpiRunner = &dpi.Runner{}
	dpi.CleanupStrayProcesses()
	go a.runSystray()

	go a.runDpiConnectivityWatchdog()

	// Синхронизируем процессы с сохранённым конфигом (в т.ч. гасим «осиротевший» sing-box).
	go func() {
		time.Sleep(120 * time.Millisecond)
		if err := a.runApplySync(false, 0); err != nil && a.ctx != nil {
			runtime.LogErrorf(a.ctx, "startup runtime sync: %v", err)
		}
	}()

	go func() {
		time.Sleep(200 * time.Millisecond)
		a.warmRoutingDpiPanelCaches()
	}()

	go a.emitRealtimeStatsLoop()
	go a.runSleepWakeWatchdog()
}

// runSleepWakeWatchdog обнаруживает выход из сна/гибернации по скачку настенных часов.
// Wails/WebView2 при пробуждении может «залипнуть»: фоновые горутины висят, GPU-контекст
// теряется, UI не отвечает. Перехват WM_POWERBROADCAST в Wails неудобен, поэтому используем
// надёжный кросс-механизм: тикаем раз в 2с и если реальная пауза была заметно больше ожидаемой
// (значит система спала) — выполняем восстановление и просим фронт переинициализироваться.
func (a *App) runSleepWakeWatchdog() {
	const tick = 2 * time.Second
	// Порог: если между тиками прошло сильно больше tick — был сон.
	const sleepThreshold = 8 * time.Second
	last := time.Now()
	for {
		time.Sleep(tick)
		now := time.Now()
		gap := now.Sub(last)
		last = now
		if gap < sleepThreshold {
			continue
		}
		// Восстановление обёрнуто в recover: лучше упасть с логом, чем зависнуть.
		func() {
			defer func() {
				if r := recover(); r != nil {
					bootlog.Append(fmt.Sprintf("sleep-wake recovery panic: %v", r))
				}
			}()
			a.recoverFromSleep(gap)
		}()
	}
}

// recoverFromSleep перезапускает фоновые проверки, чинит трей и просит WebView2 переподключиться.
func (a *App) recoverFromSleep(gap time.Duration) {
	bootlog.Append(fmt.Sprintf("resume from sleep detected (gap=%s) — recovering", gap.Round(time.Second)))
	if a.ctx == nil {
		return
	}
	// 1. Пере-применить иконку и подпись трея (после сна Shell_NotifyIcon часто «залипает»).
	refreshSystrayAfterResume()
	// 2. Сказать фронту переинициализировать таймеры/слушатели.
	runtime.EventsEmit(a.ctx, "system-resumed", gap.String())
	// 3. Принудительно перезагрузить WebView2: после сна GPU-контекст часто теряется и
	//    JS-цикл «залипает» — EventsEmit до него не доходит. Reload пересоздаёт renderer,
	//    при этом фронт заново тянет состояние через GetUIState/GetServiceHealth на маунте.
	go func() {
		defer func() {
			if r := recover(); r != nil {
				bootlog.Append(fmt.Sprintf("resume reload panic: %v", r))
			}
		}()
		time.Sleep(400 * time.Millisecond) // дать ОС досоздать сетевой стек
		runtime.WindowReloadApp(a.ctx)
	}()
}

func (a *App) emitRealtimeStatsLoop() {
	t := time.NewTicker(time.Second)
	defer t.Stop()
	for {
		select {
		case <-a.ctx.Done():
			return
		case <-t.C:
			if a.ctx == nil {
				continue
			}
			st := a.GetRealtimeStats()
			b, err := json.Marshal(st)
			if err != nil {
				continue
			}
			runtime.EventsEmit(a.ctx, realtimeStatsEvent, string(b))
		}
	}
}

func (a *App) invalidateRoutingDpiPanelCaches() {
	a.panelCacheMu.Lock()
	a.dpiListsCached = nil
	a.dpiListsCacheReady = false
	a.routingSplitCached = nil
	a.routingSplitCacheReady = false
	a.panelCacheMu.Unlock()
}

func (a *App) warmRoutingDpiPanelCaches() {
	// Список custom-dpi должен существовать и быть синхронизирован с правилами маршрутизации.
	a.mu.Lock()
	if a.st != nil {
		snap := a.st.Snapshot()
		_ = dpi.SetCustomDpiEntries(config.CustomDpiEntriesForRules(&snap))
	} else {
		_, _ = dpi.EnsureCustomDpiList()
	}
	a.mu.Unlock()

	a.invalidateRoutingDpiPanelCaches()
	_, _ = a.GetAvailableDpiLists()
	_, _ = a.GetRoutingSplitView()
}

// ShutdownCleanup останавливает DPI и ядро при полном выходе из приложения.
// Идемпотентен: при WM_ENDSESSION и OnShutdown подряд выполняется один раз.
func (a *App) ShutdownCleanup() {
	a.shutdownOnce.Do(func() {
		a.mu.Lock()
		defer a.mu.Unlock()
		port := 8080
		if a.st != nil {
			port = a.st.Snapshot().ListenPort()
		}
		if a.dpiRunner != nil {
			_ = a.dpiRunner.Stop()
		}
		dpi.CleanupStrayProcesses()
		if a.coreManager != nil {
			a.coreManager.ForceStopForShutdown(port)
		}
	})
}

// Greet returns a greeting for the given name
func (a *App) Greet(name string) string {
	return fmt.Sprintf("Hello %s, It's show time!", name)
}

func (a *App) GetUIState() UIState {
	a.mu.Lock()
	defer a.mu.Unlock()
	if a.st == nil {
		return UIState{}
	}
	s := a.st.Snapshot()
	status := ""
	if a.coreManager != nil {
		status = a.coreManager.Status()
	}
	dpiRunning := false
	if a.dpiRunner != nil {
		dpiRunning = a.dpiRunner.IsRunning()
	}
	effectiveServers := s.EffectiveServers()
	servers := make([]UIServer, 0, len(effectiveServers))
	for _, it := range effectiveServers {
		srv := it
		prem := models.ServerMatchesAuthorPremium(it, s.PremiumServerNameSubstrings)
		servers = append(servers, UIServer{
			ID:                    it.ID,
			Name:                  it.Name,
			Host:                  it.Host,
			Port:                  it.Port,
			Protocol:              multiserver.ProtocolOf(&srv),
			Source:                it.Source,
			FromSubscription:      it.FromSubscription,
			LastCheckOK:           it.LastCheckOK,
			LastCheckLatencyMs:    it.LastCheckLatencyMs,
			LastCheckDownloadMbps: it.LastCheckDownloadMbps,
			Recommended:           it.Recommended,
			Premium:               prem,
		})
	}

	deviceShort := ""
	deviceFull := ""
	a.deviceOnce.Do(func() {
		if h, err := licensing.DeviceHash(); err == nil && h != "" {
			if len(h) >= 16 {
				a.cachedDeviceHID = h[:16] + "…"
			} else {
				a.cachedDeviceHID = h
			}
			a.cachedDeviceFull = h
		}
	})
	deviceShort = a.cachedDeviceHID
	deviceFull = a.cachedDeviceFull

	return UIState{
		ProxyEnabled:              s.ProxyEnabled,
		ProxyPort:                 s.ListenPort(),
		DPIEnabled:                s.DPIEnabled,
		Elevated:                  isProcessElevated(),
		ProxyMode:                 string(s.ProxyMode),
		ActiveDPIProfile:          s.ActiveDPIProfile,
		RoutingProfile:            s.RoutingProfile,
		ActiveServerID:            s.ActiveServerID,
		CoreStatus:                status,
		DPIRunning:                dpiRunning,
		DnsPreset:                 s.DnsPreset,
		DnsPrimary:                s.DnsPrimary,
		DnsSecondary:              s.DnsSecondary,
		SubscriptionURL:           s.SubscriptionURL,
		SubscriptionExpiresUnix:   s.SubscriptionExpiresUnix,
		SubscriptionProviderLabel: s.SubscriptionProviderLabel,
		TelegramInfoURL:           s.TelegramInfoURL,
		TelegramChannelName:       s.TelegramChannelName,
		LicenseTier:               s.LicenseTier,
		LicenseExpiresUnix:        s.LicenseExpiresUnix,
		LicenseActive:             s.LicenseActive(),
		DeviceIDShort:             deviceShort,
		DeviceIDFull:              deviceFull,
		LicenseAPIBase:            s.LicenseAPIBase,
		PremiumServerMarkers:      append([]string(nil), s.PremiumServerNameSubstrings...),
		Servers:                   servers,
		DPIProfiles:               s.DPIProfiles,
		RoutingProfiles:           s.RoutingProfiles,
		StrictRoutingMode:         s.StrictRoutingMode,
	}
}

func (a *App) RelaunchAsAdmin() error {
	if isProcessElevated() {
		return nil
	}
	exePath, err := os.Executable()
	if err != nil {
		return err
	}
	cmd := fmt.Sprintf("Start-Process -FilePath '%s' -Verb RunAs", psQuote(exePath))
	if len(os.Args) > 1 {
		cmd += " -ArgumentList @(" + strings.Join(quotePSArgs(os.Args[1:]), ", ") + ")"
	}
	ps := exec.Command("powershell.exe", "-NoProfile", "-ExecutionPolicy", "Bypass", "-WindowStyle", "Hidden", "-Command", cmd)
	ps.SysProcAttr = &syscall.SysProcAttr{
		HideWindow:    true,
		CreationFlags: 0x08000000, // CREATE_NO_WINDOW
	}
	if err := ps.Start(); err != nil {
		return err
	}
	runtime.Quit(a.ctx)
	return nil
}

func (a *App) SetProxyEnabled(enabled bool) error {
	return a.updateConfig(func(cfg *models.AppConfig) { cfg.ProxyEnabled = enabled })
}

func (a *App) SetDPIEnabled(enabled bool) error {
	return a.updateConfig(func(cfg *models.AppConfig) { cfg.DPIEnabled = enabled })
}

func (a *App) SetStrictRoutingMode(enabled bool) error {
	err := a.updateConfig(func(cfg *models.AppConfig) { cfg.StrictRoutingMode = enabled })
	if err == nil {
		a.invalidateRoutingDpiPanelCaches()
	}
	return err
}

func (a *App) SetProxyMode(mode string) error {
	m := models.ProxyMode(mode)
	if m != models.ProxyModeSystem && m != models.ProxyModeTun && m != models.ProxyModeDirect {
		return fmt.Errorf("unsupported proxy mode: %s", mode)
	}
	return a.updateConfig(func(cfg *models.AppConfig) { cfg.ProxyMode = m })
}

func (a *App) SetActiveServer(id string) error {
	return a.updateConfig(func(cfg *models.AppConfig) { cfg.ActiveServerID = id })
}

func (a *App) SetSubscriptionURL(raw string) error {
	return a.updateConfig(func(cfg *models.AppConfig) { cfg.SubscriptionURL = strings.TrimSpace(raw) })
}

func (a *App) SetActiveDPIProfile(name string) error {
	return a.updateConfig(func(cfg *models.AppConfig) { cfg.ActiveDPIProfile = name })
}

func (a *App) SetRoutingProfile(name string) error {
	err := a.updateConfig(func(cfg *models.AppConfig) { cfg.RoutingProfile = name })
	if err == nil {
		a.invalidateRoutingDpiPanelCaches()
	}
	return err
}

func (a *App) SetProxyPort(port int) error {
	if port < 1 || port > 65535 {
		return fmt.Errorf("порт должен быть в диапазоне 1–65535")
	}
	return a.updateConfig(func(cfg *models.AppConfig) { cfg.ProxyPort = port })
}

func (a *App) SetDnsSettings(preset, primary, secondary string) error {
	return a.updateConfig(func(cfg *models.AppConfig) {
		cfg.DnsPreset = preset
		cfg.DnsPrimary = primary
		cfg.DnsSecondary = secondary
	})
}

func (a *App) ApplyDnsNow() error {
	a.mu.Lock()
	snap := a.st.Snapshot()
	a.mu.Unlock()
	return netdns.ApplyWindows(snap.DnsPreset, snap.DnsPrimary, snap.DnsSecondary)
}

func (a *App) SetTelegramInfoURL(u string) error {
	return a.updateConfig(func(cfg *models.AppConfig) { cfg.TelegramInfoURL = strings.TrimSpace(u) })
}

func (a *App) SetTelegramChannelName(name string) error {
	return a.updateConfig(func(cfg *models.AppConfig) { cfg.TelegramChannelName = strings.TrimSpace(name) })
}

func (a *App) SetLicenseAPIBase(u string) error {
	return a.updateConfig(func(cfg *models.AppConfig) { cfg.LicenseAPIBase = strings.TrimSpace(u) })
}

// SetPremiumServerMarkers задаёт подстроки для маркировки авторских Premium-узлов (по строке на элемент или через запятую).
func (a *App) SetPremiumServerMarkers(blob string) error {
	var parts []string
	for _, line := range strings.Split(blob, "\n") {
		line = strings.TrimSpace(line)
		if line == "" {
			continue
		}
		for _, chunk := range strings.Split(line, ",") {
			chunk = strings.TrimSpace(chunk)
			if chunk != "" {
				parts = append(parts, chunk)
			}
		}
	}
	return a.updateConfig(func(cfg *models.AppConfig) { cfg.PremiumServerNameSubstrings = parts })
}

// ActivateLicense обменивает код на параметры авторской подписки через HTTP API (см. docs/LICENSING.md).
func (a *App) ActivateLicense(code string) error {
	a.mu.Lock()
	base := ""
	if a.st != nil {
		base = strings.TrimSpace(a.st.Snapshot().LicenseAPIBase)
	}
	a.mu.Unlock()

	resp, err := licensing.ActivateCode(base, strings.TrimSpace(code), a.GetAppVersion())
	if err != nil {
		return err
	}

	a.mu.Lock()
	defer a.mu.Unlock()
	if a.st == nil {
		return fmt.Errorf("app state is not initialized")
	}
	a.st.Update(func(cfg *models.AppConfig) {
		cfg.LicenseTier = strings.TrimSpace(resp.Tier)
		cfg.LicenseExpiresUnix = resp.ValidUntilUnix
		cfg.LicenseToken = strings.TrimSpace(resp.Token)
	})
	if err := config.Save(a.st.Snapshot()); err != nil {
		return err
	}
	if a.ctx != nil {
		payload, _ := json.Marshal(map[string]any{"ok": true, "tier": resp.Tier})
		runtime.EventsEmit(a.ctx, "license-applied", string(payload))
	}
	return nil
}

func (a *App) GetAppVersion() string {
	a.mu.Lock()
	defer a.mu.Unlock()
	if a.appVersion != "" {
		return a.appVersion
	}
	return "dev"
}

func (a *App) ShouldExitOnClose() bool {
	a.mu.Lock()
	defer a.mu.Unlock()
	return a.forceQuit
}

// CloseToTray скрывает главное окно в трей и показывает уведомление.
func (a *App) CloseToTray() {
	if a.ctx != nil {
		runtime.WindowHide(a.ctx)
		notifyTrayHidden()
	}
}

// RequestQuit завершает приложение целиком (выход из трея или аварийный выход).
// forceQuit выставляется ДО Quit, иначе OnBeforeClose увидит false и просто скроет окно в трей.
func (a *App) RequestQuit() {
	a.mu.Lock()
	a.forceQuit = true
	ctx := a.ctx
	a.mu.Unlock()
	if ctx == nil {
		// Вызвали до OnStartup — завершаем процесс напрямую, иначе кнопка «закрыть» не сработает.
		go func() {
			time.Sleep(150 * time.Millisecond)
			os.Exit(0)
		}()
		return
	}
	runtime.Quit(ctx)
}

// refreshSystrayAfterResume повторно применяет иконку и подпись трея после выхода из сна,
// когда Shell_NotifyIcon мог «залипнуть» (иконка есть, но не реагирует).
func refreshSystrayAfterResume() {
	defer func() { _ = recover() }() // systray мог быть ещё не инициализирован
	if ico := trayIconICOBytes(); len(ico) > 0 {
		systray.SetTemplateIcon(ico, ico)
	} else {
		systray.SetTemplateIcon(icon.Data, icon.Data)
	}
	systray.SetTooltip("Flux — в трее")
}

func (a *App) runSystray() {
	systray.Run(func() {
		if ico := trayIconICOBytes(); len(ico) > 0 {
			systray.SetTemplateIcon(ico, ico)
		} else {
			systray.SetTemplateIcon(icon.Data, icon.Data)
		}
		systray.SetTooltip("Flux — в трее")
		mOpen := systray.AddMenuItem("Открыть окно", "")
		mQuit := systray.AddMenuItem("Выход", "")
		go func() {
			for range mOpen.ClickedCh {
				if a.ctx != nil {
					runtime.WindowUnminimise(a.ctx)
					runtime.Show(a.ctx)
				}
			}
		}()
		go func() {
			for range mQuit.ClickedCh {
				a.RequestQuit()
			}
		}()
	}, func() {})
}

// AddVLESS добавляет один узел из прокси-ссылки любого поддерживаемого протокола.
func (a *App) AddVLESS(raw string) error {
	raw = strings.TrimSpace(raw)
	// Пользователь мог вставить несколько ссылок сразу — добавляем все.
	if lines := strings.Split(raw, "\n"); len(lines) > 1 {
		added := 0
		var lastErr error
		for _, line := range lines {
			line = strings.TrimSpace(line)
			if line == "" || !multiserver.IsProxyLink(line) {
				continue
			}
			if err := a.AddVLESS(line); err != nil {
				lastErr = err
				continue
			}
			added++
		}
		if added == 0 {
			if lastErr != nil {
				return lastErr
			}
			return fmt.Errorf("в тексте не найдено прокси-ссылок")
		}
		return nil
	}
	server, err := multiserver.ParseProxyLink(raw)
	if err != nil {
		return err
	}
	server.Source = "manual"
	return a.updateConfig(func(cfg *models.AppConfig) {
		if cfg.SubscriptionURL != "" || len(cfg.VlessServers) > 0 {
			cfg.VlessServers = append(cfg.VlessServers, *server)
		} else {
			cfg.Servers = append(cfg.Servers, *server)
		}
		if cfg.ActiveServerID == "" {
			cfg.ActiveServerID = server.ID
		}
	})
}

// ImportServersFromText импортирует произвольный текст: ссылки, base64-список, Clash YAML,
// sing-box/Xray JSON. Возвращает количество добавленных узлов.
func (a *App) ImportServersFromText(raw string) (int, error) {
	raw = strings.TrimSpace(raw)
	if raw == "" {
		return 0, fmt.Errorf("пустой текст для импорта")
	}
	links, err := multiserver.ExtractSubscriptionLinks(raw)
	if err != nil {
		return 0, err
	}
	added := 0
	var firstErr error
	for _, link := range links {
		server, err := multiserver.ParseProxyLink(link)
		if err != nil {
			if firstErr == nil {
				firstErr = err
			}
			continue
		}
		server.Source = "manual"
		if err := a.updateConfig(func(cfg *models.AppConfig) {
			if cfg.SubscriptionURL != "" || len(cfg.VlessServers) > 0 {
				cfg.VlessServers = append(cfg.VlessServers, *server)
			} else {
				cfg.Servers = append(cfg.Servers, *server)
			}
			if cfg.ActiveServerID == "" {
				cfg.ActiveServerID = server.ID
			}
		}); err != nil {
			if firstErr == nil {
				firstErr = err
			}
			continue
		}
		added++
	}
	if added == 0 {
		if firstErr != nil {
			return 0, fmt.Errorf("не удалось импортировать ни одного узла: %w", firstErr)
		}
		return 0, fmt.Errorf("в тексте не найдено прокси-ссылок")
	}
	return added, nil
}

func (a *App) ImportServerFromClipboard() error {
	raw, err := runtime.ClipboardGetText(a.ctx)
	if err != nil {
		return err
	}
	raw = strings.TrimSpace(raw)
	if raw == "" {
		return fmt.Errorf("буфер обмена пуст")
	}
	// 1. Ссылка на подписку (http/https), которую нельзя перепутать с прокси-ссылкой.
	if isSubscriptionURLLine(raw) {
		if err := a.SetSubscriptionURL(raw); err != nil {
			return err
		}
		return a.RefreshSubscription()
	}
	// 2. Одиночная или множественная прокси-ссылка.
	if multiserver.IsProxyLink(raw) {
		return a.AddVLESS(raw)
	}
	// 3. Готовый список узлов / base64 / Clash YAML / JSON — импортируем как узлы.
	if links, err := multiserver.ExtractSubscriptionLinks(raw); err == nil && len(links) > 0 {
		_, impErr := a.ImportServersFromText(raw)
		return impErr
	}
	return fmt.Errorf("не распознан формат — нужна ссылка vless://, vmess://, trojan://, ss://, hysteria2://, tuic://, anytls://, ссылка подписки https://… или их список")
}

// isSubscriptionURLLine определяет, является ли строка HTTP(S)-ссылкой подписки.
func isSubscriptionURLLine(raw string) bool {
	low := strings.ToLower(strings.TrimSpace(raw))
	if !strings.HasPrefix(low, "http://") && !strings.HasPrefix(low, "https://") {
		return false
	}
	return !multiserver.IsProxyLink(raw)
}

// ClearSubscription удаляет URL подписки и все узлы, полученные из подписки (ручные остаются).
func (a *App) ClearSubscription() error {
	return a.updateConfig(func(cfg *models.AppConfig) {
		cfg.SubscriptionURL = ""
		cfg.SubscriptionExpiresUnix = 0
		cfg.SubscriptionProviderLabel = ""
		var next []models.VlessServer
		for _, s := range cfg.VlessServers {
			if !s.FromSubscription {
				next = append(next, s)
			}
		}
		cfg.VlessServers = next
		if cfg.ActiveServerID == "" {
			return
		}
		ok := containsServerID(cfg.VlessServers, cfg.ActiveServerID)
		if !ok {
			for _, s := range cfg.Servers {
				if s.ID == cfg.ActiveServerID {
					ok = true
					break
				}
			}
		}
		if ok {
			return
		}
		cfg.ActiveServerID = ""
		if len(cfg.VlessServers) > 0 {
			cfg.ActiveServerID = cfg.VlessServers[0].ID
		} else if len(cfg.Servers) > 0 {
			cfg.ActiveServerID = cfg.Servers[0].ID
		}
	})
}

// RetestServer перепроверяет ping/доступность одного узла и обновляет метрики в конфиге.
func (a *App) RetestServer(id string) error {
	id = strings.TrimSpace(id)
	if id == "" {
		return fmt.Errorf("пустой id")
	}
	a.mu.Lock()
	if a.st == nil {
		a.mu.Unlock()
		return fmt.Errorf("app state is not initialized")
	}
	snap := a.st.Snapshot()
	a.mu.Unlock()
	var found models.VlessServer
	for _, s := range snap.EffectiveServers() {
		if s.ID == id {
			found = s
			break
		}
	}
	if found.ID == "" {
		return fmt.Errorf("сервер не найден")
	}
	cp := found
	valid := multiserver.ValidateAndTestServers([]*models.VlessServer{&cp}, 3*time.Second)
	if len(valid) == 0 {
		cp.LastCheckOK = false
		cp.LastCheckedAt = time.Now()
	} else {
		cp = *valid[0]
	}
	return a.updateConfig(func(cfg *models.AppConfig) {
		replaceServerInConfig(cfg, cp)
	})
}

// RetestServersBatch перепроверяет ping и метрики для списка узлов (параллельно, до 8 одновременно).
func (a *App) RetestServersBatch(ids []string) error {
	seen := map[string]struct{}{}
	clean := make([]string, 0, len(ids))
	for _, id := range ids {
		id = strings.TrimSpace(id)
		if id == "" {
			continue
		}
		if _, ok := seen[id]; ok {
			continue
		}
		seen[id] = struct{}{}
		clean = append(clean, id)
	}
	if len(clean) == 0 {
		return fmt.Errorf("нет серверов для проверки")
	}
	if len(clean) > 500 {
		clean = clean[:500]
	}
	a.mu.Lock()
	if a.st == nil {
		a.mu.Unlock()
		return fmt.Errorf("app state is not initialized")
	}
	snap := a.st.Snapshot()
	a.mu.Unlock()

	var mu sync.Mutex
	var wg sync.WaitGroup
	sem := make(chan struct{}, 24)
	results := make([]models.VlessServer, 0, len(clean))
	for _, id := range clean {
		wg.Add(1)
		go func(id string) {
			defer wg.Done()
			sem <- struct{}{}
			defer func() { <-sem }()
			var found models.VlessServer
			for _, s := range snap.EffectiveServers() {
				if s.ID == id {
					found = s
					break
				}
			}
			if found.ID == "" {
				return
			}
			cp := found
			multiserver.ProbeServerMetrics(&cp, 3*time.Second)
			mu.Lock()
			results = append(results, cp)
			mu.Unlock()
		}(id)
	}
	wg.Wait()
	if len(results) == 0 {
		return fmt.Errorf("ни один сервер не найден")
	}
	return a.updateConfig(func(cfg *models.AppConfig) {
		for _, cp := range results {
			replaceServerInConfig(cfg, cp)
		}
	})
}

func replaceServerInConfig(cfg *models.AppConfig, srv models.VlessServer) {
	for i := range cfg.VlessServers {
		if cfg.VlessServers[i].ID == srv.ID {
			cfg.VlessServers[i] = srv
			return
		}
	}
	for i := range cfg.Servers {
		if cfg.Servers[i].ID == srv.ID {
			cfg.Servers[i] = models.Server(srv)
			return
		}
	}
}

// SubscriptionPreview — результат проверки ссылки подписки без сохранения в конфиг.
type SubscriptionPreview struct {
	OK            bool   `json:"ok"`
	URL           string `json:"url"`
	Format        string `json:"format"`
	ProviderLabel string `json:"providerLabel"`
	Name          string `json:"name"`
	Total         int    `json:"total"`
	Vless         int    `json:"vless"`
	VMess         int    `json:"vmess"`
	Trojan        int    `json:"trojan"`
	Shadowsocks   int    `json:"shadowsocks"`
	Hysteria2     int    `json:"hysteria2"`
	Hysteria1     int    `json:"hysteria"`
	Tuic          int    `json:"tuic"`
	AnyTLS        int    `json:"anytls"`
	Socks         int    `json:"socks"`
	HTTP          int    `json:"http"`
	Wireguard     int    `json:"wireguard"`
	Other         int    `json:"other"`
	ExpiresUnix   int64  `json:"expiresUnix"`
	TotalBytes    int64  `json:"totalBytes"`
	UsedBytes     int64  `json:"usedBytes"`
	Error         string `json:"error,omitempty"`
}

// PreviewSubscription скачивает подписку по произвольной ссылке и разбирает её,
// не изменяя сохранённый конфиг. Нужен, чтобы показать пользователю,
// что именно придёт по ссылке, до её применения.
func (a *App) PreviewSubscription(rawURL string) SubscriptionPreview {
	preview := SubscriptionPreview{URL: strings.TrimSpace(rawURL)}
	a.mu.Lock()
	proxyPort := 8080
	if a.st != nil {
		proxyPort = a.st.Snapshot().ListenPort()
	}
	a.mu.Unlock()

	links, meta, err := multiserver.FetchSubscriptionWithProxy(preview.URL, proxyPort)
	if err != nil {
		preview.Error = err.Error()
		return preview
	}
	preview.Format = meta.Format
	preview.ProviderLabel = meta.ProviderLabel
	preview.Name = meta.Name
	preview.ExpiresUnix = meta.ExpiresUnix
	preview.TotalBytes = meta.TotalBytes
	preview.UsedBytes = meta.UploadBytes + meta.DownloadBytes
	seen := make(map[string]bool, len(links))
	for _, link := range links {
		srv, err := multiserver.ParseProxyLink(link)
		if err != nil {
			continue
		}
		if seen[srv.ID] {
			continue
		}
		seen[srv.ID] = true
		preview.Total++
		switch multiserver.ProtocolOf(srv) {
		case "vless":
			preview.Vless++
		case "vmess":
			preview.VMess++
		case "trojan":
			preview.Trojan++
		case "shadowsocks":
			preview.Shadowsocks++
		case "hysteria2":
			preview.Hysteria2++
		case "hysteria":
			preview.Hysteria1++
		case "tuic":
			preview.Tuic++
		case "anytls":
			preview.AnyTLS++
		case "socks":
			preview.Socks++
		case "http":
			preview.HTTP++
		case "wireguard":
			preview.Wireguard++
		default:
			preview.Other++
		}
	}
	preview.OK = preview.Total > 0
	if !preview.OK {
		preview.Error = "в подписке не найдено поддерживаемых узлов"
	}
	return preview
}

// ParseServerLink разбирает одну прокси-ссылку и возвращает её поля для предпросмотра
// в диалоге добавления узла. Конфиг не изменяется.
func (a *App) ParseServerLink(raw string) (map[string]string, error) {
	srv, err := multiserver.ParseProxyLink(strings.TrimSpace(raw))
	if err != nil {
		return nil, err
	}
	params := make(map[string]string, len(srv.Params)+4)
	for k, v := range srv.Params {
		params[k] = v
	}
	params["protocol"] = multiserver.ProtocolOf(srv)
	return map[string]string{
		"id":       srv.ID,
		"name":     srv.Name,
		"host":     srv.Host,
		"port":     srv.Port,
		"protocol": multiserver.ProtocolOf(srv),
		"security": strings.ToLower(strings.TrimSpace(srv.Params["security"])),
		"network":  strings.ToLower(strings.TrimSpace(srv.Params["type"])),
		"params":   fmt.Sprintf("%v", params),
	}, nil
}

func (a *App) RefreshSubscription() error {
	a.mu.Lock()
	if a.st == nil {
		a.mu.Unlock()
		return fmt.Errorf("app state is not initialized")
	}
	snap := a.st.Snapshot()
	subscriptionURL := strings.TrimSpace(snap.SubscriptionURL)
	proxyPort := snap.ListenPort()
	a.mu.Unlock()
	if subscriptionURL == "" {
		return fmt.Errorf("subscription URL is empty")
	}
	rawLinks, subMeta, fetchErr := multiserver.FetchSubscriptionWithProxy(subscriptionURL, proxyPort)

	var servers []*models.VlessServer
	if fetchErr == nil && len(rawLinks) > 0 {
		servers = make([]*models.VlessServer, 0, len(rawLinks))
		for _, link := range rawLinks {
			srv, err := multiserver.ParseProxyLink(link)
			if err != nil {
				continue
			}
			srv.FromSubscription = true
			srv.Source = "subscription"
			servers = append(servers, srv)
		}
		if len(servers) == 0 {
			fetchErr = fmt.Errorf("в подписке не найдено поддерживаемых узлов")
		}
	}
	if fetchErr != nil || len(servers) == 0 {
		// Анти-ТСПУ фоллбэк: если провайдер заблокировал URL подписки, берём серверы из локального кэша
		// (даже если его TTL истёк) или из текущего списка и проверяем, какие из них живы прямо сейчас!
		cache, _, _ := multiserver.LoadCache(subscriptionURL)
		if cache != nil && len(cache.Servers) > 0 {
			for i := range cache.Servers {
				cp := cache.Servers[i]
				cp.FromSubscription = true
				cp.Source = "subscription"
				servers = append(servers, &cp)
			}
		}
		if len(servers) == 0 && len(snap.VlessServers) > 0 {
			for i := range snap.VlessServers {
				cp := snap.VlessServers[i]
				servers = append(servers, &cp)
			}
		}
		if len(servers) == 0 {
			return fmt.Errorf("сервер подписки заблокирован провайдером (ТСПУ), а локальный кэш пуст: %w", fetchErr)
		}
	}

	valid := multiserver.ValidateAndTestServers(servers, 2500*time.Millisecond)
	seenValid := make(map[string]bool, len(valid))
	for _, srv := range valid {
		seenValid[srv.ID] = true
	}
	// Сохраняем также остальные валидные узлы подписки в конец списка (на провайдерах вроде «Подряд»
	// ТСПУ может дропать часть массовых SYN-пакетов при проверке 150+ серверов одновременно,
	// хотя сами узлы рабочие через Reality/Hysteria2/gRPC).
	for _, srv := range servers {
		if srv != nil && !seenValid[srv.ID] {
			seenValid[srv.ID] = true
			valid = append(valid, srv)
		}
	}
	flat := make([]models.VlessServer, 0, len(valid))
	for _, srv := range valid {
		flat = append(flat, *srv)
	}
	for _, srv := range snap.VlessServers {
		if !srv.FromSubscription {
			flat = append(flat, srv)
		}
	}
	for _, srv := range snap.Servers {
		vs := models.VlessServer(srv)
		vs.FromSubscription = false
		flat = append(flat, vs)
	}
	flat = dedupeVlessServers(flat)

	if fetchErr == nil {
		subscriptionOnly := make([]models.VlessServer, 0, len(valid))
		for _, srv := range valid {
			subscriptionOnly = append(subscriptionOnly, *srv)
		}
		subscriptionOnly = dedupeVlessServers(subscriptionOnly)
		_ = multiserver.SaveCache(subscriptionURL, subscriptionOnly)
	}

	err := a.updateConfig(func(cfg *models.AppConfig) {
		activeID := cfg.ActiveServerID
		if subMeta.ExpiresUnix != 0 {
			cfg.SubscriptionExpiresUnix = subMeta.ExpiresUnix
		}
		if subMeta.ProviderLabel != "" {
			cfg.SubscriptionProviderLabel = subMeta.ProviderLabel
		}
		cfg.VlessServers = flat
		if len(cfg.Servers) > 0 {
			cfg.Servers = nil
		}
		if activeID == "" || !containsServerID(flat, activeID) {
			for _, srv := range flat {
				if srv.Recommended {
					cfg.ActiveServerID = srv.ID
					return
				}
			}
			cfg.ActiveServerID = flat[0].ID
		}
	})
	return err
}

func (a *App) UseCachedSubscription() error {
	a.mu.Lock()
	if a.st == nil {
		a.mu.Unlock()
		return fmt.Errorf("app state is not initialized")
	}
	snap := a.st.Snapshot()
	subscriptionURL := strings.TrimSpace(snap.SubscriptionURL)
	a.mu.Unlock()
	cache, _, err := multiserver.LoadCache(subscriptionURL)
	if err != nil {
		return err
	}
	if cache == nil || len(cache.Servers) == 0 {
		return fmt.Errorf("локальный кэш подписки пуст")
	}
	// Проверяем живые серверы из кэша, чтобы сразу поставить рабочий Recommended узел
	ptrs := make([]*models.VlessServer, 0, len(cache.Servers))
	for i := range cache.Servers {
		cp := cache.Servers[i]
		ptrs = append(ptrs, &cp)
	}
	valid := multiserver.ValidateAndTestServers(ptrs, 1500*time.Millisecond)
	var restored []models.VlessServer
	if len(valid) > 0 {
		for _, v := range valid {
			restored = append(restored, *v)
		}
	} else {
		restored = append(restored, cache.Servers...)
	}
	return a.updateConfig(func(cfg *models.AppConfig) {
		cfg.VlessServers = append([]models.VlessServer{}, restored...)
		for _, srv := range snap.VlessServers {
			if !srv.FromSubscription {
				cfg.VlessServers = append(cfg.VlessServers, srv)
			}
		}
		cfg.VlessServers = dedupeVlessServers(cfg.VlessServers)
		for _, srv := range cfg.VlessServers {
			if srv.Recommended {
				cfg.ActiveServerID = srv.ID
				return
			}
		}
		if cfg.ActiveServerID == "" || !containsServerID(cfg.VlessServers, cfg.ActiveServerID) {
			cfg.ActiveServerID = cfg.VlessServers[0].ID
		}
	})
}

func (a *App) RemoveServer(id string) error {
	if strings.TrimSpace(id) == "" {
		return fmt.Errorf("пустой id сервера")
	}
	return a.updateConfig(func(cfg *models.AppConfig) {
		var next []models.Server
		for _, s := range cfg.Servers {
			if s.ID != id {
				next = append(next, s)
			}
		}
		cfg.Servers = next
		var subNext []models.VlessServer
		for _, s := range cfg.VlessServers {
			if s.ID != id {
				subNext = append(subNext, s)
			}
		}
		cfg.VlessServers = subNext
		if cfg.ActiveServerID == id {
			cfg.ActiveServerID = ""
			if len(cfg.VlessServers) > 0 {
				cfg.ActiveServerID = cfg.VlessServers[0].ID
			} else if len(next) > 0 {
				cfg.ActiveServerID = next[0].ID
			}
		}
	})
}

func (a *App) GetServiceHealth() ServiceHealth {
	a.mu.Lock()
	var snap models.AppConfig
	if a.st != nil {
		snap = a.st.Snapshot()
	}
	dpiRun := a.dpiRunner != nil && a.dpiRunner.IsRunning()
	dpiConnWarn := a.dpiConnectivityWarn
	dpiConnDetail := a.dpiConnectivityDetail
	var dpiConnUnix int64
	if !a.dpiConnectivityChecked.IsZero() {
		dpiConnUnix = a.dpiConnectivityChecked.Unix()
	}
	a.mu.Unlock()

	listenPort := snap.ListenPort()
	out := ServiceHealth{
		DPIEnabled:                 snap.DPIEnabled,
		DPIRunning:                 dpiRun,
		ProxyEnabled:               snap.ProxyEnabled,
		ProxyListenPort:            listenPort,
		ActiveDPIProfile:           snap.ActiveDPIProfile,
		DpiConnectivityWarn:        dpiConnWarn,
		DpiConnectivityDetail:      dpiConnDetail,
		DpiConnectivityCheckedUnix: dpiConnUnix,
	}
	if prof, ok := findDPIProfile(snap, snap.ActiveDPIProfile); ok {
		out.DpiTool = prof.Tool
		out.DpiConfigPath = prof.ConfigPath
	}

	if srv, ok := snap.FindActiveServer(); ok {
		out.ActiveServer = srv.Name + " · " + net.JoinHostPort(srv.Host, srv.Port)
		addr := net.JoinHostPort(srv.Host, srv.Port)
		conn, err := net.DialTimeout("tcp", addr, 4*time.Second)
		if err != nil {
			out.VLESSRemote = trimErr(err)
		} else {
			_ = conn.Close()
			out.VLESSRemote = "ok"
		}
	} else {
		out.ActiveServer = "не выбран"
		out.VLESSRemote = "нет активного сервера"
	}

	out.ProxyLocalOK = false
	localAddr := net.JoinHostPort("127.0.0.1", strconv.Itoa(listenPort))
	if snap.ProxyEnabled {
		if c, err := net.DialTimeout("tcp", localAddr, 900*time.Millisecond); err == nil {
			_ = c.Close()
			out.ProxyLocalOK = true
		}
	} else {
		if c, err := net.DialTimeout("tcp", localAddr, 900*time.Millisecond); err == nil {
			_ = c.Close()
			out.ProxyPortOccupied = true
		}
	}

	return out
}

func (a *App) runDpiConnectivityWatchdog() {
	if a.ctx == nil {
		return
	}
	timer := time.NewTimer(30 * time.Second)
	defer timer.Stop()
	for {
		select {
		case <-a.ctx.Done():
			return
		case <-timer.C:
			a.probeDpiConnectivityOnce()
			timer.Reset(15 * time.Minute)
		}
	}
}

func (a *App) setDpiConnectivityResult(warn bool, detail string, ctx context.Context) {
	a.mu.Lock()
	prev := a.dpiConnectivityWarn
	a.dpiConnectivityWarn = warn
	a.dpiConnectivityDetail = detail
	a.dpiConnectivityChecked = time.Now()
	a.mu.Unlock()
	if ctx == nil || prev == warn {
		return
	}
	b, err := json.Marshal(map[string]any{"warn": warn, "detail": detail})
	if err != nil {
		return
	}
	runtime.EventsEmit(ctx, dpiConnectivityEvent, string(b))
}

func (a *App) probeDpiConnectivityOnce() {
	a.mu.Lock()
	var enabled bool
	if a.st != nil {
		enabled = a.st.Snapshot().DPIEnabled
	}
	running := a.dpiRunner != nil && a.dpiRunner.IsRunning()
	ctx := a.ctx
	a.mu.Unlock()

	if !enabled || !running || ctx == nil {
		a.setDpiConnectivityResult(false, "", ctx)
		return
	}

	const probeURL = "https://www.google.com/generate_204"
	reqCtx, cancel := context.WithTimeout(ctx, 12*time.Second)
	defer cancel()
	req, err := http.NewRequestWithContext(reqCtx, http.MethodGet, probeURL, nil)
	if err != nil {
		a.setDpiConnectivityResult(true, trimErr(err), ctx)
		return
	}
	req.Header.Set("User-Agent", "PulseShield-DpiWatchdog/1.0")

	client := &http.Client{}
	resp, err := client.Do(req)
	warn := false
	detail := ""
	if err != nil {
		warn = true
		detail = trimErr(err)
	} else {
		_, _ = io.Copy(io.Discard, resp.Body)
		_ = resp.Body.Close()
		if resp.StatusCode != http.StatusNoContent {
			warn = true
			detail = fmt.Sprintf("HTTP %d", resp.StatusCode)
		}
	}
	a.setDpiConnectivityResult(warn, detail, ctx)
}

func trimErr(err error) string {
	if err == nil {
		return ""
	}
	s := err.Error()
	if len(s) > 96 {
		return s[:96] + "…"
	}
	return s
}

func (a *App) ApplyRuntime() error {
	return a.runApplySync(false, 0)
}

// ApplyRuntimeDeferred применяет ядро/DPI в фоне и шлёт события runtime-apply (не блокирует Wails-RPC).
// Возвращает монотонный job id (в JSON событий как "job") для сопоставления на фронте.
func (a *App) ApplyRuntimeDeferred() (uint64, error) {
	if a.ctx == nil {
		return 0, fmt.Errorf("приложение ещё не инициализировано")
	}
	id := a.applyJobSeq.Add(1)
	go func(j uint64) { _ = a.runApplySync(true, j) }(id)
	return id, nil
}

func (a *App) StopAll() error {
	a.mu.Lock()
	if a.st == nil {
		a.mu.Unlock()
		return nil
	}
	a.st.Update(func(cfg *models.AppConfig) {
		cfg.ProxyEnabled = false
		cfg.DPIEnabled = false
	})
	err := config.Save(a.st.Snapshot())
	a.mu.Unlock()
	if err != nil {
		return err
	}
	return a.runApplySync(false, 0)
}

func (a *App) ApplyRoutingJSON(content string) error {
	a.mu.Lock()
	defer a.mu.Unlock()
	if a.st == nil {
		return fmt.Errorf("app state is not initialized")
	}
	snap := a.st.Snapshot()

	var profile models.RoutingProfile
	if err := json.Unmarshal([]byte(content), &profile); err != nil || len(profile.Rules) == 0 {
		var rules []models.RouteRule
		if err2 := json.Unmarshal([]byte(content), &rules); err2 != nil {
			return fmt.Errorf("invalid routing JSON")
		}
		profile = models.RoutingProfile{Rules: rules}
	}

	// Импортированный маршрут — ВСЕГДА отдельный профиль; встроенные пресеты
	// (default/selective-proxy/gaming-friendly/rf-smart-split) не перезаписываются.
	builtin := map[string]bool{"default": true, "selective-proxy": true, "gaming-friendly": true, "rf-smart-split": true}
	existing := map[string]bool{}
	for _, p := range snap.RoutingProfiles {
		existing[p.Name] = true
	}
	name := strings.TrimSpace(profile.Name)
	if name == "" || builtin[name] {
		base := "импортированный"
		if name != "" && builtin[name] {
			base = name + "-копия"
		}
		candidate := base
		for i := 2; existing[candidate]; i++ {
			candidate = fmt.Sprintf("%s-%d", base, i)
		}
		name = candidate
	}
	profile.Name = name

	a.st.Update(func(cfg *models.AppConfig) {
		replaced := false
		for i := range cfg.RoutingProfiles {
			if cfg.RoutingProfiles[i].Name == name {
				cfg.RoutingProfiles[i] = profile // обновить одноимённый пользовательский профиль
				replaced = true
				break
			}
		}
		if !replaced {
			cfg.RoutingProfiles = append(cfg.RoutingProfiles, profile)
		}
		cfg.RoutingProfile = profile.Name // переключиться на импортированный
	})
	return config.Save(a.st.Snapshot())
}

func (a *App) OpenRoutingJSON() (*ImportedFile, error) {
	path, err := runtime.OpenFileDialog(a.ctx, runtime.OpenDialogOptions{
		Title: "Открыть JSON маршрутизации",
		Filters: []runtime.FileFilter{
			{DisplayName: "JSON Files", Pattern: "*.json"},
			{DisplayName: "All Files", Pattern: "*.*"},
		},
	})
	if err != nil {
		return nil, err
	}
	if path == "" {
		return nil, nil
	}
	raw, err := os.ReadFile(path)
	if err != nil {
		return nil, err
	}
	return &ImportedFile{Path: path, Content: string(raw)}, nil
}

func (a *App) SaveRoutingJSON(content string) (string, error) {
	path, err := runtime.SaveFileDialog(a.ctx, runtime.SaveDialogOptions{
		Title:           "Сохранить JSON маршрутизации",
		DefaultFilename: "routing-profile.json",
		Filters: []runtime.FileFilter{
			{DisplayName: "JSON Files", Pattern: "*.json"},
		},
	})
	if err != nil {
		return "", err
	}
	if path == "" {
		return "", nil
	}
	if !strings.HasSuffix(strings.ToLower(path), ".json") {
		path += ".json"
	}
	if err := os.WriteFile(path, []byte(content), 0o644); err != nil {
		return "", err
	}
	return path, nil
}

func (a *App) ImportDPIConfig() (string, error) {
	src, err := runtime.OpenFileDialog(a.ctx, runtime.OpenDialogOptions{
		Title: "Импорт конфига DPI",
		Filters: []runtime.FileFilter{
			{DisplayName: "Text/Config", Pattern: "*.txt;*.yaml;*.yml;*.json;*.conf"},
			{DisplayName: "All Files", Pattern: "*.*"},
		},
	})
	if err != nil {
		return "", err
	}
	if src == "" {
		return "", nil
	}

	raw, err := os.ReadFile(src)
	if err != nil {
		return "", err
	}
	ext := strings.ToLower(filepath.Ext(src))
	base := filepath.Base(src)
	tool := "goodbyedpi"
	var dst string
	var profileName string
	if ext == ".txt" || ext == ".conf" {
		tool = "zapret_winws2"
		// For winws2 presets, keep custom files inside zapret/presets so relative
		// references to lists/iplist work exactly like bundled presets.
		zapretDir := findZapretBaseDir()
		if zapretDir == "" {
			return "", fmt.Errorf("zapret folder not found near app")
		}
		dstDir := filepath.Join(zapretDir, "presets", "custom")
		if err := os.MkdirAll(dstDir, 0o755); err != nil {
			return "", err
		}
		dst = filepath.Join(dstDir, base)
		profileName = "zapret:custom:" + strings.TrimSuffix(base, ext)
	} else {
		local := os.Getenv("LOCALAPPDATA")
		if local == "" {
			var home string
			home, err = os.UserHomeDir()
			if err != nil {
				return "", err
			}
			local = filepath.Join(home, "AppData", "Local")
		}
		dstDir := filepath.Join(local, "PulseShield", "dpi-custom")
		if err := os.MkdirAll(dstDir, 0o755); err != nil {
			return "", err
		}
		dst = filepath.Join(dstDir, base)
		profileName = "custom-import:" + strings.TrimSuffix(base, ext)
	}
	if err := os.WriteFile(dst, raw, 0o644); err != nil {
		return "", err
	}

	if ext == ".txt" || ext == ".conf" {
		rel, relErr := filepath.Rel(findZapretBaseDir(), dst)
		if relErr == nil {
			dst = filepath.ToSlash(rel)
		}
	}

	args := []string{"-c", "{config}"}
	if tool == "zapret_winws2" {
		if err := a.ResyncDPIProfilesFromDisk(); err != nil {
			return "", err
		}
		if err := a.updateConfig(func(cfg *models.AppConfig) {
			cfg.ActiveDPIProfile = profileName
		}); err != nil {
			return "", err
		}
		return dst, nil
	}
	if err := a.updateConfig(func(cfg *models.AppConfig) {
		for _, p := range cfg.DPIProfiles {
			if p.Name == profileName {
				cfg.ActiveDPIProfile = profileName
				return
			}
		}
		cfg.DPIProfiles = append(cfg.DPIProfiles, models.DPIProfile{
			Name:       profileName,
			Tool:       tool,
			ConfigPath: dst,
			Args:       args,
		})
		cfg.ActiveDPIProfile = profileName
	}); err != nil {
		return "", err
	}
	return dst, nil
}

func (a *App) ProgramGuide() string {
	return programGuideRU
}

func (a *App) ResyncDPIProfilesFromDisk() error {
	a.mu.Lock()
	defer a.mu.Unlock()
	if a.st == nil {
		return fmt.Errorf("приложение не готово")
	}
	synced := a.st.Snapshot()
	config.SyncZapretPresetProfiles(&synced)
	a.st.Update(func(cfg *models.AppConfig) {
		*cfg = synced
	})
	return config.Save(a.st.Snapshot())
}

// RunDpiPresetTests кратко проверяет кандидатов Gamingnew-Stable и Gamingnew: останавливает текущий DPI,
// запускает тестовый winws2, затем при необходимости восстанавливает рабочий DPI.
func (a *App) RunDpiPresetTests() ([]DpiPresetTestResult, error) {
	zapretDir := findZapretBaseDir()
	if zapretDir == "" {
		return nil, fmt.Errorf("папка zapret не найдена рядом с программой")
	}

	a.mu.Lock()
	if a.st == nil {
		a.mu.Unlock()
		return nil, fmt.Errorf("состояние приложения не инициализировано")
	}
	wasEnabled := a.st.Snapshot().DPIEnabled
	if a.dpiRunner != nil {
		_ = a.dpiRunner.Stop()
	}
	a.mu.Unlock()

	time.Sleep(450 * time.Millisecond)

	type cand struct {
		label string
		rel   string
	}
	candidates := []cand{
		{"Gamingnew-Stable", "presets/Gamingnew-Stable.txt"},
		{"Gamingnew", "presets/Gamingnew.txt"},
	}
	out := make([]DpiPresetTestResult, 0, len(candidates))
	for _, c := range candidates {
		pn := zapretProfileNameFromConfigPath(c.rel)
		full := filepath.Join(zapretDir, filepath.FromSlash(c.rel))
		if _, err := os.Stat(full); err != nil {
			out = append(out, DpiPresetTestResult{
				Label: c.label, SourceRel: c.rel, ProfileName: pn,
				Ok: false, Detail: "файл пресета не найден",
			})
			continue
		}
		ok, detail, err := dpi.TestWinws2Preset(c.rel, 1300*time.Millisecond)
		if err != nil {
			out = append(out, DpiPresetTestResult{
				Label: c.label, SourceRel: c.rel, ProfileName: pn,
				Ok: false, Detail: err.Error(),
			})
			continue
		}
		out = append(out, DpiPresetTestResult{
			Label: c.label, SourceRel: c.rel, ProfileName: pn,
			Ok: ok, Detail: detail,
		})
	}

	a.mu.Lock()
	needReapply := wasEnabled && a.st != nil && a.st.Snapshot().DPIEnabled
	a.mu.Unlock()
	if needReapply {
		if err := a.runApplySync(false, 0); err != nil {
			return out, fmt.Errorf("тесты выполнены; не удалось снова запустить DPI: %w", err)
		}
	}
	return out, nil
}

// summarizeProbe превращает сводку проб в строку для таблицы результатов.
func summarizeProbe(p dpi.ProbeSummary) string {
	if p.Total == 0 {
		return "пробы не выполнялись"
	}
	var b strings.Builder
	fmt.Fprintf(&b, "доступно %d/%d", p.OK, p.Total)
	if p.AvgMs > 0 {
		fmt.Fprintf(&b, ", средняя задержка %d мс", p.AvgMs)
	}
	if len(p.FailedFor) > 0 {
		b.WriteString("; не открылось: ")
		b.WriteString(strings.Join(p.FailedFor, ", "))
	}
	s := b.String()
	if len(s) > 800 {
		s = s[:797] + "…"
	}
	return s
}

// RunDpiAdaptiveTuning генерирует набор кандидатных пресетов в памяти (на базе встроенных эталонов zapret),
// для каждого — фильтрация по включённым спискам из конфига, опционально strategies.lua, затем winws2 и HTTP-пробы;
// лучший вариант сохраняется в presets/auto/.
func (a *App) RunDpiAdaptiveTuning(gamingLua bool) (*DpiTuneResult, error) {
	zapretDir := findZapretBaseDir()
	if zapretDir == "" {
		return nil, fmt.Errorf("папка zapret не найдена рядом с программой")
	}

	a.mu.Lock()
	if a.st == nil {
		a.mu.Unlock()
		return nil, fmt.Errorf("состояние приложения не инициализировано")
	}
	synced := a.st.Snapshot()
	config.SyncZapretPresetProfiles(&synced)
	wasEnabled := synced.DPIEnabled
	prevProfile := synced.ActiveDPIProfile
	if a.dpiRunner != nil {
		_ = a.dpiRunner.Stop()
	}
	a.mu.Unlock()
	time.Sleep(500 * time.Millisecond)

	gen := dpi.DefaultGeneratedTuningCandidates()
	// Пробы идут параллельно, поэтому таймаут можно держать коротким:
	// это ускоряет перебор кандидатов без потери точности.
	targets := dpi.DefaultProbeTargets()
	probeTO := 6 * time.Second

	out := make([]DpiTuneCandidateResult, 0, len(gen))
	bestHits := -1
	var bestAvgMs int64
	var bestBody []byte
	bestSource := ""

	autoDir := filepath.Join(zapretDir, "presets", "auto")
	if err := os.MkdirAll(autoDir, 0o755); err != nil {
		return nil, err
	}

	listSnap := synced

	for idx, c := range gen {
		func() {
			body := c.Build()
			text := strings.ReplaceAll(string(body), "\r\n", "\n")
			if gamingLua {
				text = dpi.EnsureStrategiesLuaInit(text)
			}
			text = dpi.FilterPresetLinesByEnabledLists(text, &listSnap)
			// Правила с целью «обход DPI» живут в custom-dpi.txt; подключаем его к кандидату,
			// иначе подбор оценит пресет без пользовательских доменов.
			text = dpi.InjectCustomDpiHostlist(text)
			body = []byte(text)
			fname := fmt.Sprintf("_pulse_tune_%03d_%s.txt", idx, dpi.SanitizeTuneFileName(c.ID))
			full := filepath.Join(autoDir, fname)
			rel := filepath.ToSlash(filepath.Join("presets", "auto", fname))

			if err := os.WriteFile(full, body, 0o644); err != nil {
				out = append(out, DpiTuneCandidateResult{
					Label: c.Label, SourceRel: "generated:" + c.ID, ProfileName: "",
					Detail: "запись временного файла: " + err.Error(),
				})
				return
			}
			defer func() { _ = os.Remove(full) }()

			pn := zapretProfileNameFromConfigPath(rel)
			alive, aliveDet, err := dpi.TestWinws2Preset(rel, 1200*time.Millisecond)
			if err != nil {
				out = append(out, DpiTuneCandidateResult{
					Label: c.Label, SourceRel: "generated:" + c.ID, ProfileName: pn,
					Detail: err.Error(),
				})
				return
			}
			if !alive {
				out = append(out, DpiTuneCandidateResult{
					Label: c.Label, SourceRel: "generated:" + c.ID, ProfileName: pn,
					Detail: aliveDet,
				})
				return
			}

			a.mu.Lock()
			fresh := a.st.Snapshot()
			config.SyncZapretPresetProfiles(&fresh)
			prof, ok := findZapretProfileByConfigRel(fresh, rel)
			_ = a.dpiRunner.Stop()
			a.mu.Unlock()
			if !ok {
				out = append(out, DpiTuneCandidateResult{
					Label: c.Label, SourceRel: "generated:" + c.ID, ProfileName: pn, PresetAlive: true,
					Detail: "профиль не найден в конфиге после синхронизации",
				})
				return
			}

			a.mu.Lock()
			if a.dpiRunner == nil {
				a.mu.Unlock()
				out = append(out, DpiTuneCandidateResult{
					Label: c.Label, SourceRel: "generated:" + c.ID, ProfileName: pn, PresetAlive: true,
					Detail: "модуль DPI не инициализирован",
				})
				return
			}
			if err := a.dpiRunner.Start(prof, &fresh); err != nil {
				a.mu.Unlock()
				out = append(out, DpiTuneCandidateResult{
					Label: c.Label, SourceRel: "generated:" + c.ID, ProfileName: pn, PresetAlive: true,
					Detail: "запуск DPI: " + err.Error(),
				})
				return
			}
			a.mu.Unlock()

			// Даём winws2 подняться и применить фильтры, прежде чем мерить.
			time.Sleep(2 * time.Second)
			probe := dpi.RunProbeTargets(targets, probeTO)
			probeText := summarizeProbe(probe)

			a.mu.Lock()
			if a.dpiRunner != nil {
				_ = a.dpiRunner.Stop()
			}
			a.mu.Unlock()
			time.Sleep(350 * time.Millisecond)

			out = append(out, DpiTuneCandidateResult{
				Label: c.Label, SourceRel: "generated:" + c.ID, ProfileName: pn,
				PresetAlive: true, ProbeHits: probe.OK, ProbeTotal: probe.Total,
				Detail: aliveDet + " | " + probeText,
			})
			// Ранжирование: больше успешных проб — лучше. При равенстве выигрывает
			// кандидат с меньшей средней задержкой.
			if probe.OK > bestHits || (probe.OK == bestHits && bestHits > 0 && probe.AvgMs > 0 &&
				(bestAvgMs == 0 || probe.AvgMs < bestAvgMs)) {
				bestHits = probe.OK
				bestAvgMs = probe.AvgMs
				bestBody = append([]byte(nil), body...)
				bestSource = "generated:" + c.ID
			}
		}()
	}

	res := &DpiTuneResult{Candidates: out, BestSourceRel: bestSource}
	if bestSource == "" {
		for _, r := range out {
			if r.PresetAlive && strings.HasPrefix(r.SourceRel, "generated:") {
				res.BestSourceRel = r.SourceRel
				break
			}
		}
	}

	if len(bestBody) == 0 {
		a.restoreDpiAfterTune(wasEnabled, prevProfile)
		res.Message = "Автоподбор не смог сгенерировать и проверить пресет — смотрите таблицу кандидатов."
		return res, nil
	}

	final := append([]byte(nil), bestBody...)
	final = append([]byte(fmt.Sprintf(
		"# Сохранено после диагностики: %s\n",
		time.Now().Format("2006-01-02 15:04:05"),
	)), final...)

	profileName, msg, err := a.applyDpiGeneratedBytes(final, true)
	if err != nil {
		a.restoreDpiAfterTune(wasEnabled, prevProfile)
		res.Message = "Не удалось сохранить новый профиль."
		return res, err
	}
	res.NewProfileName = profileName
	res.Message = msg
	res.BestSourceRel = bestSource
	return res, nil
}

// applyDpiGeneratedBytes сохраняет уже сформированный текст пресета (полностью сгенерированный в памяти) в presets/auto/<timestamp>.txt.
func (a *App) applyDpiGeneratedBytes(raw []byte, enableDPI bool) (profileName string, msg string, err error) {
	zapretDir := findZapretBaseDir()
	if zapretDir == "" {
		return "", "", fmt.Errorf("папка zapret не найдена рядом с программой")
	}
	if len(bytes.TrimSpace(raw)) == 0 {
		return "", "", fmt.Errorf("пустой пресет")
	}
	ts := time.Now().Format("20060102150405")
	autoDir := filepath.Join(zapretDir, "presets", "auto")
	if err := os.MkdirAll(autoDir, 0o755); err != nil {
		return "", "", err
	}
	dstFile := ts + ".txt"
	dstPath := filepath.Join(autoDir, dstFile)
	if err := os.WriteFile(dstPath, raw, 0o644); err != nil {
		return "", "", err
	}
	relForName := filepath.ToSlash(filepath.Join("presets", "auto", dstFile))
	pn := zapretProfileNameFromConfigPath(relForName)

	a.mu.Lock()
	if a.st == nil {
		a.mu.Unlock()
		return "", "", fmt.Errorf("состояние приложения не инициализировано")
	}
	a.st.Update(func(cfg *models.AppConfig) {
		config.SyncZapretPresetProfiles(cfg)
		cfg.ActiveDPIProfile = pn
		if enableDPI {
			cfg.DPIEnabled = true
		}
	})
	if _, ok := findDPIProfile(a.st.Snapshot(), pn); !ok {
		a.mu.Unlock()
		return "", "", fmt.Errorf("профиль %s не найден после синхронизации", pn)
	}
	if err := config.Save(a.st.Snapshot()); err != nil {
		a.mu.Unlock()
		return "", "", err
	}
	a.mu.Unlock()

	if err := a.runApplySync(false, 0); err != nil {
		return "", "", err
	}
	return pn, fmt.Sprintf("Создан новый профиль %s: параметры подобраны диагностикой и сохранены.", pn), nil
}

func (a *App) applyDpiAutoSnapshotBytes(sourceConfigRel string, prepend []byte, enableDPI bool) (profileName string, msg string, err error) {
	sourceConfigRel = filepath.ToSlash(strings.TrimSpace(sourceConfigRel))
	if sourceConfigRel == "" {
		return "", "", fmt.Errorf("не указан файл пресета")
	}
	zapretDir := findZapretBaseDir()
	if zapretDir == "" {
		return "", "", fmt.Errorf("папка zapret не найдена рядом с программой")
	}
	src := filepath.Join(zapretDir, filepath.FromSlash(sourceConfigRel))
	orig, err := os.ReadFile(src)
	if err != nil {
		return "", "", err
	}
	raw := orig
	if len(prepend) > 0 {
		raw = append(append([]byte{}, prepend...), orig...)
	}
	// Подключаем пользовательский список обхода DPI, чтобы снимок не терял
	// домены из правил с действием dpi-bypass.
	raw = []byte(dpi.InjectCustomDpiHostlist(string(raw)))
	if _, err := dpi.EnsureCustomDpiList(); err != nil {
		return "", "", fmt.Errorf("список custom-dpi.txt: %w", err)
	}

	ts := time.Now().Format("20060102150405")
	autoDir := filepath.Join(zapretDir, "presets", "auto")
	if err := os.MkdirAll(autoDir, 0o755); err != nil {
		return "", "", err
	}
	dstFile := ts + ".txt"
	dstPath := filepath.Join(autoDir, dstFile)
	if err := os.WriteFile(dstPath, raw, 0o644); err != nil {
		return "", "", err
	}
	relForName := filepath.ToSlash(filepath.Join("presets", "auto", dstFile))
	pn := zapretProfileNameFromConfigPath(relForName)

	a.mu.Lock()
	if a.st == nil {
		a.mu.Unlock()
		return "", "", fmt.Errorf("состояние приложения не инициализировано")
	}
	a.st.Update(func(cfg *models.AppConfig) {
		config.SyncZapretPresetProfiles(cfg)
		cfg.ActiveDPIProfile = pn
		if enableDPI {
			cfg.DPIEnabled = true
		}
	})
	if _, ok := findDPIProfile(a.st.Snapshot(), pn); !ok {
		a.mu.Unlock()
		return "", "", fmt.Errorf("профиль %s не найден после синхронизации", pn)
	}
	if err := config.Save(a.st.Snapshot()); err != nil {
		a.mu.Unlock()
		return "", "", err
	}
	a.mu.Unlock()

	if err := a.runApplySync(false, 0); err != nil {
		return "", "", err
	}
	return pn, fmt.Sprintf("Создан новый профиль %s: параметры подобраны диагностикой и сохранены.", pn), nil
}

// ApplyDpiAutoSnapshot копирует указанный пресет в новый файл presets/auto/<timestamp>.txt — профиль zapret:auto:<timestamp>.
func (a *App) ApplyDpiAutoSnapshot(sourceConfigRel string) (string, error) {
	_, msg, err := a.applyDpiAutoSnapshotBytes(sourceConfigRel, nil, false)
	return msg, err
}

// ApplyDpiAutotunePreset — быстрый путь: первый доступный из Stable/Gamingnew → auto-снимок zapret:auto:…
func (a *App) ApplyDpiAutotunePreset() (string, error) {
	zapretDir := findZapretBaseDir()
	if zapretDir == "" {
		return "", fmt.Errorf("папка zapret не найдена рядом с программой")
	}
	for _, c := range []string{"presets/Gamingnew-Stable.txt", "presets/Gamingnew.txt"} {
		if _, err := os.Stat(filepath.Join(zapretDir, filepath.FromSlash(c))); err == nil {
			return a.ApplyDpiAutoSnapshot(c)
		}
	}
	return "", fmt.Errorf("не найдены Gamingnew-Stable.txt или Gamingnew.txt в zapret/presets")
}

// ExportDpiPreset сохраняет копию пресета в файл по выбору пользователя (экспорт .txt).
func (a *App) ExportDpiPreset(sourceConfigRel string) (string, error) {
	sourceConfigRel = filepath.ToSlash(strings.TrimSpace(sourceConfigRel))
	if sourceConfigRel == "" {
		return "", fmt.Errorf("не указан пресет")
	}
	zapretDir := findZapretBaseDir()
	if zapretDir == "" {
		return "", fmt.Errorf("папка zapret не найдена рядом с программой")
	}
	src := filepath.Join(zapretDir, filepath.FromSlash(sourceConfigRel))
	raw, err := os.ReadFile(src)
	if err != nil {
		return "", err
	}
	name := filepath.Base(sourceConfigRel)
	path, err := runtime.SaveFileDialog(a.ctx, runtime.SaveDialogOptions{
		Title:           "Экспорт пресета DPI",
		DefaultFilename: name,
		Filters: []runtime.FileFilter{
			{DisplayName: "Текст", Pattern: "*.txt"},
			{DisplayName: "Все файлы", Pattern: "*.*"},
		},
	})
	if err != nil {
		return "", err
	}
	if path == "" {
		return "", nil
	}
	if err := os.WriteFile(path, raw, 0o644); err != nil {
		return "", err
	}
	return path, nil
}

func (a *App) ClearLogFiles() error {
	a.mu.Lock()
	defer a.mu.Unlock()
	corePath := a.resolveCoreLogPath()
	if corePath != "" {
		if err := os.WriteFile(corePath, []byte{}, 0o644); err != nil {
			return err
		}
	}
	if a.dpiRunner != nil {
		if p := a.dpiRunner.LastLogPath(); p != "" {
			if err := os.WriteFile(p, []byte{}, 0o644); err != nil {
				return err
			}
		}
	}
	return nil
}

func systemProxyStatusLine(snap models.AppConfig) string {
	if snap.ProxyMode != models.ProxyModeSystem {
		return "ℹ️ Системный прокси Windows: не используется (режим ядра: " + string(snap.ProxyMode) + ")"
	}
	if !snap.ProxyEnabled {
		return "ℹ️ Системный прокси Windows: приложение не включает прокси (локальный прокси выключен)"
	}
	ws, err := winproxy.Read()
	if err != nil {
		return "⚠️ Системный прокси Windows: не удалось прочитать настройки: " + err.Error()
	}
	if ws.ProxyEnable == 0 {
		return "⚠️ Системный прокси Windows: в реестре выключен — при режиме system_proxy включите прокси в приложении и нажмите «Применить»"
	}
	port := snap.ListenPort()
	expected := "127.0.0.1:" + strconv.Itoa(port)
	low := strings.ToLower(ws.ProxyServer)
	if strings.Contains(low, strings.ToLower(expected)) {
		return "✅ Системный прокси Windows: указывает на " + expected
	}
	return "⚠️ Системный прокси Windows: ожидается " + expected + ", сейчас: " + ws.ProxyServer
}

func (a *App) RunSelfCheckText() (string, error) {
	rep, err := diagnostics.Run()
	if err != nil {
		return "", err
	}
	health := a.GetServiceHealth()
	var snap models.AppConfig
	a.mu.Lock()
	if a.st != nil {
		snap = a.st.Snapshot()
	}
	a.mu.Unlock()
	ui := a.GetUIState()
	coreLog := ""
	if a.coreManager != nil {
		coreLog = a.coreManager.LastLogPath()
	}
	dpiLog := ""
	if a.dpiRunner != nil {
		dpiLog = a.dpiRunner.LastLogPath()
	}

	var cfgLine string
	if err := config.ValidateAppConfig(snap); err != nil {
		cfgLine = "❌ Конфигурация: " + err.Error()
	} else {
		cfgLine = "✅ Конфигурация маршрутизации валидна"
	}

	var sb strings.Builder
	sb.WriteString("=== Чек-лист ===\n")
	sb.WriteString(cfgLine + "\n")
	sb.WriteString(systemProxyStatusLine(snap) + "\n")

	if health.ProxyPortOccupied && !health.ProxyEnabled {
		sb.WriteString(fmt.Sprintf("⚠️ Порт mixed %d занят другим процессом, пока локальный прокси выключен.\n", health.ProxyListenPort))
	} else if health.ProxyEnabled {
		if health.ProxyLocalOK {
			sb.WriteString(fmt.Sprintf("✅ Локальный mixed-прокси отвечает на порту %d.\n", health.ProxyListenPort))
		} else {
			sb.WriteString(fmt.Sprintf("⚠️ Прокси включён, но порт %d не отвечает — проверьте ядро и порт.\n", health.ProxyListenPort))
		}
	} else {
		sb.WriteString("ℹ️ Локальный mixed-прокси выключен.\n")
	}

	switch health.VLESSRemote {
	case "ok":
		sb.WriteString("✅ TCP до активного VLESS-сервера: доступен.\n")
	default:
		if strings.Contains(health.VLESSRemote, "нет активного") || strings.Contains(health.VLESSRemote, "не выбран") {
			sb.WriteString("ℹ️ VLESS: нет активного сервера.\n")
		} else {
			sb.WriteString("⚠️ TCP до VLESS: " + health.VLESSRemote + "\n")
		}
	}

	if snap.DPIEnabled && health.DPIRunning {
		sb.WriteString("✅ DPI: процесс запущен.\n")
		if health.DpiConnectivityWarn {
			sb.WriteString("⚠️ DPI health: запрос generate_204 не прошёл — ТСПУ/маршрут могли измениться; запустите автоподбор DPI на вкладке DPI.\n")
			if d := strings.TrimSpace(health.DpiConnectivityDetail); d != "" {
				sb.WriteString("   Деталь: " + d + "\n")
			}
		}
	} else if snap.DPIEnabled {
		sb.WriteString("⚠️ DPI: включён в настройках, процесс не найден.\n")
	} else {
		sb.WriteString("ℹ️ DPI: выключен.\n")
	}

	sb.WriteString("\n=== Детали ===\n")
	sb.WriteString(fmt.Sprintf(
		"DataDir: %s\nConfig: %s\nLogs: %s\n"+
			"Ядро (sing-box) рядом: %v\nПапка DPI рядом: %v\n"+
			"Статус ядра: %s\nПрокси: %v · DPI: %v (процесс: %v)\n"+
			"Профиль маршрутизации: %s\n"+
			"Лог ядра: %s\nЛог DPI: %s\n",
		rep.DataDir,
		rep.ConfigPath,
		rep.LogsDir,
		rep.HasCore,
		rep.HasDpiFolder,
		ui.CoreStatus,
		ui.ProxyEnabled,
		ui.DPIEnabled,
		ui.DPIRunning,
		ui.RoutingProfile,
		coreLog,
		dpiLog,
	))
	return sb.String(), nil
}

func (a *App) CreateDiagnosticsBundle() (string, error) {
	res, err := diagnostics.CreateBundle()
	if err != nil {
		return "", err
	}
	return res.ZipPath, nil
}

func (a *App) OpenLogsFolder() error {
	rep, err := diagnostics.Run()
	if err != nil {
		return err
	}
	return exec.Command("explorer.exe", rep.LogsDir).Start()
}

func (a *App) resolveCoreLogPath() string {
	dataDir, err := config.EnsureDataDir()
	if err != nil {
		if a.coreManager != nil {
			return a.coreManager.LastLogPath()
		}
		return ""
	}
	sb := filepath.Join(dataDir, "runtime", "sing-box.log")
	if _, err := os.Stat(sb); err == nil {
		return sb
	}
	if a.coreManager != nil {
		if p := a.coreManager.LastLogPath(); p != "" {
			return p
		}
	}
	return sb
}

func (a *App) GetLogsState() LogsState {
	out := LogsState{}
	out.CoreLogPath = a.resolveCoreLogPath()
	if a.dpiRunner != nil {
		out.DPILogPath = a.dpiRunner.LastLogPath()
	}
	if out.CoreLogPath != "" {
		coreFull := tailFileLines(out.CoreLogPath, 400)
		out.CoreLogTail = tailLastLines(coreFull, 200)
		out.CoreErrors = filterErrorLines(coreFull, 80)
		out.CoreConnLines = filterConnDebugLines(coreFull, 120)
	}
	if out.DPILogPath != "" {
		dpiFull := tailFileLines(out.DPILogPath, 400)
		out.DPILogTail = tailLastLines(dpiFull, 200)
		out.DPIErrors = filterErrorLines(dpiFull, 80)
		out.DPIConnLines = filterConnDebugLines(dpiFull, 120)
	}
	return out
}

func tailLastLines(s string, maxLines int) string {
	lines := strings.Split(s, "\n")
	if len(lines) <= maxLines {
		return s
	}
	return strings.Join(lines[len(lines)-maxLines:], "\n")
}

func filterErrorLines(text string, maxLines int) string {
	var out []string
	for _, line := range strings.Split(text, "\n") {
		if logLineLooksImportant(line) {
			out = append(out, line)
		}
		if len(out) >= maxLines {
			break
		}
	}
	if len(out) == 0 {
		return ""
	}
	return strings.Join(out, "\n")
}

// logLineLooksImportant отбирает ошибки/варнинги из stderr sing-box, winws2 и сообщений Go.
func logLineLooksImportant(line string) bool {
	low := strings.ToLower(line)
	if strings.Contains(low, "error") || strings.Contains(low, "fatal") ||
		strings.Contains(low, "failed") || strings.Contains(low, "panic") ||
		strings.Contains(low, "ошибка") || strings.Contains(low, "критич") {
		return true
	}
	if strings.Contains(low, "warn") || strings.Contains(low, "warning") {
		return true
	}
	if strings.Contains(line, "\"level\":\"error\"") || strings.Contains(line, "\"level\":\"warn\"") ||
		strings.Contains(line, "\"level\": \"error\"") || strings.Contains(line, "\"level\": \"warn\"") {
		return true
	}
	if strings.Contains(low, "[error]") || strings.Contains(low, "[err]") ||
		strings.Contains(low, "[fatal]") || strings.Contains(low, "[ftl]") ||
		strings.Contains(low, "[warn]") || strings.Contains(low, "[wrn]") {
		return true
	}
	if strings.Contains(low, "eaddrinuse") || strings.Contains(low, "access is denied") ||
		strings.Contains(low, "отказано в доступе") || strings.Contains(low, "не найден") && strings.Contains(low, "exe") {
		return true
	}
	return false
}

func filterConnDebugLines(text string, maxLines int) string {
	var out []string
	for _, line := range strings.Split(text, "\n") {
		low := strings.ToLower(line)
		if strings.Contains(low, "dial") || strings.Contains(low, "connect") ||
			strings.Contains(low, "tls") || strings.Contains(low, "handshake") ||
			strings.Contains(low, "dns") || strings.Contains(low, "domain") ||
			strings.Contains(low, "resolved") || strings.Contains(low, "sniff") ||
			strings.Contains(low, "route") || strings.Contains(low, "vless") ||
			strings.Contains(low, "inbound") || strings.Contains(low, "outbound") ||
			strings.Contains(low, "socks") || strings.Contains(low, "router") ||
			(strings.Contains(low, "tcp") && strings.Contains(low, "127.0.0.1")) {
			out = append(out, line)
		}
		if len(out) >= maxLines {
			break
		}
	}
	if len(out) == 0 {
		return ""
	}
	return strings.Join(out, "\n")
}

func snapshotClashConnectionsRows(apiSecret string) []ActiveConnectionSnapshot {
	raw, err := core.FetchClashConnections(apiSecret)
	if err != nil || len(raw) == 0 {
		return nil
	}
	out := make([]ActiveConnectionSnapshot, 0, len(raw))
	for _, c := range raw {
		out = append(out, ActiveConnectionSnapshot{
			ID:       c.ID,
			Process:  c.Process,
			Remote:   c.Remote,
			Upload:   c.Upload,
			Download: c.Download,
		})
	}
	return out
}

// GetActiveConnections — живые соединения sing-box (Clash API), для монитора на дашборде.
func (a *App) GetActiveConnections() []ActiveConnectionSnapshot {
	a.mu.Lock()
	var apiSecret string
	proxyOn := false
	if a.st != nil {
		snap := a.st.Snapshot()
		proxyOn = snap.ProxyEnabled
		_, apiSecret = snap.LocalProxyAuth()
	}
	a.mu.Unlock()
	if !proxyOn || a.coreManager == nil {
		return nil
	}
	st := strings.TrimSpace(a.coreManager.Status())
	if st == "" || strings.HasPrefix(strings.ToLower(st), "error") {
		return nil
	}
	if !strings.HasPrefix(st, "running |") || strings.Contains(st, "no core") {
		return nil
	}
	return snapshotClashConnectionsRows(apiSecret)
}

func (a *App) GetRealtimeStats() RealtimeStats {
	a.mu.Lock()
	proxyOn := a.st != nil && a.st.Snapshot().ProxyEnabled

	var status string
	if a.coreManager != nil {
		status = strings.TrimSpace(a.coreManager.Status())
	}
	if a.dpiRunner != nil && a.dpiRunner.IsRunning() {
		if status == "" {
			status = "dpi running"
		} else {
			status += " | dpi running"
		}
	}

	coreUp := a.coreManager != nil && a.coreManager.CoreProcessRunning()
	stLow := strings.ToLower(status)
	statusRunning := strings.HasPrefix(strings.TrimSpace(status), "running |") &&
		!strings.Contains(stLow, "no core") && !strings.HasPrefix(stLow, "error")
	connected := proxyOn && a.coreManager != nil && (coreUp || statusRunning)

	if !proxyOn {
		a.rtLastAt = time.Time{}
		a.rtLastRecv = 0
		a.rtLastSent = 0
		a.rtSessionStart = time.Time{}
		a.rtSessionBaseRecv = 0
		a.rtSessionBaseSent = 0
		a.mu.Unlock()
		return RealtimeStats{
			Connected:             false,
			Status:                status,
			DownloadBps:           0,
			UploadBps:             0,
			ProxyProcessDetails:   nil,
			ActiveConnections:     nil,
			SessionDownloadBytes:  0,
			SessionUploadBytes:    0,
			SessionElapsedSeconds: 0,
		}
	}
	a.mu.Unlock()

	recv, sent, err := readNetstatTotals()
	now := time.Now()

	var downBps, upBps float64
	var sessDl, sessUl, sessSecs int64

	a.mu.Lock()
	if err == nil {
		if a.rtSessionStart.IsZero() {
			a.rtSessionStart = now
			a.rtSessionBaseRecv = recv
			a.rtSessionBaseSent = sent
		}
		if recv >= a.rtSessionBaseRecv {
			sessDl = int64(recv - a.rtSessionBaseRecv)
		}
		if sent >= a.rtSessionBaseSent {
			sessUl = int64(sent - a.rtSessionBaseSent)
		}
		sessSecs = int64(now.Sub(a.rtSessionStart).Seconds())
		if sessSecs < 0 {
			sessSecs = 0
		}
	}
	if err == nil && !a.rtLastAt.IsZero() && recv >= a.rtLastRecv && sent >= a.rtLastSent {
		dt := now.Sub(a.rtLastAt).Seconds()
		if dt > 0 {
			downBps = float64(recv-a.rtLastRecv) / dt
			upBps = float64(sent-a.rtLastSent) / dt
		}
	}
	if err == nil {
		a.rtLastRecv = recv
		a.rtLastSent = sent
		a.rtLastAt = now
	}
	a.mu.Unlock()

	return RealtimeStats{
		Connected:             connected,
		Status:                status,
		DownloadBps:           downBps,
		UploadBps:             upBps,
		ProxyProcessDetails:   nil,
		ActiveConnections:     nil,
		SessionDownloadBytes:  sessDl,
		SessionUploadBytes:    sessUl,
		SessionElapsedSeconds: sessSecs,
	}
}

func (a *App) updateConfig(mut func(cfg *models.AppConfig)) error {
	a.mu.Lock()
	defer a.mu.Unlock()
	if a.st == nil {
		return fmt.Errorf("app state is not initialized")
	}
	a.st.Update(mut)
	return config.Save(a.st.Snapshot())
}

func (a *App) emitRuntimeApplyJSON(job uint64, payload map[string]any) {
	if a.ctx == nil {
		return
	}
	if job != 0 {
		payload["job"] = job
	}
	b, err := json.Marshal(payload)
	if err != nil {
		return
	}
	runtime.EventsEmit(a.ctx, runtimeApplyEvent, string(b))
}

func (a *App) emitRuntimeWatchdogJSON(reason, detail string) {
	if a.ctx == nil {
		return
	}
	b, err := json.Marshal(map[string]string{"reason": reason, "detail": detail})
	if err != nil {
		return
	}
	runtime.EventsEmit(a.ctx, runtimeWatchdogEvent, string(b))
}

// applyHeavy применяет ядро и DPI без удержания основного замка приложения (a.mu).
func (a *App) applyHeavy(snap models.AppConfig, emit bool, job uint64) error {
	if emit {
		a.emitRuntimeApplyJSON(job, map[string]any{"phase": "core", "detail": "применение ядра", "done": false})
	}
	if a.coreManager != nil {
		if err := a.coreManager.Apply(snap); err != nil {
			return fmt.Errorf("ядро/прокси: %w", err)
		}
	}
	if emit {
		a.emitRuntimeApplyJSON(job, map[string]any{"phase": "core", "detail": "ядро готово", "done": false})
	}

	if a.dpiRunner != nil {
		if emit {
			a.emitRuntimeApplyJSON(job, map[string]any{"phase": "dpi", "detail": "", "done": false})
		}
		if snap.DPIEnabled {
			if prof, ok := findDPIProfile(snap, snap.ActiveDPIProfile); ok {
				if err := a.dpiRunner.Start(prof, &snap); err != nil {
					return fmt.Errorf("DPI: %w", err)
				}
			} else {
				return fmt.Errorf("профиль DPI «%s» не найден — выберите профиль на вкладке DPI или отключите DPI", snap.ActiveDPIProfile)
			}
		} else {
			_ = a.dpiRunner.Stop()
		}
		if emit {
			a.emitRuntimeApplyJSON(job, map[string]any{"phase": "dpi", "detail": "DPI готов", "done": false})
		}
	}
	if emit {
		a.emitRuntimeApplyJSON(job, map[string]any{"phase": "probe", "detail": "проверка порта", "done": false})
	}
	go func() {
		a.warmRoutingDpiPanelCaches()
	}()
	return nil
}

// runApplySync сериализует тяжёлое применение (applyMu). Снимок конфига берётся под коротким a.mu.
// При emit=true шлётся прогресс runtime-apply и после успеха планируется user-space watchdog.
// job ненулевой только для Deferred — попадает в JSON событий, чтобы UI отличил завершённые очереди apply.
func (a *App) runApplySync(emit bool, job uint64) error {
	a.applyMu.Lock()
	defer a.applyMu.Unlock()

	a.mu.Lock()
	if a.st == nil {
		a.mu.Unlock()
		err := fmt.Errorf("app state is not initialized")
		if emit {
			a.emitRuntimeApplyJSON(job, map[string]any{"phase": "", "detail": "", "done": true, "error": err.Error()})
		}
		return err
	}
	var secretChanged bool
	a.st.Update(func(c *models.AppConfig) {
		secretChanged = config.EnsureLocalInboundSecrets(c)
	})
	snap := a.st.Snapshot()
	a.mu.Unlock()
	if secretChanged {
		if err := config.Save(snap); err != nil {
			return fmt.Errorf("сохранение локальных секретов mixed/clash_api: %w", err)
		}
	}

	if !emit {
		if err := a.applyHeavy(snap, false, 0); err != nil {
			return err
		}
		a.scheduleWatchdog(snap)
		return nil
	}

	a.emitRuntimeApplyJSON(job, map[string]any{"phase": "", "detail": "старт", "done": false})
	err := a.applyHeavy(snap, true, job)
	if err != nil {
		a.emitRuntimeApplyJSON(job, map[string]any{"phase": "", "detail": "", "done": true, "error": err.Error()})
		return err
	}
	a.emitRuntimeApplyJSON(job, map[string]any{"phase": "", "detail": "готово", "done": true, "error": ""})
	a.scheduleWatchdog(snap)
	return nil
}

func (a *App) scheduleWatchdog(expected models.AppConfig) {
	gen := atomic.AddUint64(&a.watchdogGen, 1)
	go func(g uint64, snap models.AppConfig) {
		time.Sleep(8 * time.Second)
		if atomic.LoadUint64(&a.watchdogGen) != g {
			return
		}
		a.runWatchdogOnce(snap)
	}(gen, expected)
}

// runWatchdogOnce проверяет локальный mixed-порт и живость процесса DPI (user-space).
func (a *App) runWatchdogOnce(expected models.AppConfig) {
	a.mu.Lock()
	if a.st == nil {
		a.mu.Unlock()
		return
	}
	cur := a.st.Snapshot()
	a.mu.Unlock()

	if !cur.ProxyEnabled && !cur.DPIEnabled {
		return
	}
	if cur.ProxyEnabled != expected.ProxyEnabled || cur.DPIEnabled != expected.DPIEnabled {
		return
	}
	if cur.ProxyPort != expected.ProxyPort || cur.ActiveDPIProfile != expected.ActiveDPIProfile {
		return
	}

	var failReason, failDetail string
	if cur.ProxyEnabled {
		addr := net.JoinHostPort("127.0.0.1", strconv.Itoa(cur.ListenPort()))
		conn, err := net.DialTimeout("tcp", addr, 2*time.Second)
		if err != nil {
			failReason = "proxy_listen"
			failDetail = "Локальный прокси не отвечает на " + addr + ": " + trimErr(err)
		} else {
			_ = conn.Close()
		}
	}
	if failReason == "" && cur.DPIEnabled && a.dpiRunner != nil {
		if !a.dpiRunner.ProcessAlive() {
			failReason = "dpi_process"
			failDetail = "Процесс DPI не найден — при включённом DPI это указывает на сбой запуска."
		}
	}
	if failReason == "" {
		return
	}

	a.emitRuntimeWatchdogJSON(failReason, failDetail)
	_ = a.StopAll()
}

func findDPIProfile(cfg models.AppConfig, name string) (models.DPIProfile, bool) {
	for _, p := range cfg.DPIProfiles {
		if p.Name == name {
			return p, true
		}
	}
	return models.DPIProfile{}, false
}

func containsServerID(servers []models.VlessServer, id string) bool {
	for _, srv := range servers {
		if srv.ID == id {
			return true
		}
	}
	return false
}

func dedupeVlessServers(servers []models.VlessServer) []models.VlessServer {
	if len(servers) <= 1 {
		return servers
	}
	seen := make(map[string]struct{}, len(servers))
	out := make([]models.VlessServer, 0, len(servers))
	for _, srv := range servers {
		if srv.ID == "" {
			continue
		}
		if _, ok := seen[srv.ID]; ok {
			continue
		}
		seen[srv.ID] = struct{}{}
		out = append(out, srv)
	}
	return out
}

func isProcessElevated() bool {
	var token windows.Token
	if err := windows.OpenProcessToken(windows.CurrentProcess(), windows.TOKEN_QUERY, &token); err != nil {
		return false
	}
	defer token.Close()
	return token.IsElevated()
}

func findZapretBaseDir() string {
	if dir, ok := config.ResolveZapretDirNearExe(); ok {
		return dir
	}
	return ""
}

func tailFileLines(path string, maxLines int) string {
	f, err := os.Open(path)
	if err != nil {
		return ""
	}
	defer f.Close()
	var lines []string
	sc := bufio.NewScanner(f)
	for sc.Scan() {
		lines = append(lines, sc.Text())
		if len(lines) > maxLines {
			lines = lines[1:]
		}
	}
	return strings.Join(lines, "\n")
}

func readNetstatTotals() (recv uint64, sent uint64, err error) {
	if r, s, e := readNetAdapterTotalsPowerShell(); e == nil {
		return r, s, nil
	}
	if r, s, e := readNetstatTotalsEnglish(); e == nil {
		return r, s, nil
	}
	return readNetstatTotalsHeuristic()
}

// Sum bytes across active adapters (works with Russian locale; netstat -e often does not).
func readNetAdapterTotalsPowerShell() (recv uint64, sent uint64, err error) {
	ps := `$ErrorActionPreference='SilentlyContinue'; ` +
		`$s = Get-NetAdapter | Where-Object { $_.Status -eq 'Up' } | Get-NetAdapterStatistics; ` +
		`if ($null -eq $s) { exit 2 }; ` +
		`$r = ($s | Measure-Object ReceivedBytes -Sum).Sum; ` +
		`$t = ($s | Measure-Object SentBytes -Sum).Sum; ` +
		`Write-Output "$r $t"`
	out, err := runHiddenOutput("powershell.exe", "-NoProfile", "-NonInteractive", "-Command", ps)
	if err != nil {
		return 0, 0, err
	}
	fields := strings.Fields(strings.TrimSpace(string(out)))
	if len(fields) < 2 {
		return 0, 0, fmt.Errorf("powershell net adapter parse")
	}
	recv, err1 := strconv.ParseUint(strings.ReplaceAll(fields[0], ",", ""), 10, 64)
	sent, err2 := strconv.ParseUint(strings.ReplaceAll(fields[1], ",", ""), 10, 64)
	if err1 != nil || err2 != nil {
		return 0, 0, fmt.Errorf("powershell number parse")
	}
	return recv, sent, nil
}

func readNetstatTotalsEnglish() (recv uint64, sent uint64, err error) {
	out, err := runHiddenOutput("cmd.exe", "/c", "chcp 437>nul && netstat -e")
	if err != nil {
		return 0, 0, err
	}
	return parseNetstatBytesLine(out)
}

func parseNetstatBytesLine(out []byte) (recv uint64, sent uint64, err error) {
	sc := bufio.NewScanner(bytes.NewReader(out))
	for sc.Scan() {
		line := sc.Text()
		lower := strings.ToLower(line)
		if !strings.Contains(lower, "bytes") && !strings.Contains(lower, "байт") {
			continue
		}
		nums := extractUintTokens(line)
		if len(nums) >= 2 {
			return nums[0], nums[1], nil
		}
	}
	return 0, 0, fmt.Errorf("netstat bytes line not found")
}

// Last resort: pick the data row with two largest byte-like counters (locale-independent).
func readNetstatTotalsHeuristic() (recv uint64, sent uint64, err error) {
	out, err := runHiddenOutput("netstat.exe", "-e")
	if err != nil {
		return 0, 0, err
	}
	var bestR, bestS uint64
	var bestScore uint64
	sc := bufio.NewScanner(bytes.NewReader(out))
	for sc.Scan() {
		line := sc.Text()
		nums := extractUintTokens(line)
		if len(nums) < 2 {
			continue
		}
		r, s := nums[0], nums[1]
		if r < 1024 || s < 1024 {
			continue
		}
		score := r + s
		if score > bestScore {
			bestScore = score
			bestR, bestS = r, s
		}
	}
	if bestScore == 0 {
		return 0, 0, fmt.Errorf("netstat heuristic failed")
	}
	return bestR, bestS, nil
}

func extractUintTokens(s string) []uint64 {
	var out []uint64
	var b strings.Builder
	flush := func() {
		if b.Len() == 0 {
			return
		}
		v, err := strconv.ParseUint(b.String(), 10, 64)
		if err == nil {
			out = append(out, v)
		}
		b.Reset()
	}
	for _, r := range s {
		if r >= '0' && r <= '9' {
			b.WriteRune(r)
		} else {
			flush()
		}
	}
	flush()
	return out
}

func listProxyProcessDetails(proxyPort int) []ProxyProcessDetail {
	if proxyPort < 1 || proxyPort > 65535 {
		proxyPort = 8080
	}
	portSuffix := ":" + strconv.Itoa(proxyPort)
	out, err := runHiddenOutput("cmd.exe", "/c", "chcp 437>nul && netstat.exe -ano -p tcp")
	if err != nil {
		out, err = runHiddenOutput("netstat.exe", "-ano", "-p", "tcp")
		if err != nil {
			return nil
		}
	}
	counts := map[string]int{}
	sc := bufio.NewScanner(bytes.NewReader(out))
	for sc.Scan() {
		parts := strings.Fields(sc.Text())
		if len(parts) < 5 || !strings.EqualFold(parts[0], "TCP") {
			continue
		}
		local := parts[1]
		foreign := parts[2]
		state := parts[3]
		pid := parts[len(parts)-1]
		if !tcpStateIsEstablished(state) {
			continue
		}
		if !(isProxyEndpoint(local, portSuffix) || isProxyEndpoint(foreign, portSuffix)) {
			continue
		}
		counts[pid]++
	}

	var details []ProxyProcessDetail
	for pid, n := range counts {
		name := processNameByPID(pid)
		if name == "" {
			continue
		}
		details = append(details, ProxyProcessDetail{
			Name:      name,
			PID:       pid,
			ConnCount: n,
		})
	}
	sort.Slice(details, func(i, j int) bool {
		if details[i].Name != details[j].Name {
			return details[i].Name < details[j].Name
		}
		return details[i].PID < details[j].PID
	})
	return details
}

func isProxyEndpoint(addr string, portSuffix string) bool {
	addr = strings.ToLower(strings.TrimSpace(addr))
	if !strings.HasSuffix(addr, strings.ToLower(portSuffix)) {
		return false
	}
	return strings.Contains(addr, "127.0.0.1") || strings.Contains(addr, "[::1]") ||
		strings.HasPrefix(addr, "localhost:") || strings.HasPrefix(addr, "[::ffff:127.0.0.1]:")
}

func tcpStateIsEstablished(state string) bool {
	u := strings.ToUpper(strings.TrimSpace(state))
	if strings.Contains(u, "ESTABLISH") {
		return true
	}
	// Russian locale netstat
	low := strings.ToLower(state)
	return strings.Contains(low, "установлен")
}

func processNameByPID(pid string) string {
	out, err := runHiddenOutput("tasklist.exe", "/FI", "PID eq "+pid, "/FO", "CSV", "/NH")
	if err != nil {
		return ""
	}
	line := strings.TrimSpace(string(out))
	if line == "" || strings.Contains(strings.ToLower(line), "no tasks") {
		return ""
	}
	line = strings.Trim(line, "\"")
	cols := strings.Split(line, "\",\"")
	if len(cols) == 0 {
		return ""
	}
	return cols[0]
}

func runHiddenOutput(name string, args ...string) ([]byte, error) {
	cmd := exec.Command(name, args...)
	cmd.SysProcAttr = &syscall.SysProcAttr{
		HideWindow:    true,
		CreationFlags: 0x08000000, // CREATE_NO_WINDOW
	}
	return cmd.Output()
}
