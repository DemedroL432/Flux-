import { useApp } from "../context/AppContext";
import { Icon } from "../icons";
import { formatBytes, formatClock, formatLatency, formatMbps, latencyTone } from "../format";
import { ScreenHead, Switch } from "../ui";

const SHORTCUTS = [
  { key: "srv", label: "Сервер", icon: "server", sheet: "server" as const },
  { key: "route", label: "Маршруты", icon: "routes", onClick: "routes" as const },
  { key: "test", label: "Тест", icon: "gauge", sheet: "speed" as const },
  { key: "dpi", label: "DPI", icon: "dpi", sheet: "tune-dpi" as const },
];

export function ShieldScreen() {
  const {
    proxyEnabled, dpiEnabled, setProxy, setDpi, busy, proxyPhase,
    health, rt, activeServer, uiState, openSheet, setActive,
    routingProfile,
  } = useApp();

  const proxyUp = proxyEnabled && !!health?.proxyLocalOk;
  const dpiUp = dpiEnabled && !!health?.dpiRunning;
  const connecting = proxyPhase === "connecting" && proxyEnabled && !proxyUp;

  let title = "Защита выключена";
  let sub = "Трафик идёт напрямую, как без приложения.";
  let stable = "Готов к запуску";

  if (connecting) {
    title = "Поднимаем туннель…";
    sub = "Применяем конфигурацию sing-box и DPI.";
    stable = "Идёт подключение";
  } else if (proxyUp && dpiUp) {
    title = "Максимальная защита";
    sub = "Прокси и DPI работают вместе — маршрут решают правила.";
    stable = "Работает стабильно";
  } else if (proxyUp) {
    title = "VPN активен";
    sub = activeServer ? `${activeServer.name} · ${routingLabel(routingProfile)}` : routingLabel(routingProfile);
    stable = "Туннель поднят";
  } else if (dpiUp) {
    title = "Обход DPI активен";
    sub = "Трафик идёт напрямую с маскировкой пакетов.";
    stable = "Работает стабильно";
  } else if (proxyEnabled || dpiEnabled) {
    title = "Запускается…";
    sub = "Модуль включён, но процесс ещё не поднялся.";
    stable = "Ожидание процесса";
  }

  const heroClass = proxyUp || dpiUp ? "hero" : connecting ? "hero connecting" : "hero idle";
  const down = formatMbps(Number(rt?.downloadBps) || 0);
  const up = formatMbps(Number(rt?.uploadBps) || 0);
  const total = (Number(rt?.sessionDownloadBytes) || 0) + (Number(rt?.sessionUploadBytes) || 0);

  return (
    <>
      <ScreenHead
        title="Защита"
        sub="Два независимых движка — прокси и обход DPI"
        right={
          <button type="button" className="chip mute" onClick={() => openSheet("stats")}>
            Диагностика <Icon name="chev" size={14} />
          </button>
        }
      />

      <div className="scr scroll">
        <section className={heroClass}>
          <span className="hero-led" aria-hidden />
          <div className="hero-body">
            <h2 className="hero-state">{title}</h2>
            <p className="hero-sub">{sub}</p>
            <span className="hero-stable">
              <Icon name={proxyUp || dpiUp ? "wave" : connecting ? "refresh" : "info"} size={15} />
              {stable}
            </span>
          </div>
        </section>

        {/* Два независимых движка: у каждого своя кнопка включения. */}
        <div className="modes">
          <section className={`mode${proxyEnabled ? " on" : ""}`} style={{ ["--tone" as string]: "var(--a3)" }}>
            <span className={`iconbox ${proxyEnabled ? "violet" : "grey"}`}>
              <Icon name="vpn" />
            </span>
            <div className="mb">
              <h2>Прокси</h2>
              <p className={`mstate${proxyUp ? " on" : proxyEnabled ? " warn" : ""}`}>
                {proxyUp ? "Активен" : proxyEnabled ? "Запускается" : "Выключен"}
              </p>
              <small className={latencyTone(Number(activeServer?.lastCheckLatencyMs)) || undefined}>
                {activeServer ? formatLatency(Number(activeServer.lastCheckLatencyMs)) : "сервер не выбран"}
              </small>
            </div>
            <Switch
              on={proxyEnabled}
              onChange={(v) => void setProxy(v)}
              label="Прокси"
              disabled={busy}
            />
          </section>

          <section className={`mode${dpiEnabled ? " on" : ""}`} style={{ ["--tone" as string]: "var(--ok)" }}>
            <span className={`iconbox ${dpiEnabled ? "mint" : "grey"}`}>
              <Icon name="dpi" />
            </span>
            <div className="mb">
              <h2>Обход DPI</h2>
              <p className={`mstate${dpiUp ? " on" : dpiEnabled ? " warn" : ""}`}>
                {dpiUp ? "Активен" : dpiEnabled ? "Запускается" : "Выключен"}
              </p>
              <small>{uiState?.activeDpiProfile ? "профиль выбран" : "профиль не выбран"}</small>
            </div>
            <Switch
              on={dpiEnabled}
              onChange={(v) => void setDpi(v)}
              label="Обход DPI"
              disabled={busy}
            />
          </section>
        </div>

        <div className="shortcuts">
          {SHORTCUTS.map((s) => (
            <button
              key={s.key}
              type="button"
              className="shortcut"
              disabled={busy}
              onClick={() => {
                if ("sheet" in s && s.sheet) openSheet(s.sheet);
                else if ("onClick" in s && s.onClick) setActive(s.onClick);
              }}
            >
              <Icon name={s.icon} />
              <span>{s.label}</span>
              <span className="chev">›</span>
            </button>
          ))}
        </div>

        <div className="tile traffic">
          <div className="tile-head" style={{ padding: "0 0 2px" }}>
            <h3>Трафик</h3>
            <button type="button" className="more" onClick={() => openSheet("stats")}>
              Подробнее <Icon name="chev" size={14} />
            </button>
          </div>
          <div className="stats3">
            <div className="stat3">
              <div className="label">Скачивание</div>
              <div className="value">
                {down}
                <small>Мбит/с</small>
              </div>
              <div className="line" />
            </div>
            <div className="stat3">
              <div className="label">Отдача</div>
              <div className="value">
                {up}
                <small>Мбит/с</small>
              </div>
              <div className="line purp" />
            </div>
            <div className="stat3">
              <div className="label">Сессия</div>
              <div className="value time">{formatClock(Number(rt?.sessionElapsedSeconds) || 0)}</div>
              <div className="line gray" />
            </div>
          </div>
          <div className="transfer">
            <Icon name="download" size={14} style={{ color: "var(--f2)" }} />
            <strong>Передано за сессию</strong>
            <span className="tr-r">{formatBytes(total)}</span>
          </div>
        </div>

        <button type="button" className="route-bar" onClick={() => setActive("routes")}>
          <span className="iconbox gold">
            <Icon name="routes" />
          </span>
          <span className="rb-main">
            <b>Маршрут</b>
            <p>{routingLabel(routingProfile)} · {routingProfile === "all-direct" ? "напрямую" : "через правила"}</p>
          </span>
          <span className="textbtn">
            Настроить <Icon name="chev" size={14} />
          </span>
        </button>
      </div>
    </>
  );
}

/** Человеческая подпись режима маршрутизации. */
export function routingLabel(name: string): string {
  switch (name) {
    case "full-proxy":
      return "Полный прокси";
    case "manual":
      return "Ручные правила";
    case "default":
      return "Полный прокси";
    case "rf-smart-split":
    case "selective-proxy":
    case "gaming-friendly":
    case "all-direct":
    case "bypass-russia":
      return "Ручные правила";
    default:
      return name || "Ручные правила";
  }
}
