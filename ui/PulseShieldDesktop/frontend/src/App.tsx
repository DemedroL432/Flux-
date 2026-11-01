import { useEffect, useRef, type ReactNode } from "react";
import { AppProvider, useApp } from "./context/AppContext";
import { Icon } from "./icons";
import { Rail, TopBar, Toasts, BusyBar, Sheet } from "./ui";
import { NAV_ITEMS, SHEET_TITLES, type NavKey, type SheetKey } from "./types";
import { ShieldScreen } from "./pages/ShieldScreen";
import { RoutesScreen } from "./pages/RoutesScreen";
import { TuningScreen } from "./pages/TuningScreen";
import { MoreScreen } from "./pages/MoreScreen";
import { ServersScreen } from "./pages/ServersScreen";
import { TuneDpiSheet } from "./sheets/TuneDpiSheet";
import {
  ServerSheet, RuleSheet, VerifySheet, PortSheet, DnsSheet, ThemeSheet,
  StatsSheet, LogsSheet, LicenseSheet, AboutSheet, UpdatesSheet,
  GuideSheet, SpeedSheet, SubscriptionSheet,
} from "./sheets";

/** Нижний док с разделами и двумя независимыми переключателями движков. */
function Dock() {
  const { active, setActive, proxyEnabled, dpiEnabled, busy } = useApp();
  const anyOn = proxyEnabled || dpiEnabled;

  return (
    <nav className={`dock${anyOn ? "" : " off"}${busy ? " busy" : ""}`} aria-label="Основная навигация">
      {NAV_ITEMS.map((item) => (
        <button
          key={item.key}
          type="button"
          className={`dock-btn${active === item.key ? " on" : ""}`}
          onClick={() => setActive(item.key)}
          aria-current={active === item.key ? "page" : undefined}
        >
          <Icon name={item.icon} />
          <span>{item.label}</span>
        </button>
      ))}
    </nav>
  );
}

const SHEETS: Record<SheetKey, () => ReactNode> = {
  server: () => <ServerSheet />,
  rule: () => <RuleSheet />,
  verify: () => <VerifySheet />,
  "tune-dpi": () => <TuneDpiSheet />,
  stats: () => <StatsSheet />,
  logs: () => <LogsSheet />,
  license: () => <LicenseSheet />,
  about: () => <AboutSheet />,
  updates: () => <UpdatesSheet />,
  guide: () => <GuideSheet />,
  speed: () => <SpeedSheet />,
  dns: () => <DnsSheet />,
  port: () => <PortSheet />,
  subscription: () => <SubscriptionSheet />,
  theme: () => <ThemeSheet />,
};

function SheetLayer() {
  const { sheet, closeSheet, openSheet } = useApp();

  useEffect(() => {
    const w = window as unknown as { __qaOpenSheet?: (k: SheetKey) => void };
    w.__qaOpenSheet = openSheet;
    return () => {
      delete w.__qaOpenSheet;
    };
  }, [openSheet]);

  if (!sheet) return null;
  const body = SHEETS[sheet];
  if (!body) return null;

  return (
    <Sheet title={SHEET_TITLES[sheet]} onClose={closeSheet}>
      {body()}
    </Sheet>
  );
}

function Scene() {
  const { active } = useApp();
  const ref = useRef<HTMLDivElement>(null);

  useEffect(() => {
    const el = ref.current;
    if (!el) return;
    el.classList.remove("in-right", "in-left");
    void el.offsetWidth;
    el.classList.add("in-right");
  }, [active]);

  const screens: Record<NavKey, ReactNode> = {
    shield: <ShieldScreen />,
    servers: <ServersScreen />,
    routes: <RoutesScreen />,
    tuning: <TuningScreen />,
    more: <MoreScreen />,
  };

  return (
    <div className="sys-body">
      <div className="scene">
        <div className="screen" ref={ref}>
          {screens[active]}
        </div>
      </div>
    </div>
  );
}

function Shell() {
  const { busy, active } = useApp();
  const title = NAV_ITEMS.find((n) => n.key === active)?.label || "Flux";

  return (
    <div className="win">
      <Rail />
      <div className="sys">
        <TopBar title={title} />
        <BusyBar show={busy} />
        <Scene />
        <Dock />
        <SheetLayer />
        <Toasts />
      </div>
    </div>
  );
}

export function AppShell() {
  return (
    <AppProvider>
      <Shell />
    </AppProvider>
  );
}

export default AppShell;
