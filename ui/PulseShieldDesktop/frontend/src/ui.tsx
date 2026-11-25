import React from "react";
import { useApp } from "./context/AppContext";
import { NAV_ITEMS, NavKey } from "./types";
import { Icon } from "./icons";
import { formatClock } from "./format";

export function routingLabel(name?: string): string {
  switch (name) {
    case "manual":
      return "Выборочно";
    case "full-proxy":
      return "Всё, кроме РФ";
    case "global":
      return "Всё (100% GEO)";
    default:
      return name || "Выборочно";
  }
}

export function Switch({
  checked,
  disabled,
  label,
  onChange,
}: {
  checked: boolean;
  disabled?: boolean;
  label: string;
  onChange: (next: boolean) => void;
}) {
  return (
    <button
      type="button"
      role="switch"
      aria-checked={checked}
      aria-label={label}
      disabled={disabled}
      className={`fx-sw ${checked ? "on" : ""}`}
      onClick={(e) => {
        e.stopPropagation();
        onChange(!checked);
      }}
    />
  );
}

export function Sidebar() {
  const {
    active,
    setActive,
    appVersion,
    uiState,
    proxyEnabled,
    dpiEnabled,
    proxyPhase,
    setProxy,
    setDpi,
    activeServer,
    servers,
    splitRows,
    dpiLists,
    busy,
  } = useApp();

  const dpiEnabledCount = dpiLists.filter((l) => Boolean(l.enabled)).length;
  const countFor = (key: NavKey): string | undefined => {
    if (key === "servers" && servers.length > 0) return String(servers.length);
    if (key === "routes" && splitRows.length > 0) return String(splitRows.length);
    if (key === "dpi" && dpiLists.length > 0) return `${dpiEnabledCount}/${dpiLists.length}`;
    return undefined;
  };

  const activeServerName = String(activeServer?.name ?? "");
  const proxySub =
    proxyPhase === "connecting"
      ? "Подключение…"
      : proxyEnabled
        ? activeServerName || "Запущен"
        : "Остановлен";
  const dpiSub = dpiEnabled
    ? String(uiState?.activeDpiProfile || "Активен")
    : "Остановлен";

  return (
    <aside className="fx-side">
      <div
        className="fx-brand"
        style={{ ["--wails-draggable" as string]: "drag" }}
      >
        <div className="fx-brand-id">
          <span className="fx-brand-mark" aria-hidden="true">
            <svg viewBox="0 0 16 16" width="12" height="12" fill="none">
              <path
                d="M8 1.5L13.5 4.2V8.4C13.5 11.4 11.1 13.7 8 14.6C4.9 13.7 2.5 11.4 2.5 8.4V4.2L8 1.5Z"
                stroke="currentColor"
                strokeWidth="1.5"
                strokeLinejoin="round"
              />
              <circle cx="8" cy="8" r="1.8" fill="currentColor" />
            </svg>
          </span>
          <span className="fx-brand-name">Flux</span>
        </div>
        <span className="fx-ver">v{appVersion || "0.7.7"}</span>
      </div>

      <div className="fx-engines">
        <div className="fx-side-label">Движки</div>

        <div className="fx-eng-row">
          <div className="fx-eng-main">
            <span
              className={`fx-dot ${
                proxyPhase === "connected"
                  ? "on"
                  : proxyPhase === "connecting"
                    ? "connecting"
                    : ""
              }`}
            />
            <div className="fx-eng-text">
              <div className="fx-eng-title">Прокси</div>
              <div className="fx-eng-sub" title={proxySub}>
                {proxySub}
              </div>
            </div>
          </div>
          <Switch
            checked={proxyEnabled}
            disabled={busy}
            label="Переключить прокси"
            onChange={(next) => void setProxy(next)}
          />
        </div>

        <div className="fx-eng-row">
          <div className="fx-eng-main">
            <span className={`fx-dot ${dpiEnabled ? "on" : ""}`} />
            <div className="fx-eng-text">
              <div className="fx-eng-title">Обход DPI</div>
              <div className="fx-eng-sub" title={dpiSub}>
                {dpiSub}
              </div>
            </div>
          </div>
          <Switch
            checked={dpiEnabled}
            disabled={busy}
            label="Переключить обход DPI"
            onChange={(next) => void setDpi(next)}
          />
        </div>
      </div>

      <nav className="fx-nav" aria-label="Навигация">
        {NAV_ITEMS.map((item) => {
          const count = countFor(item.key);
          return (
            <button
              key={item.key}
              type="button"
              className={`fx-nav-btn ${active === item.key ? "active" : ""}`}
              onClick={() => setActive(item.key)}
            >
              <Icon name={item.icon} />
              <span className="fx-nav-label">{item.label}</span>
              {count ? <span className="fx-nav-count">{count}</span> : null}
            </button>
          );
        })}
      </nav>

      <div className="fx-side-foot">
        <span className="fx-foot-pill">
          <span className={`fx-foot-dot ${uiState?.elevated ? "elev" : ""}`} />
          <span>{uiState?.elevated ? "Администратор" : "Обычные права"}</span>
        </span>
        <span className="fx-mono">:{Number(uiState?.proxyPort ?? 8080)}</span>
      </div>
    </aside>
  );
}

export function TopBar() {
  const {
    active,
    rt,
    proxyEnabled,
    dpiEnabled,
    servers,
    splitRows,
    uiState,
    appVersion,
    updateProgress,
    closeToTray,
  } = useApp();

  const currentNav = NAV_ITEMS.find((n) => n.key === active);
  const title = currentNav?.label ?? "Flux";

  let subtitle = "";
  if (active === "shield") subtitle = "sing-box · winws2";
  else if (active === "servers") subtitle = `${servers.length} узлов`;
  else if (active === "routes") {
    subtitle = `${routingLabel(String(uiState?.routingProfile ?? ""))} · ${splitRows.length} правил`;
  } else if (active === "dpi") {
    subtitle = String(uiState?.activeDpiProfile || "Профили и списки");
  } else if (active === "net") {
    subtitle = `Flux v${appVersion || "0.7.7"}`;
  }

  const isLive = Boolean(rt?.connected || proxyEnabled || dpiEnabled);
  const elapsed = Number(rt?.sessionElapsedSeconds ?? 0);

  return (
    <header
      className="fx-top"
      style={{ ["--wails-draggable" as string]: "drag" }}
    >
      <div className="fx-top-left">
        <span className="fx-top-title">{title}</span>
        {subtitle ? <span className="fx-top-sub">{subtitle}</span> : null}
      </div>

      <div
        className="fx-top-right"
        style={{ ["--wails-draggable" as string]: "no-drag" }}
      >
        {updateProgress && (
          <div className="fx-session-pill live" title="Автоматическое обновление Flux">
            <span
              className="fx-dot on"
              style={{ marginTop: 0, width: 6, height: 6 }}
            />
            <span>
              {updateProgress.phase === "unpack"
                ? "Распаковка…"
                : updateProgress.phase === "install"
                  ? "Установка…"
                  : updateProgress.pct >= 0
                    ? `Обновление ${updateProgress.pct}%`
                    : "Скачивание…"}
            </span>
          </div>
        )}

        <div className={`fx-session-pill ${isLive ? "live" : ""}`}>
          <span
            className={`fx-dot ${isLive ? "on" : ""}`}
            style={{ marginTop: 0, width: 6, height: 6 }}
          />
          <span>{isLive ? formatClock(elapsed) : "Простой"}</span>
        </div>

        <div className="fx-win-btns">
          <button
            type="button"
            className="fx-win-btn"
            title="Свернуть"
            aria-label="Свернуть окно"
            onClick={() => {
              void import("../wailsjs/runtime/runtime").then((m) =>
                m.WindowMinimise(),
              );
            }}
          >
            —
          </button>
          <button
            type="button"
            className="fx-win-btn"
            title="Развернуть"
            aria-label="Развернуть окно"
            onClick={() => {
              void import("../wailsjs/runtime/runtime").then((m) =>
                m.WindowToggleMaximise(),
              );
            }}
          >
            □
          </button>
          <button
            type="button"
            className="fx-win-btn close"
            title="Свернуть в трей"
            aria-label="Свернуть в трей"
            onClick={() => void closeToTray()}
          >
            ×
          </button>
        </div>
      </div>
    </header>
  );
}

export function BusyLine() {
  const { busy } = useApp();
  return <div className={`fx-busy-line ${busy ? "active" : ""}`} aria-hidden="true" />;
}

export function Toasts() {
  const { toast, dismissToast, setActive } = useApp();
  if (!toast) return null;

  const variantClass = toast.variant === "error" ? "error" : "ok";

  return (
    <div className="fx-toast-stack" role="status" aria-live="polite">
      <div className={`fx-toast ${variantClass}`}>
        <span style={{ flex: 1 }}>{toast.text}</span>
        {toast.openLogs ? (
          <button
            type="button"
            className="fx-btn fx-btn-sm"
            onClick={() => {
              setActive("net");
              dismissToast();
            }}
          >
            Логи
          </button>
        ) : null}
        <button
          type="button"
          className="fx-toast-close"
          aria-label="Закрыть уведомление"
          onClick={dismissToast}
        >
          ×
        </button>
      </div>
    </div>
  );
}
