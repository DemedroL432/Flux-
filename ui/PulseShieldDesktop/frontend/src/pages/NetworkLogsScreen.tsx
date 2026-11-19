import React, { useEffect, useState } from "react";
import { useApp } from "../context/AppContext";
import { formatDate } from "../format";
import { ACCENTS, THEMES } from "../themes";
import { LogSource, LogViewMode } from "../types";
import { logPanelContent, logPanelPlaceholder } from "../utils";

type Any = Record<string, unknown>;
type NetSubTab = "logs" | "network" | "diag";

const DNS_PRESETS: {
  key: string;
  label: string;
  primary: string;
  secondary: string;
}[] = [
  { key: "dhcp", label: "Авто (DHCP / системный)", primary: "", secondary: "" },
  { key: "cloudflare", label: "Cloudflare (1.1.1.1)", primary: "1.1.1.1", secondary: "1.0.0.1" },
  { key: "google", label: "Google (8.8.8.8)", primary: "8.8.8.8", secondary: "8.8.4.4" },
  { key: "quad9", label: "Quad9 (9.9.9.9)", primary: "9.9.9.9", secondary: "149.112.112.112" },
  { key: "yandex", label: "Яндекс (77.88.8.8)", primary: "77.88.8.8", secondary: "77.88.8.1" },
  { key: "adguard", label: "AdGuard (94.140.14.14)", primary: "94.140.14.14", secondary: "94.140.15.15" },
  { key: "comss", label: "Comss (83.220.169.155)", primary: "83.220.169.155", secondary: "212.109.195.93" },
  { key: "xbox", label: "Xbox / Smart (111.88.96.50)", primary: "111.88.96.50", secondary: "111.88.96.51" },
  { key: "opendns", label: "OpenDNS (208.67.222.222)", primary: "208.67.222.222", secondary: "208.67.220.220" },
  { key: "custom", label: "Свой DNS", primary: "", secondary: "" },
];

function renderSelfCheck(data: unknown): React.ReactNode {
  if (!data) {
    return (
      <div style={{ color: "var(--text-muted)", fontSize: 12.5 }}>
        Нажмите «Самопроверка», чтобы протестировать локальные бинарники, права
        WinDivert и сетевые службы.
      </div>
    );
  }
  if (typeof data === "string") {
    return <pre className="fx-pre" style={{ padding: 0 }}>{data}</pre>;
  }
  const obj = data as Any;
  const checks = Array.isArray(obj.checks) ? (obj.checks as Any[]) : null;
  if (checks && checks.length > 0) {
    return (
      <div className="fx-list-wrap" style={{ borderTop: "none" }}>
        {checks.map((chk, i) => {
          const ok = Boolean(chk.ok ?? chk.passed);
          return (
            <div key={i} className="fx-trow" style={{ paddingLeft: 0, paddingRight: 0 }}>
              <span className={`fx-tag ${ok ? "good" : "bad"}`}>
                {ok ? "OK" : "Сбой"}
              </span>
              <span style={{ width: 200, fontWeight: 500 }}>
                {String(chk.name || chk.title || `Проверка ${i + 1}`)}
              </span>
              <span
                className="fx-cell-grow fx-mono fx-select-text"
                style={{ fontSize: 12, color: "var(--text-secondary)" }}
              >
                {String(chk.detail || chk.message || chk.error || "—")}
              </span>
            </div>
          );
        })}
      </div>
    );
  }
  return (
    <pre className="fx-pre" style={{ padding: 0 }}>
      {JSON.stringify(data, null, 2)}
    </pre>
  );
}

export function NetworkLogsScreen() {
  const {
    uiState,
    appVersion,
    logs,
    clearLogs,
    openLogsFolder,
    proxyPortDraft,
    applyProxyPort,
    applyDns,
    theme,
    setTheme,
    accent,
    setAccent,
    activateLicense,
    checkUpdates,
    updateInfo,
    installUpdate,
    selfCheck,
    runSelfCheck,
    exportBundle,
    guide,
    quit,
    notify,
    busy,
  } = useApp();

  const [tab, setTab] = useState<NetSubTab>("logs");

  // Logs tab state
  const [logSource, setLogSource] = useState<LogSource>("dpi");
  const [logMode, setLogMode] = useState<LogViewMode>("full");

  // Network tab state
  const [portInput, setPortInput] = useState(
    String(proxyPortDraft || uiState?.proxyPort || 8080),
  );
  const [dnsPreset, setDnsPreset] = useState(
    String(uiState?.dnsPreset || "dhcp"),
  );
  const [dnsPrimary, setDnsPrimary] = useState(
    String(uiState?.dnsPrimary ?? ""),
  );
  const [dnsSecondary, setDnsSecondary] = useState(
    String(uiState?.dnsSecondary ?? ""),
  );
  const [licenseCode, setLicenseCode] = useState("");

  useEffect(() => {
    setPortInput(String(uiState?.proxyPort || proxyPortDraft || 8080));
  }, [uiState?.proxyPort, proxyPortDraft]);

  useEffect(() => {
    setDnsPreset(String(uiState?.dnsPreset || "dhcp"));
    setDnsPrimary(String(uiState?.dnsPrimary ?? ""));
    setDnsSecondary(String(uiState?.dnsSecondary ?? ""));
  }, [uiState?.dnsPreset, uiState?.dnsPrimary, uiState?.dnsSecondary]);

  const logText = logPanelContent(logs, logSource, logMode);
  const logFallback = logPanelPlaceholder(logSource, logMode);
  const activeLogPath =
    logSource === "dpi"
      ? String(logs?.dpiLogPath ?? "")
      : String(logs?.coreLogPath ?? "");

  async function handleCopyLog() {
    const content = logText || logFallback;
    if (!content) return;
    try {
      await navigator.clipboard.writeText(content);
      notify("Содержимое журнала скопировано", "info");
    } catch {
      notify("Не удалось скопировать в буфер обмена", "error");
    }
  }

  function handlePresetChange(nextKey: string) {
    setDnsPreset(nextKey);
    const found = DNS_PRESETS.find((p) => p.key === nextKey);
    if (found && nextKey !== "custom") {
      setDnsPrimary(found.primary);
      setDnsSecondary(found.secondary);
    }
  }

  const licExpires = Number(uiState?.licenseExpiresUnix ?? 0);

  return (
    <div>
      {/* Sub-navigation */}
      <div className="fx-sec-head" style={{ marginBottom: 18 }}>
        <div className="fx-seg" role="tablist" aria-label="Разделы сети и диагностики">
          <button
            type="button"
            role="tab"
            aria-selected={tab === "logs"}
            className={`fx-seg-btn ${tab === "logs" ? "active" : ""}`}
            onClick={() => setTab("logs")}
          >
            Журнал
          </button>
          <button
            type="button"
            role="tab"
            aria-selected={tab === "network"}
            className={`fx-seg-btn ${tab === "network" ? "active" : ""}`}
            onClick={() => setTab("network")}
          >
            Параметры сети
          </button>
          <button
            type="button"
            role="tab"
            aria-selected={tab === "diag"}
            className={`fx-seg-btn ${tab === "diag" ? "active" : ""}`}
            onClick={() => setTab("diag")}
          >
            Диагностика и обновления
          </button>
        </div>

        <span className="fx-mono" style={{ fontSize: 11.5, color: "var(--text-muted)" }}>
          Flux v{appVersion || "0.7.6"}
        </span>
      </div>

      {/* Tab 1: Logs */}
      {tab === "logs" && (
        <section className="fx-section fx-section-first">
          <div className="fx-sec-head">
            <div className="fx-bar">
              <div className="fx-seg" role="group" aria-label="Источник логов">
                <button
                  type="button"
                  className={`fx-seg-btn ${logSource === "dpi" ? "active" : ""}`}
                  onClick={() => setLogSource("dpi")}
                >
                  DPI · winws2
                </button>
                <button
                  type="button"
                  className={`fx-seg-btn ${logSource === "core" ? "active" : ""}`}
                  onClick={() => setLogSource("core")}
                >
                  Ядро · sing-box
                </button>
              </div>

              <div className="fx-seg" role="group" aria-label="Режим журнала">
                <button
                  type="button"
                  className={`fx-seg-btn ${logMode === "full" ? "active" : ""}`}
                  onClick={() => setLogMode("full")}
                >
                  События
                </button>
                <button
                  type="button"
                  className={`fx-seg-btn ${logMode === "errors" ? "active" : ""}`}
                  onClick={() => setLogMode("errors")}
                >
                  Ошибки
                </button>
                <button
                  type="button"
                  className={`fx-seg-btn ${logMode === "filter" ? "active" : ""}`}
                  onClick={() => setLogMode("filter")}
                >
                  Соединения
                </button>
              </div>
            </div>

            <div className="fx-bar">
              <button
                type="button"
                className="fx-btn fx-btn-sm"
                onClick={() => void openLogsFolder()}
              >
                Открыть папку
              </button>
              <button
                type="button"
                className="fx-btn fx-btn-sm"
                onClick={() => void handleCopyLog()}
              >
                Копировать
              </button>
              <button
                type="button"
                className="fx-btn fx-btn-sm fx-btn-danger"
                disabled={busy}
                onClick={() => void clearLogs()}
              >
                Очистить
              </button>
            </div>
          </div>

          {activeLogPath && (
            <div
              className="fx-mono fx-truncate fx-select-text"
              style={{
                fontSize: 11,
                color: "var(--text-muted)",
                marginBottom: 8,
              }}
              title={activeLogPath}
            >
              Файл: {activeLogPath}
            </div>
          )}

          <div className="fx-well fx-log-box">
            <pre className="fx-pre" style={{ flex: 1 }}>
              {logText || logFallback}
            </pre>
          </div>
        </section>
      )}

      {/* Tab 2: Network Settings & App Preferences */}
      {tab === "network" && (
        <div className="fx-split-2">
          {/* Left Column: Port & DNS */}
          <div className="fx-col">
            <section className="fx-section fx-section-first">
              <div className="fx-sec-head">
                <div>
                  <div className="fx-sec-title">ЛОКАЛЬНЫЙ ПРОКСИ-ПОРТ</div>
                  <div className="fx-sec-desc">
                    Порт входящего локального SOCKS5/HTTP-прокси (127.0.0.1).
                  </div>
                </div>
              </div>

              <form
                className="fx-bar"
                onSubmit={(e) => {
                  e.preventDefault();
                  const num = Number(portInput);
                  if (num >= 1024 && num <= 65535) {
                    void applyProxyPort(num);
                  } else {
                    notify("Укажите корректный порт от 1024 до 65535", "error");
                  }
                }}
              >
                <input
                  type="number"
                  min={1024}
                  max={65535}
                  className="fx-well-input mono"
                  style={{ width: 130 }}
                  value={portInput}
                  onChange={(e) => setPortInput(e.target.value)}
                />
                <button
                  type="submit"
                  className="fx-btn fx-btn-primary"
                  disabled={busy}
                >
                  Применить порт
                </button>
              </form>
            </section>

            <section className="fx-section">
              <div className="fx-sec-head">
                <div>
                  <div className="fx-sec-title">ПАРАМЕТРЫ DNS</div>
                  <div className="fx-sec-desc">
                    Используется ядром sing-box для резолва доменов при работе
                    прокси.
                  </div>
                </div>
              </div>

              <div className="fx-bar" style={{ marginBottom: 10 }}>
                <select
                  className="fx-well-select fx-bar-grow"
                  value={dnsPreset}
                  onChange={(e) => handlePresetChange(e.target.value)}
                  aria-label="Пресет DNS"
                >
                  {DNS_PRESETS.map((p) => (
                    <option key={p.key} value={p.key}>
                      {p.label}
                    </option>
                  ))}
                </select>
              </div>

              <div className="fx-bar" style={{ marginBottom: 10 }}>
                <input
                  type="text"
                  className="fx-well-input mono fx-bar-grow"
                  placeholder="Основной DNS (напр. 1.1.1.1)"
                  value={dnsPrimary}
                  disabled={dnsPreset === "dhcp"}
                  onChange={(e) => {
                    setDnsPrimary(e.target.value);
                    if (dnsPreset !== "custom") setDnsPreset("custom");
                  }}
                />
                <input
                  type="text"
                  className="fx-well-input mono fx-bar-grow"
                  placeholder="Резервный DNS (напр. 1.0.0.1)"
                  value={dnsSecondary}
                  disabled={dnsPreset === "dhcp"}
                  onChange={(e) => {
                    setDnsSecondary(e.target.value);
                    if (dnsPreset !== "custom") setDnsPreset("custom");
                  }}
                />
              </div>

              <button
                type="button"
                className="fx-btn fx-btn-primary"
                disabled={busy}
                onClick={() => void applyDns(dnsPreset, dnsPrimary, dnsSecondary)}
              >
                Применить DNS
              </button>
            </section>
          </div>

          {/* Right Column: Theme, License, Process */}
          <div className="fx-col fx-col-divided">
            <section className="fx-section fx-section-first">
              <div className="fx-sec-head">
                <div>
                  <div className="fx-sec-title">ТЕМА И АКЦЕНТ</div>
                  <div className="fx-sec-desc">
                    Визуальная палитра интерфейса рабочего окна.
                  </div>
                </div>
              </div>

              <div style={{ display: "flex", flexDirection: "column", gap: 10 }}>
                <div className="fx-seg" role="group" aria-label="Тема">
                  {THEMES.map((t) => (
                    <button
                      key={t.key}
                      type="button"
                      className={`fx-seg-btn ${theme === t.key ? "active" : ""}`}
                      onClick={() => setTheme(t.key)}
                      title={t.hint}
                    >
                      {t.name}
                    </button>
                  ))}
                </div>

                <div className="fx-bar" role="group" aria-label="Акцентный цвет" style={{ gap: 6 }}>
                  {ACCENTS.map((a) => (
                    <button
                      key={a.key}
                      type="button"
                      className={`fx-btn fx-btn-sm ${accent === a.key ? "fx-btn-primary" : ""}`}
                      onClick={() => setAccent(a.key)}
                    >
                      <span
                        aria-hidden="true"
                        style={{
                          width: 8,
                          height: 8,
                          borderRadius: "99px",
                          background: a.color,
                          border: "1px solid var(--hairline-strong)",
                          flexShrink: 0,
                        }}
                      />
                      <span>{a.name}</span>
                    </button>
                  ))}
                </div>
              </div>
            </section>

            <section className="fx-section">
              <div className="fx-sec-head">
                <div className="fx-sec-title">ЛИЦЕНЗИЯ И УСТРОЙСТВО</div>
                <span
                  className={`fx-tag ${uiState?.licenseActive ? "good" : ""}`}
                >
                  {String(uiState?.licenseTier || "Free")}
                </span>
              </div>

              <div className="fx-kv" style={{ marginBottom: 10 }}>
                <div className="fx-kv-row">
                  <span className="fx-kv-k">ID устройства</span>
                  <span
                    className="fx-kv-v fx-mono fx-select-text"
                    title={String(uiState?.deviceIdFull ?? "")}
                  >
                    {String(uiState?.deviceIdShort || uiState?.deviceIdFull || "—")}
                  </span>
                </div>
                {licExpires > 0 && (
                  <div className="fx-kv-row">
                    <span className="fx-kv-k">Действует до</span>
                    <span className="fx-kv-v fx-mono">
                      {formatDate(licExpires)}
                    </span>
                  </div>
                )}
              </div>

              <form
                className="fx-bar"
                onSubmit={(e) => {
                  e.preventDefault();
                  if (licenseCode.trim()) {
                    void activateLicense(licenseCode.trim()).then(() =>
                      setLicenseCode(""),
                    );
                  }
                }}
              >
                <input
                  type="text"
                  className="fx-well-input mono fx-bar-grow"
                  placeholder="Ключ активации…"
                  value={licenseCode}
                  onChange={(e) => setLicenseCode(e.target.value)}
                />
                <button
                  type="submit"
                  className="fx-btn"
                  disabled={busy || !licenseCode.trim()}
                >
                  Активировать
                </button>
              </form>
            </section>

            <section className="fx-section">
              <div className="fx-sec-head">
                <div>
                  <div className="fx-sec-title">ПРИЛОЖЕНИЕ</div>
                  <div className="fx-sec-desc">
                    Полная остановка движков и закрытие фонового процесса в трее.
                  </div>
                </div>
                <button
                  type="button"
                  className="fx-btn fx-btn-danger"
                  onClick={() => void quit()}
                >
                  Выйти из Flux
                </button>
              </div>
            </section>
          </div>
        </div>
      )}

      {/* Tab 3: Diagnostics & Updates */}
      {tab === "diag" && (
        <div>
          <section className="fx-section fx-section-first">
            <div className="fx-sec-head">
              <div>
                <div className="fx-sec-title">ОБНОВЛЕНИЯ ПРИЛОЖЕНИЯ</div>
                <div className="fx-sec-desc">
                  Установлена версия{" "}
                  <span className="fx-mono">v{appVersion || "0.7.6"}</span>
                </div>
              </div>

              <div className="fx-bar">
                <button
                  type="button"
                  className="fx-btn fx-btn-primary"
                  disabled={busy}
                  onClick={() => void checkUpdates()}
                >
                  Проверить обновления
                </button>
                {updateInfo && !updateInfo.upToDate && Boolean(updateInfo.ok) && (
                  <button
                    type="button"
                    className="fx-btn"
                    disabled={busy}
                    onClick={() => void installUpdate()}
                  >
                    Установить обновление
                  </button>
                )}
              </div>
            </div>

            {updateInfo && (
              <div className="fx-well fx-well-pad">
                <div className="fx-bar" style={{ justifyContent: "space-between" }}>
                  <span>{String(updateInfo.message || "Проверка выполнена.")}</span>
                  <span
                    className={`fx-tag ${
                      updateInfo.upToDate
                        ? "good"
                        : updateInfo.ok
                          ? "accent"
                          : "bad"
                    }`}
                  >
                    {updateInfo.upToDate
                      ? "Актуально"
                      : String(
                          updateInfo.latestTag ||
                            updateInfo.latestVersion ||
                            "Статус",
                        )}
                  </span>
                </div>
              </div>
            )}
          </section>

          <section className="fx-section">
            <div className="fx-sec-head">
              <div>
                <div className="fx-sec-title">САМОПРОВЕРКА И ЭКСПОРТ ДИАГНОСТИКИ</div>
                <div className="fx-sec-desc">
                  Проверка компонентов sing-box / winws2 и выгрузка обезличенного
                  архива для разбора проблем.
                </div>
              </div>

              <div className="fx-bar">
                <button
                  type="button"
                  className="fx-btn"
                  disabled={busy}
                  onClick={() => void runSelfCheck()}
                >
                  Самопроверка
                </button>
                <button
                  type="button"
                  className="fx-btn"
                  disabled={busy}
                  onClick={() => void exportBundle()}
                >
                  Экспорт архива
                </button>
              </div>
            </div>

            <div className="fx-well fx-well-pad">{renderSelfCheck(selfCheck)}</div>
          </section>

          {guide && (
            <section className="fx-section">
              <div className="fx-sec-head">
                <div className="fx-sec-title">СПРАВКА И ПАМЯТКА</div>
              </div>
              <div className="fx-well">
                <pre className="fx-pre" style={{ maxHeight: 280 }}>
                  {guide}
                </pre>
              </div>
            </section>
          )}
        </div>
      )}
    </div>
  );
}
