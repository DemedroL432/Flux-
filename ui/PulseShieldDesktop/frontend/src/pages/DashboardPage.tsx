import { useApp } from "../context/AppContext";
import {
  formatSessionClock,
  formatBps,
  ROUTING_MODE_SHORT,
  routingProfileLabel,
} from "../utils";
import { OpenLogsFolder, CreateDiagnosticsBundle } from "../../wailsjs/go/main/App";
import { CountryFlag } from "../components/CountryFlag";

function ArrowSvg({ color }: { color?: string }) {
  return (
    <svg width="10" height="10" viewBox="0 0 12 12" fill="none" stroke={color || "currentColor"} strokeWidth="2.2" strokeLinecap="round" strokeLinejoin="round" style={{ display: "inline-block", verticalAlign: "middle" }}>
      <path d="M2 6h8M7 3l3 3-3 3" />
    </svg>
  );
}

export function DashboardPage() {
  const {
    active, setActive,
    proxyEnabled, dpiEnabled,
    setProxy, setDpi,
    busy, proxyPhase,
    uiState, health, rt,
    showVlessProtectedStrip,
    setToast,
    setRouting,
    runRetestServersBatch,
    sortedServers, speedHistory,
  } = useApp();

  if (active !== "dashboard") return null;

  const proxyConnecting = proxyPhase === "connecting";
  const rk = uiState?.routingProfile ?? "bypass-russia";
  const routingLabel = (rk && rk in ROUTING_MODE_SHORT ? ROUTING_MODE_SHORT[rk] : "") || routingProfileLabel(rk || "—");
  const server = sortedServers.find((s: any) => s.id === uiState?.activeServerId);
  const location = server?.name || "Сервер не выбран";
  const host = server ? `${server.host}:${server.port}` : "—";
  const proto = "VLESS";
  const activeDpiProfile = health?.activeDpiProfile || uiState?.activeDpiProfile || "zapret:auto";

  // Protection states
  const proxyUp = showVlessProtectedStrip;
  const dpiUp = !!(dpiEnabled && health?.dpiRunning);
  const isMasterActive = proxyUp || dpiUp;

  let masterTitle = "Защита выключена";
  let masterSub = "Прокси и DPI выключены. Трафик идёт напрямую.";
  if (proxyUp && dpiUp) {
    masterTitle = "Максимальная защита активна";
    masterSub = "Прокси и DPI запущены. Трафик зависит от правил маршрутизации.";
  } else if (proxyUp) {
    masterTitle = "VPN защита активна";
    masterSub = `Туннель поднят · ${location}`;
  } else if (dpiUp) {
    masterTitle = "DPI обход активен";
    masterSub = "Процесс Zapret запущен. Доступность зависит от сети и профиля.";
  } else if (proxyConnecting) {
    masterTitle = "Подключение…";
    masterSub = "Поднимаем туннель VLESS";
  }

  const coreRunning = !!health?.proxyLocalOk;
  const dpiRunning = !!health?.dpiRunning;
  const dnsRunning = false; // DNS health is not supplied by the backend.

  async function handleToggleMaster() {
    if (isMasterActive) {
      await setProxy(false);
      await setDpi(false);
    } else {
      await setProxy(true);
      await setDpi(true);
    }
  }

  async function handleSpeedtest() {
    try {
      setToast({ text: "Тестирование задержки серверов…", variant: "info" });
      const ids = (sortedServers || []).map((s: any) => s.id || s.name);
      await runRetestServersBatch(ids);
      setToast({ text: "Пинг серверов обновлён", variant: "info" });
    } catch {
      setToast({ text: "Ошибка проверки серверов", variant: "error" });
    }
  }

  async function handleExportDiag() {
    try {
      const path = await CreateDiagnosticsBundle();
      setToast({ text: "Архив сохранён: " + path, variant: "info" });
    } catch (e) {
      setToast({ text: "Ошибка экспорта: " + String(e), variant: "error" });
    }
  }

  return (
    <div className="flux-dashboard-grid">
      {/* ── Left Column ── */}
      <div className="flux-dashboard-left-col">
        {/* Top Master Status Card */}
        <div className={`flux-master-card${isMasterActive ? "" : " off"}`}>
          <div className="flux-master-left">
            <div className="flux-master-shield">
              <svg width="24" height="24" viewBox="0 0 24 24" fill="none" stroke="currentColor" strokeWidth="2.2" strokeLinecap="round" strokeLinejoin="round">
                <path d="M12 3 5 6v5c0 4.5 3 7.5 7 9 4-1.5 7-4.5 7-9V6l-7-3Z" />
                {isMasterActive && <path d="M9 12l2 2 4-4" />}
              </svg>
            </div>
            <div className="flux-master-titles">
              <h2 className="flux-master-h1">{masterTitle}</h2>
              <span className="flux-master-sub">{masterSub}</span>
              <div className="flux-master-dots-row">
                <div className="flux-master-dot-item">
                  <span className={proxyUp ? "flux-dot-active" : "flux-dot-inactive"} />
                  <span>Прокси</span>
                </div>
                <div className="flux-master-dot-item">
                  <span className={dpiUp ? "flux-dot-active" : "flux-dot-inactive"} />
                  <span>DPI</span>
                </div>
                <div className="flux-master-dot-item">
                  <span className="flux-dot-active" />
                  <span>Маршрутизация</span>
                </div>
                <div className="flux-master-dot-item">
                  <span className={dnsRunning ? "flux-dot-active" : "flux-dot-inactive"} />
                  <span>DNS: не проверен</span>
                </div>
              </div>
            </div>
          </div>

          <div className="flux-master-right">
            <div className="flux-master-server-pill">
              <span className="flux-master-flag"><CountryFlag name={location} width={22} height={15} /></span>
              <div className="flux-master-server-meta">
                <span className="flux-master-server-name">{location}</span>
                <span className="flux-master-server-host">{host}</span>
                <span className="flux-master-server-proto">{proto}</span>
              </div>
            </div>

            <div className="flux-master-stop-col">
              <span className="flux-master-uptime">
                Время работы: {formatSessionClock(Number(rt?.sessionElapsedSeconds || 0))}
              </span>
              <button
                type="button"
                className={`flux-btn-master-action${isMasterActive ? " stop" : " start"}`}
                disabled={busy}
                onClick={() => void handleToggleMaster()}
              >
                {isMasterActive ? (
                  <>
                    <span className="flux-btn-square-icon" />
                    <span>Остановить всё</span>
                  </>
                ) : (
                  <>
                    <svg width="9" height="9" viewBox="0 0 24 24" fill="none" stroke="currentColor" strokeWidth="2.5">
                      <polygon points="5 3 19 12 5 21 5 3" />
                    </svg>
                    <span>Включить всё</span>
                  </>
                )}
              </button>
            </div>
          </div>
        </div>

        {/* Dual Control Modules */}
        <div className="flux-controls-row">
          {/* Proxy Control */}
          <div className={`flux-control-card${proxyEnabled ? " active" : ""}`}>
            <div className="flux-control-header">
              <div className="flux-control-title-wrap">
                <div className="flux-control-icon">
                  <svg width="18" height="18" viewBox="0 0 24 24" fill="none" stroke="currentColor" strokeWidth="2">
                    <circle cx="12" cy="12" r="9" />
                    <path d="M12 3c3.4 4 3.4 14 0 18M3 12h18M5.5 7.5c4 1.5 9 1.5 13 0M5.5 16.5c4-1.5 9-1.5 13 0" />
                  </svg>
                </div>
                <div className="flux-control-titles">
                  <span className="flux-control-title">Прокси / VPN</span>
                  <span className="flux-control-sub">Подключение через VLESS</span>
                </div>
              </div>

              <label className="flux-switch">
                <input
                  type="checkbox"
                  checked={proxyEnabled}
                  disabled={busy || proxyConnecting}
                  onChange={(e) => void setProxy(e.target.checked)}
                />
                <span className="flux-switch-slider" />
              </label>
            </div>

            <div
              className="flux-control-selector"
              onClick={() => setActive("servers")}
              title="Выбрать сервер"
            >
              <div className="flux-selector-left">
                <CountryFlag name={location} width={22} height={15} />
                <div className="flux-selector-meta">
                  <span className="flux-selector-name">{location}</span>
                  <span className="flux-selector-host">{host}</span>
                </div>
              </div>
              <div className="flux-selector-right">
                <span className="flux-ping-badge">{server?.lastCheckOk ? `${server.lastCheckLatencyMs} ms` : "Не измерен"}</span>
                <span className="flux-selector-chevron">›</span>
              </div>
            </div>

            <div className="flux-control-meta-grid">
              <div className="flux-control-meta-item">
                <span className="flux-control-meta-label">Протокол</span>
                <span className="flux-control-meta-val">{proto}</span>
              </div>
              <div className="flux-control-meta-item">
                <span className="flux-control-meta-label">Маршрутизация</span>
                <span className="flux-control-meta-val">{routingLabel}</span>
              </div>
            </div>

            <div className="flux-control-actions">
              <button
                type="button"
                className="flux-btn-control-action"
                onClick={() => setActive("servers")}
              >
                <svg width="14" height="14" viewBox="0 0 24 24" fill="none" stroke="currentColor" strokeWidth="2">
                  <rect x="3" y="4" width="18" height="6" rx="2" />
                  <rect x="3" y="14" width="18" height="6" rx="2" />
                </svg>
                <span>Выбрать сервер</span>
                <span style={{ marginLeft: "auto" }}>→</span>
              </button>
              <button
                type="button"
                className="flux-btn-icon-more"
                onClick={() => setActive("settings")}
                title="Настройки сети"
              >
                •••
              </button>
            </div>
          </div>

          {/* DPI Control */}
          <div className={`flux-control-card${dpiEnabled ? " active-dpi" : ""}`}>
            <div className="flux-control-header">
              <div className="flux-control-title-wrap">
                <div className="flux-control-icon dpi">
                  <svg width="18" height="18" viewBox="0 0 24 24" fill="none" stroke="currentColor" strokeWidth="2">
                    <path d="M12 3 5 6v5c0 4.5 3 7.5 7 9 4-1.5 7-4.5 7-9V6l-7-3Z" />
                  </svg>
                </div>
                <div className="flux-control-titles">
                  <span className="flux-control-title">DPI / Обход блокировок</span>
                  <span className="flux-control-sub">Zapret (WinDivert)</span>
                </div>
              </div>

              <label className="flux-switch dpi">
                <input
                  type="checkbox"
                  checked={dpiEnabled}
                  disabled={busy}
                  onChange={(e) => void setDpi(e.target.checked)}
                />
                <span className="flux-switch-slider" />
              </label>
            </div>

            <div
              className="flux-control-selector"
              onClick={() => setActive("dpi")}
              title="Настройки профиля DPI"
            >
              <div className="flux-selector-left">
                <div className="flux-selector-meta">
                  <span className="flux-selector-name">Профиль: {activeDpiProfile}</span>
                  <span className="flux-selector-host">Анти-ТСПУ стратегия</span>
                </div>
              </div>
              <div className="flux-selector-right">
                <span className="flux-selector-chevron">›</span>
              </div>
            </div>

            <div className="flux-control-meta-grid">
              <div className="flux-control-meta-item">
                <span className="flux-control-meta-label">Состояние</span>
                <span className="flux-control-meta-val" style={{ color: dpiRunning ? "var(--flux-mint)" : "inherit" }}>
                  {dpiRunning ? "● Работает" : "Остановлен"}
                </span>
              </div>
              <div className="flux-control-meta-item">
                <span className="flux-control-meta-label">Процесс</span>
                <span className="flux-control-meta-val">
                  {dpiRunning ? "Процесс запущен" : "Остановлен"}
                </span>
              </div>
            </div>

            <div className="flux-control-actions">
              <button
                type="button"
                className="flux-btn-control-action"
                onClick={() => setActive("dpi")}
              >
                <svg width="14" height="14" viewBox="0 0 24 24" fill="none" stroke="currentColor" strokeWidth="2">
                  <path d="M12 3 5 6v5c0 4.5 3 7.5 7 9 4-1.5 7-4.5 7-9V6l-7-3Z" />
                </svg>
                <span>Настройки DPI</span>
                <span style={{ marginLeft: "auto" }}>→</span>
              </button>
              <button
                type="button"
                className="flux-btn-icon-more"
                onClick={() => setActive("dpi")}
                title="Дополнительные опции DPI"
              >
                •••
              </button>
            </div>
          </div>
        </div>

        {/* Traffic Chart Card */}
        <div className="flux-card flux-traffic-card">
          <div className="flux-traffic-header">
            <div className="flux-traffic-title-group">
              <div className="flux-traffic-brand-title">
                <svg width="15" height="15" viewBox="0 0 24 24" fill="none" stroke="currentColor" strokeWidth="2" style={{ color: "var(--flux-cyan)" }}>
                  <path d="M18 20V10M12 20V4M6 20v-6" />
                </svg>
                <span>Сетевой трафик</span>
              </div>

              <div className="flux-speed-badge down">
                <span className="flux-speed-circle down">↓</span>
                <span className="flux-speed-num">{formatBps(rt.downloadBps)}</span>
                <span className="flux-speed-desc">Скачивание</span>
              </div>

              <div className="flux-speed-badge up">
                <span className="flux-speed-circle up">↑</span>
                <span className="flux-speed-num">{formatBps(rt.uploadBps)}</span>
                <span className="flux-speed-desc">Отдача</span>
              </div>
            </div>

            <span>Последние 60 измерений · суммарный трафик</span>
          </div>

          <div className="flux-traffic-chart-area">
            <svg className="flux-traffic-svg" viewBox="0 0 100 45" preserveAspectRatio="none">
              <defs>
                <linearGradient id="cyan-wave-fill" x1="0" y1="0" x2="0" y2="45" gradientUnits="userSpaceOnUse">
                  <stop offset="0%" stopColor="#38bdf8" stopOpacity="0.45" />
                  <stop offset="100%" stopColor="#38bdf8" stopOpacity="0.0" />
                </linearGradient>
                <linearGradient id="purple-wave-fill" x1="0" y1="0" x2="0" y2="45" gradientUnits="userSpaceOnUse">
                  <stop offset="0%" stopColor="#d9b9ff" stopOpacity="0.30" />
                  <stop offset="100%" stopColor="#d9b9ff" stopOpacity="0.0" />
                </linearGradient>
              </defs>

              {/* Grid Lines */}
              <line className="flux-chart-grid-line" x1="0" y1="12" x2="100" y2="12" vectorEffect="non-scaling-stroke" />
              <line className="flux-chart-grid-line" x1="0" y1="26" x2="100" y2="26" vectorEffect="non-scaling-stroke" />
              <line className="flux-chart-baseline" x1="0" y1="41" x2="100" y2="41" vectorEffect="non-scaling-stroke" />

              <polyline
                points={speedHistory.map((v, i) => `${i * 100 / Math.max(speedHistory.length - 1, 1)},${41 - Math.max(0, v) * 38 / Math.max(...speedHistory, 1)}`).join(" ")}
                fill="none" stroke="#38bdf8" strokeWidth="1.8" vectorEffect="non-scaling-stroke"
              />
            </svg>

            {/* Y-axis Labels on the right */}
            <div className="flux-chart-y-axis">
              <span>{formatBps(Math.max(...speedHistory, 0))}</span>
              <span>{formatBps(Math.max(...speedHistory, 0) / 2)}</span>
              <span>0</span>
            </div>

            {/* X-axis Labels on bottom */}
            <div className="flux-chart-x-axis">
              <span>Ранее</span>
              <span>50</span>
              <span>40</span>
              <span>30</span>
              <span>20</span>
              <span>10</span>
              <span>0</span>
            </div>
          </div>
        </div>

        {/* Service Health Card (Upper row: header + button, Lower row: 4 strict columns) */}
        <div className="flux-health-strip">
          <div className="flux-health-header">
            <div className="flux-health-title-wrap">
              <svg width="15" height="15" viewBox="0 0 24 24" fill="none" stroke="currentColor" strokeWidth="2" style={{ color: "var(--flux-cyan)" }}>
                <circle cx="12" cy="12" r="3" />
                <path d="M19.4 15a1.65 1.65 0 0 0 .33 1.82l.06.06a2 2 0 0 1-2.83 2.83l-.06-.06a1.65 1.65 0 0 0-1.82-.33 1.65 1.65 0 0 0-1 1.51V21a2 2 0 0 1-2 2 2 2 0 0 1-2-2v-.09A1.65 1.65 0 0 0 9 19.4a1.65 1.65 0 0 0-1.82.33l-.06.06a2 2 0 0 1-2.83-2.83l.06-.06a1.65 1.65 0 0 0 .33-1.82 1.65 1.65 0 0 0-1.51-1H3a2 2 0 0 1-2-2 2 2 0 0 1 2-2h.09A1.65 1.65 0 0 0 4.6 9a1.65 1.65 0 0 0-.33-1.82l-.06-.06a2 2 0 0 1 2.83-2.83l.06.06a1.65 1.65 0 0 0 1.82.33H9a1.65 1.65 0 0 0 1-1.51V3a2 2 0 0 1 2-2 2 2 0 0 1 2 2v.09a1.65 1.65 0 0 0 1 1.51 1.65 1.65 0 0 0 1.82-.33l.06-.06a2 2 0 0 1 2.83 2.83l-.06.06a1.65 1.65 0 0 0-.33 1.82V9a1.65 1.65 0 0 0 1.51 1H21a2 2 0 0 1 2 2 2 2 0 0 1-2 2h-.09a1.65 1.65 0 0 0-1.51 1z" />
              </svg>
              <span className="flux-health-title">Состояние служб</span>
            </div>
            <button
              type="button"
              className="flux-btn-health-more"
              onClick={() => setActive("diagnostics")}
            >
              Подробнее →
            </button>
          </div>

          <div className="flux-health-grid">
            {/* Column 1: sing-box */}
            <div className="flux-health-col">
              <span className="flux-health-head">sing-box</span>
              <div className="flux-health-status-wrap">
                <span className={coreRunning ? "flux-dot-active" : "flux-dot-inactive"} />
                <span className="flux-health-status" style={{ color: coreRunning ? "var(--flux-mint)" : "var(--flux-text-muted)" }}>
                  {coreRunning ? "Работает" : "Остановлен"}
                </span>
              </div>
              <span className="flux-health-version">v1.10.7</span>
            </div>

            {/* Column 2: WinDivert */}
            <div className="flux-health-col">
              <span className="flux-health-head">WinDivert</span>
              <div className="flux-health-status-wrap">
                <span className={dpiRunning ? "flux-dot-active" : "flux-dot-inactive"} />
                <span className="flux-health-status" style={{ color: dpiRunning ? "var(--flux-mint)" : "var(--flux-text-muted)" }}>
                  {dpiRunning ? "Активен" : "Неактивен"}
                </span>
              </div>
              <span className="flux-health-version">v2.2.2</span>
            </div>

            {/* Column 3: Системный прокси */}
            <div className="flux-health-col">
              <span className="flux-health-head">Системный прокси</span>
              <div className="flux-health-status-wrap">
                <span className={proxyUp ? "flux-dot-active" : "flux-dot-inactive"} />
                <span className="flux-health-status" style={{ color: proxyUp ? "var(--flux-mint)" : "var(--flux-text-muted)" }}>
                  {proxyUp ? "Включен" : "Выключен"}
                </span>
              </div>
              <span className="flux-health-version">
                {"Порт: " + (health?.proxyListenPort || uiState?.proxyPort || 8080)}
              </span>
            </div>

            {/* Column 4: DNS */}
            <div className="flux-health-col">
              <span className="flux-health-head">DNS</span>
              <div className="flux-health-status-wrap">
                <span className={dnsRunning ? "flux-dot-active" : "flux-dot-inactive"} />
                <span className="flux-health-status" style={{ color: dnsRunning ? "var(--flux-mint)" : "var(--flux-red)" }}>
                  {dnsRunning ? "Работает" : "Сбой"}
                </span>
              </div>
              <span className="flux-health-version">Google (8.8.8.8)</span>
            </div>
          </div>
        </div>
      </div>

      {/* ── Right Column: UNIFIED MONOLITH PANEL ── */}
      <div className="flux-dashboard-right-col">
        <div className="flux-right-monolith">
          {/* Section 1: Маршрутизация */}
          <div className="flux-monolith-section">
            <div className="flux-monolith-header">
              <div className="flux-monolith-title-wrap">
                <svg width="16" height="16" viewBox="0 0 24 24" fill="none" stroke="currentColor" strokeWidth="2" style={{ color: "var(--flux-cyan)" }}>
                  <circle cx="12" cy="5" r="2.5" />
                  <circle cx="5" cy="19" r="2.5" />
                  <circle cx="19" cy="19" r="2.5" />
                  <path d="M12 7.5v4M12 11.5L6.5 16.5M12 11.5l5.5 5" />
                </svg>
                <span className="flux-monolith-title">Маршрутизация</span>
              </div>
              <span className="flux-monolith-sub">Куда уходит трафик</span>
            </div>

            <div className="flux-select-wrap">
              <select
                className="flux-routing-mode-select"
                value={rk}
                onChange={(e) => void setRouting(e.target.value)}
              >
                <option value="bypass-russia">Обход РФ (Только блокировки через VPN)</option>
                <option value="all-proxy">Полный прокси (Весь трафик)</option>
                <option value="all-direct">Прямой (Весь трафик без VPN)</option>
              </select>
            </div>

            <div className="flux-routing-flow-list">
              <div className="flux-routing-flow-row">
                <span className="flux-flow-label">
                  <ArrowSvg color="var(--flux-mint)" />
                  <span>Весь трафик</span>
                </span>
                <span className="flux-flow-dst">
                  <ArrowSvg color="var(--flux-blue)" />
                  <span>через VPN</span>
                </span>
              </div>
              <div className="flux-routing-flow-row">
                <span className="flux-flow-label">
                  <ArrowSvg color="var(--flux-red)" />
                  <span>Российские сайты</span>
                </span>
                <span className="flux-flow-dst">
                  <ArrowSvg color={rk === "bypass-russia" ? "var(--flux-mint)" : "var(--flux-blue)"} />
                  <span>{rk === "bypass-russia" ? "напрямую" : "через VPN"}</span>
                </span>
              </div>
              <div className="flux-routing-flow-row">
                <span className="flux-flow-label">
                  <ArrowSvg color="var(--flux-cyan)" />
                  <span>YouTube, Discord</span>
                </span>
                <span className="flux-flow-dst">
                  <ArrowSvg color="var(--flux-blue)" />
                  <span>через VPN</span>
                </span>
              </div>
              <div className="flux-routing-flow-row">
                <span className="flux-flow-label">
                  <ArrowSvg color="var(--flux-purple)" />
                  <span>Остальные сайты</span>
                </span>
                <span className="flux-flow-dst">
                  <ArrowSvg color="var(--flux-blue)" />
                  <span>{rk === "all-direct" ? "напрямую" : "через VPN"}</span>
                </span>
              </div>
            </div>

            <button
              type="button"
              className="flux-btn-monolith-action"
              onClick={() => setActive("network")}
            >
              <span>Настроить правила</span>
              <span className="flux-btn-arrow">→</span>
            </button>
          </div>

          {/* Divider */}
          <div className="flux-monolith-divider" />

          {/* Section 2: Статистика сессии */}
          <div className="flux-monolith-section">
            <div className="flux-monolith-header">
              <div className="flux-monolith-title-wrap">
                <svg width="15" height="15" viewBox="0 0 24 24" fill="none" stroke="currentColor" strokeWidth="2" style={{ color: "var(--flux-cyan)" }}>
                  <circle cx="12" cy="12" r="9" />
                  <path d="M12 7v5l3 3" />
                </svg>
                <span className="flux-monolith-title">Статистика сессии</span>
              </div>
            </div>

            <div className="flux-session-stats-list">
              <div className="flux-session-stat-row">
                <div className="flux-session-stat-lbl">
                  <span className="flux-stat-badge down">↓</span>
                  <span>Скачано</span>
                </div>
                <span className="flux-session-stat-val">1.6 GB</span>
              </div>
              <div className="flux-session-stat-row">
                <div className="flux-session-stat-lbl">
                  <span className="flux-stat-badge up">↑</span>
                  <span>Отдано</span>
                </div>
                <span className="flux-session-stat-val">141 MB</span>
              </div>
              <div className="flux-session-stat-row">
                <div className="flux-session-stat-lbl">
                  <span className="flux-stat-badge clock">⏱</span>
                  <span>Время сессии</span>
                </div>
                <span className="flux-session-stat-val">{formatSessionClock(Number(rt?.sessionElapsedSeconds || 0))}</span>
              </div>
            </div>
          </div>

          {/* Divider */}
          <div className="flux-monolith-divider" />

          {/* Section 3: Быстрые действия */}
          <div className="flux-monolith-section">
            <div className="flux-monolith-header">
              <div className="flux-monolith-title-wrap">
                <svg width="15" height="15" viewBox="0 0 24 24" fill="none" stroke="currentColor" strokeWidth="2" style={{ color: "var(--flux-cyan)" }}>
                  <polygon points="13 2 3 14 12 14 11 22 21 10 12 10 13 2" />
                </svg>
                <span className="flux-monolith-title">Быстрые действия</span>
              </div>
            </div>

            <div className="flux-actions-grid">
              <button
                type="button"
                className="flux-btn-quick-action"
                onClick={() => void handleSpeedtest()}
                title="Проверить пинг всех серверов"
              >
                <svg width="14" height="14" viewBox="0 0 24 24" fill="none" stroke="currentColor" strokeWidth="2" style={{ color: "var(--flux-cyan)" }}>
                  <circle cx="12" cy="12" r="10" />
                  <path d="M12 6v6l4 2" />
                </svg>
                <span>Тест скорости</span>
              </button>

              <button
                type="button"
                className="flux-btn-quick-action"
                onClick={() => setActive("diagnostics")}
                title="Перейти к полной диагностике"
              >
                <svg width="14" height="14" viewBox="0 0 24 24" fill="none" stroke="currentColor" strokeWidth="2" style={{ color: "var(--flux-cyan)" }}>
                  <path d="M12 22s8-4 8-10V5l-8-3-8 3v7c0 6 8 10 8 10z" />
                  <path d="M9 12l2 2 4-4" />
                </svg>
                <span>Проверить систему</span>
              </button>

              <button
                type="button"
                className="flux-btn-quick-action"
                onClick={() => void OpenLogsFolder().catch(() => {})}
                title="Открыть папку с логами"
              >
                <svg width="14" height="14" viewBox="0 0 24 24" fill="none" stroke="currentColor" strokeWidth="2" style={{ color: "var(--flux-cyan)" }}>
                  <path d="M14 2H6a2 2 0 0 0-2 2v16a2 2 0 0 0 2 2h12a2 2 0 0 0 2-2V8z" />
                  <polyline points="14 2 14 8 20 8" />
                </svg>
                <span>Открыть логи</span>
              </button>

              <button
                type="button"
                className="flux-btn-quick-action"
                onClick={() => void handleExportDiag()}
                title="Создать диагностический zip-архив"
              >
                <svg width="14" height="14" viewBox="0 0 24 24" fill="none" stroke="currentColor" strokeWidth="2" style={{ color: "var(--flux-cyan)" }}>
                  <path d="M21 15v4a2 2 0 0 1-2 2H5a2 2 0 0 1-2-2v-4M7 10l5 5 5-5M12 15V3" />
                </svg>
                <span>Экспорт отчёта</span>
              </button>
            </div>
          </div>
        </div>
      </div>
    </div>
  );
}
