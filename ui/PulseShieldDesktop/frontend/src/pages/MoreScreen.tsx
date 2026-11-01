import { useCallback, useEffect, useMemo, useState } from "react";
import { useApp } from "../context/AppContext";
import { Icon } from "../icons";
import { ScreenHead } from "../ui";

type DpiList = { name: string; description: string; enabled: boolean };
type Cat = "custom" | "media" | "social" | "games" | "other";

const CUSTOM_LIST = "custom-dpi";

const CATS: { id: Cat; label: string }[] = [
  { id: "custom", label: "Мои домены" },
  { id: "media", label: "Медиа" },
  { id: "social", label: "Соцсети" },
  { id: "games", label: "Игры" },
  { id: "other", label: "Прочее" },
];

function categoryOf(name: string): Cat {
  const n = name.replace(/\.txt$/i, "").toLowerCase();
  if (n === CUSTOM_LIST) return "custom";
  if (["steam", "roblox", "riot", "lol", "tankix", "itch", "games"].some((k) => n.includes(k))) return "games";
  if (["instagram", "facebook", "twitter", "whatsapp"].some((k) => n.includes(k))) return "social";
  if (
    ["youtube", "google", "discord", "telegram", "twitch", "soundcloud", "github", "amazon",
      "claude", "speedtest", "rutracker", "rutor", "netflix"].some((k) => n.includes(k))
  ) return "media";
  return "other";
}

export function MoreScreen() {
  const {
    uiState, openSheet, setActive, appVersion, stopAll, busy,
    dpiLists, refreshDpiLists, toggleDpiList, readDpiList, writeDpiList, listsPath,
  } = useApp();

  const [tab, setTab] = useState<Cat>("custom");
  const [editing, setEditing] = useState<string | null>(null);
  const [text, setText] = useState("");
  const [path, setPath] = useState("");
  const [loading, setLoading] = useState(false);
  const [saving, setSaving] = useState(false);

  useEffect(() => {
    void refreshDpiLists();
  }, [refreshDpiLists]);

  const byCat = useMemo(() => {
    const map: Record<Cat, DpiList[]> = { custom: [], media: [], social: [], games: [], other: [] };
    for (const l of (dpiLists || []) as DpiList[]) {
      map[categoryOf(l.name)].push(l);
    }
    return map;
  }, [dpiLists]);

  const openEditor = useCallback(
    async (name: string) => {
      setEditing(name);
      setLoading(true);
      try {
        const c = await readDpiList(name);
        setText(((c?.entries as string[]) || []).join("\n"));
        setPath(String(c?.path || ""));
      } finally {
        setLoading(false);
      }
    },
    [readDpiList]
  );

  const save = useCallback(async () => {
    if (!editing) return;
    setSaving(true);
    try {
      const entries = text.split(/\r?\n/).map((s) => s.trim()).filter(Boolean);
      await writeDpiList(editing, entries);
      setEditing(null);
    } finally {
      setSaving(false);
    }
  }, [editing, text, writeDpiList]);

  const lineCount = text.split(/\r?\n/).filter((s) => s.trim()).length;

  // Сброс редактора при уходе с экрана.
  useEffect(() => {
    setEditing(null);
  }, []);

  return (
    <>
      <ScreenHead title="Ещё" sub="Списки обхода, диагностика и помощь" />

      <div className="scr scroll">
        {editing ? (
          <div className="tile">
            <div className="tile-head">
              <button type="button" className="btn sm soft" onClick={() => setEditing(null)} disabled={saving}>
                ← Назад
              </button>
              <h3 style={{ marginLeft: 4 }}>{editing}</h3>
              <button type="button" className="btn sm primary" onClick={() => void save()} disabled={saving || busy}>
                {saving ? "Сохраняю…" : "Сохранить"}
              </button>
            </div>

            <p className="sheet-sub" style={{ margin: "0 0 8px" }}>
              {lineCount} записей{path ? ` · ${path}` : ""}
            </p>

            <textarea
              className="field mono"
              style={{ width: "100%", minHeight: 220, resize: "vertical", lineHeight: 1.6 }}
              value={text}
              spellCheck={false}
              placeholder={"youtube.com\ndiscord.com"}
              onChange={(e) => setText(e.target.value)}
              aria-label="Домены списка"
            />

            <p className="sheet-sub" style={{ marginTop: 8, marginBottom: 0 }}>
              По одному домену в строке. Ссылки вида https://site.com/path принимаются — останется только домен.
              Пустые строки и дубликаты отбрасываются.
            </p>
          </div>
        ) : (
          <>
            <div className="tile">
              <div className="tile-head">
                <h3>Списки обхода</h3>
                <span className="mute">{(dpiLists as DpiList[])?.length || 0} шт.</span>
              </div>

              <div className="seg" style={{ marginBottom: 10, flexWrap: "wrap" }}>
                {CATS.map((c) => {
                  const count = byCat[c.id].length;
                  const on = byCat[c.id].filter((l) => l.enabled).length;
                  return (
                    <button
                      key={c.id}
                      type="button"
                      className={tab === c.id ? "on" : ""}
                      onClick={() => setTab(c.id)}
                    >
                      {c.label}
                      {count > 0 ? ` ${on}/${count}` : ""}
                    </button>
                  );
                })}
              </div>

              {byCat[tab].length === 0 ? (
                <div className="empty">
                  <Icon name="info" size={20} />
                  <p>
                    {tab === "custom"
                      ? "Пока пусто. Добавьте правило с действием «Обход DPI» — домен появится здесь сам."
                      : "В этой группе нет списков."}
                  </p>
                </div>
              ) : (
                <div className="row-list">
                  {byCat[tab].map((l) => {
                    const label = l.name.replace(/\.txt$/i, "");
                    return (
                      <div key={l.name} className="server-row">
                        <button
                          type="button"
                          className={`sw${l.enabled ? " on" : ""}`}
                          role="switch"
                          aria-checked={l.enabled}
                          aria-label={label}
                          disabled={busy}
                          onClick={() => void toggleDpiList(l.name, !l.enabled)}
                        />
                        <span className="sm">
                          <b>{label}</b>
                          <small>
                            {l.name === CUSTOM_LIST ? "ваши правила обхода" : l.description || "список доменов"}
                          </small>
                        </span>
                        <button
                          type="button"
                          className="btn-edit"
                          title="Изменить адреса"
                          aria-label={`Изменить список ${label}`}
                          disabled={busy || loading}
                          onClick={() => void openEditor(l.name)}
                        >
                          <Icon name="sliders" size={14} />
                        </button>
                      </div>
                    );
                  })}
                </div>
              )}

              {listsPath ? (
                <p
                  className="sheet-sub"
                  title={listsPath}
                  style={{ margin: "10px 16px 12px", whiteSpace: "nowrap", overflow: "hidden", textOverflow: "ellipsis" }}
                >
                  Файлы: {listsPath.split(/[/\\]/).slice(-2).join("\\") || listsPath}
                </p>
              ) : null}
            </div>

            <div className="more-grid">
              <button type="button" className="more-tile" onClick={() => openSheet("subscription")}>
                <span className="mi"><Icon name="link" /></span>
                <b>Подписка</b>
                <small>любая ссылка или ключ</small>
              </button>
              <button type="button" className="more-tile" onClick={() => setActive("servers")}>
                <span className="mi"><Icon name="server" /></span>
                <b>Серверы</b>
                <small>выбор и пинг узлов</small>
              </button>
              <button type="button" className="more-tile" onClick={() => openSheet("stats")}>
                <span className="mi"><Icon name="pulse" /></span>
                <b>Диагностика</b>
                <small>самопроверка и отчёт</small>
              </button>
              <button type="button" className="more-tile" onClick={() => openSheet("logs")}>
                <span className="mi"><Icon name="log" /></span>
                <b>Журнал</b>
                <small>ядро, DPI, соединения</small>
              </button>
              <button type="button" className="more-tile" onClick={() => openSheet("tune-dpi")}>
                <span className="mi"><Icon name="dpi" /></span>
                <b>Обход DPI</b>
                <small>профиль и автоподбор</small>
              </button>
              <button type="button" className="more-tile" onClick={() => openSheet("updates")}>
                <span className="mi"><Icon name="download" /></span>
                <b>Обновления</b>
                <small>{appVersion ? `версия v${appVersion}` : "проверка обновлений"}</small>
              </button>
            </div>

            <div className="tile">
              <div className="row-l">
                <span className="rl-main">
                  <span className="rl-t">О программе</span>
                  <span className="rl-s">Flux · sing-box + zapret</span>
                </span>
                <span className="rl-r">{appVersion ? `v${appVersion}` : ""}</span>
                <button type="button" className="btn sm soft" onClick={() => openSheet("about")}>
                  <Icon name="chev" size={14} />
                </button>
              </div>
              <div className="row-l">
                <span className="rl-main">
                  <span className="rl-t">Остановить всё</span>
                  <span className="rl-s">аварийно выключить прокси и DPI</span>
                </span>
                <button type="button" className="btn sm danger" disabled={busy} onClick={() => void stopAll()}>
                  Стоп
                </button>
              </div>
            </div>

            {uiState?.subscriptionUrl ? (
              <div className="tile">
                <div
                  className="row-l"
                  onClick={() => openSheet("subscription")}
                  style={{ cursor: "pointer" }}
                >
                  <span className="ri"><Icon name="link" /></span>
                  <span className="rl-main">
                    <span className="rl-t">Подписка</span>
                    <span className="rl-s" title={String(uiState.subscriptionUrl)}>{String(uiState.subscriptionUrl)}</span>
                  </span>
                  <Icon name="chev" size={14} />
                </div>
              </div>
            ) : null}
          </>
        )}
      </div>
    </>
  );
}
