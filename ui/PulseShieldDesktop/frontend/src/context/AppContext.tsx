import { createContext, ReactNode, useCallback, useContext, useEffect, useId, useMemo, useRef, useState } from "react";
import {
  ActivateLicense,
  AddVLESS,
  ApplyAutoUpdate,
  ApplyDnsNow,
  ApplyRoutingJSON,
  ApplyRuntime,
  ApplyRuntimeDeferred,
  ClearLogFiles,
  ClearSubscription,
  CreateDiagnosticsBundle,
  GetAppVersion,
  GetLogsState,
  GetRealtimeStats,
  GetServiceHealth,
  GetUIState,
  GetUpdateInfo,
  ImportDPIConfig,
  ImportServerFromClipboard,
  OpenRoutingJSON,
  ProgramGuide,
  RefreshSubscription,
  RemoveServer,
  ResyncDPIProfilesFromDisk,
  RetestServersBatch,
  RunDpiAdaptiveTuning,
  RunSelfCheckText,
  SaveRoutingJSON,
  SetActiveDPIProfile,
  SetActiveServer,
  SetDPIEnabled,
  SetDnsSettings,
  SetLicenseAPIBase,
  SetProxyEnabled,
  SetProxyPort,
  SetRoutingProfile,
  SetSubscriptionURL,
  SetTelegramChannelName,
  SetTelegramInfoURL,
  StopAll,
} from "../../wailsjs/go/main/App";
import { main } from "../../wailsjs/go/models";
import { BrowserOpenURL, EventsOn } from "../../wailsjs/runtime/runtime";
import {
  NavKey,
  ThemeKey,
  ToastState,
  ChartPt,
  LogSource,
  LogViewMode,
} from "../types";
import {
  applyAccentHex,
  cubicBezierSegmentsThroughPoints,
  DEEP_FLOW_TENSION,
  formatRuntimeFailure,
  humanizeRuntimeError,
  logPanelContent,
  logPanelPlaceholder,
  ONBOARDING_KEY,
  parseActivationPaste,
  playSoftChime,
  routingPresetTooltip,
  routingProfileLabel,
  serverPingLabel,
  uniqueRoutingProfiles,
  formatSubscriptionExpiry,
  formatBps,
  formatBytes,
  formatSessionClock,
  ROUTING_MODE_SHORT,
  BUILTIN_ROUTING_KEYS,
} from "../utils";

export type AppContextValue = {
  // Navigation
  active: NavKey;
  setActive: (k: NavKey) => void;

  // Toggle states
  proxyEnabled: boolean;
  dpiEnabled: boolean;

  // Remote data
  // eslint-disable-next-line @typescript-eslint/no-explicit-any
  uiState: any;
  // eslint-disable-next-line @typescript-eslint/no-explicit-any
  health: any;
  rt: {
    connected: boolean;
    status: string;
    downloadBps: number;
    uploadBps: number;
    sessionDownloadBytes: number;
    sessionUploadBytes: number;
    sessionElapsedSeconds: number;
    proxyProcessDetails: unknown[];
  };
  speedHistory: number[];
  logs: {
    coreLogTail: string;
    dpiLogTail: string;
    coreLogPath: string;
    dpiLogPath: string;
    coreErrors: string;
    dpiErrors: string;
    coreConnLines: string;
    dpiConnLines: string;
  };
  appVersion: string;
  updateHint: string;
  updateInfo: {
    ok: boolean;
    releaseUrl: string;
    downloadUrl: string;
    installerDownloadUrl: string;
    upToDate: boolean;
    latestTag: string;
  } | null;

  // Visual
  theme: ThemeKey;
  setTheme: (t: ThemeKey) => void;
  accentDraft: string;
  setAccentDraft: (c: string) => void;
  themeDraft: ThemeKey;
  setThemeDraft: (t: ThemeKey) => void;

  // Loading states
  busy: boolean;
  proxyPhase: "idle" | "connecting" | "error";
  toast: ToastState | null;
  setToast: (t: ToastState | null) => void;
  showOnboarding: boolean;
  setShowOnboarding: (v: boolean) => void;
  globalBusy: string | null;
  selfCheckBusy: boolean;
  testingServerId: string | null;

  // DPI tune
  dpiTuneGamingLua: boolean;
  setDpiTuneGamingLua: (v: boolean) => void;
  dpiTuneResult: main.DpiTuneResult | null;

  // Actions
  notify: (text: string, variant?: "info" | "error", openLogs?: boolean) => void;
  refreshState: () => Promise<void>;
  pollHealth: () => Promise<void>;
  setProxy: (enabled: boolean) => Promise<void>;
  setDpi: (enabled: boolean) => Promise<void>;
  setRouting: (name: string) => Promise<void>;
  setActiveServer: (id: string) => Promise<void>;
  removeServer: (id: string) => Promise<void>;
  setActiveDPI: (name: string) => Promise<void>;
  runDpiAdaptiveTuning: (opts?: { gamingLua?: boolean }) => Promise<void>;
  runRetestServersBatch: (ids: string[]) => Promise<void>;
  runRetestSingleServer: (id: string) => Promise<void>;
  refreshSubscriptionServers: (url: string) => Promise<void>;
  importServerFromClipboard: () => Promise<void>;
  clearSubscriptionFromConfig: () => Promise<void>;
  addServerFromInput: (vlessUrl: string) => Promise<void>;
  resetRoutingToRfSmart: () => Promise<void>;
  importRouting: () => Promise<void>;
  exportRouting: () => Promise<void>;
  importDpiConfig: () => Promise<void>;
  stopAll: () => Promise<void>;
  loadGuide: () => Promise<void>;
  clearLogFiles: () => Promise<void>;
  refreshLogs: () => Promise<void>;
  runSelfCheck: () => Promise<void>;
  createBundle: () => Promise<void>;
  runUpdateCheck: () => Promise<void>;
  runAutoUpdate: () => Promise<void>;
  runInstallerUpdate: () => Promise<void>;
  applyVisualSettings: (theme: ThemeKey, accent: string) => Promise<void>;
  applyProxyPort: (port: number) => Promise<void>;
  saveDnsAndApply: (preset: string, primary: string, secondary: string) => Promise<void>;
  saveTelegramMeta: (url: string, name: string) => Promise<void>;
  activateFromBotLink: (raw: string) => Promise<void>;

  // Text output
  diagText: string;
  guideText: string;

  // Derived/computed
  // eslint-disable-next-line @typescript-eslint/no-explicit-any
  routingProfilesUnique: any[];
  // eslint-disable-next-line @typescript-eslint/no-explicit-any
  activeRoutingProfileMeta: any;
  serverStats: { total: number; recommended: number; online: number; subscription: number; manual: number };
  // eslint-disable-next-line @typescript-eslint/no-explicit-any
  sortedServers: any[];
  // eslint-disable-next-line @typescript-eslint/no-explicit-any
  manualServers: any[];
  // eslint-disable-next-line @typescript-eslint/no-explicit-any
  subscriptionServers: any[];
  dashboardHints: { key: string; text: string; nav: NavKey }[];
  showVlessProtectedStrip: boolean;
  chartPaths: { linePath: string; areaPath: string; max: number };
  chartSmooth: number[];
  rtChartId: string;
};

export const AppContext = createContext<AppContextValue | null>(null);

export function useApp(): AppContextValue {
  const ctx = useContext(AppContext);
  if (!ctx) throw new Error("useApp must be used within AppProvider");
  return ctx;
}

export function AppProvider({ children }: { children: ReactNode }) {
  const [active, setActive] = useState<NavKey>("dashboard");
  const [proxyEnabled, setProxyEnabledState] = useState(true);
  const [dpiEnabled, setDpiEnabledState] = useState(false);
  const [guideText, setGuideText] = useState("");
  const [diagText, setDiagText] = useState("");
  const [uiState, setUiState] = useState<unknown>(null);
  const [theme, setTheme] = useState<ThemeKey>("dark-neon");
  const [themeDraft, setThemeDraft] = useState<ThemeKey>("dark-neon");
  const [accentDraft, setAccentDraft] = useState("#78f65c");
  const [rt, setRt] = useState<{
    connected: boolean;
    status: string;
    downloadBps: number;
    uploadBps: number;
    sessionDownloadBytes: number;
    sessionUploadBytes: number;
    sessionElapsedSeconds: number;
    proxyProcessDetails: unknown[];
  }>({ connected: false, status: "", downloadBps: 0, uploadBps: 0, sessionDownloadBytes: 0, sessionUploadBytes: 0, sessionElapsedSeconds: 0, proxyProcessDetails: [] });
  const [speedHistory, setSpeedHistory] = useState<number[]>([]);
  const [logs, setLogs] = useState({
    coreLogTail: "",
    dpiLogTail: "",
    coreLogPath: "",
    dpiLogPath: "",
    coreErrors: "",
    dpiErrors: "",
    coreConnLines: "",
    dpiConnLines: "",
  });
  const [health, setHealth] = useState<unknown>(null);
  const [appVersion, setAppVersion] = useState("");
  const [updateHint, setUpdateHint] = useState("");
  const [updateInfo, setUpdateInfo] = useState<{
    ok: boolean;
    releaseUrl: string;
    downloadUrl: string;
    installerDownloadUrl: string;
    upToDate: boolean;
    latestTag: string;
  } | null>(null);

  const [proxyPhase, setProxyPhase] = useState<"idle" | "connecting" | "error">("idle");
  const proxyApplyDeadlineRef = useRef<number | null>(null);
  const waitingDeferredApplyRef = useRef(false);
  const proxyDeferJobExpectedRef = useRef(0);

  const prevProxyRef = useRef<boolean | null>(null);
  const [dpiTuneGamingLua, setDpiTuneGamingLua] = useState(true);
  const [dpiTuneResult, setDpiTuneResult] = useState<main.DpiTuneResult | null>(null);
  const [selfCheckBusy, setSelfCheckBusy] = useState(false);
  const [testingServerId, setTestingServerId] = useState<string | null>(null);
  const [toast, setToast] = useState<ToastState | null>(null);

  const notify = useCallback((text: string, variant: "info" | "error" = "info", openLogs?: boolean) => {
    setToast({ text, variant, openLogs, dismissible: true });
  }, []);

  const tunnelOkCelebrationSent = useRef(false);
  useEffect(() => {
    if (!health) return;
    const h = health as { proxyEnabled?: boolean; proxyLocalOk?: boolean; vlessRemote?: string };
    const tunnelUp = !!(h.proxyEnabled && h.proxyLocalOk && h.vlessRemote === "ok");
    if (tunnelUp && !tunnelOkCelebrationSent.current) {
      tunnelOkCelebrationSent.current = true;
      notify("Поздравляем! Вы покинули Чебурнет.");
      playSoftChime("success");
    }
    if (!tunnelUp) {
      tunnelOkCelebrationSent.current = false;
    }
  }, [health, notify]);

  const [showOnboarding, setShowOnboarding] = useState(false);
  const [globalBusy, setGlobalBusy] = useState<string | null>(null);

  async function withBusy(label: string, fn: () => Promise<void>) {
    setGlobalBusy(label);
    try {
      await fn();
    } finally {
      setGlobalBusy(null);
    }
  }

  useEffect(() => {
    const saved = localStorage.getItem("pulseshield.theme") as ThemeKey | null;
    if (saved) {
      setTheme(saved);
      setThemeDraft(saved);
    }
    const savedAccent = localStorage.getItem("pulseshield.accentHex");
    if (savedAccent && /^#[0-9a-fA-F]{6}$/.test(savedAccent)) {
      setAccentDraft(savedAccent);
      applyAccentHex(savedAccent);
    }
    GetUIState().then((s) => {
      setUiState(s);
      const st = s as unknown as Record<string, unknown>;
      setProxyEnabledState(!!st.proxyEnabled);
      setDpiEnabledState(!!st.dpiEnabled);
    }).catch(() => {});
    GetAppVersion().then(setAppVersion).catch(() => setAppVersion("dev"));
  }, []);

  useEffect(() => {
    const off = EventsOn("update-progress", (json: string) => {
      try {
        const o = JSON.parse(json) as { phase: string; pct: number };
        if (o.phase === "download") {
          if (typeof o.pct === "number" && o.pct >= 0) {
            setGlobalBusy(`Скачивание обновления… ${o.pct}%`);
          } else {
            setGlobalBusy("Скачивание обновления…");
          }
        } else if (o.phase === "unpack") {
          setGlobalBusy("Распаковка и подготовка файлов…");
        } else if (o.phase === "install") {
          setGlobalBusy("Запуск тихого установщика…");
        }
      } catch {
        /* ignore malformed payload */
      }
    });
    return () => off();
  }, []);

  useEffect(() => {
    document.documentElement.toggleAttribute("data-proxy-on", proxyEnabled);
  }, [proxyEnabled]);

  useEffect(() => {
    if (prevProxyRef.current === null) {
      prevProxyRef.current = proxyEnabled;
      return;
    }
    if (prevProxyRef.current !== proxyEnabled) {
      playSoftChime("toggle");
    }
    prevProxyRef.current = proxyEnabled;
  }, [proxyEnabled]);

  useEffect(() => {
    if (!proxyEnabled) {
      setSpeedHistory([]);
    }
  }, [proxyEnabled]);

  useEffect(() => {
    if (!toast) return;
    const ms =
      toast.variant === "error" && toast.openLogs ? 8000 : toast.variant === "error" ? 6000 : 3200;
    const t = setTimeout(() => setToast(null), ms);
    return () => clearTimeout(t);
  }, [toast]);

  useEffect(() => {
    if (localStorage.getItem(ONBOARDING_KEY) === "1") return;
    const t = setTimeout(() => setShowOnboarding(true), 500);
    return () => clearTimeout(t);
  }, []);

  useEffect(() => {
    document.documentElement.setAttribute("data-theme", theme);
    localStorage.setItem("pulseshield.theme", theme);
    const a = localStorage.getItem("pulseshield.accentHex");
    if (a && /^#[0-9a-fA-F]{6}$/.test(a)) {
      applyAccentHex(a);
    }
  }, [theme]);

  useEffect(() => {
    const applyRt = (raw: Record<string, unknown>) => {
      setRt(raw as typeof rt);
      const total = (Number(raw.downloadBps) || 0) + (Number(raw.uploadBps) || 0);
      setSpeedHistory((prev) => [...prev.slice(-59), total]);
    };
    void GetRealtimeStats()
      .then((s) => applyRt(s as unknown as Record<string, unknown>))
      .catch(() => {});
    const off = EventsOn("realtime-stats", (json: string) => {
      try {
        applyRt(JSON.parse(json) as Record<string, unknown>);
      } catch {
        /* ignore */
      }
    });
    return () => off();
  }, []);

  const lastHealthRef = useRef<string>("");
  const pollHealth = useCallback(async () => {
    try {
      const h = await GetServiceHealth();
      const key = JSON.stringify(h);
      if (key !== lastHealthRef.current) {
        lastHealthRef.current = key;
        setHealth(h);
      }
    } catch {}
  }, []);
  const pollHealthRef = useRef(pollHealth);
  pollHealthRef.current = pollHealth;

  useEffect(() => {
    void pollHealth();
    const id = setInterval(pollHealth, 4000);
    return () => clearInterval(id);
  }, [pollHealth]);

  useEffect(() => {
    const off = EventsOn("dpi-connectivity", () => {
      void GetServiceHealth().then(setHealth).catch(() => {});
    });
    return () => off();
  }, []);

  // Восстановление после выхода из сна: бэкенд шлёт "system-resumed".
  // WebView2 после сна может «залипнуть» — принудительно обновляем данные и,
  // если документ перестал отвечать (GPU/контекст потерян), мягко перезагружаем UI.
  useEffect(() => {
    const off = EventsOn("system-resumed", () => {
      // Перезапрашиваем состояние и здоровье (фоновые таймеры могли встать).
      lastHealthRef.current = "";
      void GetServiceHealth().then(setHealth).catch(() => {});
      void refreshState();
      void GetRealtimeStats().then((s) => setRt(s as typeof rt)).catch(() => {});
      // Если за 1.5с после resume рендер не ожил — перезагрузить webview (состояние с бэкенда).
      const probe = document.documentElement;
      const before = probe.clientHeight;
      window.setTimeout(() => {
        try {
          if (document.hidden || probe.clientHeight === 0 || before === 0) {
            window.location.reload();
          }
        } catch {
          window.location.reload();
        }
      }, 1500);
    });
    return () => off();
  }, []); // eslint-disable-line react-hooks/exhaustive-deps

  useEffect(() => {
    if (active !== "settings") return;
    setThemeDraft(theme);
    const a = localStorage.getItem("pulseshield.accentHex");
    if (a && /^#[0-9a-fA-F]{6}$/.test(a)) {
      setAccentDraft(a);
    }
  }, [active, theme]);

  const dpiResyncDoneRef = useRef(false);
  useEffect(() => {
    if (active !== "dpi") return;
    if (dpiResyncDoneRef.current) return; // only once per session
    dpiResyncDoneRef.current = true;
    ResyncDPIProfilesFromDisk()
      .then(() => void refreshState())
      .catch(() => {});
  }, [active]); // eslint-disable-line react-hooks/exhaustive-deps

  useEffect(() => {
    if (active !== "logs") return;
    const pollLogs = async () => {
      try {
        const l = await GetLogsState();
        setLogs(l as typeof logs);
      } catch {}
    };
    pollLogs();
    const id = setInterval(pollLogs, 3000);
    return () => clearInterval(id);
  }, [active]);

  async function refreshState() {
    try {
      const s = await GetUIState();
      setUiState(s);
      const st = s as unknown as Record<string, unknown>;
      setProxyEnabledState(!!st.proxyEnabled);
      setDpiEnabledState(!!st.dpiEnabled);
    } catch {}
  }

  const refreshStateRef = useRef(() => {
    void refreshState();
  });
  refreshStateRef.current = () => void refreshState();

  useEffect(() => {
    const offApply = EventsOn("runtime-apply", (json: string) => {
      let o: { done?: boolean; error?: string; job?: number };
      try {
        o = JSON.parse(json);
      } catch {
        return;
      }
      if (!o?.done) return;
      void pollHealthRef.current();
      if (!waitingDeferredApplyRef.current) return;
      if (o.job != null && Number(o.job) !== proxyDeferJobExpectedRef.current) {
        return;
      }
      waitingDeferredApplyRef.current = false;
      if (proxyApplyDeadlineRef.current != null) {
        window.clearTimeout(proxyApplyDeadlineRef.current);
        proxyApplyDeadlineRef.current = null;
      }
      proxyDeferJobExpectedRef.current = 0;
      const err = typeof o.error === "string" ? o.error.trim() : "";
      if (err) {
        setProxyPhase("error");
        notify(humanizeRuntimeError(err), "error", true);
      } else {
        setProxyPhase("idle");
      }
      void refreshStateRef.current();
    });
    const offWatch = EventsOn("runtime-watchdog", (json: string) => {
      waitingDeferredApplyRef.current = false;
      proxyDeferJobExpectedRef.current = 0;
      if (proxyApplyDeadlineRef.current != null) {
        window.clearTimeout(proxyApplyDeadlineRef.current);
        proxyApplyDeadlineRef.current = null;
      }
      setProxyPhase("idle");
      let detail = "";
      try {
        const o = JSON.parse(json) as { reason?: string; detail?: string };
        detail = (o.detail || o.reason || "").trim();
      } catch {
        /* ignore */
      }
      notify(
        detail
          ? `Восстановление после зависания сетевого режима: ${detail}`
          : "Восстановление после зависания сетевого режима — прокси и DPI остановлены.",
        "error",
        true,
      );
      void refreshStateRef.current();
    });
    const offLic = EventsOn("license-applied", () => {
      void refreshStateRef.current();
    });
    return () => {
      offApply();
      offWatch();
      offLic();
    };
  }, [notify]);

  async function setProxy(enabled: boolean) {
    setProxyEnabledState(enabled);
    setProxyPhase("connecting");
    waitingDeferredApplyRef.current = true;
    if (proxyApplyDeadlineRef.current != null) {
      window.clearTimeout(proxyApplyDeadlineRef.current);
      proxyApplyDeadlineRef.current = null;
    }
    proxyApplyDeadlineRef.current = window.setTimeout(() => {
      proxyApplyDeadlineRef.current = null;
      if (!waitingDeferredApplyRef.current) return;
      waitingDeferredApplyRef.current = false;
      proxyDeferJobExpectedRef.current = 0;
      setProxyPhase("error");
      notify("Таймаут применения прокси — проверьте лог ядра и попробуйте снова.", "error", true);
      void refreshState();
    }, 56000);

    try {
      await SetProxyEnabled(enabled);
      const job = await ApplyRuntimeDeferred();
      proxyDeferJobExpectedRef.current = Number(job) || 0;
    } catch (e: unknown) {
      waitingDeferredApplyRef.current = false;
      proxyDeferJobExpectedRef.current = 0;
      if (proxyApplyDeadlineRef.current != null) {
        window.clearTimeout(proxyApplyDeadlineRef.current);
        proxyApplyDeadlineRef.current = null;
      }
      setProxyPhase("error");
      notify(formatRuntimeFailure("Прокси", e), "error", true);
      await refreshState();
    }
  }

  async function setDpi(enabled: boolean) {
    setDpiEnabledState(enabled);
    await withBusy("Применяю DPI и конфигурацию…", async () => {
      try {
        await SetDPIEnabled(enabled);
        await ApplyRuntime();
        await refreshState();
        await pollHealthRef.current();
      } catch (e: unknown) {
        const line = formatRuntimeFailure("DPI", e);
        notify(line, "error", true);
        await refreshState();
        await pollHealthRef.current();
      }
    });
  }

  async function setRouting(name: string) {
    await withBusy("Переключаю профиль маршрутизации…", async () => {
      try {
        await SetRoutingProfile(name);
        await ApplyRuntime();
        notify("Профиль маршрутизации переключён и применён к ядру");
        await refreshState();
        await pollHealthRef.current();
      } catch (e: unknown) {
        notify(`Ошибка выбора профиля маршрутизации: ${String(e)}`, "error");
      }
    });
  }

  async function setActiveServer(id: string) {
    const st = uiState as Record<string, unknown> | null;
    const srvs = (st?.servers as unknown[]) ?? [];
    const sel = (srvs as Record<string, unknown>[]).find((x) => x.id === id);
    if (sel?.premium && !(st as Record<string, unknown>)?.licenseActive) {
      notify("Этот узел отмечен как Premium от авторов — активируйте поддержку на вкладке «Поддержка».", "error");
      return;
    }
    await withBusy("Переключаю сервер и ядро…", async () => {
      try {
        await SetActiveServer(id);
        const s = await GetUIState();
        if ((s as unknown as Record<string, unknown>).proxyEnabled) {
          await ApplyRuntime();
        }
        notify("Активный сервер обновлён");
        await refreshState();
      } catch (e: unknown) {
        const line = formatRuntimeFailure("Сервер", e);
        notify(line, "error", true);
        await refreshState();
      }
    });
  }

  async function removeServer(id: string) {
    if (!window.confirm("Удалить этот сервер из списка?")) return;
    await withBusy("Удаляю сервер и применяю конфигурацию…", async () => {
      try {
        await RemoveServer(id);
        await ApplyRuntime();
        notify("Сервер удалён, конфигурация применена");
        await refreshState();
      } catch (e: unknown) {
        const line = formatRuntimeFailure("Удаление сервера", e);
        notify(line, "error", true);
        await refreshState();
      }
    });
  }

  async function setActiveDPI(name: string) {
    await withBusy("Переключаю профиль DPI…", async () => {
      try {
        await SetActiveDPIProfile(name);
        await ApplyRuntime();
        await refreshState();
      } catch (e: unknown) {
        const line = formatRuntimeFailure("Профиль DPI", e);
        notify(line, "error", true);
        await refreshState();
      }
    });
  }

  async function runDpiAdaptiveTuning(opts?: { gamingLua?: boolean }) {
    const gamingLua = opts?.gamingLua ?? dpiTuneGamingLua;
    await withBusy("Диагностика DPI и проверка доступа к сервисам…", async () => {
      try {
        const r = await RunDpiAdaptiveTuning(gamingLua);
        setDpiTuneResult(r);
        if (r.message) {
          notify(r.message, r.newProfileName ? "info" : "error", r.newProfileName ? undefined : true);
        }
        await refreshState();
      } catch (e: unknown) {
        const line = formatRuntimeFailure("Автоподбор DPI", e);
        notify(line, "error", true);
        await refreshState();
      }
    });
  }

  async function runRetestServersBatch(ids: string[]) {
    const clean = ids.map((id) => String(id || "").trim()).filter(Boolean);
    if (clean.length === 0) {
      notify("Нет серверов для проверки", "error");
      return;
    }
    await withBusy(`Тестирую скорость и пинг (${clean.length} узлов)…`, async () => {
      try {
        await RetestServersBatch(clean);
        notify(`Тест скорости завершён (${clean.length} узлов)`);
        await refreshState();
      } catch (e: unknown) {
        notify(formatRuntimeFailure("Проверка серверов", e), "error");
      }
    });
  }

  async function runRetestSingleServer(id: string) {
    const clean = String(id || "").trim();
    if (!clean || testingServerId) return;
    setTestingServerId(clean);
    try {
      await RetestServersBatch([clean]);
      const s = await GetUIState();
      setUiState(s);
      const st = s as unknown as Record<string, unknown>;
      const srvs = (st?.servers as Record<string, unknown>[]) ?? [];
      const found = srvs.find((x) => x.id === clean);
      if (found) {
        if (found.lastCheckOk) {
          const mbps = Number(found.lastCheckDownloadMbps) || 0;
          notify(`${found.name}: ${found.lastCheckLatencyMs} ms${mbps > 0 ? ` · ~${mbps.toFixed(0)} Мбит/с` : ""}`);
        } else {
          notify(`${found.name}: сервер не ответил`, "error");
        }
      } else {
        notify("Сервер проверен");
      }
    } catch (e: unknown) {
      notify(formatRuntimeFailure("Проверка сервера", e), "error");
    } finally {
      setTestingServerId(null);
    }
  }

  async function refreshSubscriptionServers(url: string) {
    if (!url.trim()) {
      notify("Укажите Subscription URL", "error");
      return;
    }
    await withBusy("Обновляю подписку и проверяю серверы…", async () => {
      try {
        await SetSubscriptionURL(url.trim());
        await RefreshSubscription();
        notify("Подписка обновлена, список серверов загружен");
        await refreshState();
      } catch (e: unknown) {
        const line = formatRuntimeFailure("Подписка", e);
        notify(line, "error", true);
      }
    });
  }

  async function importServerFromClipboard() {
    await withBusy("Импортирую из буфера обмена…", async () => {
      try {
        await ImportServerFromClipboard();
        await refreshState();
        notify("Импорт из буфера выполнен (vless:// или ссылка подписки)");
      } catch (e: unknown) {
        const line = formatRuntimeFailure("Импорт из буфера", e);
        notify(line, "error");
      }
    });
  }

  async function clearSubscriptionFromConfig() {
    if (!window.confirm("Удалить подписку и все узлы из неё? Одиночные ключи останутся.")) return;
    await withBusy("Удаляю подписку…", async () => {
      try {
        await ClearSubscription();
        notify("Подписка удалена из конфигурации");
        await refreshState();
      } catch (e: unknown) {
        notify(formatRuntimeFailure("Подписка", e), "error");
      }
    });
  }

  async function addServerFromInput(vlessUrl: string) {
    if (!vlessUrl.trim()) return;
    await withBusy("Добавляю сервер…", async () => {
      try {
        await AddVLESS(vlessUrl.trim());
        notify("Сервер добавлен из VLESS");
        await refreshState();
      } catch (e: unknown) {
        notify(`Ошибка добавления сервера: ${String(e)}`, "error");
      }
    });
  }

  async function resetRoutingToRfSmart() {
    await withBusy("Сбрасываю профиль на rf-smart-split…", async () => {
      try {
        await SetRoutingProfile("rf-smart-split");
        await ApplyRuntime();
        notify("Активен профиль «РФ: умный сплит»");
        await refreshState();
      } catch (e: unknown) {
        const line = formatRuntimeFailure("Маршрутизация", e);
        notify(line, "error");
      }
    });
  }

  async function importRouting() {
    await withBusy("Импортирую маршрутизацию…", async () => {
      try {
        const f = await OpenRoutingJSON();
        if (!f) return;
        await ApplyRoutingJSON((f as { content: string }).content);
        notify(`Маршрутизация импортирована из ${(f as { path: string }).path}`);
        if (proxyEnabled) {
          await ApplyRuntime();
        }
        await refreshState();
      } catch (e: unknown) {
        notify(`Ошибка импорта маршрутизации: ${String(e)}`, "error");
      }
    });
  }

  async function exportRouting() {
    await withBusy("Сохраняю JSON маршрутизации…", async () => {
      try {
        const s = await GetUIState();
        const st = s as unknown as Record<string, unknown>;
        const uniq = uniqueRoutingProfiles(st.routingProfiles as unknown[]);
        const p = uniq.find((x: unknown) => (x as Record<string, unknown>).name === st.routingProfile) || uniq[0];
        if (!p) {
          notify("Нет профиля для экспорта", "error");
          return;
        }
        const json = JSON.stringify(p, null, 2);
        const pathOut = await SaveRoutingJSON(json);
        if (!pathOut) return;
        notify(`Маршрутизация экспортирована: ${pathOut}`);
      } catch (e: unknown) {
        notify(`Ошибка экспорта маршрутизации: ${String(e)}`, "error");
      }
    });
  }

  async function importDpiConfig() {
    await withBusy("Импортирую DPI-конфиг…", async () => {
      try {
        const p = await ImportDPIConfig();
        if (!p) return;
        notify(`DPI-конфиг импортирован: ${p}`);
        await refreshState();
      } catch (e: unknown) {
        notify(`Ошибка импорта DPI: ${String(e)}`, "error");
      }
    });
  }

  async function stopAll() {
    await withBusy("Останавливаю все сервисы…", async () => {
      try {
        await StopAll();
        setProxyEnabledState(false);
        setDpiEnabledState(false);
        notify("Все сервисы остановлены");
        await refreshState();
      } catch (e: unknown) {
        const line = formatRuntimeFailure("Стоп всё", e);
        notify(line, "error", true);
        await refreshState();
      }
    });
  }

  async function loadGuide() {
    try {
      const txt = await ProgramGuide();
      setGuideText(txt);
      setActive("guide");
    } catch (e: unknown) {
      notify(`Ошибка загрузки справочника: ${String(e)}`, "error");
    }
  }

  async function clearLogFiles() {
    await withBusy("Очищаю файлы логов…", async () => {
      try {
        await ClearLogFiles();
        const l = await GetLogsState();
        setLogs(l as typeof logs);
        notify("Файлы логов на диске очищены");
        playSoftChime("success");
      } catch (e: unknown) {
        notify(`Очистка логов: ${String(e)}`, "error");
      }
    });
  }

  async function refreshLogs() {
    try {
      const l = await GetLogsState();
      setLogs(l as typeof logs);
    } catch {
      /* ignore */
    }
  }

  async function runSelfCheck() {
    setSelfCheckBusy(true);
    setDiagText("Самопроверка выполняется (обычно несколько секунд, не мгновенно)…");
    notify("Самопроверка запущена — дождитесь результата.");
    try {
      const t = await RunSelfCheckText();
      setDiagText(t);
      notify("Самопроверка завершена");
    } catch (e: unknown) {
      setDiagText("");
      notify(`Ошибка самопроверки: ${String(e)}`, "error");
    } finally {
      setSelfCheckBusy(false);
    }
  }

  async function createBundle() {
    await withBusy("Создаю диагностический архив…", async () => {
      try {
        const p = await CreateDiagnosticsBundle();
        notify(`Диагностический архив создан: ${p}`);
      } catch (e: unknown) {
        notify(`Ошибка создания архива: ${String(e)}`, "error");
      }
    });
  }

  async function runUpdateCheck() {
    await withBusy("Проверяю обновления на GitHub…", async () => {
      try {
        const info = await GetUpdateInfo();
        setUpdateInfo({
          ok: !!info.ok,
          releaseUrl: info.releaseUrl || "",
          downloadUrl: info.downloadUrl || "",
          installerDownloadUrl: info.installerDownloadUrl || "",
          upToDate: !!info.upToDate,
          latestTag: info.latestTag || "",
        });
        const msg = info.message || (info.upToDate ? `У вас актуальная версия (${appVersion})` : `Доступна новая версия: ${info.latestTag}`);
        setUpdateHint(msg);
        if (info.ok) {
          if (!info.upToDate) {
            playSoftChime("success");
          }
          notify(msg, "info");
        } else {
          notify(msg, "error");
        }
      } catch (e: unknown) {
        const errMsg = `Ошибка проверки обновлений: ${String(e)}`;
        setUpdateInfo(null);
        setUpdateHint(errMsg);
        notify(errMsg, "error");
      }
    });
  }

  async function runAutoUpdate() {
    if (!updateInfo?.downloadUrl && !updateInfo?.installerDownloadUrl) return;
    const targetUrl = updateInfo.downloadUrl?.toLowerCase().endsWith(".zip")
      ? updateInfo.downloadUrl
      : updateInfo.installerDownloadUrl || updateInfo.downloadUrl;
    setGlobalBusy("Скачивание обновления…");
    try {
      await ApplyAutoUpdate(targetUrl);
    } catch (e: unknown) {
      notify(`Обновление: ${String(e)}`, "error");
    } finally {
      setGlobalBusy(null);
    }
  }

  async function runInstallerUpdate() {
    if (!updateInfo?.installerDownloadUrl) return;
    setGlobalBusy("Скачивание установщика…");
    try {
      await ApplyAutoUpdate(updateInfo.installerDownloadUrl);
    } catch (e: unknown) {
      notify(`Установщик: ${String(e)}`, "error");
    } finally {
      setGlobalBusy(null);
    }
  }

  async function applyVisualSettings(themeVal: ThemeKey, accent: string) {
    await withBusy("Применяю тему и акцент…", async () => {
      await new Promise((r) => setTimeout(r, 220));
      setTheme(themeVal);
      setThemeDraft(themeVal);
      localStorage.setItem("pulseshield.theme", themeVal);
      localStorage.setItem("pulseshield.accentHex", accent);
      setAccentDraft(accent);
      applyAccentHex(accent);
      playSoftChime("success");
      setToast({ text: "Тема и акцент применены", variant: "info", dismissible: true });
    });
  }

  async function applyProxyPort(port: number) {
    const p = Math.floor(Number(port));
    if (p < 1 || p > 65535) {
      notify("Порт должен быть от 1 до 65535", "error");
      return;
    }
    await withBusy("Меняю порт и перезапускаю ядро…", async () => {
      try {
        await SetProxyPort(p);
        await ApplyRuntime();
        notify(`Порт локального прокси: ${p}. Переподключите приложения при необходимости.`);
        await refreshState();
      } catch (e: unknown) {
        const line = formatRuntimeFailure("Порт / ядро", e);
        notify(line, "error", true);
        await refreshState();
      }
    });
  }

  async function saveDnsAndApply(preset: string, primary: string, secondary: string) {
    await withBusy("Применяю DNS к адаптерам…", async () => {
      try {
        await SetDnsSettings(preset, primary, secondary);
        await ApplyDnsNow();
        const label = preset === "dhcp" ? "Авто (DHCP)" : preset === "xbox" ? "Xbox DNS" : "Свои DNS";
        notify(`Настройки DNS «${label}» применены к активным адаптерам.`);
        await refreshState();
      } catch (e: unknown) {
        notify(`DNS: ${String(e)}`, "error");
      }
    });
  }

  async function saveTelegramMeta(url: string, name: string) {
    await withBusy("Сохраняю настройки Telegram…", async () => {
      try {
        await SetTelegramInfoURL(url);
        await SetTelegramChannelName(name);
        notify("Ссылка и название Telegram сохранены");
        await refreshState();
      } catch (e: unknown) {
        notify(String(e), "error");
      }
    });
  }

  async function activateFromBotLink(raw: string) {
    let parsed: { code: string; apiBase?: string };
    try {
      parsed = parseActivationPaste(raw);
    } catch (e: unknown) {
      notify(e instanceof Error ? e.message : String(e), "error");
      return;
    }
    await withBusy("Активирую лицензию…", async () => {
      try {
        if (parsed.apiBase?.trim()) {
          await SetLicenseAPIBase(parsed.apiBase.trim());
        }
        await ActivateLicense(parsed.code);
        notify("Лицензия активирована.");
        await refreshState();
      } catch (e: unknown) {
        notify(e instanceof Error ? e.message : String(e), "error");
      }
    });
  }

  // Derived / computed
  const rtChartId = useId().replace(/:/g, "");

  const chartSmooth = useMemo(() => {
    const h = speedHistory;
    if (h.length === 0) return [];
    if (h.length < 3) return h.slice();
    return h.map((v, i) => {
      const prev = h[i - 1] ?? v;
      const next = h[i + 1] ?? v;
      return (prev + v + next) / 3;
    });
  }, [speedHistory]);

  const chartPaths = useMemo(() => {
    if (chartSmooth.length === 0) {
      return { linePath: "", areaPath: "", max: 1 };
    }
    const max = Math.max(...chartSmooth, 1) * 0.62;
    const n = chartSmooth.length;
    const pts: ChartPt[] = chartSmooth.map((v, i) => {
      const x = (i / Math.max(n - 1, 1)) * 100;
      const y = 100 - (v / max) * 84 - 10;
      return { x, y };
    });
    const curves = cubicBezierSegmentsThroughPoints(pts, DEEP_FLOW_TENSION);
    const linePath = `M ${pts[0].x},${pts[0].y}${curves}`;
    const areaPath = `M 0,100 L ${pts[0].x},${pts[0].y}${curves} L 100,100 Z`;
    return { linePath, areaPath, max };
  }, [chartSmooth]);

  const busy = !!globalBusy;

  const routingProfilesUnique = useMemo(
    () => uniqueRoutingProfiles((uiState as Record<string, unknown> | null)?.routingProfiles as unknown[]),
    [(uiState as Record<string, unknown> | null)?.routingProfiles],
  );

  const activeRoutingProfileMeta = useMemo(() => {
    const st = uiState as Record<string, unknown> | null;
    if (!routingProfilesUnique.length || !st?.routingProfile) return null;
    const target =
      st.routingProfile === "bypass-russia"
        ? "rf-smart-split"
        : st.routingProfile === "all-proxy"
        ? "default"
        : st.routingProfile;
    return (
      routingProfilesUnique.find((x: unknown) => {
        const name = (x as Record<string, unknown>).name;
        return name === st.routingProfile || name === target;
      }) || null
    );
  }, [(uiState as Record<string, unknown> | null)?.routingProfile, routingProfilesUnique]);

  const serverStats = useMemo(() => {
    const st = uiState as Record<string, unknown> | null;
    const servers = Array.isArray(st?.servers) ? (st as Record<string, unknown[]>).servers : [];
    const recommended = (servers as Record<string, unknown>[]).filter((s) => s?.recommended).length;
    const online = (servers as Record<string, unknown>[]).filter((s) => s?.lastCheckOk).length;
    const subscription = (servers as Record<string, unknown>[]).filter((s) => s?.fromSubscription).length;
    const manual = servers.length - subscription;
    return { total: servers.length, recommended, online, subscription, manual };
  }, [(uiState as Record<string, unknown> | null)?.servers]);

  const sortedServers = useMemo(() => {
    const st = uiState as Record<string, unknown> | null;
    const servers = Array.isArray(st?.servers) ? [...(st as Record<string, unknown[]>).servers] : [];
    return (servers as Record<string, unknown>[]).sort((a, b) => {
      const aPing = Number(a?.lastCheckLatencyMs || Number.MAX_SAFE_INTEGER);
      const bPing = Number(b?.lastCheckLatencyMs || Number.MAX_SAFE_INTEGER);
      if (aPing !== bPing) return aPing - bPing;
      const aSpeed = Number(a?.lastCheckDownloadMbps || 0);
      const bSpeed = Number(b?.lastCheckDownloadMbps || 0);
      return bSpeed - aSpeed;
    });
  }, [(uiState as Record<string, unknown> | null)?.servers]);

  const manualServers = useMemo(
    () => sortedServers.filter((s: unknown) => !(s as Record<string, unknown>)?.fromSubscription),
    [sortedServers],
  );
  const subscriptionServers = useMemo(
    () => sortedServers.filter((s: unknown) => (s as Record<string, unknown>)?.fromSubscription),
    [sortedServers],
  );

  const dashboardHints = useMemo(() => {
    const out: { key: string; text: string; nav: NavKey }[] = [];
    const st = uiState as Record<string, unknown> | null;
    const h = health as Record<string, unknown> | null;
    if (!st) return out;
    const servers = (st.servers as unknown[]) || [];
    if (servers.length === 0) {
      out.push({ key: "no-vless", text: "Добавьте VLESS-сервер — без него прокси не сможет построить туннель.", nav: "servers" });
    } else if (proxyEnabled && !st.activeServerId) {
      out.push({ key: "no-active", text: "Выберите активный сервер в списке на вкладке «Серверы».", nav: "servers" });
    }
    if (h && proxyEnabled && h.vlessRemote && h.vlessRemote !== "ok" && !String(h.vlessRemote).includes("нет активного")) {
      out.push({ key: "vless-bad", text: "Прокси включён, но до VLESS-сервера нет TCP — проверьте интернет, хост и порт.", nav: "servers" });
    }
    if (h && dpiEnabled && !h.dpiRunning) {
      out.push({ key: "dpi-down", text: "DPI включён в настройках, но процесс не найден — смотрите логи DPI или запуск от администратора.", nav: "logs" });
    }
    if (h?.dpiEnabled && h?.dpiRunning && h?.dpiConnectivityWarn) {
      out.push({
        key: "dpi-probe-warn",
        text: "Проверка DPI (generate_204) не прошла — возможно, провайдер подкрутил фильтрацию. Откройте вкладку DPI и запустите автоподбор заново.",
        nav: "dpi",
      });
    }
    if (h && !proxyEnabled && h.proxyPortOccupied) {
      out.push({ key: "port-occupied", text: "Прокси выключен, но порт всё ещё слушает — смените порт в настройках или перезапустите приложение.", nav: "settings" });
    }
    return out.slice(0, 2);
  }, [uiState, health, proxyEnabled, dpiEnabled]);

  const showVlessProtectedStrip = useMemo(() => {
    const h = health as Record<string, unknown> | null;
    return !!(h && h.proxyEnabled && h.proxyLocalOk && h.vlessRemote === "ok");
  }, [health]);

  const value: AppContextValue = {
    active,
    setActive,
    proxyEnabled,
    dpiEnabled,
    uiState,
    health,
    rt,
    speedHistory,
    logs,
    appVersion,
    updateHint,
    updateInfo,
    theme,
    setTheme,
    accentDraft,
    setAccentDraft,
    themeDraft,
    setThemeDraft,
    busy,
    proxyPhase,
    toast,
    setToast,
    showOnboarding,
    setShowOnboarding,
    globalBusy,
    selfCheckBusy,
    testingServerId,
    dpiTuneGamingLua,
    setDpiTuneGamingLua,
    dpiTuneResult,
    notify,
    refreshState,
    pollHealth,
    setProxy,
    setDpi,
    setRouting,
    setActiveServer,
    removeServer,
    setActiveDPI,
    runDpiAdaptiveTuning,
    runRetestServersBatch,
    runRetestSingleServer,
    refreshSubscriptionServers,
    importServerFromClipboard,
    clearSubscriptionFromConfig,
    addServerFromInput,
    resetRoutingToRfSmart,
    importRouting,
    exportRouting,
    importDpiConfig,
    stopAll,
    loadGuide,
    clearLogFiles,
    refreshLogs,
    runSelfCheck,
    createBundle,
    runUpdateCheck,
    runAutoUpdate,
    runInstallerUpdate,
    applyVisualSettings,
    applyProxyPort,
    saveDnsAndApply,
    saveTelegramMeta,
    activateFromBotLink,
    diagText,
    guideText,
    routingProfilesUnique,
    activeRoutingProfileMeta,
    serverStats,
    sortedServers,
    manualServers,
    subscriptionServers,
    dashboardHints,
    showVlessProtectedStrip,
    chartPaths,
    chartSmooth,
    rtChartId,
  };

  return <AppContext.Provider value={value}>{children}</AppContext.Provider>;
}


