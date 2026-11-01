import { useCallback, useEffect, useMemo, useState } from "react";
import { useApp } from "../context/AppContext";
import { Icon } from "../icons";
import { ScreenHead } from "../ui";
import { detectCountry } from "../flags";
import { formatLatency, latencyTone } from "../format";

type Server = {
  id?: string;
  name?: string;
  host?: string;
  port?: string;
  protocol?: string;
  source?: string;
  fromSubscription?: boolean;
  lastCheckOk?: boolean;
  lastCheckLatencyMs?: number;
  lastCheckDownloadMbps?: number;
  recommended?: boolean;
  premium?: boolean;
};

export type SortKey = "ping" | "name" | "country";

const SORTS: { key: SortKey; label: string }[] = [
  { key: "ping", label: "По пингу" },
  { key: "name", label: "По имени" },
  { key: "country", label: "По стране" },
];

/** Сравнение по пингу: неотвечающие узлы всегда в конце. */
function comparePing(a: Server, b: Server): number {
  const pa = Number(a.lastCheckLatencyMs) || 0;
  const pb = Number(b.lastCheckLatencyMs) || 0;
  const oka = a.lastCheckOk && pa > 0;
  const okb = b.lastCheckOk && pb > 0;
  if (oka !== okb) return oka ? -1 : 1;
  if (oka && okb && pa !== pb) return pa - pb;
  return String(a.name || a.host || "").localeCompare(String(b.name || b.host || ""), "ru");
}

/** Сравнение по стране: сначала сгруппированные по стране, внутри — по пингу. */
function compareCountry(a: Server, b: Server): number {
  const ca = detectCountry(a.name, a.host)?.name || "яяя";
  const cb = detectCountry(b.name, b.host)?.name || "яяя";
  const d = ca.localeCompare(cb, "ru");
  if (d !== 0) return d;
  return comparePing(a, b);
}

export function useServerList() {
  const { servers } = useApp();
  const [query, setQuery] = useState("");
  const [sort, setSort] = useState<SortKey>("ping");
  const [groupByCountry, setGroupByCountry] = useState(false);

  const all = useMemo(() => ((servers || []) as Server[]).slice(), [servers]);

  const filtered = useMemo(() => {
    const q = query.trim().toLowerCase();
    if (!q) return all;
    return all.filter((s) => {
      const country = detectCountry(s.name, s.host)?.name.toLowerCase() || "";
      const hay = `${s.name || ""} ${s.host || ""} ${s.protocol || ""} ${country}`.toLowerCase();
      return hay.includes(q);
    });
  }, [all, query]);

  const sorted = useMemo(() => {
    const out = filtered.slice();
    if (sort === "ping") out.sort(comparePing);
    else if (sort === "country") out.sort(compareCountry);
    else out.sort((a, b) => String(a.name || "").localeCompare(String(b.name || ""), "ru"));
    return out;
  }, [filtered, sort]);

  const groups = useMemo(() => {
    if (!groupByCountry) return null;
    const map = new Map<string, Server[]>();
    for (const s of sorted) {
      const key = detectCountry(s.name, s.host)?.name || "Не определена";
      const arr = map.get(key);
      if (arr) arr.push(s);
      else map.set(key, [s]);
    }
    return Array.from(map.entries());
  }, [sorted, groupByCountry]);

  return { query, setQuery, sort, setSort, groupByCountry, setGroupByCountry, all, sorted, groups, total: all.length, shown: sorted.length };
}

export function ServersScreen() {
  const {
    busy, selectServer, removeServer, retestAll, openSheet, activeServer, importFromClipboard,
  } = useApp();
  const list = useServerList();
  const [confirmId, setConfirmId] = useState<string | null>(null);

  // Сбрасываем подтверждение удаления при уходе с экрана.
  useEffect(() => () => setConfirmId(null), []);

  const onRemove = useCallback(
    async (id: string) => {
      await removeServer(id);
      setConfirmId(null);
    },
    [removeServer]
  );

  const renderRow = (s: Server) => {
    const id = String(s.id || "");
    const isActive = !!activeServer && String(activeServer.id) === id;
    const country = detectCountry(s.name, s.host);
    const tone = latencyTone(Number(s.lastCheckLatencyMs));
    const ping = Number(s.lastCheckLatencyMs) || 0;
    const ok = !!s.lastCheckOk && ping > 0;

    return (
      <div key={id} className={`server-row${isActive ? " sel" : ""}`}>
        <button
          type="button"
          className="srv-pick"
          disabled={busy}
          onClick={() => void selectServer(id)}
          aria-label={`Выбрать ${s.name || s.host}`}
        >
          <span className="srv-flag" title={country?.name || "Страна не определена"}>
            {country ? country.flag : "🌐"}
          </span>
          <span className="srv-main">
            <span className="srv-name">
              {String(s.name || s.host || "")}
              {isActive ? <span className="srv-active-dot" title="Активный узел" /> : null}
            </span>
            <span className="srv-sub">
              {String(s.host || "")}
              {s.port ? `:${s.port}` : ""}
              {s.protocol ? ` · ${s.protocol}` : ""}
            </span>
          </span>
        </button>

        <span className={`srv-ping ${tone}`}>
          {ok ? formatLatency(ping) : s.lastCheckOk === false ? "нет ответа" : "—"}
        </span>

        <button
          type="button"
          className={`rule-del${confirmId === id ? " confirm" : ""}`}
          title={confirmId === id ? "Нажмите ещё раз, чтобы удалить" : "Удалить узел"}
          aria-label={`Удалить ${s.name || s.host}`}
          disabled={busy}
          onClick={() => (confirmId === id ? void onRemove(id) : setConfirmId(id))}
          onBlur={() => setConfirmId((v) => (v === id ? null : v))}
        >
          <Icon name={confirmId === id ? "warn" : "trash"} size={14} />
        </button>
      </div>
    );
  };

  return (
    <>
      <ScreenHead
        title="Серверы"
        sub={`${list.total} узлов${list.shown !== list.total ? ` · найдено ${list.shown}` : ""}`}
        right={
          <button type="button" className="chip mute" onClick={() => openSheet("server")}>
            Подписка <Icon name="chev" size={14} />
          </button>
        }
      />

      <div className="scr scroll">
        <div className="tile">
          <div className="field-row">
            <input
              className="field"
              placeholder="Поиск по имени, хосту или стране"
              value={list.query}
              onChange={(e) => list.setQuery(e.target.value)}
              aria-label="Поиск сервера"
            />
            {list.query ? (
              <button type="button" className="btn sm soft" onClick={() => list.setQuery("")} title="Сбросить поиск">
                <Icon name="close" size={14} />
              </button>
            ) : null}
          </div>

          <div className="seg" role="tablist" aria-label="Сортировка серверов">
            {SORTS.map((s) => (
              <button
                key={s.key}
                type="button"
                role="tab"
                aria-selected={list.sort === s.key}
                className={list.sort === s.key ? "on" : ""}
                onClick={() => list.setSort(s.key)}
              >
                {s.label}
              </button>
            ))}
          </div>

          <div className="srv-actions">
            <button type="button" className="btn sm soft" disabled={busy} onClick={() => void retestAll()}>
              <Icon name="refresh" size={14} /> Проверить все
            </button>
            <button type="button" className="btn sm soft" disabled={busy} onClick={() => void importFromClipboard()}>
              <Icon name="copy" size={14} /> Из буфера
            </button>
            <button type="button" className="btn sm soft" onClick={() => openSheet("speed")}>
              <Icon name="gauge" size={14} /> Тест
            </button>
          </div>
        </div>

        <div className="tile">
          <div className="tile-head">
            <h3>Узлы</h3>
            <button
              type="button"
              className="more"
              onClick={() => list.setGroupByCountry(!list.groupByCountry)}
              title="Группировать по стране"
            >
              {list.groupByCountry ? "Без групп" : "По странам"}
            </button>
          </div>

          {list.total === 0 ? (
            <div className="empty">
              <Icon name="server" size={22} />
              <p>Серверов нет. Импортируйте подписку или вставьте ссылку вручную.</p>
              <div style={{ display: "flex", gap: 8, flexWrap: "wrap", justifyContent: "center" }}>
                <button type="button" className="btn sm primary" onClick={() => openSheet("subscription")}>
                  Подписка и ключи
                </button>
                <button type="button" className="btn sm soft" onClick={() => openSheet("server")}>
                  Добавить узлы
                </button>
              </div>
            </div>
          ) : list.sorted.length === 0 ? (
            <div className="empty">
              <Icon name="info" size={20} />
              <p>По запросу «{list.query}» ничего не найдено.</p>
              <button type="button" className="btn sm soft" onClick={() => list.setQuery("")}>
                Сбросить поиск
              </button>
            </div>
          ) : list.groups ? (
            <div className="srv-groups">
              {list.groups.map(([country, items]) => (
                <div key={country} className="srv-group">
                  <div className="srv-group-head">
                    <span>{country}</span>
                    <span className="mute">{items.length}</span>
                  </div>
                  {items.map(renderRow)}
                </div>
              ))}
            </div>
          ) : (
            <div className="row-list">{list.sorted.map(renderRow)}</div>
          )}
        </div>
      </div>
    </>
  );
}
