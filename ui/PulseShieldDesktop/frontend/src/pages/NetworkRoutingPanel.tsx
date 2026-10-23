import { useCallback, useEffect, useMemo, useState } from "react";
import { AddUserRoutingRule, ApplyRuntime, GetRoutingSplitView, RemoveUserRoutingRule, SetStrictRoutingMode } from "../../wailsjs/go/main/App";
import { PageTip } from "../components/PageTip";

type Props = {
  visible: boolean;
  busy: boolean;
  proxyEnabled: boolean;
  strictRoutingMode: boolean;
  refreshParent?: () => Promise<void>;
  notify: (msg: string, variant?: "info" | "error", openLogs?: boolean) => void;
  onImportRouting?: () => void | Promise<void>;
  onExportRouting?: () => void | Promise<void>;
};

type SplitRow = { pattern: string; target: string; system: boolean; label?: string; source?: string };

type TargetAction = "proxy" | "direct" | "dpi-bypass" | "block";

const TARGET_OPTIONS: { value: TargetAction; label: string; hint: string }[] = [
  { value: "proxy",     label: "Через прокси",  hint: "Трафик уходит через VLESS-сервер" },
  { value: "direct",    label: "Напрямую",       hint: "Трафик идёт минуя прокси" },
  { value: "dpi-bypass",label: "DPI-обход",      hint: "Прямой выход; результат обхода зависит от запущенного DPI и профиля" },
  { value: "block",     label: "Блокировать",    hint: "Соединения на этот адрес отклоняются" },
];

const SPLIT_HELP =
  "Порядок в ядре: процессы из карты → ваши правила → активный профиль → библиотека (если не «чистый гейминг»). Первое совпадение побеждает; без совпадений смотрите final профиля и proxyMode (ROUTING_GUIDE.md).";

function GamingShieldIcon() {
  return (
    <svg viewBox="0 0 24 24" fill="none" xmlns="http://www.w3.org/2000/svg" aria-hidden>
      <path
        d="M12 3 19 6.2v5.3c0 4.5-2.6 8-7 10.8L12 21l-2-1.2C5.6 16 3 12.5 3 8V6.2L12 3Z"
        stroke="currentColor"
        strokeWidth="1.35"
        strokeLinejoin="round"
      />
      <path d="M13.25 8H10.5l2.1 4.8h-2l-.85 3.5 4.55-5.6h-2.05l1.9-2.7Z" fill="currentColor" />
    </svg>
  );
}

function rowMatches(row: SplitRow, q: string): boolean {
  if (!q) return true;
  const hay = `${row.pattern} ${row.label || ""} ${row.target}`.toLowerCase();
  return hay.includes(q);
}

function targetLabel(t: string) {
  switch (t) {
    case "dpi-bypass": return "DPI";
    case "block":      return "блок";
    case "proxy":      return "proxy";
    case "direct":     return "direct";
    default:           return t;
  }
}

function rowClass(src: string | undefined) {
  if (src === "user")    return "routing-split-li routing-split-li--user";
  if (src === "library") return "routing-split-li routing-split-li--library";
  if (src === "profile") return "routing-split-li routing-split-li--profile";
  return "routing-split-li";
}

function sourceBadge(src: string | undefined) {
  if (src === "user")    return <span className="routing-src-badge routing-src-user">вы</span>;
  if (src === "profile") return <span className="routing-src-badge routing-src-profile">профиль</span>;
  if (src === "library")
    return <span className="routing-src-badge routing-src-library" title="Библиотека Flux">🔒</span>;
  return null;
}

export function NetworkRoutingPanel({
  visible,
  busy,
  proxyEnabled,
  strictRoutingMode,
  refreshParent,
  notify,
  onImportRouting,
  onExportRouting,
}: Props) {
  const [entries, setEntries] = useState<SplitRow[]>([]);
  const [pattern, setPattern] = useState("");
  const [targetDraft, setTargetDraft] = useState<TargetAction>("proxy");
  const [filterQ, setFilterQ] = useState("");
  const [loading, setLoading] = useState(false);
  const [loadError, setLoadError] = useState<string | null>(null);

  const reload = useCallback(async () => {
    setLoading(true);
    setLoadError(null);
    try {
      const rows = await GetRoutingSplitView();
      setEntries(rows || []);
    } catch (e: unknown) {
      setEntries([]);
      const msg = e instanceof Error ? e.message : String(e);
      setLoadError(msg.trim() || "Ошибка загрузки данных");
      notify(`Маршруты: ${msg}`, "error", true);
    } finally {
      setLoading(false);
    }
  }, [notify]);

  useEffect(() => {
    if (!visible) return;
    void reload();
  }, [visible, reload, strictRoutingMode]);

  const { proxyRows, directRows, blockedRows } = useMemo(() => {
    const proxy: SplitRow[] = [];
    const direct: SplitRow[] = [];
    const blocked: SplitRow[] = [];
    for (const e of entries) {
      if (e.target === "proxy")                           proxy.push(e);
      else if (e.target === "block")                     blocked.push(e);
      else if (e.target === "direct" || e.target === "dpi-bypass") direct.push(e);
      else direct.push(e);
    }
    return { proxyRows: proxy, directRows: direct, blockedRows: blocked };
  }, [entries]);

  const qNorm = filterQ.trim().toLowerCase();
  const proxyFiltered   = useMemo(() => proxyRows.filter((r) => rowMatches(r, qNorm)),   [proxyRows,   qNorm]);
  const directFiltered  = useMemo(() => directRows.filter((r) => rowMatches(r, qNorm)),  [directRows,  qNorm]);
  const blockedFiltered = useMemo(() => blockedRows.filter((r) => rowMatches(r, qNorm)), [blockedRows, qNorm]);

  async function addRule() {
    const p = pattern.trim();
    if (!p) {
      notify("Введите домен, .суффикс, *.host или CIDR", "error");
      return;
    }
    try {
      await AddUserRoutingRule(p, targetDraft);
      notify(`Добавлено: ${p} → ${targetDraft}`);
      setPattern("");
      await reload();
      if (proxyEnabled) await ApplyRuntime();
    } catch (e: unknown) {
      notify(String(e), "error", true);
    }
  }

  async function removeRow(row: SplitRow) {
    if (row.source !== "user") return;
    try {
      await RemoveUserRoutingRule(row.pattern);
      notify(`Удалено: ${row.pattern}`);
      await reload();
      if (proxyEnabled) await ApplyRuntime();
    } catch (e: unknown) {
      notify(String(e), "error", true);
    }
  }

  async function toggleStrict(next: boolean) {
    try {
      await SetStrictRoutingMode(next);
      notify(next ? "Чистый гейминг: библиотека не подмешивается" : "Библиотека снова подмешивается");
      await reload();
      if (refreshParent) await refreshParent();
      if (proxyEnabled) await ApplyRuntime();
    } catch (e: unknown) {
      notify(String(e), "error", true);
    }
  }

  function renderRows(rows: SplitRow[]) {
    if (rows.length === 0) {
      return <p className="muted routing-split-empty-col">Нет правил</p>;
    }
    return (
      <ul className="routing-split-ul routing-split-ul-single">
        {rows.map((r, i) => (
          <li key={`${r.source}-${i}-${r.pattern}-${r.label}`} className={rowClass(r.source)}>
            <code>{r.pattern}</code>
            <span className="muted routing-split-meta">
              · {targetLabel(r.target)}
              {sourceBadge(r.source)}
            </span>
            {r.source === "user" ? (
              <button
                type="button"
                className="btn btn-soft btn-compact routing-split-rm"
                disabled={busy}
                onClick={() => void removeRow(r)}
                title="Удалить"
              >
                ✕
              </button>
            ) : null}
          </li>
        ))}
      </ul>
    );
  }

  const selectedTarget = TARGET_OPTIONS.find((o) => o.value === targetDraft);

  return (
    <section className="routing-split-shell">
      <h4 className="routing-split-heading dpi-section-title">
        Умный сплит
        <PageTip text={SPLIT_HELP} label="Маршрутизация" />
      </h4>

      <button
        type="button"
        role="switch"
        aria-checked={strictRoutingMode}
        disabled={busy}
        className={"routing-game-card" + (strictRoutingMode ? " routing-game-card--on" : "")}
        onClick={() => void toggleStrict(!strictRoutingMode)}
      >
        <span className="routing-game-card-icon"><GamingShieldIcon /></span>
        <div className="routing-game-card-body">
          <div className="routing-game-card-title-row">
            <span className="routing-game-card-title">Чистый гейминг</span>
            <PageTip
              text="Не подмешивать встроенную библиотеку Flux — только карта процессов, ваши правила и активный профиль."
              label="Чистый гейминг"
            />
          </div>
          <p className="routing-game-card-sub">Использовать только ваши правила и профиль</p>
        </div>
      </button>

      <div className="routing-split-toolbar">
        <input
          type="search"
          className="routing-split-search"
          placeholder="Поиск по доменам, суффиксам, CIDR…"
          value={filterQ}
          disabled={busy}
          onChange={(e) => setFilterQ(e.target.value)}
          aria-label="Поиск по правилам"
        />
        {onImportRouting ? (
          <button type="button" className="btn btn-soft btn-compact" disabled={busy} onClick={() => void onImportRouting()}>
            Импорт
          </button>
        ) : null}
        {onExportRouting ? (
          <button type="button" className="btn btn-soft btn-compact" disabled={busy} onClick={() => void onExportRouting()}>
            Экспорт
          </button>
        ) : null}
      </div>

      {loadError ? (
        <div className="routing-split-error-row muted">
          <span>{loadError}</span>
          <button type="button" className="btn btn-soft btn-compact" disabled={busy || loading} onClick={() => void reload()}>
            Повторить
          </button>
        </div>
      ) : null}
      {loading ? <p className="muted">Загрузка…</p> : null}

      <details className="routing-split-collapse" open>
        <summary className="routing-split-summary">
          <span className="routing-split-summary-title">Текущие правила</span>
          <span className="routing-split-summary-meta">
            {proxyFiltered.length} прокси · {directFiltered.length} напрямую
            {blockedFiltered.length > 0 ? ` · ${blockedFiltered.length} блок` : ""}
          </span>
        </summary>

        <div className="routing-split-grid routing-split-grid-network">
          <div className="routing-split-col">
            <div className="routing-split-col-title">Через прокси ({proxyFiltered.length})</div>
            <div className="routing-split-table-wrap routing-split-table-wrap-dual">{renderRows(proxyFiltered)}</div>
          </div>
          <div className="routing-split-col">
            <div className="routing-split-col-title">Напрямую / DPI ({directFiltered.length})</div>
            <div className="routing-split-table-wrap routing-split-table-wrap-dual">{renderRows(directFiltered)}</div>
          </div>
          {blockedFiltered.length > 0 ? (
            <div className="routing-split-col routing-split-col--blocked">
              <div className="routing-split-col-title">Заблокировано ({blockedFiltered.length})</div>
              <div className="routing-split-table-wrap routing-split-table-wrap-dual">{renderRows(blockedFiltered)}</div>
            </div>
          ) : null}
        </div>
      </details>

      <div className="routing-split-add routing-split-add-sticky">
        <input
          className="routing-split-input"
          placeholder="Домен, .суффикс, *.host или CIDR"
          value={pattern}
          disabled={busy}
          onChange={(e) => setPattern(e.target.value)}
          onKeyDown={(e) => { if (e.key === "Enter") void addRule(); }}
        />
        <select
          className="select routing-split-target-select"
          value={targetDraft}
          disabled={busy}
          onChange={(e) => setTargetDraft(e.target.value as TargetAction)}
          title={selectedTarget?.hint}
        >
          {TARGET_OPTIONS.map((o) => (
            <option key={o.value} value={o.value} title={o.hint}>{o.label}</option>
          ))}
        </select>
        <button type="button" className="btn btn-primary" disabled={busy} onClick={() => void addRule()}>
          Добавить
        </button>
      </div>
      <p className="muted routing-split-add-hint">
        {selectedTarget?.hint}
      </p>
    </section>
  );
}
