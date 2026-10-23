import { useApp } from "../context/AppContext";
import { NavKey } from "../types";
import { playSoftChime } from "../utils";

function NavIcon({ kind }: { kind: string }) {
  const c = { width: 18, height: 18, viewBox: "0 0 24 24", fill: "none", stroke: "currentColor", strokeWidth: 1.8, strokeLinecap: "round" as const, strokeLinejoin: "round" as const };
  switch (kind) {
    case "dashboard":
      return (
        <svg {...c}>
          <path d="M3 10.5 12 3l9 7.5" />
          <path d="M5 9.5V20h14V9.5" />
        </svg>
      );
    case "servers":
      return (
        <svg {...c}>
          <rect x="3" y="4" width="18" height="6" rx="2" />
          <rect x="3" y="14" width="18" height="6" rx="2" />
          <circle cx="7" cy="7" r="1" fill="currentColor" stroke="none" />
          <circle cx="7" cy="17" r="1" fill="currentColor" stroke="none" />
        </svg>
      );
    case "network":
      return (
        <svg {...c}>
          <circle cx="12" cy="5" r="2.5" />
          <circle cx="5" cy="19" r="2.5" />
          <circle cx="19" cy="19" r="2.5" />
          <path d="M12 7.5v4M12 11.5L6.5 16.5M12 11.5l5.5 5" />
        </svg>
      );
    case "dpi":
      return (
        <svg {...c}>
          <path d="M12 3 5 6v5c0 4.5 3 7.5 7 9 4-1.5 7-4.5 7-9V6l-7-3Z" />
        </svg>
      );
    case "diagnostics":
      return (
        <svg {...c}>
          <rect x="2" y="3" width="20" height="14" rx="2" />
          <path d="M6 10h2l2-3 3 6 2-3h3" />
          <path d="M8 21h8M12 17v4" />
        </svg>
      );
    case "logs":
      return (
        <svg {...c}>
          <rect x="3" y="4" width="18" height="16" rx="2" />
          <path d="M7 8h10M7 12h7M7 16h5" />
        </svg>
      );
    case "settings":
      return (
        <svg {...c}>
          <circle cx="12" cy="12" r="3" />
          <path d="M19.4 15a1.65 1.65 0 0 0 .33 1.82l.06.06a2 2 0 0 1-2.83 2.83l-.06-.06a1.65 1.65 0 0 0-1.82-.33 1.65 1.65 0 0 0-1 1.51V21a2 2 0 0 1-4 0v-.09A1.65 1.65 0 0 0 9 19.4a1.65 1.65 0 0 0-1.82.33l-.06.06a2 2 0 0 1-2.83-2.83l.06-.06a1.65 1.65 0 0 0 .33-1.82 1.65 1.65 0 0 0-1.51-1H3a2 2 0 0 1 0-4h.09A1.65 1.65 0 0 0 4.6 9a1.65 1.65 0 0 0-.33-1.82l-.06-.06a2 2 0 0 1 2.83-2.83l.06.06a1.65 1.65 0 0 0 1.82.33H9a1.65 1.65 0 0 0 1-1.51V3a2 2 0 0 1 4 0v.09a1.65 1.65 0 0 0 1 1.51 1.65 1.65 0 0 0 1.82-.33l.06-.06a2 2 0 0 1 2.83 2.83l-.06.06a1.65 1.65 0 0 0-.33 1.82V9a1.65 1.65 0 0 0 1.51 1H21a2 2 0 0 1 0 4h-.09a1.65 1.65 0 0 0-1.51 1Z" />
        </svg>
      );
    case "guide":
      return (
        <svg {...c}>
          <circle cx="12" cy="12" r="9" />
          <path d="M9.09 9a3 3 0 0 1 5.83 1c0 2-3 3-3 3M12 17h.01" />
        </svg>
      );
    default:
      return (
        <svg {...c}>
          <circle cx="12" cy="12" r="8" />
        </svg>
      );
  }
}

export function DesktopSidebar() {
  const { active, setActive, appVersion, health, uiState } = useApp();

  function nav(key: NavKey) {
    playSoftChime("nav");
    setActive(key);
  }

  const dpiWarn = !!(health?.dpiEnabled && health?.dpiRunning && health?.dpiConnectivityWarn);
  const licenseType = uiState?.licenseTier || "free";
  const isLicensed = !!uiState?.licenseActive;

  return (
    <aside className="flux-sidebar" aria-label="Боковая панель навигации">
      <div className="flux-sidebar-brand" style={{ ["--wails-draggable" as any]: "drag" }}>
        <div className="flux-sidebar-logo-icon">
          <span>F</span>
        </div>
        <div className="flux-sidebar-brand-info">
          <span className="flux-sidebar-brand-name">Flux</span>
          <span className="flux-sidebar-brand-version">v{appVersion || "0.7.4"}</span>
        </div>
      </div>

      <nav className="flux-sidebar-nav" aria-label="Основное меню">
        <button
          type="button"
          className={`flux-nav-item${active === "dashboard" ? " active" : ""}`}
          onClick={() => nav("dashboard")}
        >
          <span className="flux-nav-icon"><NavIcon kind="dashboard" /></span>
          <span>Обзор</span>
        </button>

        <button
          type="button"
          className={`flux-nav-item${active === "servers" ? " active" : ""}`}
          onClick={() => nav("servers")}
        >
          <span className="flux-nav-icon"><NavIcon kind="servers" /></span>
          <span>Серверы</span>
        </button>

        <button
          type="button"
          className={`flux-nav-item${active === "network" ? " active" : ""}`}
          onClick={() => nav("network")}
        >
          <span className="flux-nav-icon"><NavIcon kind="network" /></span>
          <span>Маршрутизация</span>
        </button>

        <button
          type="button"
          className={`flux-nav-item${active === "dpi" ? " active" : ""}${dpiWarn ? " warn" : ""}`}
          onClick={() => nav("dpi")}
        >
          <span className="flux-nav-icon"><NavIcon kind="dpi" /></span>
          <span>Обход DPI</span>
        </button>

        <button
          type="button"
          className={`flux-nav-item${(active === "diagnostics" || active === "logs") ? " active" : ""}`}
          onClick={() => nav("diagnostics")}
        >
          <span className="flux-nav-icon"><NavIcon kind="diagnostics" /></span>
          <span>Диагностика</span>
        </button>

      </nav>

      <div className="flux-sidebar-footer">
        <button
          type="button"
          className={`flux-nav-item${active === "settings" ? " active" : ""}`}
          onClick={() => nav("settings")}
        >
          <span className="flux-nav-icon"><NavIcon kind="settings" /></span>
          <span>Настройки</span>
        </button>

        <button
          type="button"
          className={`flux-nav-item${active === "guide" ? " active" : ""}`}
          onClick={() => nav("guide")}
        >
          <span className="flux-nav-icon"><NavIcon kind="guide" /></span>
          <span>Справка</span>
        </button>

        <div
          className="flux-license-pill"
          role="button"
          tabIndex={0}
          onKeyDown={e => { if(e.key === "Enter" || e.key === " ") { e.preventDefault(); nav("settings"); } }}
          onClick={() => nav("settings")}
          title="Управление лицензией"
        >
          <div className="flux-license-info">
            <div style={{ display: "flex", alignItems: "center" }}>
              <span className="flux-license-dot" style={{ background: isLicensed ? "var(--flux-mint)" : "var(--flux-orange)" }} />
              <span className="flux-license-title">Лицензия: {licenseType}</span>
            </div>
            <span className="flux-license-sub">
              {isLicensed ? `до ${new Date(uiState.licenseExpiresUnix * 1000).toLocaleDateString("ru-RU")}` : "Не активирована"}
            </span>
          </div>
          <svg width="12" height="12" viewBox="0 0 24 24" fill="none" stroke="currentColor" strokeWidth="2" style={{ color: "var(--flux-text-muted)" }}>
            <path d="M9 18l6-6-6-6" />
          </svg>
        </div>
      </div>
    </aside>
  );
}
