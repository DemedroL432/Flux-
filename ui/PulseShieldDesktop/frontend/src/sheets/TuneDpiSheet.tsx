import { useMemo, useState } from "react";
import { useApp } from "../context/AppContext";
import { Icon } from "../icons";

type Candidate = {
  label?: string;
  sourceRel?: string;
  profileName?: string;
  presetAlive?: boolean;
  probeHits?: number;
  probeTotal?: number;
  detail?: string;
};

/** Полноценная шторка «Обход DPI»: профиль, автоподбор и результаты. */
export function TuneDpiSheet() {
  const {
    uiState, dpiLists, toggleDpiList, setDpiProfile, runDpiTuning, dpiTune,
    busy, listsPath, dpiEnabled, setDpi,
  } = useApp();
  const [autoBusy, setAutoBusy] = useState(false);

  const profiles = useMemo(() => ((uiState?.dpiProfiles as { name?: string; description?: string }[]) || []), [uiState]);
  const activeProfile = String(uiState?.activeDpiProfile || "");

  const candidates = useMemo<Candidate[]>(() => {
    const raw = (dpiTune as { candidates?: Candidate[] } | null)?.candidates;
    return Array.isArray(raw) ? raw : [];
  }, [dpiTune]);

  const bestRel = String((dpiTune as { bestSourceRel?: string } | null)?.bestSourceRel || "");

  async function runTuning() {
    setAutoBusy(true);
    try {
      await runDpiTuning();
    } finally {
      setAutoBusy(false);
    }
  }

  return (
    <>
      <div className="tile">
        <div className="tile-head">
          <h3>Модуль</h3>
          <button
            type="button"
            className={`sw${dpiEnabled ? " on" : ""}`}
            role="switch"
            aria-checked={dpiEnabled}
            aria-label="Обход DPI"
            disabled={busy}
            onClick={() => void setDpi(!dpiEnabled)}
          />
        </div>
        <p className="sheet-sub" style={{ margin: 0 }}>
          Работает независимо от прокси. Запуск процесса не гарантирует доступность сайтов — всё зависит
          от сети и профиля.
        </p>
      </div>

      <div className="tile">
        <div className="tile-head">
          <h3>Профиль</h3>
          <span className="mute">{profiles.length} шт.</span>
        </div>
        {profiles.length === 0 ? (
          <div className="empty">
            <Icon name="info" size={20} />
            <p>Профили не найдены. Проверьте папку zapret рядом с программой.</p>
          </div>
        ) : (
          <div className="option-list">
            {profiles.map((p) => {
              const name = String(p.name || "");
              const on = name === activeProfile;
              return (
                <button
                  key={name}
                  type="button"
                  className={`option${on ? " on" : ""}`}
                  disabled={busy}
                  onClick={() => void setDpiProfile(name)}
                >
                  <span className="option-main">
                    <strong>{name}</strong>
                    {p.description ? <small>{p.description}</small> : null}
                  </span>
                  {on ? <Icon name="check" size={14} /> : null}
                </button>
              );
            })}
          </div>
        )}
      </div>

      <div className="tile">
        <div className="tile-head">
          <h3>Автоподбор</h3>
        </div>
        <p className="sheet-sub" style={{ marginTop: 0 }}>
          Перебирает варианты пресета и проверяет их реальными запросами к YouTube, Discord, Twitch
          и другим сервисам. Лучший по числу доступных и меньшей задержке сохраняется как профиль.
        </p>
        <button
          type="button"
          className="btn primary block"
          disabled={busy || autoBusy}
          onClick={() => void runTuning()}
        >
          {autoBusy ? "Подбираю… это может занять несколько минут" : "Подобрать конфигурацию"}
        </button>

        {candidates.length > 0 ? (
          <div className="tune-table">
            {candidates.map((c, i) => {
              const hits = Number(c.probeHits) || 0;
              const total = Number(c.probeTotal) || 0;
              const best = String(c.sourceRel || "") === bestRel;
              const ratio = total > 0 ? hits / total : 0;
              return (
                <div key={`${c.sourceRel || i}`} className={`tune-row${best ? " best" : ""}`}>
                  <span className={`tune-bar${ratio >= 0.9 ? " ok" : ratio >= 0.5 ? " mid" : " bad"}`}>
                    {hits}/{total}
                  </span>
                  <span className="tune-main">
                    <strong>
                      {String(c.label || c.profileName || "")}
                      {best ? " · выбрано" : ""}
                    </strong>
                    <small>{String(c.detail || "")}</small>
                  </span>
                </div>
              );
            })}
          </div>
        ) : null}
      </div>

      <div className="tile">
        <div className="tile-head">
          <h3>Списки доменов</h3>
          <span className="mute">{(dpiLists || []).length} шт.</span>
        </div>
        {(dpiLists || []).length === 0 ? (
          <p className="sheet-sub" style={{ margin: 0 }}>Списки не найдены в {listsPath || "zapret\\lists"}.</p>
        ) : (
          <div className="row-list">
            {(dpiLists as { name?: string; description?: string; enabled?: boolean }[]).map((l) => {
              const name = String(l.name || "");
              const label = name.replace(/\.txt$/i, "");
              return (
                <div key={name} className="server-row">
                  <button
                    type="button"
                    className={`sw${l.enabled ? " on" : ""}`}
                    role="switch"
                    aria-checked={!!l.enabled}
                    aria-label={label}
                    disabled={busy}
                    onClick={() => void toggleDpiList(name, !l.enabled)}
                  />
                  <span className="sm">
                    <b>{label}</b>
                    <small>{name === "custom-dpi" ? "ваши правила обхода" : l.description || "домены"}</small>
                  </span>
                </div>
              );
            })}
          </div>
        )}
        {listsPath ? (
          <p className="sheet-sub" style={{ margin: "8px 0 0" }}>Файлы: {listsPath}</p>
        ) : null}
      </div>
    </>
  );
}
