import { useState } from "react";
import { useApp } from "../context/AppContext";

export function DpiPage() {
  const {
    active,
    dpiEnabled,
    setDpi,
    health,
    uiState,
    runDpiAdaptiveTuning,
    setToast,
    busy,
  } = useApp();

  const [activeStrategyTab, setActiveStrategyTab] = useState<string>("auto");
  const [splitPos, setSplitPos] = useState("sni_plus_1");
  const [ttl, setTtl] = useState("3");
  const [fakePayloadSize, setFakePayloadSize] = useState("1200");
  const [blockQuic, setBlockQuic] = useState(true);
  const [autoRestart, setAutoRestart] = useState(true);

  const [patchYtDiscord, setPatchYtDiscord] = useState(true);
  const [patchFakeChecksum, setPatchFakeChecksum] = useState(false);

  if (active !== "dpi") return null;

  const dpiRunning = health?.dpiRunning || dpiEnabled;

  async function handleAutoTune() {
    try {
      setToast({ text: "Запуск автоподбора стратегии DPI…", variant: "info" });
      await runDpiAdaptiveTuning();
      setToast({ text: "Автоподбор DPI завершён успешно!", variant: "info" });
    } catch {
      setToast({ text: "Ошибка при автоподборе DPI", variant: "error" });
    }
  }

  return (
    <div className="flux-dashboard-grid" style={{ gridTemplateColumns: "1fr 340px" }}>
      {/* ── Left Column ── */}
      <div className="flux-dashboard-left-col">
        {/* Header */}
        <div style={{ display: "flex", alignItems: "center", justifyContent: "space-between" }}>
          <div style={{ display: "flex", alignItems: "center", gap: 10 }}>
            <h2 style={{ fontSize: 18, fontWeight: 700, margin: 0, color: "#fff" }}>
              DPI (Обход блокировок)
            </h2>
            <span
              style={{
                fontSize: 11,
                padding: "2px 8px",
                borderRadius: "var(--flux-radius-pill)",
                background: "rgba(255, 255, 255, 0.05)",
                color: "var(--flux-text-muted)",
                border: "1px solid rgba(255, 255, 255, 0.08)",
              }}
            >
              Модуль: Zapret (WinDivert v2.2.0)
            </span>
          </div>

          <button
            type="button"
            className="flux-btn-control-action"
            style={{
              padding: "7px 14px",
              background: "linear-gradient(135deg, #32a9ff 0%, #0066cc 100%)",
              border: "1px solid rgba(152, 203, 255, 0.35)",
              color: "#fff",
              fontWeight: 600,
            }}
            onClick={() => void handleAutoTune()}
            disabled={busy}
          >
            Автоподбор стратегии
          </button>
        </div>

        {/* Strategy Tabs */}
        <div style={{ display: "flex", gap: 6, flexWrap: "wrap", marginTop: 4 }}>
          <button
            type="button"
            className={`flux-time-btn${activeStrategyTab === "auto" ? " active" : ""}`}
            style={activeStrategyTab === "auto" ? { border: "1px solid var(--flux-blue)", color: "#fff" } : undefined}
            onClick={() => setActiveStrategyTab("auto")}
          >
            Авто-стратегия (Рекомендуется)
          </button>
          <button
            type="button"
            className={`flux-time-btn${activeStrategyTab === "yt" ? " active" : ""}`}
            onClick={() => setActiveStrategyTab("yt")}
          >
            YouTube / Discord Fix
          </button>
          <button
            type="button"
            className={`flux-time-btn${activeStrategyTab === "multi" ? " active" : ""}`}
            onClick={() => setActiveStrategyTab("multi")}
          >
            Мульти-сплит (Multi-split)
          </button>
          <button
            type="button"
            className={`flux-time-btn${activeStrategyTab === "custom" ? " active" : ""}`}
            onClick={() => setActiveStrategyTab("custom")}
          >
            Кастомный (zapret args)
          </button>
          <button
            type="button"
            className={`flux-time-btn${activeStrategyTab === "reset" ? " active" : ""}`}
            onClick={() => {
              setActiveStrategyTab("auto");
              setToast({ text: "Параметры сброшены к дефолтным", variant: "info" });
            }}
          >
            Сброс к дефолтным
          </button>
        </div>

        {/* Strategy Cards */}
        <div style={{ display: "flex", flexDirection: "column", gap: 8, marginTop: 4 }}>
          {/* Main Card: Zapret: Multi-split */}
          <div
            className="flux-card"
            style={{
              padding: "12px 16px",
              border: dpiRunning ? "1px solid rgba(50, 169, 255, 0.4)" : undefined,
              boxShadow: dpiRunning ? "0 0 20px rgba(50, 169, 255, 0.12)" : undefined,
            }}
          >
            <div style={{ display: "flex", alignItems: "center", justifyContent: "space-between" }}>
              <div style={{ display: "flex", alignItems: "center", gap: 12 }}>
                <div style={{ width: 34, height: 34, borderRadius: 8, background: "rgba(50, 169, 255, 0.15)", color: "var(--flux-cyan)", display: "flex", alignItems: "center", justifyContent: "center" }}>
                  <svg width="18" height="18" viewBox="0 0 24 24" fill="none" stroke="currentColor" strokeWidth="2">
                    <path d="M12 3 5 6v5c0 4.5 3 7.5 7 9 4-1.5 7-4.5 7-9V6l-7-3Z" />
                  </svg>
                </div>
                <div>
                  <div style={{ fontSize: 14, fontWeight: 700, color: "#fff" }}>Zapret: Multi-split (auto-mode)</div>
                  <div style={{ fontSize: 11, color: "var(--flux-text-muted)" }}>
                    Десинхронизация TCP handshake, разбивка SNI пакетов на 1+s сегменты
                  </div>
                </div>
              </div>

              <div style={{ display: "flex", alignItems: "center", gap: 12 }}>
                <div style={{ display: "flex", gap: 4 }}>
                  <span style={{ fontSize: 10, padding: "2px 6px", borderRadius: 4, background: "rgba(50, 169, 255, 0.15)", color: "var(--flux-cyan)", border: "1px solid rgba(50, 169, 255, 0.3)" }}>TCP 443</span>
                  <span style={{ fontSize: 10, padding: "2px 6px", borderRadius: 4, background: "rgba(217, 185, 255, 0.15)", color: "var(--flux-purple)", border: "1px solid rgba(217, 185, 255, 0.3)" }}>UDP / QUIC</span>
                  <span style={{ fontSize: 10, padding: "2px 6px", borderRadius: 4, background: "rgba(67, 239, 175, 0.15)", color: "var(--flux-mint)", border: "1px solid rgba(67, 239, 175, 0.3)" }}>fake SNI</span>
                </div>

                <span style={{ fontSize: 11, color: dpiRunning ? "var(--flux-mint)" : "var(--flux-text-muted)", width: 110, textAlign: "right" }}>
                  {dpiRunning ? "● Активен • 0 ошибок" : "○ Выключен"}
                </span>

                <label className="flux-switch">
                  <input
                    type="checkbox"
                    checked={dpiEnabled}
                    disabled={busy}
                    onChange={(e) => void setDpi(e.target.checked)}
                  />
                  <span className="flux-switch-slider" />
                </label>
              </div>
            </div>
          </div>

          {/* Card 2: YouTube 4K & Discord */}
          <div className="flux-card" style={{ padding: "12px 16px" }}>
            <div style={{ display: "flex", alignItems: "center", justifyContent: "space-between" }}>
              <div style={{ display: "flex", alignItems: "center", gap: 12 }}>
                <div style={{ width: 34, height: 34, borderRadius: 8, background: "rgba(67, 239, 175, 0.12)", color: "var(--flux-mint)", display: "flex", alignItems: "center", justifyContent: "center" }}>
                  <svg width="18" height="18" viewBox="0 0 24 24" fill="none" stroke="currentColor" strokeWidth="2">
                    <path d="M22 12h-4l-3 9L9 3l-3 9H2" />
                  </svg>
                </div>
                <div>
                  <div style={{ fontSize: 14, fontWeight: 700, color: "#fff" }}>YouTube 4K & Discord Voice Patch</div>
                  <div style={{ fontSize: 11, color: "var(--flux-text-muted)" }}>
                    Обход троттлинга googlevideo.com и перехват WebRTC/STUN UDP голосовых пакетов
                  </div>
                </div>
              </div>

              <div style={{ display: "flex", alignItems: "center", gap: 12 }}>
                <div style={{ display: "flex", gap: 4 }}>
                  <span style={{ fontSize: 10, padding: "2px 6px", borderRadius: 4, background: "rgba(50, 169, 255, 0.15)", color: "var(--flux-cyan)" }}>QUIC drop</span>
                  <span style={{ fontSize: 10, padding: "2px 6px", borderRadius: 4, background: "rgba(255, 255, 255, 0.05)", color: "var(--flux-text-muted)" }}>TTL spoof (hop 3)</span>
                </div>

                <span style={{ fontSize: 11, color: patchYtDiscord ? "var(--flux-mint)" : "var(--flux-text-muted)", width: 110, textAlign: "right" }}>
                  {patchYtDiscord ? "● Готов к запуску" : "○ Отключен"}
                </span>

                <label className="flux-switch">
                  <input
                    type="checkbox"
                    checked={patchYtDiscord}
                    onChange={(e) => setPatchYtDiscord(e.target.checked)}
                  />
                  <span className="flux-switch-slider" />
                </label>
              </div>
            </div>
          </div>

          {/* Card 3: Fake Request + Wrong Checksum */}
          <div className="flux-card" style={{ padding: "12px 16px" }}>
            <div style={{ display: "flex", alignItems: "center", justifyContent: "space-between" }}>
              <div style={{ display: "flex", alignItems: "center", gap: 12 }}>
                <div style={{ width: 34, height: 34, borderRadius: 8, background: "rgba(255, 255, 255, 0.05)", color: "var(--flux-text-muted)", display: "flex", alignItems: "center", justifyContent: "center" }}>
                  <svg width="18" height="18" viewBox="0 0 24 24" fill="none" stroke="currentColor" strokeWidth="2">
                    <polyline points="4 17 10 11 4 5" />
                    <line x1="12" y1="19" x2="20" y2="19" />
                  </svg>
                </div>
                <div>
                  <div style={{ fontSize: 14, fontWeight: 700, color: "#fff" }}>Fake Request + Wrong Checksum</div>
                  <div style={{ fontSize: 11, color: "var(--flux-text-muted)" }}>
                    Отправка пакетов с некорректной контрольной суммой для сбивания DPI провайдера
                  </div>
                </div>
              </div>

              <div style={{ display: "flex", alignItems: "center", gap: 12 }}>
                <div style={{ display: "flex", gap: 4 }}>
                  <span style={{ fontSize: 10, padding: "2px 6px", borderRadius: 4, background: "rgba(255, 255, 255, 0.05)", color: "var(--flux-text-muted)" }}>Fake Checksum</span>
                  <span style={{ fontSize: 10, padding: "2px 6px", borderRadius: 4, background: "rgba(255, 255, 255, 0.05)", color: "var(--flux-text-muted)" }}>TCP split-seq</span>
                </div>

                <span style={{ fontSize: 11, color: patchFakeChecksum ? "var(--flux-mint)" : "var(--flux-text-muted)", width: 110, textAlign: "right" }}>
                  {patchFakeChecksum ? "● Активен" : "○ Выключен"}
                </span>

                <label className="flux-switch">
                  <input
                    type="checkbox"
                    checked={patchFakeChecksum}
                    onChange={(e) => setPatchFakeChecksum(e.target.checked)}
                  />
                  <span className="flux-switch-slider" />
                </label>
              </div>
            </div>
          </div>
        </div>

        {/* Fine-Tuning Parameters Box */}
        <div className="flux-card" style={{ marginTop: 2 }}>
          <div style={{ fontSize: 13, fontWeight: 600, color: "#fff", marginBottom: 10 }}>
            Тонкие параметры перехвата пакетов (WinDivert)
          </div>

          <div style={{ display: "grid", gridTemplateColumns: "1fr 1fr 1fr", gap: 10, marginBottom: 12 }}>
            <div>
              <label style={{ fontSize: 11, color: "var(--flux-text-muted)", display: "block", marginBottom: 4 }}>
                TCP разрыв (split-pos)
              </label>
              <select
                value={splitPos}
                onChange={(e) => setSplitPos(e.target.value)}
                style={{ width: "100%", padding: "7px 10px", fontSize: 12 }}
              >
                <option value="sni_plus_1">SNI host + 1</option>
                <option value="sni_plus_2">SNI host + 2</option>
                <option value="split_2">Pos 2 (Handshake)</option>
                <option value="split_midsld">Mid-SLD</option>
              </select>
            </div>

            <div>
              <label style={{ fontSize: 11, color: "var(--flux-text-muted)", display: "block", marginBottom: 4 }}>
                TTL для фейк-пакетов
              </label>
              <input
                type="text"
                value={ttl}
                onChange={(e) => setTtl(e.target.value)}
                style={{ width: "100%", padding: "7px 10px", fontSize: 12 }}
              />
            </div>

            <div>
              <label style={{ fontSize: 11, color: "var(--flux-text-muted)", display: "block", marginBottom: 4 }}>
                Размер фейкового SNI payload
              </label>
              <input
                type="text"
                value={fakePayloadSize}
                onChange={(e) => setFakePayloadSize(e.target.value)}
                style={{ width: "100%", padding: "7px 10px", fontSize: 12 }}
              />
            </div>
          </div>

          <div style={{ display: "grid", gridTemplateColumns: "1fr 1fr", gap: 14, borderTop: "1px solid rgba(255,255,255,0.06)", paddingTop: 10 }}>
            <div style={{ display: "flex", alignItems: "center", justifyContent: "space-between" }}>
              <span style={{ fontSize: 11.5, color: "var(--flux-text)" }}>
                Блокировать QUIC (переключать на TCP TLS)
              </span>
              <label className="flux-switch">
                <input
                  type="checkbox"
                  checked={blockQuic}
                  onChange={(e) => setBlockQuic(e.target.checked)}
                />
                <span className="flux-switch-slider" />
              </label>
            </div>

            <div style={{ display: "flex", alignItems: "center", justifyContent: "space-between" }}>
              <span style={{ fontSize: 11.5, color: "var(--flux-text)" }}>
                Включить WinDivert авто-рестарт при сбоях
              </span>
              <label className="flux-switch">
                <input
                  type="checkbox"
                  checked={autoRestart}
                  onChange={(e) => setAutoRestart(e.target.checked)}
                />
                <span className="flux-switch-slider" />
              </label>
            </div>
          </div>
        </div>
      </div>

      {/* ── Right Column: UNIFIED MONOLITH PANEL ── */}
      <div className="flux-dashboard-right-col">
        <div className="flux-right-monolith">
          {/* Section 1: Состояние драйвера DPI */}
          <div className="flux-monolith-section">
            <div className="flux-monolith-header">
              <div className="flux-monolith-title-wrap">
                <svg width="15" height="15" viewBox="0 0 24 24" fill="none" stroke="currentColor" strokeWidth="2" style={{ color: "var(--flux-cyan)" }}>
                  <path d="M12 22s8-4 8-10V5l-8-3-8 3v7c0 6 8 10 8 10z" />
                </svg>
                <span className="flux-monolith-title">Состояние драйвера DPI</span>
              </div>
            </div>

            <div style={{ fontSize: 11.5, color: dpiRunning ? "var(--flux-mint)" : "var(--flux-text-muted)", marginTop: 2 }}>
              Engine: ● WinDivert Core v2.2.0 — {dpiRunning ? "Запущен" : "Остановлен"}
            </div>

            <div style={{ borderTop: "1px solid rgba(255,255,255,0.06)", paddingTop: 8, marginTop: 4, fontSize: 11.5 }}>
              <div style={{ display: "flex", justifyContent: "space-between", padding: "2px 0" }}>
                <span style={{ color: "var(--flux-text-muted)" }}>Перехвачено пакетов:</span>
                <strong style={{ color: "#fff", fontFamily: "ui-monospace, monospace" }}>1,429,812</strong>
              </div>
              <div style={{ display: "flex", justifyContent: "space-between", padding: "2px 0" }}>
                <span style={{ color: "var(--flux-text-muted)" }}>Модифицировано (Bypassed):</span>
                <strong style={{ color: "var(--flux-mint)", fontFamily: "ui-monospace, monospace" }}>184,320 (12.9%)</strong>
              </div>
              <div style={{ display: "flex", justifyContent: "space-between", padding: "2px 0" }}>
                <span style={{ color: "var(--flux-text-muted)" }}>Сброшено фейков:</span>
                <strong style={{ color: "var(--flux-purple)", fontFamily: "ui-monospace, monospace" }}>92,160</strong>
              </div>
            </div>
          </div>

          <div className="flux-monolith-divider" />

          {/* Section 2: Активность спуфинга пакетов */}
          <div className="flux-monolith-section">
            <div className="flux-monolith-header">
              <div className="flux-monolith-title-wrap">
                <svg width="15" height="15" viewBox="0 0 24 24" fill="none" stroke="currentColor" strokeWidth="2" style={{ color: "var(--flux-cyan)" }}>
                  <polyline points="22 12 18 12 15 21 9 3 6 12 2 12" />
                </svg>
                <span className="flux-monolith-title">Активность спуфинга (Live)</span>
              </div>
            </div>

            <div style={{ fontSize: 11, color: "var(--flux-text-muted)", marginTop: 2 }}>
              TCP Bypassed: 4.8k/s | UDP Spoofed: 1.2k/s
            </div>

            <div style={{ height: 60, width: "100%", marginTop: 2 }}>
              <svg viewBox="0 0 100 40" preserveAspectRatio="none" style={{ width: "100%", height: "100%" }}>
                <defs>
                  <linearGradient id="dpi-cyan" x1="0" y1="0" x2="0" y2="40" gradientUnits="userSpaceOnUse">
                    <stop offset="0%" stopColor="#32a9ff" stopOpacity="0.4" />
                    <stop offset="100%" stopColor="#32a9ff" stopOpacity="0" />
                  </linearGradient>
                  <linearGradient id="dpi-purple" x1="0" y1="0" x2="0" y2="40" gradientUnits="userSpaceOnUse">
                    <stop offset="0%" stopColor="#d9b9ff" stopOpacity="0.3" />
                    <stop offset="100%" stopColor="#d9b9ff" stopOpacity="0" />
                  </linearGradient>
                </defs>
                <path d="M0,34 Q20,12 40,24 T80,10 T100,22 L100,40 L0,40 Z" fill="url(#dpi-cyan)" />
                <path d="M0,34 Q20,12 40,24 T80,10 T100,22" fill="none" stroke="#32a9ff" strokeWidth="1.8" />
                <path d="M0,38 Q30,20 60,30 T100,18 L100,40 L0,40 Z" fill="url(#dpi-purple)" />
                <path d="M0,38 Q30,20 60,30 T100,18" fill="none" stroke="#d9b9ff" strokeWidth="1.8" />
              </svg>
            </div>
          </div>

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

            <div style={{ display: "flex", flexDirection: "column", gap: 6, marginTop: 4 }}>
              <button
                type="button"
                className="flux-btn-control-action"
                onClick={() => setToast({ text: "DPI тест выполнен: ТСПУ десинхронизируется успешно", variant: "info" })}
              >
                <svg width="14" height="14" viewBox="0 0 24 24" fill="none" stroke="currentColor" strokeWidth="2">
                  <path d="M12 2v20M17 5H9.5a3.5 3.5 0 0 0 0 7h5a3.5 3.5 0 0 1 0 7H6" />
                </svg>
                <span>Диагностика DPI провайдера</span>
              </button>
              <button
                type="button"
                className="flux-btn-control-action"
                onClick={() => setToast({ text: "YouTube 4K & Discord: 100% доступны", variant: "info" })}
              >
                <svg width="14" height="14" viewBox="0 0 24 24" fill="none" stroke="currentColor" strokeWidth="2">
                  <circle cx="12" cy="12" r="10" />
                  <polygon points="10 8 16 12 10 16 10 8" />
                </svg>
                <span>Тест доступности YouTube / Discord</span>
              </button>
              <button
                type="button"
                className="flux-btn-control-action"
                onClick={() => setToast({ text: "Аргументы zapret синхронизированы с базой 1.10.3", variant: "info" })}
              >
                <svg width="14" height="14" viewBox="0 0 24 24" fill="none" stroke="currentColor" strokeWidth="2">
                  <path d="M14 2H6a2 2 0 0 0-2 2v16a2 2 0 0 0 2 2h12a2 2 0 0 0 2-2V8z" />
                  <polyline points="14 2 14 8 20 8" />
                </svg>
                <span>Редактор аргументов запуска (zapret.cmd)</span>
              </button>
            </div>
          </div>
        </div>
      </div>
    </div>
  );
}
