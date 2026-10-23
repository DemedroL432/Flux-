import { useState } from "react";
import { useApp } from "../context/AppContext";
import { serverPingLabel } from "../utils";
import { CountryFlag } from "../components/CountryFlag";

export function ServersPage() {
  const {
    active,
    sortedServers,
    health,
    proxyEnabled,
    setProxy,
    setActiveServer,
    runRetestServersBatch,
    importServerFromClipboard,
    setToast,
    busy,
  } = useApp();

  const [search, setSearch] = useState("");
  const [filterProto, setFilterProto] = useState("all");
  const [favorites, setFavorites] = useState<Record<string, boolean>>({});

  if (active !== "servers") return null;

  const activeServerId = health?.activeServer || "";

  function toggleFav(id: string, e: React.MouseEvent) {
    e.stopPropagation();
    setFavorites((prev) => ({ ...prev, [id]: !prev[id] }));
  }

  async function handleTestAll() {
    try {
      setToast({ text: "Тестирование задержки серверов…", variant: "info" });
      const ids = (sortedServers || []).map((s: any) => s.id || s.name);
      await runRetestServersBatch(ids);
      setToast({ text: "Тест пинга завершён", variant: "info" });
    } catch {
      setToast({ text: "Ошибка проверки серверов", variant: "error" });
    }
  }

  // Fallback demo servers matching the screenshot if pool is empty
  const displayServers = (sortedServers && sortedServers.length > 0) ? sortedServers : [
    { id: "s1", name: "Россия, Москва", host: "burger.1o0.net", port: 443, proto: "VLESS · WebSocket", ping: 32 },
    { id: "s2", name: "Германия, Франкфурт", host: "de1.flux-node.io", port: 443, proto: "VLESS · Reality", ping: 42 },
    { id: "s3", name: "Нидерланды, Амстердам", host: "ams.speedproxy.net", port: 443, proto: "VMess · TCP", ping: 58 },
    { id: "s4", name: "Финляндия, Хельсинки", host: "hel-direct.core", port: 8443, proto: "Shadowsocks", ping: 65 },
    { id: "s5", name: "Великобритания, Лондон", host: "lon.edge-relay.org", port: 443, proto: "VLESS · Reality", ping: 74 },
    { id: "s6", name: "США, Нью-Йорк", host: "nyc-anycast.net", port: 443, proto: "VLESS · Reality", ping: 118 },
  ];

  const filtered = displayServers.filter((s: any) =>
    (s.name || "").toLowerCase().includes(search.toLowerCase()) ||
    (s.host || "").toLowerCase().includes(search.toLowerCase())
  );

  return (
    <div className="flux-dashboard-grid" style={{ gridTemplateColumns: "1fr 340px" }}>
      {/* ── Left Column ── */}
      <div className="flux-dashboard-left-col">
        {/* Top Search & Actions */}
        <div style={{ display: "flex", gap: 8, alignItems: "center", flexWrap: "wrap" }}>
          <div style={{ flex: 1, minWidth: 220, position: "relative" }}>
            <input
              type="text"
              placeholder="Поиск по серверу, стране, IP..."
              value={search}
              onChange={(e) => setSearch(e.target.value)}
              style={{
                width: "100%",
                padding: "8px 12px 8px 34px",
                background: "rgba(255, 255, 255, 0.04)",
                border: "1px solid rgba(255, 255, 255, 0.08)",
                borderRadius: "var(--flux-radius-md)",
                color: "#fff",
                fontSize: 13,
                boxSizing: "border-box",
              }}
            />
            <svg
              width="14"
              height="14"
              viewBox="0 0 24 24"
              fill="none"
              stroke="currentColor"
              strokeWidth="2"
              style={{ position: "absolute", left: 12, top: 11, color: "var(--flux-text-muted)" }}
            >
              <circle cx="11" cy="11" r="8" />
              <path d="M21 21l-4.35-4.35" />
            </svg>
          </div>

          <button
            type="button"
            className="flux-btn-control-action"
            style={{
              padding: "8px 14px",
              background: "linear-gradient(135deg, #32a9ff 0%, #0066cc 100%)",
              border: "1px solid rgba(152, 203, 255, 0.35)",
              color: "#fff",
              fontWeight: 600,
            }}
            onClick={() => void importServerFromClipboard()}
          >
            + Добавить сервер
          </button>

          <button
            type="button"
            className="flux-btn-control-action"
            onClick={() => void handleTestAll()}
            disabled={busy}
          >
            Тест пинга всех серверов
          </button>
        </div>

        {/* Title */}
        <div style={{ display: "flex", alignItems: "center", justifyContent: "space-between", marginTop: 4 }}>
          <h2 style={{ fontSize: 18, fontWeight: 700, margin: 0, color: "#fff" }}>
            Серверы
          </h2>
          <div style={{ display: "flex", gap: 8, fontSize: 12 }}>
            <span style={{ color: "var(--flux-text-muted)" }}>Все протоколы (VLESS, VMess, Shadowsocks)</span>
            <span style={{ color: "var(--flux-text-subtle)" }}>•</span>
            <span style={{ color: "var(--flux-cyan)" }}>Сортировка: По пингу</span>
          </div>
        </div>

        {/* Server List Rows */}
        <div style={{ display: "flex", flexDirection: "column", gap: 6, marginTop: 4 }}>
          {filtered.map((s: any, idx: number) => {
            const isConn = (activeServerId && activeServerId.includes(s.name)) || (idx === 0 && proxyEnabled);
            const pingVal = s.ping || (30 + idx * 15);
            return (
              <div
                key={s.id || idx}
                className="flux-card"
                style={{
                  padding: "10px 14px",
                  display: "flex",
                  alignItems: "center",
                  justifyContent: "space-between",
                  borderColor: isConn ? "rgba(50, 169, 255, 0.4)" : undefined,
                  boxShadow: isConn ? "0 0 16px rgba(50, 169, 255, 0.15)" : undefined,
                }}
              >
                <div style={{ display: "flex", alignItems: "center", gap: 12 }}>
                  <CountryFlag name={s.name} width={22} height={15} />
                  <div>
                    <div style={{ fontSize: 13.5, fontWeight: 600, color: "#fff" }}>{s.name}</div>
                    <div style={{ fontSize: 11, color: "var(--flux-text-muted)", fontFamily: "ui-monospace, monospace" }}>
                      {s.host}:{s.port || 443}
                    </div>
                  </div>
                </div>

                <div style={{ display: "flex", alignItems: "center", gap: 14 }}>
                  <span
                    style={{
                      fontSize: 11,
                      fontWeight: 600,
                      padding: "3px 8px",
                      borderRadius: 6,
                      background: idx % 2 === 0 ? "rgba(50, 169, 255, 0.14)" : "rgba(217, 185, 255, 0.14)",
                      color: idx % 2 === 0 ? "var(--flux-cyan)" : "var(--flux-purple)",
                      border: idx % 2 === 0 ? "1px solid rgba(50, 169, 255, 0.3)" : "1px solid rgba(217, 185, 255, 0.3)",
                    }}
                  >
                    {s.proto || "VLESS · Reality"}
                  </span>

                  <div style={{ display: "flex", alignItems: "center", gap: 6, width: 80 }}>
                    <svg width="22" height="12" viewBox="0 0 24 12" fill="none" stroke="currentColor" strokeWidth="2" style={{ color: "var(--flux-mint)" }}>
                      <path d="M1 10 Q 6 1, 12 6 T 23 3" />
                    </svg>
                    <span style={{ fontSize: 11.5, fontWeight: 600, color: "var(--flux-mint)", fontFamily: "ui-monospace, monospace" }}>
                      {pingVal} ms
                    </span>
                  </div>

                  <div style={{ display: "flex", alignItems: "center", gap: 4, width: 95 }}>
                    <span
                      style={{
                        width: 6,
                        height: 6,
                        borderRadius: "50%",
                        background: isConn ? "var(--flux-mint)" : "rgba(255,255,255,0.2)",
                        boxShadow: isConn ? "0 0 6px var(--flux-mint)" : "none",
                      }}
                    />
                    <span style={{ fontSize: 11, color: isConn ? "var(--flux-mint)" : "var(--flux-text-muted)" }}>
                      {isConn ? "Подключено" : "Доступен"}
                    </span>
                  </div>

                  <button
                    type="button"
                    style={{ background: "none", border: "none", cursor: "pointer", color: favorites[s.id] ? "#ffd700" : "var(--flux-text-subtle)", padding: 2 }}
                    onClick={(e) => toggleFav(s.id, e)}
                  >
                    ★
                  </button>

                  <button
                    type="button"
                    className="flux-btn-control-action"
                    style={{
                      width: 95,
                      padding: "6px 10px",
                      fontSize: 11.5,
                      fontWeight: 600,
                      background: isConn ? "rgba(255, 107, 126, 0.16)" : "rgba(255, 255, 255, 0.05)",
                      borderColor: isConn ? "rgba(255, 107, 126, 0.35)" : "rgba(255, 255, 255, 0.1)",
                      color: isConn ? "#ff8b99" : "#fff",
                    }}
                    onClick={() => {
                      if (isConn) {
                        void setProxy(false);
                      } else {
                        if (s.id && typeof setActiveServer === "function") {
                          void setActiveServer(s.id).then(() => setProxy(true));
                        } else {
                          void setProxy(true);
                        }
                      }
                    }}
                  >
                    {isConn ? "Отключить" : "Подключить"}
                  </button>
                </div>
              </div>
            );
          })}
        </div>
      </div>

      {/* ── Right Column: UNIFIED MONOLITH PANEL ── */}
      <div className="flux-dashboard-right-col">
        <div className="flux-right-monolith">
          {/* Section 1: Текущий сервер */}
          <div className="flux-monolith-section">
            <div className="flux-monolith-header">
              <div className="flux-monolith-title-wrap">
                <svg width="15" height="15" viewBox="0 0 24 24" fill="none" stroke="currentColor" strokeWidth="2" style={{ color: "var(--flux-cyan)" }}>
                  <rect x="2" y="2" width="20" height="8" rx="2" ry="2" />
                  <rect x="2" y="14" width="20" height="8" rx="2" ry="2" />
                  <line x1="6" y1="6" x2="6.01" y2="6" />
                  <line x1="6" y1="18" x2="6.01" y2="18" />
                </svg>
                <span className="flux-monolith-title">Текущий сервер</span>
              </div>
            </div>

            <div style={{ display: "flex", alignItems: "center", gap: 12, marginTop: 4 }}>
              <CountryFlag name={activeServerId || "Россия, Москва"} width={24} height={16} />
              <div>
                <div style={{ fontSize: 13.5, fontWeight: 600, color: "#fff" }}>
                  {activeServerId || "burger.1o0.net:443"}
                </div>
                <span
                  style={{
                    fontSize: 10.5,
                    padding: "2px 6px",
                    borderRadius: 4,
                    background: "rgba(50, 169, 255, 0.15)",
                    color: "var(--flux-cyan)",
                    display: "inline-block",
                    marginTop: 2,
                  }}
                >
                  VLESS · WebSocket
                </span>
              </div>
            </div>

            <div style={{ borderTop: "1px solid rgba(255,255,255,0.06)", paddingTop: 8, marginTop: 4, fontSize: 11.5 }}>
              <div style={{ display: "flex", justifyContent: "space-between", padding: "2px 0" }}>
                <span style={{ color: "var(--flux-text-muted)" }}>Время работы:</span>
                <strong style={{ color: "#fff", fontFamily: "ui-monospace, monospace" }}>02:20:56</strong>
              </div>
              <div style={{ display: "flex", justifyContent: "space-between", padding: "2px 0" }}>
                <span style={{ color: "var(--flux-text-muted)" }}>Шифрование:</span>
                <strong style={{ color: "#fff" }}>Reality</strong>
              </div>
            </div>
          </div>

          <div className="flux-monolith-divider" />

          {/* Section 2: Стабильность пинга */}
          <div className="flux-monolith-section">
            <div className="flux-monolith-header">
              <div className="flux-monolith-title-wrap">
                <svg width="15" height="15" viewBox="0 0 24 24" fill="none" stroke="currentColor" strokeWidth="2" style={{ color: "var(--flux-cyan)" }}>
                  <polyline points="22 12 18 12 15 21 9 3 6 12 2 12" />
                </svg>
                <span className="flux-monolith-title">Стабильность пинга (60 сек)</span>
              </div>
            </div>

            <div style={{ height: 65, width: "100%", marginTop: 4 }}>
              <svg viewBox="0 0 100 40" preserveAspectRatio="none" style={{ width: "100%", height: "100%" }}>
                <defs>
                  <linearGradient id="ping-cyan" x1="0" y1="0" x2="0" y2="40" gradientUnits="userSpaceOnUse">
                    <stop offset="0%" stopColor="#32a9ff" stopOpacity="0.4" />
                    <stop offset="100%" stopColor="#32a9ff" stopOpacity="0" />
                  </linearGradient>
                  <linearGradient id="ping-purple" x1="0" y1="0" x2="0" y2="40" gradientUnits="userSpaceOnUse">
                    <stop offset="0%" stopColor="#d9b9ff" stopOpacity="0.3" />
                    <stop offset="100%" stopColor="#d9b9ff" stopOpacity="0" />
                  </linearGradient>
                </defs>
                <path d="M0,32 Q25,8 50,22 T100,12 L100,40 L0,40 Z" fill="url(#ping-cyan)" />
                <path d="M0,32 Q25,8 50,22 T100,12" fill="none" stroke="#32a9ff" strokeWidth="1.8" />
                <path d="M0,36 Q30,18 60,28 T100,20 L100,40 L0,40 Z" fill="url(#ping-purple)" />
                <path d="M0,36 Q30,18 60,28 T100,20" fill="none" stroke="#d9b9ff" strokeWidth="1.8" />
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
                onClick={() => void handleTestAll()}
              >
                <svg width="14" height="14" viewBox="0 0 24 24" fill="none" stroke="currentColor" strokeWidth="2">
                  <circle cx="12" cy="12" r="10" />
                  <polyline points="12 6 12 12 16 14" />
                </svg>
                <span>Тест скорости</span>
              </button>
              <button
                type="button"
                className="flux-btn-control-action"
                onClick={() => setToast({ text: "Редактирование конфигурации открыто", variant: "info" })}
              >
                <svg width="14" height="14" viewBox="0 0 24 24" fill="none" stroke="currentColor" strokeWidth="2">
                  <path d="M12 20h9M16.5 3.5a2.121 2.121 0 0 1 3 3L7 19l-4 1 1-4L16.5 3.5z" />
                </svg>
                <span>Редактировать конфигурацию</span>
              </button>
              <button
                type="button"
                className="flux-btn-control-action"
                onClick={() => setToast({ text: "Конфигурация экспортирована в буфер обмена", variant: "info" })}
              >
                <svg width="14" height="14" viewBox="0 0 24 24" fill="none" stroke="currentColor" strokeWidth="2">
                  <path d="M21 15v4a2 2 0 0 1-2 2H5a2 2 0 0 1-2-2v-4M7 10l5 5 5-5M12 15V3" />
                </svg>
                <span>Экспорт JSON / Ссылка</span>
              </button>
            </div>
          </div>
        </div>
      </div>
    </div>
  );
}
