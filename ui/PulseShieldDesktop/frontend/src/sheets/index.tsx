import { useState } from "react";
import { useApp } from "../context/AppContext";
import { Icon } from "../icons";
import { Switch } from "../ui";
import { ROUTE_TARGET_LABEL, type RouteTarget } from "../types";
import { formatBytes, formatLatency, formatMbps, latencyTone } from "../format";
import { ACCENTS, THEMES } from "../themes";

/** Шторка «Серверы и подписка». */
export function ServerSheet() {
  const {
    servers, activeServer, selectServer, removeServer, retestAll, busy,
    openSheet, applySubscription, importFromClipboard, refreshSubscription,
    clearSubscription, uiState,
  } = useApp();
  const [url, setUrl] = useState(String(uiState?.subscriptionUrl || ""));

  return (
    <>
      <div className="field-row">
        <input
          className="field mono"
          placeholder="https://provider/sub"
          value={url}
          onChange={(e) => setUrl(e.target.value)}
          aria-label="Ссылка подписки"
        />
        <button
          type="button"
          className="btn sm primary"
          disabled={busy || !url.trim()}
          onClick={() => void applySubscription(url.trim())}
        >
          Применить
        </button>
      </div>

      <div className="section-title">
        <span>Подписка</span>
      </div>
      <div className="row-list">
        <button type="button" className="row-l" disabled={busy} onClick={() => void refreshSubscription()}>
          <span className="rl-main">
            <strong>Обновить список</strong>
            <small>перечитать узлы из подписки</small>
          </span>
          <Icon name="refresh" size={15} />
        </button>
        <button type="button" className="row-l" disabled={busy} onClick={() => void importFromClipboard()}>
          <span className="rl-main">
            <strong>Импорт из буфера</strong>
            <small>ссылка, список или конфиг</small>
          </span>
          <Icon name="copy" size={15} />
        </button>
        <button type="button" className="row-l" disabled={busy} onClick={() => void clearSubscription()}>
          <span className="rl-main">
            <strong>Отвязать подписку</strong>
            <small>оставить только ручные узлы</small>
          </span>
          <Icon name="trash" size={15} />
        </button>
      </div>

      <div className="section-title">
        <span>Узлы</span>
        <button type="button" className="textbtn" disabled={busy} onClick={() => void retestAll()}>
          Проверить все
        </button>
      </div>

      {servers.length === 0 ? (
        <div className="empty">
          <Icon name="server" size={22} />
          <p>Серверов нет. Импортируйте подписку или вставьте ссылку.</p>
          <button type="button" className="btn primary block" onClick={() => openSheet("subscription")}>
            Подписка и ключи
          </button>
        </div>
      ) : (
        <div className="row-list">
          {servers.map((s) => {
            const id = String(s.id || "");
            const isActive = activeServer && String(activeServer.id) === id;
            const tone = latencyTone(Number(s.lastCheckLatencyMs));
            return (
              <div key={id} className={`server-row${isActive ? " sel" : ""}`}>
                <button
                  type="button"
                  className="rl-main"
                  disabled={busy}
                  onClick={() => void selectServer(id)}
                  style={{ background: "transparent", border: 0, textAlign: "left" }}
                >
                  <strong>
                    {String(s.name || s.host || "")}
                    {isActive ? " ✓" : ""}
                  </strong>
                  <small>
                    {String(s.host || "")}:{String(s.port || "")}
                    {s.protocol ? ` · ${String(s.protocol)}` : ""}
                  </small>
                </button>
                <span className={`rl-r ${tone}`}>{formatLatency(Number(s.lastCheckLatencyMs))}</span>
                <button
                  type="button"
                  className="rule-del"
                  title="Удалить узел"
                  disabled={busy}
                  onClick={() => void removeServer(id)}
                >
                  <Icon name="trash" size={14} />
                </button>
              </div>
            );
          })}
        </div>
      )}
    </>
  );
}

/** Шторка «Новое правило». */
export function RuleSheet() {
  const { addRule, busy, closeSheet, openSheet } = useApp();
  const [pattern, setPattern] = useState("");
  const [target, setTarget] = useState<RouteTarget>("proxy");

  const targets: RouteTarget[] = ["proxy", "dpi-bypass", "direct", "block"];

  async function submit() {
    const p = pattern.trim();
    if (!p) return;
    await addRule(p, target);
    closeSheet();
    if (target === "dpi-bypass") {
      openSheet("tune-dpi");
    }
  }

  return (
    <>
      <input
        className="field"
        placeholder="youtube.com"
        value={pattern}
        onChange={(e) => setPattern(e.target.value)}
        onKeyDown={(e) => {
          if (e.key === "Enter") void submit();
        }}
        aria-label="Домен, IP или имя процесса"
        autoFocus
      />

      <p className="sheet-sub">Домен, IP-адрес или имя программы (file.exe).</p>

      <div className="option-list">
        {targets.map((t) => (
          <button
            key={t}
            type="button"
            className={`option${target === t ? " on" : ""}`}
            onClick={() => setTarget(t)}
          >
            <span>{ROUTE_TARGET_LABEL[t]}</span>
            {target === t ? <Icon name="check" size={14} /> : null}
          </button>
        ))}
      </div>

      {target === "dpi-bypass" ? (
        <p className="sheet-sub">
          Домен попадёт в список <span className="mono">custom-dpi</span>. Обход заработает, только если включён
          модуль DPI.
        </p>
      ) : null}

      <button type="button" className="btn primary block" disabled={busy || !pattern.trim()} onClick={() => void submit()}>
        Добавить правило
      </button>
    </>
  );
}

/** Шторка «Проверить сайт». */
export function VerifySheet() {
  const { verifyTarget } = useApp();
  const [q, setQ] = useState("");
  const [res, setRes] = useState<Record<string, unknown> | null>(null);
  const [loading, setLoading] = useState(false);

  async function check() {
    if (!q.trim()) return;
    setLoading(true);
    try {
      setRes((await verifyTarget(q.trim())) as Record<string, unknown> | null);
    } finally {
      setLoading(false);
    }
  }

  return (
    <>
      <div className="field-row">
        <input
          className="field mono"
          placeholder="https://example.com или 1.2.3.4"
          value={q}
          onChange={(e) => setQ(e.target.value)}
          onKeyDown={(e) => {
            if (e.key === "Enter") void check();
          }}
          aria-label="Адрес для проверки"
        />
        <button type="button" className="btn sm primary" disabled={loading || !q.trim()} onClick={() => void check()}>
          {loading ? "…" : "Проверить"}
        </button>
      </div>

      {res ? (
        <div className="verdict">
          <div className="metric-row">
            <span>Адрес</span>
            <strong className="mono">{String(res.normalized || "")}</strong>
          </div>
          <div className="metric-row">
            <span>Пойдёт</span>
            <strong>{ROUTE_TARGET_LABEL[(res.target as RouteTarget) || "direct"]}</strong>
          </div>
          <div className="metric-row">
            <span>Источник</span>
            <strong>{String(res.source || "")}</strong>
          </div>
          <div className="metric-row">
            <span>Правило</span>
            <strong className="mono">{String(res.ruleLabel || "—")}</strong>
          </div>
          <p className="sheet-sub" style={{ margin: "8px 0 0" }}>{String(res.note || "")}</p>
        </div>
      ) : (
        <p className="sheet-sub">Проверит, какое правило сработает, и куда уйдёт трафик.</p>
      )}
    </>
  );
}

/** Шторка «Порт прокси». */
export function PortSheet() {
  const { proxyPortDraft, applyProxyPort, busy } = useApp();
  const [port, setPort] = useState(String(proxyPortDraft));

  return (
    <>
      <div className="field-row">
        <input
          className="field mono"
          type="number"
          min={1}
          max={65535}
          value={port}
          onChange={(e) => setPort(e.target.value)}
          aria-label="Порт локального прокси"
        />
        <button
          type="button"
          className="btn sm primary"
          disabled={busy}
          onClick={() => void applyProxyPort(Number(port))}
        >
          Применить
        </button>
      </div>
      <p className="sheet-sub">Порт слушает только локальный интерфейс. При активном прокси ядро перезапустится.</p>
    </>
  );
}

/** Шторка «DNS». */
export function DnsSheet() {
  const { uiState, applyDns, busy } = useApp();
  const [preset, setPreset] = useState(String(uiState?.dnsPreset || "auto"));
  const [primary, setPrimary] = useState(String(uiState?.dnsPrimary || ""));
  const [secondary, setSecondary] = useState(String(uiState?.dnsSecondary || ""));

  const presets = [
    { key: "auto", label: "Автоматически" },
    { key: "system", label: "Как в системе" },
    { key: "custom", label: "Свои серверы" },
  ];

  return (
    <>
      <div className="option-list">
        {presets.map((p) => (
          <button
            key={p.key}
            type="button"
            className={`option${preset === p.key ? " on" : ""}`}
            onClick={() => setPreset(p.key)}
          >
            <span>{p.label}</span>
            {preset === p.key ? <Icon name="check" size={14} /> : null}
          </button>
        ))}
      </div>

      {preset === "custom" ? (
        <>
          <input
            className="field mono"
            placeholder="1.1.1.1"
            value={primary}
            onChange={(e) => setPrimary(e.target.value)}
            aria-label="Основной DNS"
            style={{ marginBottom: 8 }}
          />
          <input
            className="field mono"
            placeholder="8.8.8.8"
            value={secondary}
            onChange={(e) => setSecondary(e.target.value)}
            aria-label="Дополнительный DNS"
            style={{ marginBottom: 12 }}
          />
        </>
      ) : null}

      <button
        type="button"
        className="btn primary block"
        disabled={busy}
        onClick={() => void applyDns(preset, primary, secondary)}
      >
        Применить
      </button>
    </>
  );
}

/** Шторка «Оформление». */
export function ThemeSheet() {
  const { theme, accent, setTheme, setAccent } = useApp();

  return (
    <>
      <div className="section-title">
        <span>Тема</span>
      </div>
      <div className="theme-grid">
        {THEMES.map((t) => (
          <button
            key={t.key}
            type="button"
            className={`theme-swatch${theme === t.key ? " on" : ""}`}
            onClick={() => setTheme(t.key)}
            title={t.hint}
          >
            <span className="theme-preview" style={{ background: t.preview }} />
            <span className="theme-name">{t.name}</span>
          </button>
        ))}
      </div>

      <div className="section-title">
        <span>Акцент</span>
      </div>
      <div className="accent-row">
        {ACCENTS.map((a) => (
          <button
            key={a.key}
            type="button"
            className={`accent-dot${accent === a.key ? " on" : ""}`}
            style={{ background: a.color }}
            title={a.name}
            aria-label={a.name}
            onClick={() => setAccent(a.key)}
          />
        ))}
      </div>
    </>
  );
}

/** Остальные шторки — компактные, с полезными действиями. */
export function StatsSheet() {
  const { rt, health, selfCheck, runSelfCheck, exportBundle, busy } = useApp();
  const total = (Number(rt?.sessionDownloadBytes) || 0) + (Number(rt?.sessionUploadBytes) || 0);

  return (
    <>
      <div className="metric-row">
        <span>Скачивание</span>
        <strong className="mono">{formatMbps(Number(rt?.downloadBps) || 0)} Мбит/с</strong>
      </div>
      <div className="metric-row">
        <span>Отдача</span>
        <strong className="mono">{formatMbps(Number(rt?.uploadBps) || 0)} Мбит/с</strong>
      </div>
      <div className="metric-row">
        <span>Передано за сессию</span>
        <strong className="mono">{formatBytes(total)}</strong>
      </div>
      <div className="metric-row">
        <span>Прокси</span>
        <strong className={health?.proxyLocalOk ? "ok" : "muted"}>
          {health?.proxyLocalOk ? "доступен" : "недоступен"}
        </strong>
      </div>
      <div className="metric-row">
        <span>DPI</span>
        <strong className={health?.dpiRunning ? "ok" : "muted"}>
          {health?.dpiRunning ? "запущен" : "остановлен"}
        </strong>
      </div>

      <div style={{ display: "flex", gap: 8, marginTop: 12, flexWrap: "wrap" }}>
        <button type="button" className="btn sm soft" disabled={busy} onClick={() => void runSelfCheck()}>
          <Icon name="refresh" size={14} /> Самопроверка
        </button>
        <button type="button" className="btn sm soft" onClick={() => void exportBundle()}>
          <Icon name="folder" size={14} /> Отчёт
        </button>
      </div>

      {selfCheck ? (
        <pre className="diag-box" style={{ marginTop: 10 }}>{selfCheck}</pre>
      ) : null}
    </>
  );
}

export function LogsSheet() {
  const { logs, clearLogs, openLogsFolder } = useApp();
  return (
    <>
      <div style={{ display: "flex", gap: 8, marginBottom: 10, flexWrap: "wrap" }}>
        <button type="button" className="btn sm soft" onClick={() => void openLogsFolder()}>
          <Icon name="folder" size={14} /> Папка
        </button>
        <button type="button" className="btn sm soft" onClick={() => void clearLogs()}>
          <Icon name="trash" size={14} /> Очистить
        </button>
      </div>
      <pre className="log-box">{String(logs?.coreLogTail || "Ядро ещё не писало в журнал.")}</pre>
      <pre className="log-box" style={{ marginTop: 8 }}>{String(logs?.dpiLogTail || "DPI ещё не писало в журнал.")}</pre>
      {logs?.coreErrors ? <pre className="log-box bad" style={{ marginTop: 8 }}>{String(logs.coreErrors)}</pre> : null}
    </>
  );
}

export function LicenseSheet() {
  const { uiState, activateLicense, busy } = useApp();
  const [code, setCode] = useState("");
  return (
    <>
      <div className="metric-row">
        <span>Статус</span>
        <strong className={uiState?.licenseActive ? "ok" : "muted"}>
          {uiState?.licenseActive ? "активна" : "не активирована"}
        </strong>
      </div>
      <div className="metric-row">
        <span>Тариф</span>
        <strong>{String(uiState?.licenseTier || "free")}</strong>
      </div>
      <div className="metric-row">
        <span>Устройство</span>
        <strong className="mono">{String(uiState?.deviceIdShort || "—")}</strong>
      </div>

      <div className="field-row" style={{ marginTop: 12 }}>
        <input
          className="field"
          placeholder="Ключ активации"
          value={code}
          onChange={(e) => setCode(e.target.value)}
          aria-label="Ключ активации"
        />
        <button
          type="button"
          className="btn sm primary"
          disabled={busy || !code.trim()}
          onClick={() => void activateLicense(code.trim())}
        >
          Активировать
        </button>
      </div>
    </>
  );
}

export function AboutSheet() {
  const { appVersion, guide } = useApp();
  return (
    <>
      <div className="metric-row">
        <span>Версия</span>
        <strong className="mono">{appVersion || "—"}</strong>
      </div>
      <div className="metric-row">
        <span>Ядро</span>
        <strong>sing-box</strong>
      </div>
      <div className="metric-row">
        <span>Обход DPI</span>
        <strong>Zapret</strong>
      </div>
      {guide ? <pre className="diag-box" style={{ marginTop: 12 }}>{guide}</pre> : null}
    </>
  );
}

export function UpdatesSheet() {
  const { updateInfo, checkUpdates, installUpdate, busy } = useApp();
  return (
    <>
      <button type="button" className="btn primary block" disabled={busy} onClick={() => void checkUpdates()}>
        Проверить обновления
      </button>
      {updateInfo ? (
        <div className="verdict" style={{ marginTop: 12 }}>
          <div className="metric-row">
            <span>Версия</span>
            <strong className="mono">{String(updateInfo.latestTag || "")}</strong>
          </div>
          <p className="sheet-sub" style={{ margin: "8px 0 0" }}>{String(updateInfo.message || "")}</p>
          {updateInfo.installerDownloadUrl ? (
            <button type="button" className="btn primary block" style={{ marginTop: 10 }} onClick={() => void installUpdate()}>
              Установить
            </button>
          ) : null}
        </div>
      ) : null}
    </>
  );
}

export function GuideSheet() {
  const { guide } = useApp();
  return <pre className="diag-box">{guide || "Описание появится после загрузки данных."}</pre>;
}

export function SpeedSheet() {
  const { servers, retestAll, busy } = useApp();
  return (
    <>
      <button type="button" className="btn primary block" disabled={busy} onClick={() => void retestAll()}>
        Проверить все узлы
      </button>
      <div className="row-list" style={{ marginTop: 10 }}>
        {servers.map((s) => (
          <div key={String(s.id)} className="server-row">
            <span className="rl-main">
              <strong>{String(s.name || s.host || "")}</strong>
              <small>{String(s.host || "")}</small>
            </span>
            <span className={`rl-r ${latencyTone(Number(s.lastCheckLatencyMs))}`}>
              {formatLatency(Number(s.lastCheckLatencyMs))}
            </span>
          </div>
        ))}
      </div>
    </>
  );
}

export function SubscriptionSheet() {
  const {
    uiState, previewSubscription, importText, applySubscription, busy,
  } = useApp();
  const [url, setUrl] = useState(String(uiState?.subscriptionUrl || ""));
  const [manual, setManual] = useState("");
  const [preview, setPreview] = useState<Record<string, unknown> | null>(null);

  return (
    <>
      <div className="field-row">
        <input
          className="field mono"
          placeholder="https://provider/sub"
          value={url}
          onChange={(e) => setUrl(e.target.value)}
          aria-label="Ссылка подписки"
        />
        <button
          type="button"
          className="btn sm soft"
          disabled={busy || !url.trim()}
          onClick={() => void previewSubscription(url.trim()).then(setPreview)}
        >
          Разобрать
        </button>
      </div>

      {preview ? (
        <div className="verdict">
          <div className="metric-row">
            <span>Формат</span>
            <strong>{String(preview.format || "—")}</strong>
          </div>
          <div className="metric-row">
            <span>Узлов</span>
            <strong>{String(preview.total || 0)}</strong>
          </div>
          {preview.error ? <p className="sheet-sub bad" style={{ margin: "8px 0 0" }}>{String(preview.error)}</p> : null}
        </div>
      ) : null}

      <button
        type="button"
        className="btn primary block"
        style={{ marginTop: 10 }}
        disabled={busy || !url.trim()}
        onClick={() => void applySubscription(url.trim())}
      >
        Применить подписку
      </button>

      <div className="section-title">
        <span>Или вставьте ключи вручную</span>
      </div>
      <textarea
        className="field mono"
        style={{ width: "100%", minHeight: 96, resize: "vertical" }}
        placeholder={"vless://…\nvmess://…\ntrojan://…\nss://…"}
        value={manual}
        onChange={(e) => setManual(e.target.value)}
        aria-label="Ключи вручную"
      />
      <button
        type="button"
        className="btn block"
        style={{ marginTop: 8 }}
        disabled={busy || !manual.trim()}
        onClick={() => void importText(manual.trim())}
      >
        Импортировать список
      </button>

      <p className="sheet-sub" style={{ marginTop: 12, marginBottom: 0 }}>
        Поддерживаются vless, vmess, trojan, ss (и ss2), hysteria, hysteria2, tuic, anytls, socks, http,
        wireguard — плюс подписки в формате base64, Clash YAML и sing-box/Xray JSON. SSR ядром sing-box
        не поддерживается.
      </p>
    </>
  );
}

/** Обёртка над переключателем для повторного использования в шторках. */
export { Switch };
