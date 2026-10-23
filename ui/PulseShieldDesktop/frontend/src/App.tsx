import { ConceptPages } from "./pages/ConceptPages";
import './App.css';
import './GraphiteTheme.css';
import './ConceptTheme.css';
import { AppProvider, useApp } from "./context/AppContext";
import { DesktopSidebar } from "./components/DesktopSidebar";
import { DesktopTopBar } from "./components/DesktopTopBar";
import { GuidePage } from "./pages/GuidePage";
import { LogsPage } from "./pages/LogsPage";
import { OpenLogsFolder } from "../wailsjs/go/main/App";

export function AppShell() {
  const {
    active, setActive,
    busy,
    toast, setToast,
    globalBusy,
  } = useApp();

  return (
    <div className="flux-desktop-layout" aria-busy={busy}>
      <div className="flux-desktop-bg" />
      <div className="flux-desktop-bg-overlay" />

      {/* Desktop Left Sidebar */}
      <DesktopSidebar />

      {/* Main Desktop Area */}
      <div className="flux-main-area">
        {/* Top Window Bar with section title & window controls */}
        <DesktopTopBar />

        {/* Scrollable Content Container */}
        <main className={`flux-content-scroll${active === "dashboard" ? " flux-content-dashboard" : ""}`}>
          {/* Brand Hero Banner with corrected text */}


          {/* Busy and Toast notifications */}
          {globalBusy ? (
            <div className="app-inline-busy" role="status" aria-live="polite" style={{ marginBottom: 12 }}>
              <span className="app-inline-busy-spin" aria-hidden />
              <span className="app-inline-busy-text">{globalBusy}</span>
            </div>
          ) : null}

          {toast ? (
            <div
              className={`toast-banner${toast.variant === "error" ? " toast-banner--error" : ""}`}
              role="status"
              aria-live={toast.variant === "error" ? "assertive" : "polite"}
              style={{ marginBottom: 12 }}
            >
              <span className="toast-banner-text">{toast.text}</span>
              {toast.openLogs ? (
                <button type="button" className="btn btn-soft toast-banner-logs" onClick={() => void OpenLogsFolder().catch(e => setToast({text: String(e), variant: "error"}))}>
                  Логи
                </button>
              ) : null}
              <button type="button" className="toast-banner-close" aria-label="Закрыть" onClick={() => setToast(null)}>×</button>
            </div>
          ) : null}

          {/* Pages */}





          {(active === "diagnostics" || active === "logs") && <div className="concept-actions" style={{marginBottom:16}}><button className="concept-button" onClick={() => setActive("diagnostics")}>Проверки</button><button className="concept-button" onClick={() => setActive("logs")}>Журнал</button></div>}
          <ConceptPages />
          <LogsPage />

          <GuidePage />
        </main>
      </div>
    </div>
  );
}

function App() {
  return (
    <AppProvider>
      <AppShell />
    </AppProvider>
  );
}

export default App;
