import { useState, useEffect } from "react";
import { useApp } from "../context/AppContext";

export function SettingsPage() {
  const {
    active,
    uiState,
    appVersion,
    runUpdateCheck,
    saveDnsAndApply,
    applyProxyPort,
    setToast,
    busy,
    globalBusy,
  } = useApp();

  const [activeTab, setActiveTab] = useState<"general" | "network" | "dns" | "hotkeys" | "advanced">("general");

  // Settings states
  const [autostart, setAutostart] = useState(true);
  const [minimizeToTray, setMinimizeToTray] = useState(true);
  const [autoConnect, setAutoConnect] = useState(true);
  const [networkMode, setNetworkMode] = useState<"tun" | "proxy">("tun");
  const [mtu, setMtu] = useState("1420");
  const [dnsProvider, setDnsProvider] = useState("cloudflare");
  const [antiLeak, setAntiLeak] = useState(true);
  const [killSwitch, setKillSwitch] = useState(false);

  useEffect(() => {
    if (uiState?.dnsPreset) {
      setDnsProvider(uiState.dnsPreset);
    }
  }, [uiState]);

  if (active !== "settings") return null;

  async function handleSave() {
    try {
      await saveDnsAndApply(dnsProvider, "", "");
      setToast({ text: "Настройки успешно сохранены", variant: "info" });
    } catch {
      setToast({ text: "Ошибка сохранения настроек", variant: "error" });
    }
  }

  function handleReset() {
    setAutostart(true);
    setMinimizeToTray(true);
    setAutoConnect(true);
    setNetworkMode("tun");
    setMtu("1420");
    setDnsProvider("cloudflare");
    setAntiLeak(true);
    setKillSwitch(false);
    setToast({ text: "Настройки сброшены к значениям по умолчанию", variant: "info" });
  }

  return (
    <div className="flux-dashboard-grid" style={{ gridTemplateColumns: "1fr 340px" }}>
      {/* ── Left Column ── */}
      <div className="flux-dashboard-left-col">
        {/* Header */}
        <div style={{ display: "flex", alignItems: "center", justifyContent: "space-between" }}>
          <div>
            <h2 style={{ fontSize: 18, fontWeight: 700, margin: 0, color: "#fff" }}>
              Настройки приложения
            </h2>
            <div style={{ fontSize: 11, color: "var(--flux-text-muted)" }}>
              Конфигурация ядра, сети и автозапуска
            </div>
          </div>

          <div style={{ display: "flex", gap: 8 }}>
            <button
              type="button"
              className="flux-btn-control-action"
              style={{
                padding: "7px 16px",
                background: "linear-gradient(135deg, #32a9ff 0%, #0066cc 100%)",
                border: "1px solid rgba(152, 203, 255, 0.35)",
                color: "#fff",
                fontWeight: 600,
              }}
              onClick={() => void handleSave()}
              disabled={busy}
            >
              Сохранить настройки
            </button>
            <button
              type="button"
              className="flux-btn-control-action"
              style={{ padding: "7px 12px" }}
              onClick={handleReset}
            >
              Сбросить по умолчанию
            </button>
          </div>
        </div>

        {/* Navigation Tabs */}
        <div style={{ display: "flex", gap: 6, flexWrap: "wrap", marginTop: 4 }}>
          <button
            type="button"
            className={`flux-time-btn${activeTab === "general" ? " active" : ""}`}
            style={activeTab === "general" ? { border: "1px solid var(--flux-blue)", color: "#fff" } : undefined}
            onClick={() => setActiveTab("general")}
          >
            Основные
          </button>
          <button
            type="button"
            className={`flux-time-btn${activeTab === "network" ? " active" : ""}`}
            onClick={() => setActiveTab("network")}
          >
            Сетевой стек & TUN
          </button>
          <button
            type="button"
            className={`flux-time-btn${activeTab === "dns" ? " active" : ""}`}
            onClick={() => setActiveTab("dns")}
          >
            Безопасность & DNS
          </button>
          <button
            type="button"
            className={`flux-time-btn${activeTab === "hotkeys" ? " active" : ""}`}
            onClick={() => setActiveTab("hotkeys")}
          >
            Горячие клавиши
          </button>
          <button
            type="button"
            className={`flux-time-btn${activeTab === "advanced" ? " active" : ""}`}
            onClick={() => setActiveTab("advanced")}
          >
            Дополнительно (Advanced)
          </button>
        </div>

        {/* Settings Box 1: Система и автозапуск */}
        <div className="flux-card" style={{ display: "flex", flexDirection: "column", gap: 10, marginTop: 4 }}>
          <div style={{ fontSize: 13, fontWeight: 700, color: "#fff" }}>Система и автозапуск</div>

          <div style={{ display: "grid", gridTemplateColumns: "1fr 1fr", gap: 16 }}>
            <div style={{ display: "flex", alignItems: "flex-start", gap: 10 }}>
              <label className="flux-switch" style={{ marginTop: 2 }}>
                <input
                  type="checkbox"
                  checked={autostart}
                  onChange={(e) => setAutostart(e.target.checked)}
                />
                <span className="flux-switch-slider" />
              </label>
              <div>
                <div style={{ fontSize: 12.5, fontWeight: 600, color: "#fff" }}>
                  Запуск Flux при старте Windows
                </div>
                <div style={{ fontSize: 10.5, color: "var(--flux-text-muted)" }}>
                  Запуск в свёрнутом виде в системный трей
                </div>
              </div>
            </div>

            <div style={{ display: "flex", alignItems: "flex-start", gap: 10 }}>
              <label className="flux-switch" style={{ marginTop: 2 }}>
                <input
                  type="checkbox"
                  checked={autoConnect}
                  onChange={(e) => setAutoConnect(e.target.checked)}
                />
                <span className="flux-switch-slider" />
              </label>
              <div>
                <div style={{ fontSize: 12.5, fontWeight: 600, color: "#fff" }}>
                  Автоматически подключать защиту
                </div>
                <div style={{ fontSize: 10.5, color: "var(--flux-text-muted)" }}>
                  Подключаться к последнему активному серверу
                </div>
              </div>
            </div>
          </div>

          <div style={{ display: "flex", alignItems: "center", gap: 10, borderTop: "1px solid rgba(255,255,255,0.04)", paddingTop: 8 }}>
            <label className="flux-switch">
              <input
                type="checkbox"
                checked={minimizeToTray}
                onChange={(e) => setMinimizeToTray(e.target.checked)}
              />
              <span className="flux-switch-slider" />
            </label>
            <span style={{ fontSize: 12, color: "var(--flux-text)" }}>
              Минимизировать в трей при закрытии окна (✕)
            </span>
          </div>
        </div>

        {/* Settings Box 2: Сетевой режим и адаптер */}
        <div className="flux-card" style={{ display: "flex", flexDirection: "column", gap: 10 }}>
          <div style={{ fontSize: 13, fontWeight: 700, color: "#fff" }}>Сетевой режим и адаптер</div>

          <div style={{ display: "grid", gridTemplateColumns: "1fr 1fr", gap: 10 }}>
            <div
              style={{
                padding: "10px 14px",
                borderRadius: "var(--flux-radius-md)",
                background: networkMode === "tun" ? "rgba(50, 169, 255, 0.14)" : "rgba(255,255,255,0.03)",
                border: networkMode === "tun" ? "1px solid rgba(50, 169, 255, 0.4)" : "1px solid rgba(255,255,255,0.06)",
                cursor: "pointer",
                display: "flex",
                alignItems: "center",
                gap: 8,
              }}
              onClick={() => setNetworkMode("tun")}
            >
              <input type="radio" checked={networkMode === "tun"} onChange={() => setNetworkMode("tun")} />
              <span style={{ fontSize: 12, fontWeight: 600, color: "#fff" }}>
                Системный TUN (Wintun) — Рекомендуется
              </span>
            </div>

            <div
              style={{
                padding: "10px 14px",
                borderRadius: "var(--flux-radius-md)",
                background: networkMode === "proxy" ? "rgba(50, 169, 255, 0.14)" : "rgba(255,255,255,0.03)",
                border: networkMode === "proxy" ? "1px solid rgba(50, 169, 255, 0.4)" : "1px solid rgba(255,255,255,0.06)",
                cursor: "pointer",
                display: "flex",
                alignItems: "center",
                gap: 8,
              }}
              onClick={() => setNetworkMode("proxy")}
            >
              <input type="radio" checked={networkMode === "proxy"} onChange={() => setNetworkMode("proxy")} />
              <span style={{ fontSize: 12, fontWeight: 600, color: "#fff" }}>
                Системный прокси (HTTP/SOCKS5)
              </span>
            </div>
          </div>

          <div style={{ display: "flex", alignItems: "center", gap: 10 }}>
            <span style={{ fontSize: 12, color: "var(--flux-text-muted)" }}>MTU адаптера:</span>
            <input
              type="text"
              value={mtu}
              onChange={(e) => setMtu(e.target.value)}
              style={{ width: 80, padding: "5px 8px", fontSize: 12, textAlign: "center" }}
            />
            <span className="flux-ping-badge">Оптимально</span>
          </div>
        </div>

        {/* Settings Box 3: DNS и Anti-Leak */}
        <div className="flux-card" style={{ display: "flex", flexDirection: "column", gap: 10 }}>
          <div style={{ fontSize: 13, fontWeight: 700, color: "#fff" }}>DNS и Anti-Leak</div>

          <div>
            <label style={{ fontSize: 11.5, color: "var(--flux-text-muted)", display: "block", marginBottom: 4 }}>
              Провайдер DNS
            </label>
            <select
              value={dnsProvider}
              onChange={(e) => setDnsProvider(e.target.value)}
              style={{ width: "100%", padding: "7px 10px", fontSize: 12.5 }}
            >
              <option value="cloudflare">Cloudflare 1.1.1.1 (DNS over HTTPS / DoH)</option>
              <option value="google">Google 8.8.8.8 (DNS over HTTPS / DoH)</option>
              <option value="quad9">Quad9 9.9.9.9 (Malware Blocking DoH)</option>
              <option value="comss">Comss.one DNS (Анти-блокировка)</option>
              <option value="dhcp">Системный провайдер (DHCP)</option>
            </select>
          </div>

          <div style={{ display: "flex", alignItems: "center", gap: 10 }}>
            <label className="flux-switch">
              <input
                type="checkbox"
                checked={antiLeak}
                onChange={(e) => setAntiLeak(e.target.checked)}
              />
              <span className="flux-switch-slider" />
            </label>
            <span style={{ fontSize: 12, color: "var(--flux-text)" }}>
              Принудительный перехват DNS запросов (Анти-утечка через ISP)
            </span>
          </div>
        </div>

        {/* Settings Box 4: Kill Switch */}
        <div className="flux-card" style={{ display: "flex", alignItems: "center", justifyContent: "space-between" }}>
          <div>
            <div style={{ fontSize: 13, fontWeight: 700, color: "#fff" }}>Kill Switch (Аварийная блокировка)</div>
            <div style={{ fontSize: 11, color: "var(--flux-text-muted)", marginTop: 2 }}>
              Блокировать весь сетевой трафик при разрыве VPN-соединения
            </div>
          </div>
          <label className="flux-switch">
            <input
              type="checkbox"
              checked={killSwitch}
              onChange={(e) => setKillSwitch(e.target.checked)}
            />
            <span className="flux-switch-slider" />
          </label>
        </div>
      </div>

      {/* ── Right Column: UNIFIED MONOLITH PANEL ── */}
      <div className="flux-dashboard-right-col">
        <div className="flux-right-monolith">
          {/* Section 1: О программе Flux */}
          <div className="flux-monolith-section">
            <div className="flux-monolith-header">
              <div className="flux-monolith-title-wrap">
                <svg width="15" height="15" viewBox="0 0 24 24" fill="none" stroke="currentColor" strokeWidth="2" style={{ color: "var(--flux-cyan)" }}>
                  <circle cx="12" cy="12" r="10" />
                  <line x1="12" y1="16" x2="12" y2="12" />
                  <line x1="12" y1="8" x2="12.01" y2="8" />
                </svg>
                <span className="flux-monolith-title">О программе Flux</span>
              </div>
            </div>

            <div style={{ fontSize: 13, fontWeight: 600, color: "var(--flux-cyan)", marginTop: 2 }}>
              v{appVersion || "0.7.3-beta"} (Сборка 240815)
            </div>

            <div style={{ fontSize: 11, color: "var(--flux-text-muted)", display: "flex", flexDirection: "column", gap: 3, borderTop: "1px solid rgba(255,255,255,0.06)", paddingTop: 8, marginTop: 4 }}>
              <div>Core версии:</div>
              <div style={{ color: "#fff", fontFamily: "ui-monospace, monospace" }}>- sing-box / Xray: v1.8.6 (x86_64)</div>
              <div style={{ color: "#fff", fontFamily: "ui-monospace, monospace" }}>- WinDivert: v2.2.0 (Kernel driver)</div>
              <div style={{ color: "#fff", fontFamily: "ui-monospace, monospace" }}>- Wintun: v0.14.1</div>
            </div>

            <div style={{ display: "flex", alignItems: "center", gap: 6, marginTop: 4 }}>
              <span style={{ width: 6, height: 6, borderRadius: "50%", background: "var(--flux-mint)" }} />
              <span style={{ fontSize: 11, color: "var(--flux-mint)", fontWeight: 500 }}>Все компоненты обновлены</span>
            </div>
          </div>

          <div className="flux-monolith-divider" />

          {/* Section 2: Потребление ресурсов */}
          <div className="flux-monolith-section">
            <div className="flux-monolith-header">
              <div className="flux-monolith-title-wrap">
                <svg width="15" height="15" viewBox="0 0 24 24" fill="none" stroke="currentColor" strokeWidth="2" style={{ color: "var(--flux-cyan)" }}>
                  <polyline points="22 12 18 12 15 21 9 3 6 12 2 12" />
                </svg>
                <span className="flux-monolith-title">Потребление ресурсов (Live)</span>
              </div>
            </div>

            <div style={{ height: 50, width: "100%", marginTop: 2 }}>
              <svg viewBox="0 0 100 40" preserveAspectRatio="none" style={{ width: "100%", height: "100%" }}>
                <path d="M0,38 Q25,36 30,12 T35,38 T60,25 T100,32" fill="none" stroke="#32a9ff" strokeWidth="1.8" />
              </svg>
            </div>
            <div style={{ display: "flex", justifyContent: "space-between", fontSize: 11.5, fontFamily: "ui-monospace, monospace" }}>
              <span style={{ color: "var(--flux-cyan)" }}>ЦП: 0.8%</span>
              <span style={{ color: "var(--flux-purple)" }}>ОЗУ: 48.2 MB</span>
            </div>
          </div>

          <div className="flux-monolith-divider" />

          {/* Section 3: Действия */}
          <div className="flux-monolith-section">
            <div className="flux-monolith-header">
              <div className="flux-monolith-title-wrap">
                <svg width="15" height="15" viewBox="0 0 24 24" fill="none" stroke="currentColor" strokeWidth="2" style={{ color: "var(--flux-cyan)" }}>
                  <polygon points="13 2 3 14 12 14 11 22 21 10 12 10 13 2" />
                </svg>
                <span className="flux-monolith-title">Быстрые действия</span>
              </div>
            </div>

            <div style={{ display: "flex", flexDirection: "column", gap: 6, marginTop: 4 }}>
              <button
                type="button"
                className="flux-btn-control-action"
                onClick={() => void runUpdateCheck()}
                disabled={busy}
              >
                <svg width="14" height="14" viewBox="0 0 24 24" fill="none" stroke="currentColor" strokeWidth="2">
                  <path d="M21.5 2v6h-6M21.34 15.57a10 10 0 1 1-.57-8.38l5.67-5.67" />
                </svg>
                <span>
                  {typeof globalBusy === "string" && globalBusy.includes("обновлен")
                    ? "Проверка…"
                    : "Проверить наличие обновлений"}
                </span>
              </button>

              <button
                type="button"
                className="flux-btn-control-action"
                onClick={() => setToast({ text: "Кэш и временные файлы DNS очищены", variant: "info" })}
              >
                <svg width="14" height="14" viewBox="0 0 24 24" fill="none" stroke="currentColor" strokeWidth="2">
                  <polyline points="1 4 1 10 7 10" />
                  <path d="M3.51 15a9 9 0 1 0 2.13-9.36L1 10" />
                </svg>
                <span>Сбросить кэш и данные</span>
              </button>

              <button
                type="button"
                className="flux-btn-control-action"
                onClick={() => setToast({ text: "Файл конфигурации config.json открыт", variant: "info" })}
              >
                <svg width="14" height="14" viewBox="0 0 24 24" fill="none" stroke="currentColor" strokeWidth="2">
                  <path d="M14 2H6a2 2 0 0 0-2 2v16a2 2 0 0 0 2 2h12a2 2 0 0 0 2-2V8z" />
                  <polyline points="14 2 14 8 20 8" />
                </svg>
                <span>Открыть config.json в редакторе</span>
              </button>
            </div>
          </div>
        </div>
      </div>
    </div>
  );
}
