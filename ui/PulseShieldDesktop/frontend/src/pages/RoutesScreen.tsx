import React, { useMemo, useState } from "react";
import { useApp } from "../context/AppContext";
import { ROUTE_TARGET_LABEL, RouteTarget } from "../types";
import { routingLabel } from "../ui";

type FilterTab = "all" | "user" | RouteTarget;

const DEFAULT_ROW_CAP = 60;

function targetTagTone(target: string): string {
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

function sourceLabel(source: string): string {
  switch (source) {
    case "user":
      return "Моё";
    case "profile":
      return "Профиль";
    case "library":
      return "Библиотека";
    default:
      return source || "Система";
  }
}

export function RoutesScreen() {
  const {
    uiState,
    splitRows,
    setRouting,
    addRule,
    removeRule,
    busy,
  } = useApp();

  const [pattern, setPattern] = useState("");
  const [target, setTarget] = useState<RouteTarget>("proxy");
  const [filterTab, setFilterTab] = useState<FilterTab>("all");
  const [search, setSearch] = useState("");
  const [expanded, setExpanded] = useState(false);

  const routingProfile = String(uiState?.routingProfile || "manual");
  const modeOrder: Record<string, number> = {
    manual: 0,
    "full-proxy": 1,
    global: 2,
  };
  const availableProfiles: string[] = (
    Array.isArray(uiState?.routingProfiles) && uiState.routingProfiles.length > 0
      ? (uiState.routingProfiles as unknown[])
          .map((p) =>
            typeof p === "string"
              ? p
              : String((p as Record<string, unknown>)?.name ?? ""),
          )
          .filter(Boolean)
      : ["manual", "full-proxy", "global"]
  ).sort((a, b) => (modeOrder[a] ?? 99) - (modeOrder[b] ?? 99));

  const counts = useMemo(() => {
    const res = {
      all: splitRows.length,
      user: 0,
      proxy: 0,
      "dpi-bypass": 0,
      direct: 0,
      block: 0,
    };
    for (const r of splitRows) {
      if (String(r.source ?? "") === "user") res.user += 1;
      const t = String(r.target ?? "");
      if (t === "proxy") res.proxy += 1;
      else if (t === "dpi-bypass") res["dpi-bypass"] += 1;
      else if (t === "direct") res.direct += 1;
      else if (t === "block") res.block += 1;
    }
    return res;
  }, [splitRows]);

  const filtered = useMemo(() => {
    const q = search.trim().toLowerCase();
    return splitRows.filter((row) => {
      const rowTarget = String(row.target ?? "");
      const rowSource = String(row.source ?? "");
      if (filterTab === "user" && rowSource !== "user") return false;
      if (
        filterTab !== "all" &&
        filterTab !== "user" &&
        rowTarget !== filterTab
      ) {
        return false;
      }
      if (!q) return true;
      return (
        String(row.pattern ?? "").toLowerCase().includes(q) ||
        String(row.label ?? "").toLowerCase().includes(q) ||
        rowTarget.toLowerCase().includes(q)
      );
    });
  }, [splitRows, filterTab, search]);

  const visibleRows =
    expanded || filtered.length <= DEFAULT_ROW_CAP
      ? filtered
      : filtered.slice(0, DEFAULT_ROW_CAP);

  async function handleAddRule(e: React.FormEvent) {
    e.preventDefault();
    const trimmed = pattern.trim();
    if (!trimmed) return;
    await addRule(trimmed, target);
    setPattern("");
  }

  return (
    <div>
      {/* Routing Profile Switcher */}
      <section className="fx-section fx-section-first">
        <div className="fx-sec-head">
          <div>
            <div className="fx-sec-title">РЕЖИМ МАРШРУТИЗАЦИИ</div>
            <div className="fx-sec-desc">
              {routingProfile === "global"
                ? "Весь внешний трафик (включая .ru, все программы и DNS) идёт через прокси для полного скрытия GEO."
                : routingProfile === "full-proxy"
                  ? "Весь зарубежный интернет идёт через прокси; российские сайты и банки (.ru / .рф / Госуслуги) идут напрямую."
                  : "Через прокси идут только ваши правила; весь остальной трафик идёт напрямую (Steam, Discord, YouTube и др. обрабатываются Обходом DPI по своим спискам)."}
            </div>
          </div>

          <div className="fx-seg" role="group" aria-label="Режим маршрутизации">
            {availableProfiles.map((prof) => (
              <button
                key={prof}
                type="button"
                className={`fx-seg-btn ${routingProfile === prof ? "active" : ""}`}
                disabled={busy}
                onClick={() => void setRouting(prof)}
              >
                {routingLabel(prof)}
              </button>
            ))}
          </div>
        </div>
      </section>

      {/* Inline Rule Builder */}
      <section className="fx-section">
        <div className="fx-sec-head">
          <div>
            <div className="fx-sec-title">ДОБАВИТЬ ПОЛЬЗОВАТЕЛЬСКОЕ ПРАВИЛО</div>
            <div className="fx-sec-desc">
              Пользовательские правила имеют наивысший приоритет (выше процессов и профиля). Действие «Обход DPI»
              направляет домен напрямую в обход прокси для обработки winws2.
            </div>
          </div>
        </div>

        <form className="fx-bar" onSubmit={(e) => void handleAddRule(e)}>
          <input
            type="text"
            className="fx-well-input mono fx-bar-grow"
            placeholder="Домен, *.маска, URL, IP/CIDR или process.exe"
            value={pattern}
            onChange={(e) => setPattern(e.target.value)}
          />

          <select
            className="fx-well-select"
            value={target}
            onChange={(e) => setTarget(e.target.value as RouteTarget)}
            aria-label="Действие маршрута"
          >
            {(Object.keys(ROUTE_TARGET_LABEL) as RouteTarget[]).map((key) => (
              <option key={key} value={key}>
                {ROUTE_TARGET_LABEL[key]}
              </option>
            ))}
          </select>

          <button
            type="submit"
            className="fx-btn fx-btn-primary"
            disabled={busy || !pattern.trim()}
          >
            Добавить
          </button>
        </form>
      </section>

      {/* Rules Filter & Full-bleed Table */}
      <section className="fx-section">
        <div className="fx-sec-head">
          <div className="fx-bar fx-bar-grow">
            <div className="fx-seg" role="group" aria-label="Фильтр правил">
              <button
                type="button"
                className={`fx-seg-btn ${filterTab === "all" ? "active" : ""}`}
                onClick={() => setFilterTab("all")}
              >
                Все ({counts.all})
              </button>
              <button
                type="button"
                className={`fx-seg-btn ${filterTab === "user" ? "active" : ""}`}
                onClick={() => setFilterTab("user")}
              >
                Мои ({counts.user})
              </button>
              <button
                type="button"
                className={`fx-seg-btn ${filterTab === "proxy" ? "active" : ""}`}
                onClick={() => setFilterTab("proxy")}
              >
                Прокси ({counts.proxy})
              </button>
              <button
                type="button"
                className={`fx-seg-btn ${filterTab === "dpi-bypass" ? "active" : ""}`}
                onClick={() => setFilterTab("dpi-bypass")}
              >
                DPI ({counts["dpi-bypass"]})
              </button>
              <button
                type="button"
                className={`fx-seg-btn ${filterTab === "direct" ? "active" : ""}`}
                onClick={() => setFilterTab("direct")}
              >
                Напрямую ({counts.direct})
              </button>
              <button
                type="button"
                className={`fx-seg-btn ${filterTab === "block" ? "active" : ""}`}
                onClick={() => setFilterTab("block")}
              >
                Блок ({counts.block})
              </button>
            </div>

            <input
              type="text"
              className="fx-well-input"
              style={{ width: 220, maxWidth: "100%" }}
              placeholder="Фильтр по шаблону…"
              value={search}
              onChange={(e) => setSearch(e.target.value)}
            />
          </div>

          {filtered.length > DEFAULT_ROW_CAP && (
            <button
              type="button"
              className="fx-btn fx-btn-sm"
              onClick={() => setExpanded((prev) => !prev)}
            >
              {expanded
                ? `Свернуть до ${DEFAULT_ROW_CAP}`
                : `Показать все (${filtered.length})`}
            </button>
          )}
        </div>

        <div className="fx-list-wrap">
          <div className="fx-thead">
            <span style={{ flex: 1.4, minWidth: 160 }}>Шаблон</span>
            <span style={{ width: 120 }}>Назначение</span>
            <span style={{ width: 110 }}>Источник</span>
            <span className="fx-cell-grow">Описание</span>
            <span style={{ width: 92, textAlign: "right" }}>Действие</span>
          </div>

          {visibleRows.length === 0 ? (
            <div className="fx-empty">
              {splitRows.length === 0
                ? "Правила маршрутизации отсутствуют."
                : "Нет правил, подходящих под выбранный фильтр."}
            </div>
          ) : (
            visibleRows.map((row, idx) => {
              const pat = String(row.pattern ?? "");
              const t = String(row.target ?? "direct");
              const src = String(row.source ?? "");
              const lbl = String(row.label ?? "");
              const isUser = src === "user" && !row.system;
              const userIdx = Number(row.userIndex ?? -1);

              return (
                <div key={`${src}-${pat}-${idx}`} className="fx-trow">
                  <div
                    style={{ flex: 1.4, minWidth: 160 }}
                    className="fx-mono fx-truncate fx-select-text"
                    title={pat}
                  >
                    {pat}
                  </div>

                  <div style={{ width: 120 }}>
                    <span className={`fx-tag ${targetTagTone(t)}`}>
                      {ROUTE_TARGET_LABEL[t as RouteTarget] || t}
                    </span>
                  </div>

                  <div style={{ width: 110 }}>
                    <span className={`fx-tag ${isUser ? "accent" : ""}`}>
                      {sourceLabel(src)}
                    </span>
                  </div>

                  <div
                    className="fx-cell-grow fx-truncate"
                    style={{ fontSize: 12, color: "var(--text-secondary)" }}
                    title={lbl}
                  >
                    {lbl || "—"}
                  </div>

                  <div style={{ width: 92, textAlign: "right" }}>
                    {isUser && userIdx >= 0 ? (
                      <button
                        type="button"
                        className="fx-btn fx-btn-sm fx-btn-ghost"
                        disabled={busy}
                        onClick={() => void removeRule(userIdx)}
                      >
                        Удалить
                      </button>
                    ) : (
                      <span
                        className="fx-mono"
                        style={{ fontSize: 11, color: "var(--text-muted)" }}
                      >
                        {row.system ? "система" : "—"}
                      </span>
                    )}
                  </div>
                </div>
              );
            })
          )}
        </div>
      </section>
    </div>
  );
}
