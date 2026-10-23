import { useApp } from "../context/AppContext";

export function DesktopStatusBar() {
  const { health, uiState } = useApp();

  const coreRunning = health?.proxyLocalOK || health?.coreRunning;
  const dpiRunning = health?.dpiRunning;
  const dnsRunning = health?.dnsRunning ?? true;
  const proxyPort = uiState?.proxyPort || "8080";
  const dnsPreset = uiState?.dnsPreset || "Google DoH";

  return (
    <footer className="flux-bottom-bar" aria-label="Строка состояния служб">
      <div className="flux-bottom-bar-left">
        <span>Состояние ядра:</span>
        <div style={{ display: "flex", alignItems: "center", gap: 5 }}>
          <span
            className={coreRunning ? "flux-dot-active" : "flux-dot-inactive"}
            style={{ background: coreRunning ? "var(--flux-mint)" : "var(--flux-orange)" }}
          />
          <span style={{ color: coreRunning ? "#fff" : "var(--flux-text-muted)" }}>
            sing-box Core {coreRunning ? "активен" : "остановлен"}
          </span>
        </div>

        <span>|</span>

        <div style={{ display: "flex", alignItems: "center", gap: 5 }}>
          <span
            className={dpiRunning ? "flux-dot-active" : "flux-dot-inactive"}
            style={{ background: dpiRunning ? "var(--flux-mint)" : "var(--flux-text-muted)" }}
          />
          <span style={{ color: dpiRunning ? "#fff" : "var(--flux-text-muted)" }}>
            WinDivert {dpiRunning ? "работает" : "выключен"}
          </span>
        </div>

        <span>|</span>

        <div style={{ display: "flex", alignItems: "center", gap: 5 }}>
          <span
            className={dnsRunning ? "flux-dot-active" : "flux-dot-inactive"}
            style={{ background: dnsRunning ? "var(--flux-mint)" : "var(--flux-orange)" }}
          />
          <span>DNS: {dnsPreset}</span>
        </div>
      </div>

      <div className="flux-bottom-bar-right">
        <span>Порт: {proxyPort}</span>
      </div>
    </footer>
  );
}
