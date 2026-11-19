import type { AccentKey, ThemeKey } from "./types";

export const THEMES: { key: ThemeKey; name: string; hint: string; preview: string }[] = [
  { key: "monolith", name: "Монолит", hint: "Чистый нейтральный графит без лишних рамок", preview: "#0e1013" },
  { key: "midnight", name: "Полночь", hint: "Глубокий холодный синий", preview: "#0a101b" },
  { key: "daylight", name: "Светлая", hint: "Спокойная светлая тема для работы днём", preview: "#f4f5f7" },
];

export const ACCENTS: { key: AccentKey; name: string; color: string }[] = [
  { key: "ice", name: "Нейтральный", color: "#f2f4f7" },
  { key: "mint", name: "Мята", color: "#34d399" },
  { key: "amber", name: "Янтарь", color: "#fbbf24" },
  { key: "violet", name: "Сирень", color: "#a78bfa" },
  { key: "rose", name: "Коралл", color: "#f87171" },
  { key: "lime", name: "Лайм", color: "#a3e635" },
];

export const DEFAULT_THEME: ThemeKey = "monolith";
export const DEFAULT_ACCENT: AccentKey = "ice";

const LEGACY_THEME: Record<string, ThemeKey> = {
  graphite: "monolith",
  "dark-neon": "monolith",
  dark: "monolith",
  carbon: "monolith",
  ember: "monolith",
  forest: "monolith",
  plum: "monolith",
  abyss: "monolith",
  black: "monolith",
  sunset: "monolith",
  ocean: "monolith",
  rose: "monolith",
  emerald: "monolith",
  kitty: "monolith",
  anime: "monolith",
  rkn: "monolith",
  matrix: "monolith",
  midnight: "midnight",
  light: "daylight",
  daylight: "daylight",
};

export function normalizeTheme(raw: string | null | undefined): ThemeKey {
  if (!raw) return DEFAULT_THEME;
  const known = THEMES.find((t) => t.key === raw);
  if (known) return known.key;
  return LEGACY_THEME[raw] ?? DEFAULT_THEME;
}

export function normalizeAccent(raw: string | null | undefined): AccentKey {
  if (!raw) return DEFAULT_ACCENT;
  const known = ACCENTS.find((a) => a.key === raw);
  return known ? known.key : DEFAULT_ACCENT;
}

export function isLightTheme(t: ThemeKey): boolean {
  return t === "daylight";
}

export const THEME_STORAGE_KEY = "pulseshield.theme";
export const ACCENT_STORAGE_KEY = "pulseshield.accent";
