import React, { useEffect, useMemo, useState } from "react";
import { useApp } from "../context/AppContext";
import { Switch } from "../ui";

type Any = Record<string, unknown>;
type ListCategory = "all" | "custom" | "media" | "social" | "games" | "other";

const CATEGORY_LABELS: { key: ListCategory; label: string }[] = [
  { key: "all", label: "Все" },
  { key: "custom", label: "Мои" },
  { key: "media", label: "Медиа" },
  { key: "social", label: "Соцсети" },
  { key: "games", label: "Игры" },
  { key: "other", label: "Прочее" },
];

function classifyList(name: string, description: string): Exclude<ListCategory, "all"> {
  const s = `${name} ${description}`.toLowerCase();
  if (
    s.includes("custom") ||
    s.includes("user") ||
    s.includes("my-") ||
    s.includes("мои") ||
    s.includes("пользователь")
  ) {
    return "custom";
  }
  if (
    s.includes("youtube") ||
    s.includes("twitch") ||
    s.includes("netflix") ||
    s.includes("spotify") ||
    s.includes("video") ||
    s.includes("media") ||
    s.includes("music") ||
    s.includes("rutube") ||
    s.includes("hdrezka")
  ) {
    return "media";
  }
  if (
    s.includes("discord") ||
    s.includes("telegram") ||
    s.includes("twitter") ||
    s.includes("instagram") ||
    s.includes("facebook") ||
    s.includes("meta") ||
    s.includes("tiktok") ||
    s.includes("reddit") ||
    s.includes("social") ||
    s.includes("whatsapp")
  ) {
    return "social";
  }
  if (
    s.includes("game") ||
    s.includes("steam") ||
    s.includes("roblox") ||
    s.includes("epic") ||
    s.includes("riot") ||
    s.includes("valorant") ||
    s.includes("battle") ||
    s.includes("ea") ||
    s.includes("ubisoft") ||
    s.includes("minecraft")
  ) {
    return "games";
  }
  return "other";
}

export function DpiScreen() {
  const {
    uiState,
    dpiEnabled,
    setDpi,
    dpiSubTab,
    setDpiSubTab,
    setDpiProfile,
    runDpiTuning,
    dpiTune,
    dpiLists,
    refreshDpiLists,
    toggleDpiList,
    readDpiList,
    writeDpiList,
    listsPath,
    busy,
  } = useApp();

  const [category, setCategory] = useState<ListCategory>("all");
  const [listSearch, setListSearch] = useState("");
  const [selectedName, setSelectedName] = useState("");

  const [detailMeta, setDetailMeta] = useState<Any | null>(null);
  const [editorText, setEditorText] = useState("");
  const [savedText, setSavedText] = useState("");
  const [loadingDetail, setLoadingDetail] = useState(false);

  const dpiProfiles = Array.isArray(uiState?.dpiProfiles)
    ? (uiState.dpiProfiles as Any[])
    : [];
  const activeProfile = String(uiState?.activeDpiProfile ?? "");

  const filteredLists = useMemo(() => {
    const q = listSearch.trim().toLowerCase();
    return dpiLists.filter((item) => {
      const name = String(item.name ?? "");
      const desc = String(item.description ?? "");
      if (category !== "all" && classifyList(name, desc) !== category) {
        return false;
      }
      if (!q) return true;
      return name.toLowerCase().includes(q) || desc.toLowerCase().includes(q);
    });
  }, [dpiLists, category, listSearch]);

  useEffect(() => {
    if (dpiSubTab !== "lists") return;
    if (!selectedName && filteredLists.length > 0) {
      setSelectedName(String(filteredLists[0].name ?? ""));
    }
  }, [dpiSubTab, filteredLists, selectedName]);

  useEffect(() => {
    if (dpiSubTab !== "lists" || !selectedName) return;
    let cancelled = false;
    setLoadingDetail(true);
    void readDpiList(selectedName).then((res) => {
      if (cancelled) return;
      setLoadingDetail(false);
      if (!res) {
        setDetailMeta(null);
        setEditorText("");
        setSavedText("");
        return;
      }
      setDetailMeta(res);
      const entries = Array.isArray(res.entries) ? res.entries.map(String) : [];
      const joined = entries.join("\n");
      setEditorText(joined);
      setSavedText(joined);
    });
    return () => {
      cancelled = true;
    };
  }, [dpiSubTab, selectedName, readDpiList]);

  const entryCount = useMemo(() => {
    return editorText
      .split(/\r?\n/)
      .map((s) => s.trim())
      .filter(Boolean).length;
  }, [editorText]);

  const isDirty = editorText !== savedText;
  const tooLarge = Boolean(detailMeta?.tooLarge);

  async function handleSaveList() {
    if (!selectedName || tooLarge) return;
    const entries = editorText
      .split(/\r?\n/)
      .map((s) => s.trim())
      .filter(Boolean);
    await writeDpiList(selectedName, entries);
    const updated = entries.join("\n");
    setEditorText(updated);
    setSavedText(updated);
    const refreshed = await readDpiList(selectedName);
    if (refreshed) setDetailMeta(refreshed);
  }

  async function handleReloadList() {
    if (!selectedName) return;
    setLoadingDetail(true);
    try {
      const res = await readDpiList(selectedName);
      if (res) {
        setDetailMeta(res);
        const entries = Array.isArray(res.entries)
          ? res.entries.map(String)
          : [];
        const joined = entries.join("\n");
        setEditorText(joined);
        setSavedText(joined);
      }
    } finally {
      setLoadingDetail(false);
    }
  }

  const tuneCandidates = Array.isArray(dpiTune?.candidates)
    ? ((dpiTune?.candidates ?? []) as Any[])
    : [];

  return (
    <div>
      {/* Top Sub-tabs Header */}
      <div className="fx-sec-head" style={{ marginBottom: 18 }}>
        <div className="fx-seg" role="tablist" aria-label="Разделы обхода DPI">
          <button
            type="button"
            role="tab"
            aria-selected={dpiSubTab === "profiles"}
            className={`fx-seg-btn ${dpiSubTab === "profiles" ? "active" : ""}`}
            onClick={() => setDpiSubTab("profiles")}
          >
            Профили и автоподбор
          </button>
          <button
            type="button"
            role="tab"
            aria-selected={dpiSubTab === "lists"}
            className={`fx-seg-btn ${dpiSubTab === "lists" ? "active" : ""}`}
            onClick={() => setDpiSubTab("lists")}
          >
            Списки доменов ({dpiLists.length})
          </button>
        </div>

        <div className="fx-bar">
          <span className={`fx-tag ${dpiEnabled ? "good" : ""}`}>
            winws2: {dpiEnabled ? "работает" : "остановлен"}
          </span>
          <button
            type="button"
            className={`fx-btn fx-btn-sm ${dpiEnabled ? "fx-btn-danger" : "fx-btn-primary"}`}
            disabled={busy}
            onClick={() => void setDpi(!dpiEnabled)}
          >
            {dpiEnabled ? "Выключить DPI" : "Включить DPI"}
          </button>
        </div>
      </div>

      {/* Tab 1: Profiles & Auto-Tuning */}
      {dpiSubTab === "profiles" && (
        <>
          <section className="fx-section fx-section-first">
            <div className="fx-sec-head">
              <div>
                <div className="fx-sec-title">ПРОФИЛИ СТРАТЕГИЙ ОБХОДА DPI</div>
                <div className="fx-sec-desc">
                  Выберите готовую стратегию winws2 / zapret. Активный профиль
                  применяется сразу.
                </div>
              </div>
              <span className="fx-mono" style={{ fontSize: 11.5, color: "var(--text-muted)" }}>
                Всего профилей: {dpiProfiles.length}
              </span>
            </div>

            <div className="fx-list-wrap">
              <div className="fx-thead">
                <span style={{ width: 220 }}>Профиль</span>
                <span style={{ width: 90 }}>Движок</span>
                <span className="fx-cell-grow">Конфигурация / параметры</span>
                <span style={{ width: 100, textAlign: "right" }}>Статус</span>
              </div>

              {dpiProfiles.length === 0 ? (
                <div className="fx-empty">Профили DPI не найдены.</div>
              ) : (
                dpiProfiles.map((prof) => {
                  const name = String(prof.name ?? "");
                  const tool = String(prof.tool || "winws2");
                  const cfg = String(prof.configPath ?? "");
                  const args = Array.isArray(prof.args)
                    ? prof.args.map(String).join(" ")
                    : String(prof.args ?? "");
                  const isActive = name === activeProfile;

                  return (
                    <div
                      key={name}
                      className={`fx-trow clickable ${isActive ? "selected" : ""}`}
                      onClick={() => void setDpiProfile(name)}
                    >
                      <div
                        style={{ width: 220, fontWeight: isActive ? 600 : 500 }}
                        className="fx-mono fx-truncate"
                        title={name}
                      >
                        {name}
                      </div>

                      <div style={{ width: 90 }}>
                        <span className="fx-tag">{tool}</span>
                      </div>

                      <div
                        className="fx-cell-grow fx-mono fx-truncate"
                        style={{ fontSize: 11.5, color: "var(--text-muted)" }}
                        title={cfg || args}
                      >
                        {cfg || args || "встроенный профиль"}
                      </div>

                      <div style={{ width: 100, textAlign: "right" }}>
                        {isActive ? (
                          <span className="fx-tag accent">Активен</span>
                        ) : (
                          <span
                            style={{ fontSize: 11.5, color: "var(--text-muted)" }}
                          >
                            Выбрать
                          </span>
                        )}
                      </div>
                    </div>
                  );
                })
              )}
            </div>
          </section>

          <section className="fx-section">
            <div className="fx-sec-head">
              <div>
                <div className="fx-sec-title">АВТОПОДБОР КОНФИГУРАЦИИ</div>
                <div className="fx-sec-desc">
                  Поочерёдно проверяет встроенные стратегии на доступность целевых
                  ресурсов и сохраняет лучшую в{" "}
                  <span className="fx-mono">zapret:AutoGenerate</span>{" "}
                  (перезаписывая <span className="fx-mono">presets/AutoGenerate.txt</span>).
                </div>
              </div>

              <button
                type="button"
                className="fx-btn fx-btn-primary"
                disabled={busy}
                onClick={() => void runDpiTuning()}
              >
                {busy ? "Идёт подбор…" : "Подобрать конфигурацию"}
              </button>
            </div>

            {dpiTune && (
              <div>
                <div className="fx-well fx-well-pad" style={{ marginBottom: 12 }}>
                  <div
                    className="fx-bar"
                    style={{ justifyContent: "space-between", marginBottom: 4 }}
                  >
                    <strong>Результат автоподбора</strong>
                    {Boolean(dpiTune.newProfileName) && (
                      <span className="fx-tag good fx-mono">
                        {String(dpiTune.newProfileName)}
                      </span>
                    )}
                  </div>
                  <div style={{ fontSize: 12.5, color: "var(--text-secondary)" }}>
                    {String(dpiTune.message || "Подбор завершён.")}
                  </div>
                  {Boolean(dpiTune.bestSourceRel) && (
                    <div
                      className="fx-mono"
                      style={{
                        marginTop: 4,
                        fontSize: 11.5,
                        color: "var(--text-muted)",
                      }}
                    >
                      Лучший источник: {String(dpiTune.bestSourceRel)}
                    </div>
                  )}
                </div>

                {tuneCandidates.length > 0 && (
                  <div className="fx-list-wrap">
                    <div className="fx-thead">
                      <span style={{ width: 220 }}>Кандидат</span>
                      <span style={{ width: 100 }}>Пробные ответы</span>
                      <span style={{ width: 96 }}>Статус</span>
                      <span className="fx-cell-grow">Детали</span>
                    </div>

                    {tuneCandidates.map((cand, idx) => {
                      const label = String(cand.label || cand.sourceRel || `#${idx + 1}`);
                      const profName = String(cand.profileName ?? "");
                      const hits = Number(cand.probeHits ?? 0);
                      const total = Number(cand.probeTotal ?? 0);
                      const alive = Boolean(cand.presetAlive);
                      const detail = String(cand.detail ?? "");

                      return (
                        <div
                          key={`${label}-${idx}`}
                          className={`fx-trow ${profName ? "clickable" : ""}`}
                          onClick={() => {
                            if (profName) void setDpiProfile(profName);
                          }}
                        >
                          <div style={{ width: 220 }} className="fx-mono fx-truncate">
                            {label}
                          </div>
                          <div style={{ width: 100 }} className="fx-mono">
                            {hits}/{total}
                          </div>
                          <div style={{ width: 96 }}>
                            <span className={`fx-tag ${alive ? "good" : "bad"}`}>
                              {alive ? "Пройден" : "Сбой"}
                            </span>
                          </div>
                          <div
                            className="fx-cell-grow fx-truncate"
                            style={{ fontSize: 12, color: "var(--text-secondary)" }}
                            title={detail}
                          >
                            {detail || "—"}
                          </div>
                        </div>
                      );
                    })}
                  </div>
                )}
              </div>
            )}
          </section>
        </>
      )}

      {/* Tab 2: Master-Detail Domain Lists Editor */}
      {dpiSubTab === "lists" && (
        <section className="fx-section fx-section-first">
          <div className="fx-dpi-split">
            {/* Master Column */}
            <div className="fx-dpi-master">
              <div className="fx-dpi-master-top">
                <div className="fx-bar" style={{ gap: 6 }}>
                  <input
                    type="text"
                    className="fx-well-input fx-bar-grow"
                    style={{ minWidth: 0, height: 28, fontSize: 12 }}
                    placeholder="Поиск списка…"
                    value={listSearch}
                    onChange={(e) => setListSearch(e.target.value)}
                  />
                  <button
                    type="button"
                    className="fx-btn fx-btn-sm"
                    disabled={busy}
                    onClick={() => void refreshDpiLists()}
                  >
                    Обновить
                  </button>
                </div>

                <div className="fx-bar" style={{ gap: 4 }}>
                  {CATEGORY_LABELS.map((cat) => (
                    <button
                      key={cat.key}
                      type="button"
                      className={`fx-btn fx-btn-sm ${
                        category === cat.key ? "fx-btn-primary" : "fx-btn-ghost"
                      }`}
                      style={{ height: 23, padding: "0 7px", fontSize: 11 }}
                      onClick={() => setCategory(cat.key)}
                    >
                      {cat.label}
                    </button>
                  ))}
                </div>

                {listsPath && (
                  <div
                    className="fx-mono fx-truncate fx-select-text"
                    style={{ fontSize: 10.5, color: "var(--text-muted)" }}
                    title={listsPath}
                  >
                    {listsPath}
                  </div>
                )}
              </div>

              <div className="fx-dpi-list">
                {filteredLists.length === 0 ? (
                  <div className="fx-empty">Списки не найдены.</div>
                ) : (
                  filteredLists.map((item) => {
                    const name = String(item.name ?? "");
                    const desc = String(item.description ?? "");
                    const enabled = Boolean(item.enabled);
                    const isSelected = name === selectedName;

                    return (
                      <div
                        key={name}
                        className={`fx-dpi-item ${isSelected ? "active" : ""}`}
                        onClick={() => setSelectedName(name)}
                      >
                        <div className="fx-cell-grow">
                          <div
                            className="fx-mono fx-truncate"
                            style={{
                              fontSize: 12.5,
                              fontWeight: isSelected ? 600 : 500,
                            }}
                          >
                            {name}
                          </div>
                          {desc && (
                            <div
                              className="fx-truncate"
                              style={{ fontSize: 11, color: "var(--text-muted)" }}
                              title={desc}
                            >
                              {desc}
                            </div>
                          )}
                        </div>

                        <Switch
                          checked={enabled}
                          disabled={busy}
                          label={`Включить список ${name}`}
                          onChange={(next) => void toggleDpiList(name, next)}
                        />
                      </div>
                    );
                  })
                )}
              </div>
            </div>

            {/* Detail Editor Column */}
            <div className="fx-dpi-detail">
              <div className="fx-dpi-detail-head">
                <div className="fx-cell-grow">
                  <div style={{ display: "flex", alignItems: "center", gap: 8 }}>
                    <span className="fx-mono" style={{ fontWeight: 600 }}>
                      {selectedName || "Выберите список"}
                    </span>
                    {Boolean(detailMeta?.customized) && (
                      <span className="fx-tag accent">Изменён</span>
                    )}
                    {isDirty && <span className="fx-tag warn">Не сохранено</span>}
                  </div>
                  <div
                    className="fx-truncate"
                    style={{ fontSize: 11.5, color: "var(--text-muted)" }}
                    title={String(detailMeta?.path ?? "")}
                  >
                    {selectedName
                      ? `${entryCount} записей · ${String(
                          detailMeta?.description || detailMeta?.path || "",
                        )}`
                      : "Выберите файл из списка слева для редактирования доменов"}
                  </div>
                </div>

                {selectedName && (
                  <div className="fx-bar">
                    <button
                      type="button"
                      className="fx-btn fx-btn-sm"
                      disabled={busy || loadingDetail}
                      onClick={() => void handleReloadList()}
                    >
                      Перечитать
                    </button>
                    <button
                      type="button"
                      className="fx-btn fx-btn-sm fx-btn-primary"
                      disabled={busy || loadingDetail || tooLarge || !isDirty}
                      onClick={() => void handleSaveList()}
                    >
                      Сохранить
                    </button>
                  </div>
                )}
              </div>

              {!selectedName ? (
                <div className="fx-empty">Выберите список доменов слева.</div>
              ) : loadingDetail ? (
                <div className="fx-empty">Загрузка содержимого списка…</div>
              ) : tooLarge ? (
                <div className="fx-empty">
                  Файл <span className="fx-mono">{selectedName}</span> слишком
                  большой ({Number(detailMeta?.total ?? 0)} записей) для
                  встроенного редактора. Отредактируйте его вручную по пути:{" "}
                  <span className="fx-mono fx-select-text">
                    {String(detailMeta?.path ?? listsPath)}
                  </span>
                </div>
              ) : (
                <textarea
                  className="fx-dpi-editor"
                  spellCheck={false}
                  placeholder="По одному домену или маске в строке…"
                  value={editorText}
                  onChange={(e) => setEditorText(e.target.value)}
                />
              )}
            </div>
          </div>
        </section>
      )}
    </div>
  );
}
