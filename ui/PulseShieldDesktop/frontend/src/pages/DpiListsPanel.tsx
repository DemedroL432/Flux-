import { useCallback, useEffect, useMemo, useState } from "react";
import { GetAvailableDpiLists, GetZapretListsPathHint, ToggleDpiList } from "../../wailsjs/go/main/App";

type Props = {
  visible: boolean;
  busy: boolean;
  notify: (msg: string, variant?: "info" | "error", openLogs?: boolean) => void;
  refreshState: () => Promise<void>;
};

type DpiListRow = { name: string; description: string; enabled: boolean };
type DpiCat = "media" | "social" | "games" | "other";

function listCategory(baseName: string): DpiCat {
  const n = baseName.replace(/\.txt$/i, "").toLowerCase();
  if (["steam", "roblox", "riot", "lol", "tankix", "itch", "games"].some((k) => n.includes(k))) return "games";
  if (["instagram", "facebook", "twitter", "whatsapp"].some((k) => n.includes(k))) return "social";
  if (["youtube", "google", "discord", "telegram", "twitch", "soundcloud", "github", "amazon",
       "claude", "speedtest", "rutracker", "rutor", "netflix"].some((k) => n.includes(k))) return "media";
  return "other";
}

const CAT_META: { id: DpiCat; label: string; icon: string }[] = [
  { id: "media",  label: "Медиа",    icon: "▶" },
  { id: "social", label: "Соцсети",  icon: "◎" },
  { id: "games",  label: "Игры",     icon: "▲" },
  { id: "other",  label: "Прочее",   icon: "⊞" },
];

/* Checkmark icon */
function CheckIcon() {
  return (
    <svg viewBox="0 0 12 12" width="10" height="10" fill="none" stroke="currentColor" strokeWidth="2" strokeLinecap="round" strokeLinejoin="round">
      <path d="M2 6l3 3 5-5"/>
    </svg>
  );
}

export function DpiListsPanel({ visible, busy, notify, refreshState }: Props) {
  const [lists, setLists] = useState<DpiListRow[]>([]);
  const [loading, setLoading] = useState(false);
  const [loadError, setLoadError] = useState<string | null>(null);
  const [loadedOnce, setLoadedOnce] = useState(false);
  const [listsHintPath, setListsHintPath] = useState("");
  const [tab, setTab] = useState<DpiCat>("media");

  useEffect(() => {
    void Promise.resolve().then(() => GetZapretListsPathHint())
      .then(setListsHintPath)
      .catch(() => setListsHintPath("zapret\\lists"));
  }, []);

  const reload = useCallback(async () => {
    setLoading(true);
    setLoadError(null);
    try {
      const rows = await GetAvailableDpiLists();
      setLists(rows || []);
    } catch (e: unknown) {
      setLists([]);
      setLoadError("Ошибка загрузки");
      notify(`Списки DPI: ${String(e)}`, "error", true);
    } finally {
      setLoading(false);
      setLoadedOnce(true);
    }
  }, [notify]);

  useEffect(() => {
    if (!visible) return;
    void reload();
  }, [visible, reload]);

  const byCategory = useMemo(() => {
    const result: Record<DpiCat, DpiListRow[]> = { media: [], social: [], games: [], other: [] };
    for (const L of lists) result[listCategory(L.name)].push(L);
    return result;
  }, [lists]);

  const filtered = byCategory[tab];

  async function onToggle(name: string, next: boolean) {
    try {
      await ToggleDpiList(name, next);
      notify(next ? `Список «${name}» включён` : `Список «${name}» выключен`);
      await refreshState();
      await reload();
    } catch (e: unknown) {
      notify(String(e), "error", true);
      await reload();
    }
  }

  const showGlobalEmpty = loadedOnce && !loading && lists.length === 0;

  const totalLists = lists.length;
  const totalActive = lists.filter((L) => L.enabled).length;

  return (
    <details className="dpi-lists-collapse" open>
      <summary className="dpi-lists-summary">
        <span className="dpi-lists-summary-title">Списки обхода</span>
        <span className="dpi-lists-summary-meta">
          {totalLists > 0 ? `${totalActive} из ${totalLists} активны` : "нет списков"}
        </span>
      </summary>

      <div className="dpi-lists-section">
        {/* Category tabs */}
        <div className="dpi-cat-tabs">
          {CAT_META.map((c) => {
            const count = byCategory[c.id].length;
            const activeCount = byCategory[c.id].filter((L) => L.enabled).length;
            return (
              <button
                key={c.id}
                type="button"
                className={`dpi-cat-tab${tab === c.id ? " dpi-cat-tab--on" : ""}`}
                onClick={() => setTab(c.id)}
              >
                {c.label}
                {count > 0 && (
                  <span className={`dpi-cat-badge${activeCount > 0 ? " dpi-cat-badge--on" : ""}`}>
                    {activeCount}/{count}
                  </span>
                )}
              </button>
            );
          })}
        </div>

      {/* Loading / error */}
      {loading ? <p className="muted" style={{ margin: "8px 0" }}>Загрузка…</p> : null}

      {showGlobalEmpty ? (
        <p className="muted" style={{ margin: "8px 0" }}>
          {loadError ? `${loadError}. ` : ""}
          Списки не найдены в <code className="inline-code">{listsHintPath || "zapret\\lists"}</code>.
          Убедитесь, что папка zapret/lists рядом с Flux.exe.
        </p>
      ) : null}

      {/* Tiles grid */}
      {!loading && filtered.length > 0 ? (
        <div className="dpi-lists-grid">
          {filtered.map((L) => {
            const displayName = L.name.replace(/\.txt$/i, "");
            return (
              <button
                key={L.name}
                type="button"
                className={`dpi-list-tile${L.enabled ? " dpi-list-tile--on" : ""}`}
                disabled={busy}
                title={L.description ? `${L.name}: ${L.description}` : L.name}
                onClick={() => void onToggle(L.name, !L.enabled)}
              >
                <span className="dpi-list-check">
                  {L.enabled ? (
                    <span className="dpi-list-check-icon"><CheckIcon /></span>
                  ) : null}
                </span>
                <span className="dpi-list-meta">
                  <span className="dpi-list-name">{displayName}</span>
                  {L.description ? (
                    <span className="dpi-list-count">{L.description.slice(0, 32)}</span>
                  ) : null}
                </span>
              </button>
            );
          })}
        </div>
      ) : null}

      {!loading && loadedOnce && lists.length > 0 && filtered.length === 0 ? (
        <p className="muted" style={{ margin: "6px 0" }}>В этой группе нет списков.</p>
      ) : null}
      </div>
    </details>
  );
}
