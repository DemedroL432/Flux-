import {
  createContext,
  useCallback,
  useContext,
  useEffect,
  useMemo,
  useState,
  type ReactNode,
} from "react";
import {
  ActivateLicense,
  AddUserRoutingRule,
  ApplyDnsNow,
  ApplyLatestUpdate,
  ApplyRuntime,
  CheckForUpdates,
  ClearLogFiles,
  ClearSubscription,
  CloseToTray,
  CreateDiagnosticsBundle,
  GetAppVersion,
  GetAvailableDpiLists,
  GetDpiListContent,
  GetLogsState,
  GetRealtimeStats,
  GetRoutingSplitView,
  GetServiceHealth,
  GetUIState,
  GetUpdateInfo,
  GetZapretListsPathHint,
  ImportServerFromClipboard,
  ImportServersFromText,
  OpenLogsFolder,
  PreviewSubscription,
  ProgramGuide,
  RefreshSubscription,
  RemoveServer,
  RemoveUserRoutingRuleAt,
  RequestQuit,
  RetestServer,
  RetestServersBatch,
  RunDpiAdaptiveTuning,
  RunSelfCheckText,
  SetActiveDPIProfile,
  SetActiveServer,
  SetAutoUpdateEnabled,
  SetAutostartEnabled,
  SetDPIEnabled,
  SetDnsSettings,
  SetDpiListContent,
  SetProxyEnabled,
  SetProxyMode,
  SetProxyPort,
  SetRoutingProfile,
  SetSubscriptionURL,
  ToggleDpiList,
  VerifyRouteTarget,
} from "../../wailsjs/go/main/App";
import { EventsOn } from "../../wailsjs/runtime/runtime";
import type { AccentKey, DpiSubTab, NavKey, ThemeKey, ToastState } from "../types";
import {
  ACCENT_STORAGE_KEY,
  DEFAULT_ACCENT,
  DEFAULT_THEME,
  THEME_STORAGE_KEY,
  normalizeAccent,
  normalizeTheme,
} from "../themes";

type Any = Record<string, unknown>;

type Ctx = {
  // Навигация
  active: NavKey;
  setActive: (k: NavKey) => void;
  dpiSubTab: DpiSubTab;
  setDpiSubTab: (tab: DpiSubTab) => void;
  openDpiTab: (tab: DpiSubTab) => void;

  // Тема
  theme: ThemeKey;
  accent: AccentKey;
  setTheme: (t: ThemeKey) => void;
  setAccent: (a: AccentKey) => void;

  // Состояние движков
  uiState: Any;
  health: Any;
  rt: Any;
  logs: Any;
  appVersion: string;
  busy: boolean;
  proxyPhase: "idle" | "connecting" | "connected";
  proxyEnabled: boolean;
  dpiEnabled: boolean;
  proxyMode: string;
  setProxy: (on: boolean) => Promise<void>;
  setProxyMode: (mode: string) => Promise<void>;
  setDpi: (on: boolean) => Promise<void>;
  stopAll: () => Promise<void>;

  // Серверы и маршруты
  servers: Any[];
  activeServer: Any | null;
  splitRows: Any[];
  routingProfile: string;
  setRouting: (name: string) => Promise<void>;
  addRule: (pattern: string, target: string) => Promise<void>;
  removeRule: (index: number) => Promise<void>;
  verifyTarget: (q: string) => Promise<Any | null>;
  selectServer: (id: string) => Promise<void>;
  removeServer: (id: string) => Promise<void>;
  retestServer: (id: string) => Promise<void>;
  retestAll: () => Promise<void>;

  // Подписка
  previewSubscription: (url: string) => Promise<Any | null>;
  importText: (text: string) => Promise<number>;
  importFromClipboard: () => Promise<void>;
  applySubscription: (url: string) => Promise<void>;
  refreshSubscription: () => Promise<void>;
  clearSubscription: () => Promise<void>;

  // DPI
  dpiLists: Any[];
  refreshDpiLists: () => Promise<void>;
  toggleDpiList: (name: string, on: boolean) => Promise<void>;
  readDpiList: (name: string) => Promise<Any | null>;
  writeDpiList: (name: string, entries: string[]) => Promise<void>;
  runDpiTuning: () => Promise<Any | null>;
  dpiTune: Any | null;
  setDpiProfile: (name: string) => Promise<void>;
  listsPath: string;

  // Сеть, диагностика и обновления
  selfCheck: string;
  runSelfCheck: () => Promise<void>;
  exportBundle: () => Promise<void>;
  proxyPortDraft: number;
  applyProxyPort: (p: number) => Promise<void>;
  applyDns: (preset: string, p: string, s: string) => Promise<void>;
  activateLicense: (code: string) => Promise<void>;
  autoUpdateEnabled: boolean;
  setAutoUpdateEnabled: (on: boolean) => Promise<void>;
  autostartEnabled: boolean;
  setAutostartEnabled: (on: boolean) => Promise<void>;
  checkUpdates: () => Promise<void>;
  updateInfo: Any | null;
  updateProgress: { phase: string; pct: number } | null;
  installUpdate: () => Promise<void>;
  guide: string;
  clearLogs: () => Promise<void>;
  openLogsFolder: () => Promise<void>;
  quit: () => void;
  closeToTray: () => void;

  // Уведомления
  toast: ToastState | null;
  notify: (text: string, variant?: "info" | "error") => void;
  dismissToast: () => void;
  refresh: () => Promise<void>;
};

const AppCtx = createContext<Ctx | null>(null);

function readLS(key: string): string {
  try {
    return localStorage.getItem(key) || "";
  } catch {
    return "";
  }
}

function writeLS(key: string, val: string) {
  try {
    localStorage.setItem(key, val);
  } catch {
    /* игнорируем ошибку доступа к localStorage */
  }
}

export function AppProvider({ children }: { children: ReactNode }) {
  const [active, setActiveRaw] = useState<NavKey>("shield");
  const [dpiSubTab, setDpiSubTab] = useState<DpiSubTab>("profiles");
  const [theme, setThemeRaw] = useState<ThemeKey>(() => {
    const t = normalizeTheme(readLS(THEME_STORAGE_KEY));
    return t || DEFAULT_THEME;
  });
  const [accent, setAccentRaw] = useState<AccentKey>(() => {
    const a = normalizeAccent(readLS(ACCENT_STORAGE_KEY));
    return a || DEFAULT_ACCENT;
  });

  const [uiState, setUiState] = useState<Any>({});
  const [health, setHealth] = useState<Any>({});
  const [rt, setRt] = useState<Any>({});
  const [logs, setLogs] = useState<Any>({});
  const [appVersion, setAppVersion] = useState("");
  const [busy, setBusy] = useState(false);
  const [splitRows, setSplitRows] = useState<Any[]>([]);
  const [dpiLists, setDpiLists] = useState<Any[]>([]);
  const [dpiTune, setDpiTune] = useState<Any | null>(null);
  const [selfCheck, setSelfCheck] = useState("");
  const [updateInfo, setUpdateInfo] = useState<Any | null>(null);
  const [updateProgress, setUpdateProgress] = useState<{ phase: string; pct: number } | null>(null);
  const [guide, setGuide] = useState("");
  const [listsPath, setListsPath] = useState("");
  const [toast, setToast] = useState<ToastState | null>(null);
  const [proxyPortDraft, setProxyPortDraft] = useState(8080);

  const dismissToast = useCallback(() => setToast(null), []);

  const notify = useCallback((text: string, variant: "info" | "error" = "info") => {
    setToast({ text, variant });
    window.setTimeout(() => setToast((t) => (t && t.text === text ? null : t)), 4200);
  }, []);

  const setActive = useCallback((k: NavKey) => {
    setActiveRaw(k);
  }, []);

  const openDpiTab = useCallback((tab: DpiSubTab) => {
    setDpiSubTab(tab);
    setActiveRaw("dpi");
  }, []);

  const setTheme = useCallback((t: ThemeKey) => {
    setThemeRaw(t);
    writeLS(THEME_STORAGE_KEY, t);
  }, []);

  const setAccent = useCallback((a: AccentKey) => {
    setAccentRaw(a);
    writeLS(ACCENT_STORAGE_KEY, a);
  }, []);

  useEffect(() => {
    const el = document.documentElement;
    el.setAttribute("data-theme", theme);
    el.setAttribute("data-accent", accent);
  }, [theme, accent]);

  const loadDpiLists = useCallback(async () => {
    try {
      setDpiLists(((await GetAvailableDpiLists()) || []) as unknown as Any[]);
    } catch {
      setDpiLists([]);
    }
  }, []);

  /** Перечитать состояние из бэкенда. */
  const refresh = useCallback(async () => {
    try {
      const [ui, h, r, l, rows] = await Promise.all([
        GetUIState(),
        GetServiceHealth(),
        GetRealtimeStats(),
        GetLogsState(),
        GetRoutingSplitView(),
      ]);
      setUiState((ui || {}) as unknown as Any);
      setHealth((h || {}) as unknown as Any);
      setRt((r || {}) as unknown as Any);
      setLogs((l || {}) as unknown as Any);
      setSplitRows((rows || []) as unknown as Any[]);
      const port = Number((ui as unknown as Any)?.proxyPort);
      if (port > 0) setProxyPortDraft(port);
    } catch {
      /* бэкенд ещё не готов — повторим на следующем тике */
    }
  }, []);

  useEffect(() => {
    let alive = true;
    void (async () => {
      await Promise.all([refresh(), loadDpiLists()]);
      if (!alive) return;
      try {
        setAppVersion(await GetAppVersion());
        setListsPath(await GetZapretListsPathHint());
        setGuide(await ProgramGuide());
      } catch {
        /* необязательные данные */
      }
    })();
    return () => {
      alive = false;
    };
  }, [refresh, loadDpiLists]);

  // Живые метрики каждые 2 секунды.
  useEffect(() => {
    const id = window.setInterval(() => {
      void (async () => {
        try {
          const [r, h] = await Promise.all([GetRealtimeStats(), GetServiceHealth()]);
          setRt((r || {}) as unknown as Any);
          setHealth((h || {}) as unknown as Any);
        } catch {
          /* пропускаем тик */
        }
      })();
    }, 2000);
    return () => window.clearInterval(id);
  }, []);

  // Подписка на события бэкенда.
  useEffect(() => {
    const offs: Array<() => void> = [];
    const reg = (event: string, cb: (...data: unknown[]) => void) => {
      try {
        const off = EventsOn(event, cb);
        if (typeof off === "function") offs.push(off);
      } catch {
        /* игнорируем при отсутствии runtime */
      }
    };

    reg("state-changed", () => {
      void refresh();
    });
    reg("update-available", (raw: unknown) => {
      try {
        const info = typeof raw === "string" ? (JSON.parse(raw) as Any) : ((raw || {}) as Any);
        setUpdateInfo(info);
        const verLabel = String(info?.latestTag || info?.latestVersion || "");
        if (verLabel) {
          notify(`Найдено обновление ${verLabel} — скачивание…`);
        }
      } catch {
        /* некорректный payload игнорируем */
      }
    });
    reg("update-progress", (raw: unknown) => {
      try {
        const p = typeof raw === "string" ? (JSON.parse(raw) as Any) : ((raw || {}) as Any);
        setUpdateProgress({
          phase: String(p?.phase || "download"),
          pct: Number(p?.pct ?? -1),
        });
      } catch {
        /* некорректный payload игнорируем */
      }
    });
    reg("update-error", (raw: unknown) => {
      setUpdateProgress(null);
      notify(`Ошибка обновления: ${String(raw || "неизвестная ошибка")}`, "error");
    });

    return () => {
      for (const off of offs) off();
    };
  }, [notify, refresh]);

  const proxyEnabled = !!uiState?.proxyEnabled;
  const dpiEnabled = !!uiState?.dpiEnabled;
  const proxyRunning = proxyEnabled && !!health?.proxyLocalOk;

  const proxyPhase: "idle" | "connecting" | "connected" = proxyRunning
    ? "connected"
    : proxyEnabled
      ? "connecting"
      : "idle";

  const withBusy = useCallback(async <T,>(fn: () => Promise<T>): Promise<T | null> => {
    setBusy(true);
    try {
      const res = await fn();
      await refresh();
      return res;
    } catch (e: unknown) {
      notify(String(e instanceof Error ? e.message : e), "error");
      return null;
    } finally {
      setBusy(false);
    }
  }, [notify, refresh]);

  const setProxy = useCallback(
    async (on: boolean) => {
      await withBusy(async () => {
        await SetProxyEnabled(on);
        await ApplyRuntime();
      });
    },
    [withBusy]
  );

  const setDpi = useCallback(
    async (on: boolean) => {
      await withBusy(async () => {
        await SetDPIEnabled(on);
        await ApplyRuntime();
      });
    },
    [withBusy]
  );

  const stopAll = useCallback(async () => {
    await withBusy(async () => {
      await SetProxyEnabled(false);
      await SetDPIEnabled(false);
      await ApplyRuntime();
    });
  }, [withBusy]);

  const setRouting = useCallback(
    async (name: string) => {
      await withBusy(async () => {
        await SetRoutingProfile(name);
        await ApplyRuntime();
      });
    },
    [withBusy]
  );

  const setProxyMode = useCallback(
    async (mode: string) => {
      await withBusy(async () => {
        await SetProxyMode(mode);
        await ApplyRuntime();
      });
    },
    [withBusy]
  );

  const addRule = useCallback(
    async (pattern: string, target: string) => {
      await withBusy(async () => {
        await AddUserRoutingRule(pattern, target);
        await ApplyRuntime();
        await loadDpiLists();
      });
    },
    [loadDpiLists, withBusy]
  );

  const removeRule = useCallback(
    async (index: number) => {
      await withBusy(async () => {
        await RemoveUserRoutingRuleAt(index);
        await ApplyRuntime();
        await loadDpiLists();
      });
    },
    [loadDpiLists, withBusy]
  );

  const verifyTarget = useCallback(async (q: string) => {
    try {
      return (await VerifyRouteTarget(q)) as unknown as Any;
    } catch (e: unknown) {
      notify(String(e instanceof Error ? e.message : e), "error");
      return null;
    }
  }, [notify]);

  const selectServer = useCallback(
    async (id: string) => {
      await withBusy(async () => {
        await SetActiveServer(id);
        await ApplyRuntime();
      });
    },
    [withBusy]
  );

  const removeServerFn = useCallback(
    async (id: string) => {
      await withBusy(() => RemoveServer(id));
    },
    [withBusy]
  );

  const retestServerFn = useCallback(
    async (id: string) => {
      await withBusy(() => RetestServer(id));
    },
    [withBusy]
  );

  const retestAll = useCallback(async () => {
    await withBusy(async () => {
      const ids = ((uiState?.servers as Any[]) || []).map((s) => String(s.id || ""));
      if (ids.length) await RetestServersBatch(ids);
    });
  }, [withBusy, uiState]);

  const previewSubscription = useCallback(async (url: string) => {
    try {
      return (await PreviewSubscription(url)) as unknown as Any;
    } catch (e: unknown) {
      notify(String(e instanceof Error ? e.message : e), "error");
      return null;
    }
  }, [notify]);

  const importText = useCallback(
    async (text: string) => {
      const n = await withBusy(() => ImportServersFromText(text));
      return typeof n === "number" ? n : 0;
    },
    [withBusy]
  );

  const importFromClipboard = useCallback(async () => {
    await withBusy(() => ImportServerFromClipboard());
  }, [withBusy]);

  const applySubscription = useCallback(
    async (url: string) => {
      await withBusy(async () => {
        await SetSubscriptionURL(url);
        await RefreshSubscription();
        await ApplyRuntime();
      });
    },
    [withBusy]
  );

  const refreshSubscriptionFn = useCallback(async () => {
    await withBusy(() => RefreshSubscription());
  }, [withBusy]);

  const clearSubscriptionFn = useCallback(async () => {
    await withBusy(() => ClearSubscription());
  }, [withBusy]);

  const toggleDpiList = useCallback(
    async (name: string, on: boolean) => {
      await withBusy(async () => {
        await ToggleDpiList(name, on);
        await loadDpiLists();
      });
    },
    [loadDpiLists, withBusy]
  );

  const readDpiList = useCallback(async (name: string) => {
    try {
      return (await GetDpiListContent(name)) as unknown as Any;
    } catch (e: unknown) {
      notify(String(e instanceof Error ? e.message : e), "error");
      return null;
    }
  }, [notify]);

  const writeDpiList = useCallback(
    async (name: string, entries: string[]) => {
      await withBusy(() => SetDpiListContent(name, entries));
    },
    [withBusy]
  );

  const runDpiTuning = useCallback(async () => {
    const res = await withBusy(async () => {
      notify("Подбираю конфигурацию обхода — это может занять несколько минут…");
      return await RunDpiAdaptiveTuning(false);
    });
    if (res) {
      setDpiTune(res as unknown as Any);
      notify("Подбор завершён");
    }
    return (res as unknown as Any) || null;
  }, [notify, withBusy]);

  const setDpiProfile = useCallback(
    async (name: string) => {
      await withBusy(async () => {
        await SetActiveDPIProfile(name);
        await ApplyRuntime();
      });
    },
    [withBusy]
  );

  const runSelfCheckFn = useCallback(async () => {
    setBusy(true);
    try {
      setSelfCheck(await RunSelfCheckText());
    } catch (e: unknown) {
      notify(String(e instanceof Error ? e.message : e), "error");
    } finally {
      setBusy(false);
    }
  }, [notify]);

  const exportBundle = useCallback(async () => {
    try {
      const p = await CreateDiagnosticsBundle();
      notify(`Архив сохранён: ${p}`);
    } catch (e: unknown) {
      notify(String(e instanceof Error ? e.message : e), "error");
    }
  }, [notify]);

  const applyProxyPort = useCallback(
    async (p: number) => {
      await withBusy(async () => {
        await SetProxyPort(p);
        await ApplyRuntime();
      });
    },
    [withBusy]
  );

  const applyDns = useCallback(
    async (preset: string, primary: string, secondary: string) => {
      await withBusy(async () => {
        await SetDnsSettings(preset, primary, secondary);
        await ApplyDnsNow();
        notify("DNS применён");
      });
    },
    [notify, withBusy]
  );

  const activateLicense = useCallback(
    async (code: string) => {
      await withBusy(() => ActivateLicense(code));
      notify("Ключ применён");
    },
    [notify, withBusy]
  );

  const autoUpdateEnabled = uiState?.autoUpdateEnabled !== false;
  const autostartEnabled = !!uiState?.autostartEnabled;

  const setAutoUpdateEnabled = useCallback(
    async (on: boolean) => {
      await withBusy(async () => {
        await SetAutoUpdateEnabled(on);
      });
      notify(on ? "Автообновление включено" : "Автообновление отключено");
    },
    [notify, withBusy]
  );

  const setAutostartEnabled = useCallback(
    async (on: boolean) => {
      const ok = await withBusy(async () => {
        await SetAutostartEnabled(on);
        return true;
      });
      if (ok) {
        notify(on ? "Автозапуск с Windows включён" : "Автозапуск с Windows отключён");
      }
    },
    [notify, withBusy]
  );

  const checkUpdates = useCallback(async () => {
    setBusy(true);
    try {
      await CheckForUpdates();
      setUpdateInfo((await GetUpdateInfo()) as unknown as Any);
    } catch (e: unknown) {
      notify(String(e instanceof Error ? e.message : e), "error");
    } finally {
      setBusy(false);
    }
  }, [notify]);

  const installUpdate = useCallback(async () => {
    setBusy(true);
    try {
      await ApplyLatestUpdate();
    } catch (e: unknown) {
      setUpdateProgress(null);
      notify(String(e instanceof Error ? e.message : e), "error");
    } finally {
      setBusy(false);
    }
  }, [notify]);

  const clearLogs = useCallback(async () => {
    await withBusy(() => ClearLogFiles());
  }, [withBusy]);

  const openLogsFolderFn = useCallback(async () => {
    try {
      await OpenLogsFolder();
    } catch (e: unknown) {
      notify(String(e instanceof Error ? e.message : e), "error");
    }
  }, [notify]);

  const quit = useCallback(() => {
    void RequestQuit().catch(() => {
      /* при ошибке всё равно закрываем процесс */
    });
  }, []);

  const closeToTray = useCallback(() => {
    void CloseToTray().catch(() => {
      void import("../../wailsjs/runtime/runtime").then((m) => m.WindowHide());
    });
  }, []);

  const servers = useMemo(() => ((uiState?.servers as Any[]) || []), [uiState]);
  const activeServer = useMemo(() => {
    const id = String(uiState?.activeServerId || "");
    return servers.find((s) => String(s.id) === id) || null;
  }, [servers, uiState]);

  const value: Ctx = {
    active,
    setActive,
    dpiSubTab,
    setDpiSubTab,
    openDpiTab,
    theme,
    accent,
    setTheme,
    setAccent,
    uiState,
    health,
    rt,
    logs,
    appVersion,
    busy,
    proxyPhase,
    proxyEnabled,
    dpiEnabled,
    proxyMode: String(uiState?.proxyMode || "system_proxy"),
    setProxy,
    setProxyMode,
    setDpi,
    stopAll,
    servers,
    activeServer,
    splitRows,
    routingProfile: String(uiState?.routingProfile || "manual"),
    setRouting,
    addRule,
    removeRule,
    verifyTarget,
    selectServer,
    removeServer: removeServerFn,
    retestServer: retestServerFn,
    retestAll,
    previewSubscription,
    importText,
    importFromClipboard,
    applySubscription,
    refreshSubscription: refreshSubscriptionFn,
    clearSubscription: clearSubscriptionFn,
    dpiLists,
    refreshDpiLists: loadDpiLists,
    toggleDpiList,
    readDpiList,
    writeDpiList,
    runDpiTuning,
    dpiTune,
    setDpiProfile,
    listsPath,
    selfCheck,
    runSelfCheck: runSelfCheckFn,
    exportBundle,
    proxyPortDraft,
    applyProxyPort,
    applyDns,
    activateLicense,
    autoUpdateEnabled,
    setAutoUpdateEnabled,
    autostartEnabled,
    setAutostartEnabled,
    checkUpdates,
    updateInfo,
    updateProgress,
    installUpdate,
    guide,
    clearLogs,
    openLogsFolder: openLogsFolderFn,
    quit,
    closeToTray,
    toast,
    notify,
    dismissToast,
    refresh,
  };

  return <AppCtx.Provider value={value}>{children}</AppCtx.Provider>;
}

export function useApp(): Ctx {
  const ctx = useContext(AppCtx);
  if (!ctx) throw new Error("useApp должен вызываться внутри AppProvider");
  return ctx;
}
