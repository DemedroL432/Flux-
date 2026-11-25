import React, { useState } from "react";
import { useApp } from "../context/AppContext";
import { detectCountry } from "../flags";
import {
  formatBytes,
  formatLatency,
  formatMbps,
  latencyTone,
} from "../format";
import { FlagBadge } from "../icons";
import { ROUTE_TARGET_LABEL, RouteTarget } from "../types";
import { routingLabel } from "../ui";

type Any = Record<string, unknown>;

const QUICK_PROBES = ["youtube.com", "discord.com", "chatgpt.com", "gosuslugi.ru"];

function targetBadgeTone(target: string): string {
  switch (target) {
    case "proxy":
      return "accent";
    case "dpi-bypass":
      return "good";
    case "block":
      return "bad";
    default:
      return "";
  }
}

export function ShieldScreen() {
  const {
    uiState,
    health,
    rt,
    proxyEnabled,
    dpiEnabled,
    proxyMode,
    proxyPhase,
    setProxy,
    setProxyMode,
    setDpi,
    stopAll,
    activeServer,
    servers,
    splitRows,
    dpiLists,
    setActive,
    openDpiTab,
    retestServer,
    verifyTarget,
    busy,
  } = useApp();

  const [probeQuery, setProbeQuery] = useState("");
  const [probeResult, setProbeResult] = useState<Any | null>(null);
  const [probing, setProbing] = useState(false);

  const serverName = String(activeServer?.name ?? "");
  const serverHost = String(activeServer?.host ?? "");
  const serverPort = String(activeServer?.port ?? "");
  const serverProto = String(activeServer?.protocol ?? "");
  const serverId = String(activeServer?.id ?? "");
  const latency = Number(activeServer?.lastCheckLatencyMs ?? 0);
  const country = activeServer ? detectCountry(serverName, serverHost) : null;

  const dpiEnabledCount = dpiLists.filter((l) => Boolean(l.enabled)).length;
  const routingName = routingLabel(String(uiState?.routingProfile ?? ""));
  const dlBps = Number(rt?.downloadBps ?? 0);
  const ulBps = Number(rt?.uploadBps ?? 0);
  const sessDl = Number(rt?.sessionDownloadBytes ?? 0);
  const sessUl = Number(rt?.sessionUploadBytes ?? 0);

  let statusOverline = "ДВИЖКИ ОСТАНОВЛЕНЫ";
  let statusDotClass = "";
  if (proxyPhase === "connecting") {
    statusOverline = "ПОДКЛЮЧЕНИЕ ПРОКСИ…";
    statusDotClass = "connecting";
  } else if (proxyEnabled && dpiEnabled) {
    statusOverline = "ЗАЩИТА АКТИВНА · PROXY + DPI";
    statusDotClass = "on";
  } else if (proxyEnabled) {
    statusOverline = "ПРОКСИ АКТИВЕН · SING-BOX";
    statusDotClass = "on";
  } else if (dpiEnabled) {
    statusOverline = "ОБХОД DPI АКТИВЕН · WINWS2";
    statusDotClass = "on";
  }

  const headline = activeServer
    ? serverName
    : servers.length > 0
      ? "Узел не выбран"
      : "Добавьте сервер или включите DPI";

  const heroSub = [
    `Режим: ${routingName}`,
    uiState?.activeDpiProfile
      ? `DPI: ${String(uiState.activeDpiProfile)}`
      : "DPI профиль не выбран",
    `Локальный порт: :${Number(uiState?.proxyPort ?? 8080)}`,
  ].join(" · ");

  async function handleProbe(targetQ?: string) {
    const q = (targetQ ?? probeQuery).trim();
    if (!q) return;
    if (targetQ) setProbeQuery(targetQ);
    setProbing(true);
    try {
      const res = await verifyTarget(q);
      if (res) setProbeResult(res);
    } finally {
      setProbing(false);
    }
  }

  return (
    <div>
      {/* Borderless Hero */}
      <section className="fx-hero">
        <div className="fx-hero-row">
          <div>
            <div className="fx-hero-status">
              <span
                className={`fx-dot ${statusDotClass}`}
                style={{ marginTop: 0 }}
              />
              <span>{statusOverline}</span>
            </div>
            <h1 className="fx-hero-title">{headline}</h1>
            <p className="fx-hero-sub">{heroSub}</p>
          </div>

          <div className="fx-actions">
            {!proxyEnabled && (
              <button
                type="button"
                className="fx-btn fx-btn-primary"
                disabled={busy}
                onClick={() => void setProxy(true)}
              >
                Включить прокси
              </button>
            )}
            {!dpiEnabled && (
              <button
                type="button"
                className={`fx-btn ${proxyEnabled ? "fx-btn-primary" : ""}`}
                disabled={busy}
                onClick={() => void setDpi(true)}
              >
                Включить DPI
              </button>
            )}
            {(proxyEnabled || dpiEnabled) && (
              <button
                type="button"
                className="fx-btn fx-btn-danger"
                disabled={busy}
                onClick={() => void stopAll()}
              >
                Остановить всё
              </button>
            )}
          </div>
        </div>
      </section>

      {/* 4-column Hairline Telemetry Strip */}
      <section className="fx-metrics-4" aria-label="Телеметрия">
        <div className="fx-metric">
          <div className="fx-metric-k">Входящий</div>
          <div className="fx-metric-v">{formatMbps(dlBps)}</div>
          <div className="fx-metric-s">Сессия: {formatBytes(sessDl)}</div>
        </div>

        <div className="fx-metric">
          <div className="fx-metric-k">Исходящий</div>
          <div className="fx-metric-v">{formatMbps(ulBps)}</div>
          <div className="fx-metric-s">Сессия: {formatBytes(sessUl)}</div>
        </div>

        <div className="fx-metric">
          <div className="fx-metric-k">Пинг узла</div>
          <div className={`fx-metric-v fx-lat ${latencyTone(latency)}`}>
            {latency > 0 ? formatLatency(latency) : "—"}
          </div>
          <div className="fx-metric-s">
            {activeServer ? serverName : "Узел не выбран"}
          </div>
        </div>

        <div className="fx-metric">
          <div className="fx-metric-k">Маршрутизация</div>
          <div className="fx-metric-v" style={{ fontSize: 15, fontFamily: "var(--font-sans)" }}>
            {routingName}
          </div>
          <div className="fx-metric-s">{splitRows.length} правил</div>
        </div>
      </section>

      {/* 2-Column Open Split: PROXY // SING-BOX | DPI // WINWS2 */}
      <section className="fx-section">
        <div className="fx-split-2">
          {/* Proxy Column */}
          <div className="fx-col">
            <div className="fx-sec-head">
              <span className="fx-sec-title">PROXY // SING-BOX</span>
              <span
                className={`fx-tag ${
                  proxyPhase === "connected"
                    ? "good"
                    : proxyPhase === "connecting"
                      ? "warn"
                      : ""
                }`}
              >
                {proxyPhase === "connected"
                  ? "Подключён"
                  : proxyPhase === "connecting"
                    ? "Подключение…"
                    : "Отключён"}
              </span>
            </div>

            {activeServer ? (
              <div
                style={{
                  display: "flex",
                  alignItems: "center",
                  gap: 10,
                  marginBottom: 14,
                }}
              >
                <FlagBadge code={country?.code} name={country?.name} />
                <div className="fx-cell-grow">
                  <div
                    className="fx-truncate"
                    style={{ fontWeight: 600, fontSize: 14 }}
                  >
                    {serverName}
                  </div>
                  <div
                    className="fx-mono fx-truncate"
                    style={{ fontSize: 11.5, color: "var(--text-muted)" }}
                  >
                    {serverProto.toUpperCase()} · {serverHost}:{serverPort}
                  </div>
                </div>
                <span className={`fx-lat ${latencyTone(latency)}`}>
                  {latency > 0 ? formatLatency(latency) : "—"}
                </span>
              </div>
            ) : (
              <div
                style={{
                  fontSize: 12.5,
                  color: "var(--text-muted)",
                  marginBottom: 14,
                }}
              >
                Активный сервер не выбран. Добавьте подписку или ключ на экране
                «Серверы».
              </div>
            )}

            <div className="fx-kv">
              <div className="fx-kv-row">
                <span className="fx-kv-k">Режим захвата</span>
                <div className="fx-seg" role="group" aria-label="Режим захвата трафика">
                  <button
                    type="button"
                    className={`fx-seg-btn ${proxyMode !== "tun" ? "active" : ""}`}
                    disabled={busy}
                    onClick={() => void setProxyMode("system_proxy")}
                    title="HTTP/SOCKS5 системный прокси Windows (без прав администратора)"
                  >
                    Системный прокси
                  </button>
                  <button
                    type="button"
                    className={`fx-seg-btn ${proxyMode === "tun" ? "active" : ""}`}
                    disabled={busy}
                    onClick={() => void setProxyMode("tun")}
                    title={
                      uiState?.elevated
                        ? "Виртуальный VPN-адаптер: перехватывает все программы, UDP/WebRTC и IPv6"
                        : "Требуется запуск Flux от имени администратора"
                    }
                  >
                    TUN (VPN)
                  </button>
                </div>
              </div>
              <div className="fx-kv-row">
                <span className="fx-kv-k">Локальный прокси</span>
                <span className="fx-kv-v fx-mono">
                  127.0.0.1:{Number(uiState?.proxyPort ?? 8080)} ·{" "}
                  {health?.proxyLocalOk ? "активен" : "остановлен"}
                </span>
              </div>
              <div className="fx-kv-row">
                <span className="fx-kv-k">Режим маршрута</span>
                <span className="fx-kv-v">{routingName}</span>
              </div>
              <div className="fx-kv-row">
                <span className="fx-kv-k">Источник</span>
                <span className="fx-kv-v">
                  {String(uiState?.subscriptionProviderLabel || "Локальные ключи")}
                </span>
              </div>
            </div>

            <div className="fx-bar" style={{ marginTop: 14 }}>
              <button
                type="button"
                className="fx-btn fx-btn-sm"
                onClick={() => setActive("servers")}
              >
                Сменить узел
              </button>
              {serverId && (
                <button
                  type="button"
                  className="fx-btn fx-btn-sm"
                  disabled={busy}
                  onClick={() => void retestServer(serverId)}
                >
                  Замерить пинг
                </button>
              )}
              <button
                type="button"
                className="fx-btn fx-btn-sm fx-btn-ghost"
                onClick={() => setActive("routes")}
              >
                Правила ({splitRows.length})
              </button>
            </div>
          </div>

          {/* DPI Column */}
          <div className="fx-col fx-col-divided">
            <div className="fx-sec-head">
              <span className="fx-sec-title">DPI // WINWS2</span>
              <span className={`fx-tag ${dpiEnabled ? "good" : ""}`}>
                {dpiEnabled ? "Активен" : "Отключён"}
              </span>
            </div>

            <div className="fx-kv">
              <div className="fx-kv-row">
                <span className="fx-kv-k">Профиль обхода</span>
                <span className="fx-kv-v fx-mono">
                  {String(uiState?.activeDpiProfile || "не выбран")}
                </span>
              </div>
              <div className="fx-kv-row">
                <span className="fx-kv-k">Движок</span>
                <span className="fx-kv-v fx-mono">
                  {String(health?.dpiTool || "winws2")}
                </span>
              </div>
              <div className="fx-kv-row">
                <span className="fx-kv-k">Списки доменов</span>
                <span className="fx-kv-v fx-mono">
                  {dpiEnabledCount} из {dpiLists.length} активно
                </span>
              </div>
              <div className="fx-kv-row">
                <span className="fx-kv-k">Системные права</span>
                <span className="fx-kv-v">
                  {uiState?.elevated
                    ? "Администратор (WinDivert готов)"
                    : "Обычный пользователь"}
                </span>
              </div>
            </div>

            <div className="fx-bar" style={{ marginTop: 14 }}>
              <button
                type="button"
                className="fx-btn fx-btn-sm"
                onClick={() => openDpiTab("profiles")}
              >
                Профили и автоподбор
              </button>
              <button
                type="button"
                className="fx-btn fx-btn-sm"
                onClick={() => openDpiTab("lists")}
              >
                Редактор списков
              </button>
            </div>
          </div>
        </div>
      </section>

      {/* Inline Route Probe */}
      <section className="fx-section">
        <div className="fx-sec-head">
          <div>
            <div className="fx-sec-title">ПРОВЕРКА МАРШРУТА ДОМЕНА</div>
            <div className="fx-sec-desc">
              Узнайте итоговое направление трафика с учётом активных правил,
              прокси и обхода DPI.
            </div>
          </div>
          <div className="fx-bar">
            {QUICK_PROBES.map((domain) => (
              <button
                key={domain}
                type="button"
                className="fx-btn fx-btn-sm fx-btn-ghost fx-mono"
                disabled={probing}
                onClick={() => void handleProbe(domain)}
              >
                {domain}
              </button>
            ))}
          </div>
        </div>

        <form
          className="fx-bar"
          onSubmit={(e) => {
            e.preventDefault();
            void handleProbe();
          }}
        >
          <input
            type="text"
            className="fx-well-input mono fx-bar-grow"
            placeholder="Введите домен, URL или IP (например youtube.com или 1.1.1.1)"
            value={probeQuery}
            onChange={(e) => setProbeQuery(e.target.value)}
          />
          <button
            type="submit"
            className="fx-btn fx-btn-primary"
            disabled={probing || !probeQuery.trim()}
          >
            {probing ? "Проверка…" : "Проверить"}
          </button>
        </form>

        {probeResult && (
          <div className="fx-well fx-well-pad" style={{ marginTop: 10 }}>
            <div
              style={{
                display: "flex",
                alignItems: "center",
                justifyContent: "space-between",
                gap: 12,
                flexWrap: "wrap",
              }}
            >
              <div style={{ display: "flex", alignItems: "center", gap: 8 }}>
                <span className="fx-mono fx-select-text" style={{ fontWeight: 600 }}>
                  {String(probeResult.normalized || probeResult.query || "")}
                </span>
                <span
                  className={`fx-tag ${targetBadgeTone(String(probeResult.target ?? ""))}`}
                >
                  {ROUTE_TARGET_LABEL[
                    String(probeResult.target ?? "") as RouteTarget
                  ] || String(probeResult.target || "—")}
                </span>
              </div>
              <span className="fx-mono" style={{ fontSize: 11.5, color: "var(--text-muted)" }}>
                {probeResult.final
                  ? "по умолчанию"
                  : String(
                      probeResult.ruleLabel || probeResult.source || "правило",
                    )}
              </span>
            </div>
            {Boolean(probeResult.note) && (
              <div
                style={{
                  marginTop: 6,
                  fontSize: 12,
                  color: "var(--text-secondary)",
                }}
              >
                {String(probeResult.note)}
              </div>
            )}
          </div>
        )}
      </section>
    </div>
  );
}
