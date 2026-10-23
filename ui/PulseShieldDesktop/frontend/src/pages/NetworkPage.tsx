import { useState } from "react";
import { useApp } from "../context/AppContext";

type RuleItem = {
  id: string;
  name: string;
  condition: string;
  domains: string[];
  action: "vpn_dpi" | "vpn" | "direct" | "reject";
  actionLabel: string;
  traffic: string;
  priority: number;
  enabled: boolean;
  arrowColor: string;
  badgeStyle: "vpn_dpi" | "vpn" | "direct" | "bypass";
};

const INITIAL_RULES: RuleItem[] = [
  {
    id: "rule_01",
    name: "Заблокированные ресурсы (РКН)",
    condition: "geosite:category-ru-blocked",
    domains: ["geosite:category-ru-blocked"],
    action: "vpn",
    actionLabel: "Через VPN (VLESS)",
    traffic: "3.4 GB",
    priority: 1,
    enabled: true,
    arrowColor: "#43efaf",
    badgeStyle: "vpn",
  },
  {
    id: "rule_02",
    name: "YouTube, Discord, Instagram",
    condition: "domain:youtube.com, discord.com...",
    domains: [
      "domain:youtube.com",
      "domain:googlevideo.com",
      "domain:discord.com",
      "domain:discordapp.com",
      "domain:instagram.com",
      "domain:cdninstagram.com",
    ],
    action: "vpn_dpi",
    actionLabel: "Через VPN + DPI",
    traffic: "14.2 GB",
    priority: 2,
    enabled: true,
    arrowColor: "#32a9ff",
    badgeStyle: "vpn_dpi",
  },
  {
    id: "rule_03",
    name: "Российские сервисы (Банки, Госуслуги)",
    condition: "geoip:ru, geosite:ru",
    domains: ["geoip:ru", "geosite:ru"],
    action: "direct",
    actionLabel: "Прямое подключение (Direct)",
    traffic: "8.1 GB",
    priority: 3,
    enabled: true,
    arrowColor: "#d9b9ff",
    badgeStyle: "direct",
  },
  {
    id: "rule_04",
    name: "Torrent / P2P трафик",
    condition: "protocol:bittorrent",
    domains: ["protocol:bittorrent"],
    action: "direct",
    actionLabel: "Прямое соединение (Bypass)",
    traffic: "21.6 GB",
    priority: 4,
    enabled: true,
    arrowColor: "#32a9ff",
    badgeStyle: "bypass",
  },
  {
    id: "rule_05",
    name: "Локальная сеть (LAN)",
    condition: "geoip:private",
    domains: ["geoip:private"],
    action: "direct",
    actionLabel: "Прямое подключение",
    traffic: "512 MB",
    priority: 5,
    enabled: true,
    arrowColor: "#89919d",
    badgeStyle: "direct",
  },
  {
    id: "rule_06",
    name: "Остальной внешний трафик",
    condition: "default / fallback",
    domains: ["default"],
    action: "direct",
    actionLabel: "Прямое подключение",
    traffic: "1.2 GB",
    priority: 99,
    enabled: true,
    arrowColor: "#89919d",
    badgeStyle: "direct",
  },
];

export function NetworkPage() {
  const { active, uiState, setRouting, exportRouting, importRouting, setToast } = useApp();

  const [activeTab, setActiveTab] = useState<"all" | "smart" | "whitelist" | "custom" | "import">("smart");
  const [search, setSearch] = useState("");
  const [rules, setRules] = useState<RuleItem[]>(INITIAL_RULES);
  const [editingRule, setEditingRule] = useState<RuleItem | null>(null);

  if (active !== "network") return null;

  const currentMode = uiState?.routingProfile || "bypass-russia";

  function toggleRule(id: string, e: React.MouseEvent) {
    e.stopPropagation();
    setRules((prev) =>
      prev.map((r) => (r.id === id ? { ...r, enabled: !r.enabled } : r))
    );
  }

  function handleSaveEdit(updated: RuleItem) {
    setRules((prev) => prev.map((r) => (r.id === updated.id ? updated : r)));
    setEditingRule(null);
    setToast({ text: `Правило «${updated.name}» сохранено`, variant: "info" });
  }

  const filteredRules = rules.filter((r) =>
    r.name.toLowerCase().includes(search.toLowerCase()) ||
    r.condition.toLowerCase().includes(search.toLowerCase())
  );

  return (
    <div className="flux-dashboard-grid" style={{ gridTemplateColumns: "1fr 340px" }}>
      {/* ── Left Column ── */}
      <div className="flux-dashboard-left-col">
        {/* Search bar & Add rule */}
        <div style={{ display: "flex", gap: 10, alignItems: "center" }}>
          <div style={{ flex: 1, position: "relative" }}>
            <input
              type="text"
              placeholder="Поиск по правилу, домену, IP..."
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
              width: "auto",
              padding: "8px 16px",
              background: "linear-gradient(135deg, #32a9ff 0%, #0066cc 100%)",
              border: "1px solid rgba(152, 203, 255, 0.35)",
              color: "#fff",
              fontWeight: 600,
            }}
            onClick={() => {
              setEditingRule({
                id: `rule_${Date.now()}`,
                name: "Новое правило",
                condition: "domain:example.com",
                domains: ["domain:example.com"],
                action: "vpn",
                actionLabel: "Через VPN (VLESS)",
                traffic: "0 B",
                priority: rules.length + 1,
                enabled: true,
                arrowColor: "#32a9ff",
                badgeStyle: "vpn",
              });
            }}
          >
            + Добавить правило
          </button>
        </div>

        {/* Title & Tabs */}
        <div style={{ display: "flex", alignItems: "center", gap: 8, marginTop: 4, flexWrap: "wrap" }}>
          <h2 style={{ fontSize: 18, fontWeight: 700, margin: "0 8px 0 0", color: "#fff" }}>
            Маршрутизация
          </h2>
          <div style={{ display: "flex", gap: 6, flexWrap: "wrap" }}>
            <button
              type="button"
              className={`flux-time-btn${activeTab === "all" ? " active" : ""}`}
              onClick={() => {
                setActiveTab("all");
                void setRouting("all-proxy");
              }}
            >
              Полный прокси (Весь трафик)
            </button>
            <button
              type="button"
              className={`flux-time-btn${activeTab === "smart" ? " active" : ""}`}
              style={activeTab === "smart" ? { border: "1px solid var(--flux-blue)", color: "#fff" } : undefined}
              onClick={() => {
                setActiveTab("smart");
                void setRouting("bypass-russia");
              }}
            >
              Умный обход (Только заблокированные)
            </button>
            <button
              type="button"
              className={`flux-time-btn${activeTab === "whitelist" ? " active" : ""}`}
              onClick={() => setActiveTab("whitelist")}
            >
              Белый список
            </button>
            <button
              type="button"
              className={`flux-time-btn${activeTab === "custom" ? " active" : ""}`}
              onClick={() => setActiveTab("custom")}
            >
              Пользовательский
            </button>
            <button
              type="button"
              className={`flux-time-btn${activeTab === "import" ? " active" : ""}`}
              onClick={() => {
                setActiveTab("import");
                void importRouting();
              }}
            >
              Импорт правил (geo / dat)
            </button>
          </div>
        </div>

        {/* Rules Table */}
        <div style={{ display: "flex", flexDirection: "column", gap: 6, marginTop: 6 }}>
          {filteredRules.map((r) => {
            return (
              <div
                key={r.id}
                className="flux-card"
                style={{
                  padding: "10px 14px",
                  display: "flex",
                  alignItems: "center",
                  justifyContent: "space-between",
                  cursor: "pointer",
                }}
                onClick={() => setEditingRule(r)}
              >
                <div style={{ display: "flex", alignItems: "center", gap: 14 }}>
                  <span style={{ fontSize: 16, color: r.arrowColor, fontWeight: 700 }}>→</span>
                  <div>
                    <div style={{ fontSize: 13.5, fontWeight: 600, color: "#fff" }}>{r.name}</div>
                    <div style={{ fontSize: 11, color: "var(--flux-text-muted)", fontFamily: "ui-monospace, monospace" }}>
                      {r.condition}
                    </div>
                  </div>
                </div>

                <div style={{ display: "flex", alignItems: "center", gap: 18 }}>
                  <span
                    style={{
                      fontSize: 11.5,
                      fontWeight: 600,
                      padding: "4px 10px",
                      borderRadius: "var(--flux-radius-sm)",
                      background:
                        r.badgeStyle === "vpn_dpi"
                          ? "rgba(50, 169, 255, 0.16)"
                          : r.badgeStyle === "vpn"
                          ? "rgba(67, 239, 175, 0.14)"
                          : "rgba(255, 255, 255, 0.05)",
                      color:
                        r.badgeStyle === "vpn_dpi"
                          ? "var(--flux-cyan)"
                          : r.badgeStyle === "vpn"
                          ? "var(--flux-mint)"
                          : "var(--flux-text-muted)",
                      border:
                        r.badgeStyle === "vpn_dpi"
                          ? "1px solid rgba(50, 169, 255, 0.35)"
                          : r.badgeStyle === "vpn"
                          ? "1px solid rgba(67, 239, 175, 0.35)"
                          : "1px solid rgba(255, 255, 255, 0.08)",
                    }}
                  >
                    {r.actionLabel}
                  </span>

                  <span style={{ fontSize: 11.5, color: "var(--flux-text-muted)", width: 55, textAlign: "right", fontFamily: "ui-monospace, monospace" }}>
                    {r.traffic}
                  </span>

                  <span style={{ fontSize: 11.5, color: "var(--flux-text-subtle)", width: 20, textAlign: "center" }}>
                    {r.priority}
                  </span>

                  <label className="flux-switch" onClick={(e) => e.stopPropagation()}>
                    <input
                      type="checkbox"
                      checked={r.enabled}
                      onChange={(e) => toggleRule(r.id, e as any)}
                    />
                    <span className="flux-switch-slider" />
                  </label>
                </div>
              </div>
            );
          })}
        </div>
      </div>

      {/* ── Right Column: UNIFIED MONOLITH PANEL ── */}
      <div className="flux-dashboard-right-col">
        <div className="flux-right-monolith">
          {/* Section 1: Текущий режим */}
          <div className="flux-monolith-section">
            <div className="flux-monolith-header">
              <div className="flux-monolith-title-wrap">
                <svg width="15" height="15" viewBox="0 0 24 24" fill="none" stroke="currentColor" strokeWidth="2" style={{ color: "var(--flux-cyan)" }}>
                  <circle cx="12" cy="12" r="10" />
                  <path d="M12 2a14.5 14.5 0 0 0 0 20 14.5 14.5 0 0 0 0-20M2 12h20" />
                </svg>
                <span className="flux-monolith-title">
                  Текущий режим: {currentMode === "bypass-russia" ? "Умный обход" : currentMode}
                </span>
              </div>
            </div>

            <div
              style={{
                padding: "10px 10px",
                background: "rgba(0, 0, 0, 0.25)",
                borderRadius: "var(--flux-radius-md)",
                display: "flex",
                alignItems: "center",
                justifyContent: "space-between",
                fontSize: 11.5,
                marginTop: 2,
              }}
            >
              <span>Трафик →</span>
              <div
                style={{
                  padding: "5px 10px",
                  background: "rgba(255, 255, 255, 0.06)",
                  border: "1px solid rgba(255, 255, 255, 0.1)",
                  borderRadius: 8,
                  textAlign: "center",
                  fontWeight: 600,
                  color: "#fff",
                }}
              >
                Filter<br />(DPI/Geo)
              </div>
              <div style={{ display: "flex", flexDirection: "column", gap: 2, fontSize: 11 }}>
                <span style={{ color: "var(--flux-purple)" }}>68% Direct</span>
                <span style={{ color: "var(--flux-cyan)" }}>32% VPN</span>
              </div>
            </div>

            <div style={{ fontSize: 10.5, color: "var(--flux-text-muted)", textAlign: "center", marginTop: 2 }}>
              DoH (Cloudflare 1.1.1.1) • Анти-DNS Leak
            </div>
          </div>

          <div className="flux-monolith-divider" />

          {/* Section 2: Распределение трафика */}
          <div className="flux-monolith-section">
            <div className="flux-monolith-header">
              <div className="flux-monolith-title-wrap">
                <svg width="15" height="15" viewBox="0 0 24 24" fill="none" stroke="currentColor" strokeWidth="2" style={{ color: "var(--flux-cyan)" }}>
                  <path d="M21.21 15.89A10 10 0 1 1 8 2.83" />
                  <path d="M22 12A10 10 0 0 0 12 2v10z" />
                </svg>
                <span className="flux-monolith-title">Распределение трафика (Live)</span>
              </div>
            </div>

            <div style={{ display: "flex", justifyContent: "space-between", fontSize: 11, marginTop: 2 }}>
              <span style={{ color: "var(--flux-cyan)" }}>VPN: 17.6 GB</span>
              <span style={{ color: "var(--flux-purple)" }}>Direct: 31.4 GB</span>
            </div>

            <div style={{ height: 60, width: "100%", position: "relative", marginTop: 2 }}>
              <svg viewBox="0 0 100 40" preserveAspectRatio="none" style={{ width: "100%", height: "100%" }}>
                <defs>
                  <linearGradient id="dist-cyan" x1="0" y1="0" x2="0" y2="40" gradientUnits="userSpaceOnUse">
                    <stop offset="0%" stopColor="#32a9ff" stopOpacity="0.4" />
                    <stop offset="100%" stopColor="#32a9ff" stopOpacity="0" />
                  </linearGradient>
                  <linearGradient id="dist-purple" x1="0" y1="0" x2="0" y2="40" gradientUnits="userSpaceOnUse">
                    <stop offset="0%" stopColor="#d9b9ff" stopOpacity="0.3" />
                    <stop offset="100%" stopColor="#d9b9ff" stopOpacity="0" />
                  </linearGradient>
                </defs>
                <path d="M0,35 Q25,10 50,25 T100,15 L100,40 L0,40 Z" fill="url(#dist-cyan)" />
                <path d="M0,35 Q25,10 50,25 T100,15" fill="none" stroke="#32a9ff" strokeWidth="1.8" />
                <path d="M0,38 Q30,22 60,32 T100,24 L100,40 L0,40 Z" fill="url(#dist-purple)" />
                <path d="M0,38 Q30,22 60,32 T100,24" fill="none" stroke="#d9b9ff" strokeWidth="1.8" />
              </svg>
            </div>
          </div>

          <div className="flux-monolith-divider" />

          {/* Section 3: Действия маршрутизации */}
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
                onClick={() => setToast({ text: "Тест маршрутизации выполнен: без аномалий", variant: "info" })}
              >
                <svg width="14" height="14" viewBox="0 0 24 24" fill="none" stroke="currentColor" strokeWidth="2">
                  <circle cx="12" cy="12" r="3" />
                  <path d="M19.4 15a1.65 1.65 0 0 0 .33 1.82l.06.06a2 2 0 0 1-2.83 2.83l-.06-.06a1.65 1.65 0 0 0-1.82-.33" />
                </svg>
                <span>Тест маршрутизации IP/Домена</span>
              </button>
              <button
                type="button"
                className="flux-btn-control-action"
                onClick={() => setToast({ text: "Базы GeoIP / Geosite обновлены до последней версии", variant: "info" })}
              >
                <svg width="14" height="14" viewBox="0 0 24 24" fill="none" stroke="currentColor" strokeWidth="2">
                  <path d="M23 4v6h-6M1 20v-6h6M3.51 9a9 9 0 0 1 14.85-3.36L23 10M1 14l4.64 4.36A9 9 0 0 0 20.49 15" />
                </svg>
                <span>Обновить базы GeoIP / Geosite</span>
              </button>
              <button
                type="button"
                className="flux-btn-control-action"
                onClick={() => void exportRouting()}
              >
                <svg width="14" height="14" viewBox="0 0 24 24" fill="none" stroke="currentColor" strokeWidth="2">
                  <path d="M21 15v4a2 2 0 0 1-2 2H5a2 2 0 0 1-2-2v-4M7 10l5 5 5-5M12 15V3" />
                </svg>
                <span>Экспорт конфигурации routing.json</span>
              </button>
            </div>
          </div>
        </div>
      </div>

      {/* ── Modal: Edit Rule ── */}
      {editingRule && (
        <div
          style={{
            position: "fixed",
            inset: 0,
            background: "rgba(0, 0, 0, 0.75)",
            backdropFilter: "blur(10px)",
            zIndex: 100,
            display: "flex",
            alignItems: "center",
            justifyContent: "center",
          }}
        >
          <div
            className="flux-card"
            style={{
              width: 580,
              maxWidth: "92vw",
              background: "#161a24",
              border: "1px solid rgba(152, 203, 255, 0.28)",
              borderRadius: "var(--flux-radius-xl)",
              padding: 24,
              boxShadow: "0 16px 48px rgba(0, 0, 0, 0.7)",
            }}
          >
            <div style={{ display: "flex", justifyContent: "space-between", alignItems: "flex-start", marginBottom: 16 }}>
              <div>
                <h3 style={{ margin: 0, fontSize: 18, fontWeight: 700, color: "#fff" }}>
                  Редактирование правила{" "}
                  <span style={{ fontSize: 12, color: "var(--flux-cyan)", fontWeight: 500 }}>
                    ID: {editingRule.id} · {editingRule.name}
                  </span>
                </h3>
                <p style={{ margin: "4px 0 0 0", fontSize: 11.5, color: "var(--flux-text-muted)" }}>
                  Настройка условий перехвата, целевых протоколов и действия для сетевого трафика
                </p>
              </div>
              <button
                type="button"
                className="flux-win-btn close"
                onClick={() => setEditingRule(null)}
                style={{ fontSize: 16 }}
              >
                ✕
              </button>
            </div>

            <div style={{ display: "grid", gridTemplateColumns: "1fr 140px", gap: 12, marginBottom: 14 }}>
              <div>
                <label style={{ fontSize: 11.5, color: "var(--flux-text-muted)", display: "block", marginBottom: 4 }}>
                  Название правила
                </label>
                <input
                  type="text"
                  value={editingRule.name}
                  onChange={(e) => setEditingRule({ ...editingRule, name: e.target.value })}
                  style={{ width: "100%", padding: "7px 10px", fontSize: 13 }}
                />
              </div>
              <div>
                <label style={{ fontSize: 11.5, color: "var(--flux-text-muted)", display: "block", marginBottom: 4 }}>
                  Приоритет
                </label>
                <input
                  type="number"
                  value={editingRule.priority}
                  onChange={(e) => setEditingRule({ ...editingRule, priority: Number(e.target.value) })}
                  style={{ width: "100%", padding: "7px 10px", fontSize: 13 }}
                />
              </div>
            </div>

            <div style={{ marginBottom: 14 }}>
              <label style={{ fontSize: 11.5, color: "var(--flux-text-muted)", display: "block", marginBottom: 6 }}>
                Условия сопоставления (Target Domains / IPs / Geosite)
              </label>
              <textarea
                rows={5}
                value={editingRule.domains.join("\n")}
                onChange={(e) =>
                  setEditingRule({
                    ...editingRule,
                    domains: e.target.value.split("\n").filter(Boolean),
                    condition: e.target.value.split("\n")[0] || "custom",
                  })
                }
                style={{
                  width: "100%",
                  padding: 10,
                  fontSize: 12,
                  fontFamily: "ui-monospace, monospace",
                  boxSizing: "border-box",
                }}
              />
            </div>

            <div style={{ marginBottom: 16 }}>
              <label style={{ fontSize: 11.5, color: "var(--flux-text-muted)", display: "block", marginBottom: 8 }}>
                Действие (Outbound Action)
              </label>
              <div style={{ display: "grid", gridTemplateColumns: "1fr 1fr 1fr", gap: 8 }}>
                <button
                  type="button"
                  className="flux-btn-control-action"
                  style={{
                    borderColor: editingRule.action === "vpn_dpi" ? "var(--flux-blue)" : undefined,
                    background: editingRule.action === "vpn_dpi" ? "rgba(50, 169, 255, 0.16)" : undefined,
                    flexDirection: "column",
                    padding: "8px 6px",
                  }}
                  onClick={() =>
                    setEditingRule({
                      ...editingRule,
                      action: "vpn_dpi",
                      actionLabel: "Через VPN + DPI",
                      badgeStyle: "vpn_dpi",
                      arrowColor: "#32a9ff",
                    })
                  }
                >
                  <strong style={{ fontSize: 12 }}>🛡 Прокси (VPN + DPI)</strong>
                  <span style={{ fontSize: 10, color: "var(--flux-text-muted)", marginTop: 2 }}>
                    VLESS Reality + DPI
                  </span>
                </button>

                <button
                  type="button"
                  className="flux-btn-control-action"
                  style={{
                    borderColor: editingRule.action === "direct" ? "var(--flux-blue)" : undefined,
                    background: editingRule.action === "direct" ? "rgba(50, 169, 255, 0.16)" : undefined,
                    flexDirection: "column",
                    padding: "8px 6px",
                  }}
                  onClick={() =>
                    setEditingRule({
                      ...editingRule,
                      action: "direct",
                      actionLabel: "Прямое подключение",
                      badgeStyle: "direct",
                      arrowColor: "#89919d",
                    })
                  }
                >
                  <strong style={{ fontSize: 12 }}>→ Прямое соединение</strong>
                  <span style={{ fontSize: 10, color: "var(--flux-text-muted)", marginTop: 2 }}>
                    В обход VPN (Direct)
                  </span>
                </button>

                <button
                  type="button"
                  className="flux-btn-control-action"
                  style={{
                    borderColor: editingRule.action === "reject" ? "var(--flux-red)" : undefined,
                    background: editingRule.action === "reject" ? "rgba(255, 107, 126, 0.16)" : undefined,
                    flexDirection: "column",
                    padding: "8px 6px",
                  }}
                  onClick={() =>
                    setEditingRule({
                      ...editingRule,
                      action: "reject",
                      actionLabel: "Блокировать (Reject)",
                      badgeStyle: "bypass",
                      arrowColor: "#ff6b7e",
                    })
                  }
                >
                  <strong style={{ fontSize: 12 }}>⊘ Блокировать</strong>
                  <span style={{ fontSize: 10, color: "var(--flux-text-muted)", marginTop: 2 }}>
                    Сброс TCP/UDP
                  </span>
                </button>
              </div>
            </div>

            <div style={{ display: "flex", justifyContent: "space-between", alignItems: "center", borderTop: "1px solid rgba(255, 255, 255, 0.08)", paddingTop: 14 }}>
              <button
                type="button"
                className="flux-btn-stop-all"
                onClick={() => {
                  setRules((prev) => prev.filter((r) => r.id !== editingRule.id));
                  setEditingRule(null);
                }}
              >
                Удалить правило
              </button>

              <div style={{ display: "flex", gap: 8 }}>
                <button
                  type="button"
                  className="flux-btn-control-action"
                  onClick={() => setEditingRule(null)}
                >
                  Отмена
                </button>
                <button
                  type="button"
                  className="flux-btn-control-action"
                  style={{ background: "var(--flux-blue)", borderColor: "var(--flux-blue)", color: "#fff", fontWeight: 600 }}
                  onClick={() => handleSaveEdit(editingRule)}
                >
                  Сохранить изменения
                </button>
              </div>
            </div>
          </div>
        </div>
      )}
    </div>
  );
}
