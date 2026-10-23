import { useApp } from "../context/AppContext";
import { WindowMinimise, WindowToggleMaximise, Quit } from "../../wailsjs/runtime/runtime";
import { RequestQuit } from "../../wailsjs/go/main/App";
import { NavKey } from "../types";

const NAV_TITLES: Record<NavKey, string> = {
  dashboard: "Обзор",
  servers: "Серверы",
  network: "Маршрутизация",
  dpi: "DPI (Обход блокировок)",
  diagnostics: "Диагностика",
  logs: "Журнал",
  settings: "Настройки",
  guide: "Справка",
  support: "Поддержка",
};

export function DesktopTopBar() {
  const { active, health, runUpdateCheck, busy, globalBusy } = useApp();

  const isProtected = health?.proxyLocalOk || health?.dpiRunning;
  const title = NAV_TITLES[active] || "Главная";

  return (
    <header className="flux-topbar" style={{ ["--wails-draggable" as any]: "drag" }}>
      <div className="flux-topbar-left" style={{ ["--wails-draggable" as any]: "drag" }}>
        <span className="flux-section-bar">|</span>
        <span className="flux-section-name">{title}</span>
      </div>

      <div className="flux-topbar-right" style={{ ["--wails-draggable" as any]: "no-drag" }}>
        <div className={`flux-status-badge${isProtected ? "" : " off"}`}>
          <span className="flux-status-badge-dot" />
          <span>{isProtected ? "Модули работают" : "Модули остановлены"}</span>
        </div>

        <button
          type="button"
          className="flux-btn-check-update"
          disabled={busy}
          onClick={() => void runUpdateCheck()}
          title="Проверить обновление"
        >
          <svg width="12" height="12" viewBox="0 0 24 24" fill="none" stroke="currentColor" strokeWidth="2">
            <path d="M21.5 2v6h-6M21.34 15.57a10 10 0 1 1-.57-8.38l5.67-5.67" />
          </svg>
          <span>
            {typeof globalBusy === "string" && globalBusy.includes("обновлен")
              ? "Проверка…"
              : "Проверить обновление"}
          </span>
        </button>

        <div className="flux-window-controls">
          <button
            type="button"
            className="flux-win-btn"
            title="Свернуть"
            onClick={() => WindowMinimise()}
          >
            <svg width="10" height="2" viewBox="0 0 10 2" fill="currentColor">
              <rect width="10" height="2" rx="1" />
            </svg>
          </button>
          <button
            type="button"
            className="flux-win-btn"
            title="Развернуть"
            onClick={() => WindowToggleMaximise()}
          >
            <svg width="9" height="9" viewBox="0 0 10 10" fill="none" stroke="currentColor" strokeWidth="1.3">
              <rect x="0.5" y="0.5" width="9" height="9" rx="1" />
            </svg>
          </button>
          <button
            type="button"
            className="flux-win-btn close"
            title="Закрыть"
            onClick={() => {
              RequestQuit().catch(() => Quit());
            }}
          >
            <svg width="10" height="10" viewBox="0 0 10 10" fill="none" stroke="currentColor" strokeWidth="1.6">
              <path d="M1 1l8 8M9 1L1 9" />
            </svg>
          </button>
        </div>
      </div>
    </header>
  );
}
