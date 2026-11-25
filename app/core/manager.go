package core

import (
	"bytes"
	"fmt"
	"log"
	"os"
	"path/filepath"
	"strings"
	"sync"

	"pulseshield/app/config"
	"pulseshield/app/models"
	"pulseshield/app/winproxy"
)

type Manager struct {
	mu            sync.Mutex
	adapter       *Adapter
	status        string
	lastLog       string
	lastLogFile   *os.File // ссылка на текущий лог-файл; закрывается при следующем рестарте
	proxySnapshot *winproxy.Snapshot
	lastConfigRaw []byte
	lastProxyMode models.ProxyMode
	lastProxyAddr string
}

func NewManager(adapter *Adapter) *Manager {
	initial := "stopped"
	if adapter == nil || adapter.binary == "" {
		initial = "core not bundled"
	}
	return &Manager{adapter: adapter, status: initial}
}

func (m *Manager) Apply(cfg models.AppConfig) error {
	m.mu.Lock()
	defer m.mu.Unlock()

	forceResetProxy := func() { _ = winproxy.ForceResetWindowsProxy() }

	if m.adapter == nil {
		if !cfg.ProxyEnabled {
			if cfg.DPIEnabled {
				m.status = "proxy off | dpi on (no bundled core)"
			} else {
				m.status = "stopped"
			}
			return nil
		}
		m.status = "core not bundled"
		return fmt.Errorf("ядро не инициализировано (внутренняя ошибка)")
	}

	if !cfg.ProxyEnabled && !cfg.DPIEnabled {
		m.proxySnapshot = nil
		if err := m.adapter.Stop(forceResetProxy); err != nil {
			return err
		}
		KillStraySingBoxOnListenPort(cfg.ListenPort())
		KillStraySingBoxOnListenPort(SingBoxClashAPIPort)
		m.status = "stopped"
		return nil
	}

	// Proxy requires core, but DPI-only mode must work independently.
	if cfg.ProxyEnabled && m.adapter.binary == "" {
		m.status = "core not bundled"
		return fmt.Errorf("рядом с программой нет sing-box.exe (папка core) — переустановите portable или сборку")
	}

	if cfg.ProxyEnabled {
		if cfg.ProxyMode == models.ProxyModeTun && !IsElevated() {
			m.status = "error: TUN требует запуск PulseShield от имени администратора"
			return fmt.Errorf("режим TUN (виртуальный адаптер) требует прав администратора — перезапустите приложение через «Запуск от имени администратора»")
		}
		dataDir, err := config.EnsureDataDir()
		if err != nil {
			m.status = "error: " + err.Error()
			return err
		}
		runDir := dataDir + "\\runtime"
		confRaw, err := RenderSingBoxConfig(cfg, runDir)
		if err != nil {
			m.status = "error: " + err.Error()
			return err
		}
		confChanged := !bytes.Equal(m.lastConfigRaw, confRaw)
		confPath, err := WriteSingBoxConfig(cfg)
		if err != nil {
			m.status = "error: " + err.Error()
			return err
		}
		if confChanged || !m.adapter.IsRunning() {
			if confChanged && m.adapter.IsRunning() {
				_, clashSec := cfg.LocalProxyAuth()
				_ = CloseAllClashConnections(clashSec)
			}
			logFile, logPath, err := OpenLogFile("core")
			if err == nil {
				// Закрываем предыдущий лог-файл: без этого каждый рестарт sing-box
				// оставлял открытый *os.File навсегда (утечка FD).
				if m.lastLogFile != nil {
					_ = m.lastLogFile.Close()
				}
				m.lastLogFile = logFile
				m.adapter.SetOutput(logFile, logFile)
				m.lastLog = logPath
			}
			if err := m.adapter.Stop(nil); err != nil {
				return err
			}
			KillStraySingBoxOnListenPort(cfg.ListenPort())
			KillStraySingBoxOnListenPort(SingBoxClashAPIPort)
			rotateOversizedSingBoxLog(filepath.Join(runDir, "sing-box.log"), 5*1024*1024)
			if err := m.adapter.Start("run", "-c", confPath); err != nil {
				m.status = "error: " + err.Error()
				return err
			}
			m.lastConfigRaw = append(m.lastConfigRaw[:0], confRaw...)
			zdir, _ := config.ResolveZapretDirNearExe()
			listsDir := ""
			if zdir != "" {
				listsDir = zdir + `\lists`
			}
			ruleN := len(config.MergeSingBoxRoutingRules(cfg, listsDir))
			_, clashSec := cfg.LocalProxyAuth()
			log.Printf("[PulseShield] sing-box запущен: routingProfile=%s proxyMode=%s strictRouting=%v mergedRules=%d final=%s listen=:%d experimental.clash_api=http://127.0.0.1:%d secret_len=%d",
				cfg.RoutingProfile, cfg.ProxyMode, cfg.StrictRoutingMode, ruleN, routeFinalOutbound(cfg), cfg.ListenPort(), SingBoxClashAPIPort, len(strings.TrimSpace(clashSec)))
			if err := PutClashProxyGroupSelection(clashSec, SingBoxClashGlobalProxyGroup, SingBoxMainOutboundTag); err != nil {
				log.Printf("[PulseShield] clash_api: PUT /proxies/%s → %q: %v", SingBoxClashGlobalProxyGroup, SingBoxMainOutboundTag, err)
			}
			if !IsElevated() {
				log.Printf("[PulseShield] Clash API /connections: чтобы sing-box заполнял поле process (имя процесса), на Windows обычно нужен запуск PulseShield от имени администратора")
			}
		}

		// System proxy handling.
		if cfg.ProxyMode == models.ProxyModeSystem {
			if m.proxySnapshot == nil {
				if snap, err := winproxy.Read(); err == nil {
					m.proxySnapshot = &snap
				}
			}
			hostport := fmt.Sprintf("127.0.0.1:%d", cfg.ListenPort())
			// No proxy auth: the local mixed inbound is unauthenticated (127.0.0.1 only),
			// and Windows can't pass credentials to browsers via the system proxy anyway.
			proxySig := hostport
			if m.lastProxyMode != cfg.ProxyMode || m.lastProxyAddr != proxySig {
				if err := winproxy.ApplyHTTPProxy(hostport, "", ""); err != nil {
					m.status = "error: " + err.Error()
					return err
				}
				m.lastProxyAddr = proxySig
			}
		} else {
			if m.lastProxyMode == models.ProxyModeSystem {
				_ = winproxy.Disable()
			}
			m.lastProxyAddr = ""
		}
		m.lastProxyMode = cfg.ProxyMode
	} else {
		// Прокси выключен: останавливаем sing-box и снимаем «осиротевшие» процессы на mixed-порту.
		if m.lastProxyMode == models.ProxyModeSystem {
			m.proxySnapshot = nil
		}
		_ = m.adapter.Stop(forceResetProxy)
		KillStraySingBoxOnListenPort(cfg.ListenPort())
		KillStraySingBoxOnListenPort(SingBoxClashAPIPort)
		m.lastConfigRaw = nil
		m.lastProxyAddr = ""
	}

	if !m.adapter.IsRunning() {
		if cfg.DPIEnabled {
			m.status = "proxy off | dpi on (sing-box stopped)"
		} else {
			m.status = "stopped"
		}
		return nil
	}
	m.status = fmt.Sprintf("running | proxy=%v mode=%s dpi=%v route=%s", cfg.ProxyEnabled, cfg.ProxyMode, cfg.DPIEnabled, cfg.RoutingProfile)
	return nil
}

func (m *Manager) Status() string {
	m.mu.Lock()
	defer m.mu.Unlock()
	return m.status
}

// CoreProcessRunning — процесс sing-box ещё жив (до того как статус строки обновится на «running»).
func (m *Manager) CoreProcessRunning() bool {
	m.mu.Lock()
	defer m.mu.Unlock()
	return m.adapter != nil && m.adapter.IsRunning()
}

func (m *Manager) LastLogPath() string {
	m.mu.Lock()
	defer m.mu.Unlock()
	return m.lastLog
}

// ForceStopForShutdown останавливает прокси и ядро при выходе из приложения (без изменения конфига).
func (m *Manager) ForceStopForShutdown(listenPort int) {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.proxySnapshot = nil
	if m.adapter != nil {
		_ = m.adapter.Stop(func() { _ = winproxy.ForceResetWindowsProxy() })
	}
	// Закрываем лог-файл sing-box при shutdown.
	if m.lastLogFile != nil {
		_ = m.lastLogFile.Close()
		m.lastLogFile = nil
	}
	KillStraySingBoxOnListenPort(listenPort)
	KillStraySingBoxOnListenPort(SingBoxClashAPIPort)
	m.status = "stopped"
}

func rotateOversizedSingBoxLog(logPath string, maxBytes int64) {
	if maxBytes <= 0 || strings.TrimSpace(logPath) == "" {
		return
	}
	info, err := os.Stat(logPath)
	if err != nil || info.Size() <= maxBytes {
		return
	}
	backup := logPath + ".prev"
	_ = os.Remove(backup)
	if err := os.Rename(logPath, backup); err != nil {
		_ = os.Truncate(logPath, 0)
	}
}
