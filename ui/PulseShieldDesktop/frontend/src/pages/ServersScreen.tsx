import React, { useEffect, useMemo, useState } from "react";
import { useApp } from "../context/AppContext";
import { detectCountry } from "../flags";
import {
  formatDate,
  formatLatency,
  formatRemaining,
  latencyTone,
} from "../format";
import { FlagBadge } from "../icons";

type Any = Record<string, unknown>;
type SortKey = "ping" | "country" | "name";

export function ServersScreen() {
  const {
    uiState,
    servers,
    activeServer,
    selectServer,
    removeServer,
    retestServer,
    retestAll,
    previewSubscription,
    importText,
    importFromClipboard,
    applySubscription,
    refreshSubscription,
    clearSubscription,
    busy,
  } = useApp();

  const [subUrl, setSubUrl] = useState(String(uiState?.subscriptionUrl ?? ""));
  const [showImportPanel, setShowImportPanel] = useState(false);
  const [manualKeys, setManualKeys] = useState("");
  const [previewData, setPreviewData] = useState<Any | null>(null);
  const [confirmClearSub, setConfirmClearSub] = useState(false);

  const [query, setQuery] = useState("");
  const [sortBy, setSortBy] = useState<SortKey>("ping");
  const [groupByCountry, setGroupByCountry] = useState(false);
  const [confirmDeleteId, setConfirmDeleteId] = useState<string | null>(null);

  useEffect(() => {
    setSubUrl(String(uiState?.subscriptionUrl ?? ""));
  }, [uiState?.subscriptionUrl]);

  const activeId = String(uiState?.activeServerId ?? activeServer?.id ?? "");
  const hasSub = Boolean(uiState?.subscriptionUrl);
  const subExpires = Number(uiState?.subscriptionExpiresUnix ?? 0);

  const enrichedServers = useMemo(() => {
    return servers.map((srv) => {
      const name = String(srv.name ?? "");
      const host = String(srv.host ?? "");
      const country = detectCountry(name, host);
      const lat = Number(srv.lastCheckLatencyMs ?? 0);
      return {
        raw: srv,
        id: String(srv.id ?? ""),
        name,
        host,
        port: String(srv.port ?? ""),
        proto: String(srv.protocol ?? ""),
        fromSub: Boolean(srv.fromSubscription),
        ok: Boolean(srv.lastCheckOk),
        latency: lat,
        countryCode: country?.code ?? "ZZ",
        countryName: country?.name ?? "Прочие регионы",
      };
    });
  }, [servers]);

  const filteredSorted = useMemo(() => {
    const q = query.trim().toLowerCase();
    const list = enrichedServers.filter((item) => {
      if (!q) return true;
      return (
        item.name.toLowerCase().includes(q) ||
        item.host.toLowerCase().includes(q) ||
        item.proto.toLowerCase().includes(q) ||
        item.countryName.toLowerCase().includes(q)
      );
    });

    list.sort((a, b) => {
      if (sortBy === "ping") {
        const la = a.latency > 0 ? a.latency : 999999;
        const lb = b.latency > 0 ? b.latency : 999999;
        if (la !== lb) return la - lb;
        return a.name.localeCompare(b.name, "ru");
      }
      if (sortBy === "country") {
        const cmp = a.countryName.localeCompare(b.countryName, "ru");
        if (cmp !== 0) return cmp;
        const la = a.latency > 0 ? a.latency : 999999;
        const lb = b.latency > 0 ? b.latency : 999999;
        return la - lb;
      }
      return a.name.localeCompare(b.name, "ru");
    });

    return list;
  }, [enrichedServers, query, sortBy]);

  const grouped = useMemo(() => {
    if (!groupByCountry) {
      return [{ code: "", title: "", items: filteredSorted }];
    }
    const map = new Map<
      string,
      { code: string; title: string; items: typeof filteredSorted }
    >();
    for (const item of filteredSorted) {
      const key = item.countryName;
      const existing = map.get(key);
      if (existing) {
        existing.items.push(item);
      } else {
        map.set(key, { code: item.countryCode, title: key, items: [item] });
      }
    }
    return Array.from(map.values());
  }, [filteredSorted, groupByCountry]);

  async function handlePreview() {
    if (!subUrl.trim()) return;
    const res = await previewSubscription(subUrl.trim());
    setPreviewData(res);
  }

  async function handleImportManual() {
    if (!manualKeys.trim()) return;
    const added = await importText(manualKeys.trim());
    if (added > 0) {
      setManualKeys("");
      setShowImportPanel(false);
    }
  }

  return (
    <div>
      {/* Subscription & Keys Top Section */}
      <section className="fx-section fx-section-first">
        <div className="fx-sec-head">
          <div>
            <div className="fx-sec-title">ПОДПИСКА И КЛЮЧИ</div>
            <div className="fx-sec-desc">
              {hasSub ? (
                <>
                  Активная подписка:{" "}
                  <span style={{ color: "var(--text-primary)" }}>
                    {String(uiState?.subscriptionProviderLabel || "Подключена")}
                  </span>
                  {subExpires > 0 && (
                    <>
                      {" "}
                      · истекает {formatDate(subExpires)} (
                      {formatRemaining(subExpires)})
                    </>
                  )}
                </>
              ) : (
                "Вставьте ссылку подписки или импортируйте ключи из буфера обмена"
              )}
            </div>
          </div>

          <button
            type="button"
            className="fx-btn fx-btn-sm fx-btn-ghost"
            onClick={() => {
              setShowImportPanel((prev) => !prev);
              setConfirmClearSub(false);
            }}
          >
            {showImportPanel ? "Скрыть ручной ввод" : "Ключи вручную / предпросмотр"}
          </button>
        </div>

        <form
          className="fx-bar"
          onSubmit={(e) => {
            e.preventDefault();
            if (subUrl.trim()) void applySubscription(subUrl.trim());
          }}
        >
          <input
            type="text"
            className="fx-well-input mono fx-bar-grow"
            placeholder="https://... (URL подписки Clash / Sing-box / VLESS)"
            value={subUrl}
            onChange={(e) => setSubUrl(e.target.value)}
          />
          <button
            type="submit"
            className="fx-btn fx-btn-primary"
            disabled={busy || !subUrl.trim()}
          >
            Применить
          </button>
          {hasSub && (
            <button
              type="button"
              className="fx-btn"
              disabled={busy}
              onClick={() => void refreshSubscription()}
            >
              Обновить
            </button>
          )}
          <button
            type="button"
            className="fx-btn"
            disabled={busy}
            onClick={() => void importFromClipboard()}
          >
            Из буфера
          </button>
          <button
            type="button"
            className="fx-btn"
            disabled={busy || servers.length === 0}
            onClick={() => void retestAll()}
          >
            Пинг всех
          </button>
        </form>

        {/* Collapsible Inline Manual Import & Preview Panel */}
        {showImportPanel && (
          <div className="fx-well fx-well-pad" style={{ marginTop: 12 }}>
            <div className="fx-sec-head" style={{ marginBottom: 8 }}>
              <span className="fx-sec-title">РУЧНОЙ ИМПОРТ И ПРЕДПРОСМОТР</span>
              <div className="fx-bar">
                <button
                  type="button"
                  className="fx-btn fx-btn-sm"
                  disabled={busy || !subUrl.trim()}
                  onClick={() => void handlePreview()}
                >
                  Предпросмотр URL
                </button>
                {hasSub &&
                  (confirmClearSub ? (
                    <>
                      <button
                        type="button"
                        className="fx-btn fx-btn-sm fx-btn-danger"
                        disabled={busy}
                        onClick={() => {
                          setConfirmClearSub(false);
                          void clearSubscription();
                        }}
                      >
                        Точно очистить подписку?
                      </button>
                      <button
                        type="button"
                        className="fx-btn fx-btn-sm fx-btn-ghost"
                        onClick={() => setConfirmClearSub(false)}
                      >
                        Отмена
                      </button>
                    </>
                  ) : (
                    <button
                      type="button"
                      className="fx-btn fx-btn-sm fx-btn-danger"
                      disabled={busy}
                      onClick={() => setConfirmClearSub(true)}
                    >
                      Очистить подписку
                    </button>
                  ))}
              </div>
            </div>

            {previewData && (
              <div
                style={{
                  padding: "8px 10px",
                  marginBottom: 10,
                  borderRadius: 5,
                  border: "1px solid var(--hairline)",
                  background: "var(--bg-canvas)",
                  fontSize: 12,
                }}
              >
                {previewData.ok ? (
                  <div className="fx-bar" style={{ justifyContent: "space-between" }}>
                    <span>
                      Провайдер:{" "}
                      <strong>
                        {String(
                          previewData.name ||
                            previewData.providerLabel ||
                            "Без названия",
                        )}
                      </strong>{" "}
                      · формат{" "}
                      <span className="fx-mono">{String(previewData.format || "—")}</span>
                    </span>
                    <span className="fx-tag good">
                      {Number(previewData.total ?? 0)} серверов
                    </span>
                  </div>
                ) : (
                  <span style={{ color: "var(--bad)" }}>
                    Ошибка предпросмотра: {String(previewData.error || "неизвестно")}
                  </span>
                )}
              </div>
            )}

            <textarea
              rows={4}
              className="fx-textarea mono"
              placeholder="Вставьте один или несколько ключей (vless://, vmess://, trojan://, ss://, hy2://) или base64-содержимое…"
              value={manualKeys}
              onChange={(e) => setManualKeys(e.target.value)}
            />
            <div className="fx-bar" style={{ justifyContent: "flex-end", marginTop: 8 }}>
              <button
                type="button"
                className="fx-btn fx-btn-primary fx-btn-sm"
                disabled={busy || !manualKeys.trim()}
                onClick={() => void handleImportManual()}
              >
                Импортировать ключи
              </button>
            </div>
          </div>
        )}
      </section>

      {/* Filter & Sort Controls */}
      <section className="fx-section">
        <div className="fx-sec-head">
          <div className="fx-bar fx-bar-grow">
            <input
              type="text"
              className="fx-well-input"
              style={{ width: 250, maxWidth: "100%" }}
              placeholder="Поиск узла по имени, адресу, стране…"
              value={query}
              onChange={(e) => setQuery(e.target.value)}
            />

            <div className="fx-seg" role="group" aria-label="Сортировка серверов">
              <button
                type="button"
                className={`fx-seg-btn ${sortBy === "ping" ? "active" : ""}`}
                onClick={() => setSortBy("ping")}
              >
                По пингу
              </button>
              <button
                type="button"
                className={`fx-seg-btn ${sortBy === "country" ? "active" : ""}`}
                onClick={() => setSortBy("country")}
              >
                По стране
              </button>
              <button
                type="button"
                className={`fx-seg-btn ${sortBy === "name" ? "active" : ""}`}
                onClick={() => setSortBy("name")}
              >
                По имени
              </button>
            </div>

            <button
              type="button"
              className={`fx-btn fx-btn-sm ${groupByCountry ? "fx-btn-primary" : ""}`}
              onClick={() => setGroupByCountry((v) => !v)}
            >
              По странам
            </button>
          </div>

          <span className="fx-mono" style={{ fontSize: 11.5, color: "var(--text-muted)" }}>
            Показано {filteredSorted.length} из {servers.length}
          </span>
        </div>

        {/* Full-bleed Hairline Table */}
        <div className="fx-list-wrap">
          <div className="fx-thead">
            <span className="fx-cell-grow">Узел</span>
            <span style={{ width: 82 }}>Протокол</span>
            <span style={{ width: 190 }}>Адрес</span>
            <span style={{ width: 78, textAlign: "right" }}>Пинг</span>
            <span style={{ width: 158, textAlign: "right" }}>Действия</span>
          </div>

          {filteredSorted.length === 0 ? (
            <div className="fx-empty">
              {servers.length === 0
                ? "Список серверов пуст. Вставьте URL подписки или ключи выше."
                : "По вашему запросу узлы не найдены."}
            </div>
          ) : (
            grouped.map((group) => (
              <React.Fragment key={group.title || "all"}>
                {groupByCountry && group.title && (
                  <div className="fx-group-head">
                    <FlagBadge code={group.code} name={group.title} />
                    <span>{group.title}</span>
                    <span className="fx-mono" style={{ opacity: 0.7 }}>
                      ({group.items.length})
                    </span>
                  </div>
                )}

                {group.items.map((item) => {
                  const isActive = item.id === activeId;
                  const isConfirmDelete = confirmDeleteId === item.id;
                  return (
                    <div
                      key={item.id}
                      className={`fx-trow clickable ${isActive ? "selected" : ""}`}
                      onClick={() => {
                        setConfirmDeleteId(null);
                        void selectServer(item.id);
                      }}
                    >
                      <div
                        className="fx-cell-grow"
                        style={{ display: "flex", alignItems: "center", gap: 10 }}
                      >
                        <FlagBadge
                          code={item.countryCode === "ZZ" ? undefined : item.countryCode}
                          name={item.countryName}
                        />
                        <span className="fx-truncate" style={{ fontWeight: isActive ? 600 : 500 }}>
                          {item.name}
                        </span>
                        {isActive && <span className="fx-tag accent">Активен</span>}
                        {!item.fromSub && (
                          <span className="fx-tag">Ручной</span>
                        )}
                      </div>

                      <div style={{ width: 82 }}>
                        <span className="fx-tag">{item.proto.toUpperCase()}</span>
                      </div>

                      <div
                        style={{ width: 190, color: "var(--text-muted)", fontSize: 12 }}
                        className="fx-mono fx-truncate fx-select-text"
                        title={`${item.host}:${item.port}`}
                      >
                        {item.host}:{item.port}
                      </div>

                      <div
                        style={{ width: 78, textAlign: "right" }}
                        className={`fx-lat ${latencyTone(item.latency)}`}
                      >
                        {item.latency > 0 ? formatLatency(item.latency) : "—"}
                      </div>

                      <div
                        style={{
                          width: 158,
                          display: "flex",
                          justifyContent: "flex-end",
                          gap: 6,
                        }}
                        onClick={(e) => e.stopPropagation()}
                      >
                        <button
                          type="button"
                          className="fx-btn fx-btn-sm fx-btn-ghost"
                          disabled={busy}
                          onClick={() => void retestServer(item.id)}
                          title="Замерить пинг узла"
                        >
                          Пинг
                        </button>

                        {isConfirmDelete ? (
                          <button
                            type="button"
                            className="fx-btn fx-btn-sm fx-btn-danger"
                            disabled={busy}
                            onClick={() => {
                              setConfirmDeleteId(null);
                              void removeServer(item.id);
                            }}
                          >
                            Удалить?
                          </button>
                        ) : (
                          <button
                            type="button"
                            className="fx-btn fx-btn-sm fx-btn-ghost"
                            disabled={busy}
                            onClick={() => setConfirmDeleteId(item.id)}
                          >
                            Удалить
                          </button>
                        )}
                      </div>
                    </div>
                  );
                })}
              </React.Fragment>
            ))
          )}
        </div>
      </section>
    </div>
  );
}
