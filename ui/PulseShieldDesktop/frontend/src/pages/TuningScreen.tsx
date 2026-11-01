import { useApp } from "../context/AppContext";
import { Icon } from "../icons";
import { ScreenHead } from "../ui";

export function TuningScreen() {
  const {
    uiState, health, appVersion, openSheet, proxyPortDraft, listsPath,
    dpiLists, health: h, setActive,
  } = useApp();

  const enabledLists = (dpiLists || []).filter((l) => !!l.enabled).length;
  const dnsPreset = String(uiState?.dnsPreset || "auto");
  const shortListsPath = listsPath ? listsPath.split(/[/\\]/).slice(-2).join("\\") : "zapret\\lists";

  return (
    <>
      <ScreenHead title="Настройка" sub="Сеть, DNS и обход DPI" />

      <div className="scr scroll">
        <div className="tile">
          <div className="tile-head">
            <h3>Прокси</h3>
            <span className="mute">sing-box</span>
          </div>
          <button type="button" className="set-row" onClick={() => openSheet("port")}>
            <span className="sr-main">
              <b>Порт локального прокси</b>
              <small>Приложение и браузер подключаются к этому порту</small>
            </span>
            <span className="sr-val">{proxyPortDraft}</span>
            <Icon name="chev" size={14} />
          </button>
          <button type="button" className="set-row" onClick={() => openSheet("dns")}>
            <span className="sr-main">
              <b>DNS</b>
              <small>Серверы имён для туннеля и обхода</small>
            </span>
            <span className="sr-val">{dnsPreset}</span>
            <Icon name="chev" size={14} />
          </button>
          <button type="button" className="set-row" onClick={() => openSheet("server")}>
            <span className="sr-main">
              <b>Серверы</b>
              <small>{(uiState?.servers as unknown[])?.length || 0} подключений</small>
            </span>
            <Icon name="chev" size={14} />
          </button>
        </div>

        <div className="tile">
          <div className="tile-head">
            <h3>Обход DPI</h3>
            <span className="mute">{h?.dpiRunning ? "работает" : "остановлен"}</span>
          </div>
          <button type="button" className="set-row" onClick={() => openSheet("tune-dpi")}>
            <span className="sr-main">
              <b>Профиль и автоподбор</b>
              <small>{String(uiState?.activeDpiProfile || "профиль не выбран")}</small>
            </span>
            <Icon name="chev" size={14} />
          </button>
          <button type="button" className="set-row" onClick={() => setActive("more")}>
            <span className="sr-main">
              <b>Списки доменов</b>
              <small>{enabledLists} активно · {shortListsPath}</small>
            </span>
            <Icon name="chev" size={14} />
          </button>
        </div>

        <div className="tile">
          <div className="tile-head">
            <h3>Приложение</h3>
            <span className="mute">{appVersion ? `v${appVersion}` : ""}</span>
          </div>
          <button type="button" className="set-row" onClick={() => openSheet("updates")}>
            <span className="sr-main">
              <b>Обновления</b>
              <small>Проверить новую версию</small>
            </span>
            <Icon name="chev" size={14} />
          </button>
          <button type="button" className="set-row" onClick={() => openSheet("license")}>
            <span className="sr-main">
              <b>Лицензия</b>
              <small>
                {uiState?.licenseActive ? `активна · ${String(uiState?.licenseTier || "")}` : "не активирована"}
              </small>
            </span>
            <Icon name="chev" size={14} />
          </button>
          <button type="button" className="set-row" onClick={() => openSheet("about")}>
            <span className="sr-main">
              <b>О программе</b>
              <small>Версия и компоненты</small>
            </span>
            <Icon name="chev" size={14} />
          </button>
        </div>

        <div className="tile">
          <div className="tile-head">
            <h3>Состояние</h3>
          </div>
          <div className="metric-row">
            <span>Локальный прокси</span>
            <strong className={health?.proxyLocalOk ? "ok" : "muted"}>
              {health?.proxyLocalOk ? "доступен" : "недоступен"}
            </strong>
          </div>
          <div className="metric-row">
            <span>Процесс DPI</span>
            <strong className={health?.dpiRunning ? "ok" : "muted"}>
              {health?.dpiRunning ? "запущен" : "остановлен"}
            </strong>
          </div>
          <div className="metric-row">
            <span>Права</span>
            <strong className={uiState?.elevated ? "ok" : "bad"}>
              {uiState?.elevated ? "администратор" : "обычные"}
            </strong>
          </div>
        </div>
      </div>
    </>
  );
}
