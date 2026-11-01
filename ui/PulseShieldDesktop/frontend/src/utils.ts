import { ChartPt, LogSource, LogViewMode } from "./types";

/** Фирменный акцент Deep Flow + натяжение линии ~ Chart.js tension */
export const DEEP_FLOW_VIOLET = "#8b5cf6";
export const DEEP_FLOW_TENSION = 0.4;

export const ACCENT_PRESETS = ["#78f65c", "#58d8ff", "#a78bfa", "#fb7185", "#fbbf24", "#38bdf8", "#34d399", "#f472b6"];

export const ONBOARDING_KEY = "pulseshield.onboardingDone";

/** Подписи двух режимов маршрутизации (имя в конфиге без изменений). */
export const ROUTING_PROFILE_LABELS: Record<string, string> = {
  "full-proxy": "Полный прокси — весь трафик через туннель, кроме локального и РФ",
  manual: "Прокси с ручными правилами — в туннель только то, что вы добавили",
};

export const BUILTIN_ROUTING_KEYS = ["full-proxy", "manual"] as const;

export const ROUTING_MODE_CARDS: { key: (typeof BUILTIN_ROUTING_KEYS)[number]; icon: string; blurb: string }[] = [
  { key: "full-proxy", icon: "◎", blurb: "Весь трафик через туннель." },
  { key: "manual", icon: "⊕", blurb: "В туннель — только ваши правила." },
];

/** Примеры «что через VPN / что напрямую» для подсказок (совпадает с логикой пресетов в ядре). */
export const ROUTING_TRAFFIC_CHIPS: Record<string, string[]> = {
  "full-proxy": [
    "Через туннель: весь интернет-трафик",
    "Напрямую: локальная сеть и сайты РФ",
  ],
  manual: [
    "Через туннель: только добавленные вами сайты и процессы",
    "Напрямую: всё остальное, включая сайты РФ",
  ],
};

export const ROUTING_MODE_SHORT: Record<string, string> = {
  "full-proxy": "Полный прокси",
  manual: "Ручные правила",
};

export function formatBps(bps: number): string {
  if (!bps || bps <= 0) return "0 KB/s";
  const kb = bps / 1024;
  if (kb < 1024) return `${kb.toFixed(1)} KB/s`;
  return `${(kb / 1024).toFixed(2)} MB/s`;
}

export function formatBytes(n: number): string {
  if (n == null || !Number.isFinite(n) || n <= 0) return "0 B";
  const u = ["B", "KB", "MB", "GB"];
  let v = n;
  let i = 0;
  while (v >= 1024 && i < u.length - 1) {
    v /= 1024;
    i++;
  }
  return `${v < 10 && i > 0 ? v.toFixed(1) : Math.round(v)} ${u[i]}`;
}

export function formatSessionClock(totalSec: number): string {
  const s = Math.max(0, Math.floor(Number(totalSec) || 0));
  const h = Math.floor(s / 3600);
  const m = Math.floor((s % 3600) / 60);
  const sec = s % 60;
  const pad = (x: number) => x.toString().padStart(2, "0");
  return `${pad(h)}:${pad(m)}:${pad(sec)}`;
}

export function serverPingLabel(s: { lastCheckOk?: boolean; lastCheckLatencyMs?: number; lastCheckDownloadMbps?: number }): string {
  if (s.lastCheckOk && s.lastCheckLatencyMs != null && Number.isFinite(Number(s.lastCheckLatencyMs))) {
    const mbps = Number(s.lastCheckDownloadMbps);
    if (mbps > 0 && Number.isFinite(mbps)) {
      return `${s.lastCheckLatencyMs} ms · ${mbps.toFixed(0)} Мбит/с`;
    }
    return `${s.lastCheckLatencyMs} ms`;
  }
  return "—";
}

export function formatSubscriptionExpiry(unixSec: number | undefined): string {
  if (unixSec == null || unixSec <= 0) return "";
  const exp = new Date(unixSec * 1000);
  const days = Math.ceil((exp.getTime() - Date.now()) / 86400000);
  if (days < 0) return `истекла ${exp.toLocaleDateString()}`;
  return `~${days} дн. · до ${exp.toLocaleDateString()}`;
}

export function humanizeRuntimeError(raw: string): string {
  const s = raw.toLowerCase();
  if (raw.includes("requires elevation") || s.includes("от администратора") || s.includes("runas")) {
    return "Нужны права администратора — запустите Flux от имени администратора.";
  }
  if (
    (s.includes("not found") || s.includes("не найден") || s.includes("missing")) &&
    (s.includes("binary") || s.includes(".exe") || s.includes("zapret") || s.includes("goodbye") || s.includes("sing-box") || s.includes("winws"))
  ) {
    return "Не найден исполняемый файл модуля — проверьте папки core, dpi и zapret рядом с программой.";
  }
  return raw;
}

export function formatRuntimeFailure(context: string, e: unknown): string {
  const raw = e instanceof Error ? e.message : String(e);
  return `${context}: ${humanizeRuntimeError(raw)}`;
}

/** Ссылка от Telegram-бота (?start=…), https://…?code=…, pulseshield://… или сырой код. */
export function parseActivationPaste(raw: string): { code: string; apiBase?: string } {
  const t = raw.trim();
  if (!t) {
    throw new Error("Вставьте ссылку от бота или код активации.");
  }
  if (/^pulseshield:\/\//i.test(t)) {
    try {
      const u = new URL(t);
      const code = u.searchParams.get("code") || u.searchParams.get("c") || u.searchParams.get("token");
      const api = u.searchParams.get("api") || u.searchParams.get("base") || u.searchParams.get("server");
      if (code?.trim()) return { code: code.trim(), ...(api?.trim() ? { apiBase: api.trim() } : {}) };
    } catch {
      /* fall through */
    }
  }

  let urlStr = t;
  if (t.includes("t.me") && !/^https?:\/\//i.test(t)) {
    urlStr = "https://" + t.replace(/^\/+/g, "");
  }

  if (/^https?:\/\//i.test(urlStr)) {
    try {
      const u = new URL(urlStr);
      const host = u.hostname.toLowerCase();
      if (host.includes("t.me") || host.includes("telegram.")) {
        const start = u.searchParams.get("start");
        if (start?.trim()) return { code: start.trim() };
      }

      const api = u.searchParams.get("api") || u.searchParams.get("base") || u.searchParams.get("server");
      const apiOpt = api?.trim() ? { apiBase: api.trim() } : {};

      const codeKeys = ["code", "c", "token", "key", "license", "activation", "activate"];
      for (const k of codeKeys) {
        const v = u.searchParams.get(k);
        if (v?.trim()) return { code: v.trim(), ...apiOpt };
      }

      if (u.hash && u.hash.length > 2) {
        const hp = new URLSearchParams(u.hash.replace(/^#\??/, ""));
        for (const k of codeKeys) {
          const v = hp.get(k);
          if (v?.trim()) {
            const hba = hp.get("api") || hp.get("base") || hp.get("server");
            return { code: v.trim(), ...(hba?.trim() ? { apiBase: hba.trim() } : apiOpt) };
          }
        }
      }
    } catch {
      /* fall through */
    }
  }

  return { code: t };
}

let lastSoundAt = 0;
const actionSounds = {
  nav: new URL('./assets/audio/nav.wav', import.meta.url).href,
  toggle: new URL('./assets/audio/toggle.wav', import.meta.url).href,
  success: new URL('./assets/audio/success.wav', import.meta.url).href,
};
let currentSound: HTMLAudioElement | null = null;
export function playSoftChime(kind: "toggle" | "nav" | "success"): void {
  if (localStorage.getItem('flux.sounds') === 'off' || Date.now() - lastSoundAt < 120) return;
  lastSoundAt = Date.now();
  currentSound?.pause();
  currentSound = new Audio(actionSounds[kind]);
  void currentSound.play().catch(() => {});
}

export function parseHex(hex: string): { r: number; g: number; b: number } | null {
  const m = /^#?([a-f\d]{2})([a-f\d]{2})([a-f\d]{2})$/i.exec(hex.trim());
  if (!m) return null;
  return { r: parseInt(m[1], 16), g: parseInt(m[2], 16), b: parseInt(m[3], 16) };
}

export function rgbToHex(r: number, g: number, b: number): string {
  const c = (n: number) => Math.max(0, Math.min(255, n)).toString(16).padStart(2, "0");
  return `#${c(r)}${c(g)}${c(b)}`;
}

export function mixTowardWhite(rgb: { r: number; g: number; b: number }, t: number): { r: number; g: number; b: number } {
  return {
    r: Math.round(rgb.r + (255 - rgb.r) * t),
    g: Math.round(rgb.g + (255 - rgb.g) * t),
    b: Math.round(rgb.b + (255 - rgb.b) * t),
  };
}

export function applyAccentHex(hex: string): void {
  const rgb = parseHex(hex);
  if (!rgb) return;
  const root = document.documentElement;
  // v3 CSS variable names
  root.style.setProperty("--accent", hex);
  const b = mixTowardWhite(rgb, 0.25);
  root.style.setProperty("--accent2", rgbToHex(b.r, b.g, b.b));
  // derived
  const dimA = rgb.r.toString(16).padStart(2,"0") + rgb.g.toString(16).padStart(2,"0") + rgb.b.toString(16).padStart(2,"0");
  root.style.setProperty("--accent-dim", `rgba(${rgb.r},${rgb.g},${rgb.b},0.11)`);
  root.style.setProperty("--accent-glow", `rgba(${rgb.r},${rgb.g},${rgb.b},0.25)`);
  // legacy compat
  root.style.setProperty("--accent-a", hex);
  root.style.setProperty("--accent-b", rgbToHex(b.r, b.g, b.b));
  void dimA;
}

export function uniqueRoutingProfiles(profiles: unknown[] | undefined): unknown[] {
  if (!profiles?.length) return [];
  const seen = new Set<string>();
  return profiles.filter((p: unknown) => {
    const n = (p as { name?: string })?.name as string;
    if (!n || seen.has(n)) return false;
    seen.add(n);
    return true;
  });
}

export function cubicBezierSegmentsThroughPoints(pts: ChartPt[], tension: number): string {
  if (pts.length < 2) return "";
  const k = Math.max(0, Math.min(1, tension)) * 2.35;
  let s = "";
  for (let i = 0; i < pts.length - 1; i++) {
    const p0 = pts[i === 0 ? i : i - 1];
    const p1 = pts[i];
    const p2 = pts[i + 1];
    const p3 = pts[i + 2 < pts.length ? i + 2 : i + 1];
    const cp1x = p1.x + ((p2.x - p0.x) / 6) * k;
    const cp1y = p1.y + ((p2.y - p0.y) / 6) * k;
    const cp2x = p2.x - ((p3.x - p1.x) / 6) * k;
    const cp2y = p2.y - ((p3.y - p1.y) / 6) * k;
    s += ` C ${cp1x},${cp1y} ${cp2x},${cp2y} ${p2.x},${p2.y}`;
  }
  return s;
}

export function logPanelContent(logs: unknown, source: LogSource, mode: LogViewMode): string {
  if (!logs) return "";
  const l = logs as Record<string, string>;
  if (source === "core") {
    if (mode === "errors") return l.coreErrors || "";
    if (mode === "filter") return l.coreConnLines || "";
    return l.coreLogTail || "";
  }
  if (mode === "errors") return l.dpiErrors || "";
  if (mode === "filter") return l.dpiConnLines || "";
  return l.dpiLogTail || "";
}

export function logPanelPlaceholder(source: LogSource, mode: LogViewMode): string {
  if (source === "core") {
    if (mode === "errors") return "Нет строк с error/fatal в выборке.";
    if (mode === "filter") return "Нет подходящих строк — включите прокси и сгенерируйте трафик.";
    return "Лог ядра пуст.";
  }
  if (mode === "errors") return "Нет строк с error/fatal в выборке.";
  if (mode === "filter") return "Нет подходящих строк.";
  return "Лог DPI пуст.";
}

export function routingPresetTooltip(profileKey: string): string {
  const chips = ROUTING_TRAFFIC_CHIPS[profileKey];
  const lines = ["Выбранный режим сразу применяется к ядру (sing-box)."];
  if (chips?.length) {
    lines.push("", "Ориентир по трафику:");
    chips.forEach((c) => lines.push(`• ${c}`));
  }
  if (profileKey === "manual") {
    lines.push("", "Ручной режим: через туннель идёт только то, что вы добавили правилами, остальное напрямую.");
  }
  if (profileKey === "full-proxy") {
    lines.push("", "Полный прокси: всё уходит в туннель, кроме локальной сети и сайтов РФ.");
  }
  if (profileKey && !(BUILTIN_ROUTING_KEYS as readonly string[]).includes(profileKey as (typeof BUILTIN_ROUTING_KEYS)[number])) {
    lines.push("", `Профиль «${profileKey}» пользовательский — полное редактирование в разделе «Продвинутые».`);
  }
  return lines.join("\n");
}

export function routingProfileLabel(name: string): string {
  return ROUTING_PROFILE_LABELS[name] ?? name;
}
