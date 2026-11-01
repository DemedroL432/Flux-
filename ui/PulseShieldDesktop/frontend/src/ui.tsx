import type { ReactNode } from "react";
import { useApp } from "./context/AppContext";
import { NAV_ITEMS } from "./types";
import { Icon } from "./icons";

/** Левая рейка навигации с логотипом, разделами и кнопкой оформления. */
export function Rail() {
  const { active, setActive, openSheet, theme } = useApp();

  return (
    <aside className="rail">
      <button
        type="button"
        className="rail-brand"
        title="Flux — на главный экран"
        onClick={() => setActive("shield")}
      >
        F
      </button>

      <nav className="rail-nav" aria-label="Разделы">
        {NAV_ITEMS.map((item) => (
          <button
            key={item.key}
            type="button"
            className={`rail-btn${active === item.key ? " active" : ""}`}
            data-tip={item.label}
            aria-label={item.label}
            aria-current={active === item.key ? "page" : undefined}
            onClick={() => setActive(item.key)}
          >
            <Icon name={item.icon} />
          </button>
        ))}
      </nav>

      <div className="rail-bottom">
        <button
          type="button"
          className="rail-theme"
          title={`Оформление · тема «${theme}»`}
          aria-label="Оформление"
          onClick={() => openSheet("theme")}
        >
          <Icon name="palette" />
        </button>
      </div>
    </aside>
  );
}

/** Шапка окна: заголовок раздела, состояние, кнопки окна. */
export function TopBar({ title, sub }: { title: string; sub?: string }) {
  const { appVersion, closeToTray, openSheet, active, setActive, proxyEnabled, dpiEnabled, health, proxyPhase } = useApp();

  const proxyUp = proxyEnabled && !!health?.proxyLocalOk;
  const dpiUp = dpiEnabled && !!health?.dpiRunning;
  const activeUp = proxyUp || dpiUp;
  const connecting = (proxyPhase === "connecting" && proxyEnabled && !proxyUp) || ((proxyEnabled || dpiEnabled) && !activeUp);

  const canGoBack = active !== "shield";

  return (
    <header className="sys-top" style={{ ["--wails-draggable" as string]: "drag" }}>
      <button
        type="button"
        className="sys-back"
        style={{
          ["--wails-draggable" as string]: "no-drag",
          opacity: canGoBack ? 1 : 0.35,
          cursor: canGoBack ? "pointer" : "default",
          pointerEvents: canGoBack ? "auto" : "none",
        }}
        onClick={() => setActive("shield")}
        disabled={!canGoBack}
        title={canGoBack ? "Назад на главный экран" : undefined}
        aria-label="Назад"
      >
        <Icon name="back" />
      </button>

      <div className="sys-brand" style={{ ["--wails-draggable" as string]: "drag" }}>
        <span className="sys-logo">F</span>
        <span>{title}</span>
        {sub ? (
          <span className="sys-sub">{sub}</span>
        ) : (
          <span className="sys-sub">{`Flux ${appVersion ? "v" + appVersion : ""}`}</span>
        )}
      </div>

      <div className="sys-actions" style={{ ["--wails-draggable" as string]: "no-drag" }}>
        <span className={`sys-state${activeUp ? " on" : connecting ? " warn" : ""}`}>
          <i className="d" />
          {activeUp ? "Защищено" : connecting ? "Запускается" : "Выключено"}
        </span>

        <button
          type="button"
          className="icon-btn"
          title="Оформление"
          aria-label="Оформление"
          onClick={() => openSheet("theme")}
        >
          <Icon name="palette" />
        </button>

        <div className="win-controls">
          <button type="button" className="wc" title="Свернуть" onClick={() => void import("../wailsjs/runtime/runtime").then((m) => m.WindowMinimise())}>
            <svg viewBox="0 0 10 10"><path d="M2 5h6" /></svg>
          </button>
          <button type="button" className="wc" title="Развернуть" onClick={() => void import("../wailsjs/runtime/runtime").then((m) => m.WindowToggleMaximise())}>
            <svg viewBox="0 0 10 10"><rect x="2" y="2" width="6" height="6" rx="1" /></svg>
          </button>
          <button type="button" className="wc close" title="Свернуть в трей" onClick={closeToTray}>
            <svg viewBox="0 0 10 10"><path d="M2.5 2.5l5 5M7.5 2.5l-5 5" /></svg>
          </button>
        </div>
      </div>
    </header>
  );
}

/** Заголовок экрана. */
export function ScreenHead({
  title,
  sub,
  right,
}: {
  title: string;
  sub?: string;
  right?: ReactNode;
}) {
  return (
    <header className="screen-head">
      <div>
        <h1>{title}</h1>
        {sub ? <p className="sub">{sub}</p> : null}
      </div>
      {right ? <div className="right">{right}</div> : null}
    </header>
  );
}

/** Переключатель. */
export function Switch({
  on,
  onChange,
  label,
  disabled,
}: {
  on: boolean;
  onChange: (v: boolean) => void;
  label: string;
  disabled?: boolean;
}) {
  return (
    <button
      type="button"
      role="switch"
      aria-checked={on}
      aria-label={label}
      disabled={disabled}
      className={`sw${on ? " on" : ""}`}
      onClick={() => onChange(!on)}
    />
  );
}

/** Полупрозрачная шторка снизу. */
export function Sheet({
  title,
  subtitle,
  onClose,
  children,
}: {
  title: string;
  subtitle?: string;
  onClose: () => void;
  children: ReactNode;
}) {
  return (
    <div className="sheet-layer open">
      <div className="scrim" onClick={onClose} />
      <section className="sheet" role="dialog" aria-modal="true" aria-label={title}>
        <header className="sheet-head">
          <div>
            <h2>{title}</h2>
            {subtitle ? <p className="sheet-sub">{subtitle}</p> : null}
          </div>
          <button type="button" className="sheet-close" aria-label="Закрыть" onClick={onClose}>
            <Icon name="close" size={16} />
          </button>
        </header>
        <div className="sheet-body">{children}</div>
      </section>
    </div>
  );
}

/** Всплывающие уведомления. */
export function Toasts() {
  const { toast } = useApp();
  if (!toast) return null;
  return (
    <div className="toast-stack">
      <div className={`toast${toast.variant === "error" ? " bad" : ""}`} role="status">
        <span className="toast-ico">
          <Icon name={toast.variant === "error" ? "warn" : "check"} size={14} />
        </span>
        <span className="toast-text">{toast.text}</span>
      </div>
    </div>
  );
}

/** Плашка загрузки. */
export function BusyBar({ show }: { show: boolean }) {
  if (!show) return null;
  return (
    <div className="busy-bar">
      <span className="spinner" />
      <span>Применяю изменения…</span>
    </div>
  );
}
