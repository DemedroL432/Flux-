import { useMemo, useState } from "react";
import { useApp } from "../context/AppContext";
import { Icon } from "../icons";
import { ScreenHead } from "../ui";
import { ROUTE_TARGET_LABEL, type RouteTarget } from "../types";
import { routingLabel } from "./ShieldScreen";

type Row = {
  pattern: string;
  target: string;
  system?: boolean;
  label?: string;
  source?: string;
  userIndex?: number;
};

/** Порядок приоритета: пользовательские правила выше профиля, профиль выше библиотеки. */
function sourceRank(src?: string): number {
  switch (src) {
    case "user":
      return 0;
    case "profile":
      return 1;
    case "library":
      return 2;
    default:
      return 3;
  }
}

const MODE_CARDS = [
  {
    key: "full-proxy",
    title: "Полный прокси",
    blurb: "Весь трафик через туннель, кроме локального и РФ",
    icon: "shield",
  },
  {
    key: "manual",
    title: "Ручные правила",
    blurb: "В туннель уходит только то, что вы добавили",
    icon: "routes",
  },
];

const FILTERS: { key: "all" | "user" | "dpi" | "proxy"; label: string }[] = [
  { key: "all", label: "Все" },
  { key: "user", label: "Мои" },
  { key: "proxy", label: "Через прокси" },
  { key: "dpi", label: "Обход DPI" },
];

export function RoutesScreen() {
  const { splitRows, routingProfile, setRouting, removeRule, busy, openSheet, uiState } = useApp();
  const [expanded, setExpanded] = useState(false);
  const [filter, setFilter] = useState<"all" | "user" | "dpi" | "proxy">("all");
  const [query, setQuery] = useState("");

  const all = useMemo(() => {
    const rows = ((splitRows || []) as unknown as Row[]).slice();
    return rows.sort((a, b) => sourceRank(a.source) - sourceRank(b.source));
  }, [splitRows]);

  const rows = useMemo(() => {
    let out = all;
    if (filter === "user") out = out.filter((r) => r.source === "user");
    else if (filter === "dpi") out = out.filter((r) => r.target === "dpi-bypass");
    else if (filter === "proxy") out = out.filter((r) => r.target === "proxy");
    const q = query.trim().toLowerCase();
    if (q) out = out.filter((r) => (r.pattern || "").toLowerCase().includes(q));
    return out;
  }, [all, filter, query]);

  const LIMIT = 8;
  const visible = expanded ? rows : rows.slice(0, LIMIT);
  const hidden = Math.max(0, rows.length - LIMIT);

  const activeMode = routingProfile === "full-proxy" || routingProfile === "default" ? "full-proxy" : "manual";

  // Сколько доменов реально уйдёт в обход DPI — это и есть содержимое custom-dpi.
  const dpiCount = useMemo(() => all.filter((r) => r.target === "dpi-bypass" && r.source === "user").length, [all]);
  const userCount = useMemo(() => all.filter((r) => r.source === "user").length, [all]);

  return (
    <>
      <ScreenHead
        title="Маршруты"
        sub="Что идёт через прокси, а что напрямую"
        right={
          <button type="button" className="chip mute" onClick={() => openSheet("verify")}>
            Проверить сайт <Icon name="chev" size={14} />
          </button>
        }
      />

      <div className="scr scroll">
        <div className="profiles">
          {MODE_CARDS.map((m) => {
            const on = activeMode === m.key;
            return (
              <button
                key={m.key}
                type="button"
                className={`profile${on ? " on" : ""}`}
                disabled={busy}
                onClick={() => void setRouting(m.key)}
                aria-pressed={on}
              >
                <span className="pi">
                  <Icon name={m.icon} />
                </span>
                <b>{m.title}</b>
                <small>{m.blurb}</small>
                <span className="tick">
                  <Icon name="check" size={12} />
                </span>
              </button>
            );
          })}
        </div>

        <div className="tile">
          <div className="tile-head">
            <h3>Правила</h3>
            <button type="button" className="more" onClick={() => openSheet("rule")} disabled={busy}>
              Добавить <Icon name="plus" size={14} />
            </button>
          </div>

          <div className="rules-toolbar">
            <input
              className="field"
              placeholder="Найти правило"
              value={query}
              onChange={(e) => setQuery(e.target.value)}
              aria-label="Поиск по правилам"
            />
          </div>

          <div className="seg" role="tablist" aria-label="Фильтр правил">
            {FILTERS.map((f) => (
              <button
                key={f.key}
                type="button"
                role="tab"
                aria-selected={filter === f.key}
                className={filter === f.key ? "on" : ""}
                onClick={() => setFilter(f.key)}
              >
                {f.label}
              </button>
            ))}
          </div>

          {rows.length === 0 ? (
            <div className="empty">
              <Icon name="routes" size={22} />
              <p>
                {all.length === 0
                  ? "Правил пока нет. Добавьте сайт, который должен идти через прокси или в обход DPI."
                  : "Ничего не нашлось по этому фильтру."}
              </p>
              {all.length === 0 ? (
                <button type="button" className="btn sm primary" onClick={() => openSheet("rule")}>
                  Добавить правило
                </button>
              ) : null}
            </div>
          ) : (
            <div className="rules-list">
              {visible.map((r) => {
                const tgt = (r.target || "direct") as RouteTarget;
                const isUser = r.source === "user";
                return (
                  <div
                    key={`${r.source}-${r.userIndex ?? r.label}-${r.pattern}`}
                    className="rule-row"
                  >
                    <span className={`src-badge ${r.source || "profile"}`}>
                      {isUser ? "моё" : r.source === "library" ? "библ." : "профиль"}
                    </span>
                    <span className="rl-t" title={r.pattern}>
                      {r.pattern}
                    </span>
                    <span className={`chip ${tgt}`}>{ROUTE_TARGET_LABEL[tgt] || tgt}</span>
                    {isUser && typeof r.userIndex === "number" ? (
                      <button
                        type="button"
                        className="rule-del"
                        title="Удалить это правило"
                        aria-label={`Удалить правило ${r.pattern}`}
                        disabled={busy}
                        onClick={() => void removeRule(r.userIndex as number)}
                      >
                        <Icon name="trash" size={14} />
                      </button>
                    ) : (
                      <span className="rule-del ghost" aria-hidden />
                    )}
                  </div>
                );
              })}
            </div>
          )}

          {hidden > 0 ? (
            <button type="button" className="btn sm soft block" onClick={() => setExpanded((v) => !v)}>
              {expanded ? "Свернуть" : `Показать все (${rows.length})`}
            </button>
          ) : null}
        </div>

        <div className="tile">
          <div className="tile-head">
            <h3>Обход DPI</h3>
            <button type="button" className="more" onClick={() => openSheet("tune-dpi")} disabled={busy}>
              Настроить <Icon name="chev" size={14} />
            </button>
          </div>
          <div className="metric-row">
            <span>Домены в custom-dpi</span>
            <strong className={dpiCount > 0 ? "ok" : "muted"}>{dpiCount}</strong>
          </div>
          <div className="metric-row">
            <span>Всего моих правил</span>
            <strong>{userCount}</strong>
          </div>
          <div className="metric-row">
            <span>Модуль DPI</span>
            <strong className={uiState?.dpiEnabled ? "ok" : "muted"}>
              {uiState?.dpiEnabled ? "включён" : "выключен"}
            </strong>
          </div>
          <p className="sheet-sub" style={{ margin: "4px 16px 12px" }}>
            Сайты с действием «Обход DPI» попадают в список custom-dpi и обходятся без прокси.
          </p>
        </div>

        <div className="tile">
          <div className="tile-head">
            <h3>Режим</h3>
            <span className="mute">{routingLabel(String(uiState?.routingProfile || ""))}</span>
          </div>
          <div className="metric-row">
            <span>Остальной трафик</span>
            <strong>{activeMode === "full-proxy" ? "через прокси" : "напрямую"}</strong>
          </div>
        </div>
      </div>
    </>
  );
}
