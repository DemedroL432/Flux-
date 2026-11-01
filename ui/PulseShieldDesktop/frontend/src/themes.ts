import type { AccentKey, ThemeKey } from "./types";

export const THEMES: { key: ThemeKey; name: string; hint: string; preview: string }[] = [
  { key: "graphite", name: "Уголь", hint: "Тёплый графит, золото и мята", preview: "radial-gradient(120% 110% at 18% 0%, #3a2a1c, #0a0908)" },
  { key: "ember", name: "Тлеющий", hint: "Тёмный янтарь и медь", preview: "radial-gradient(120% 110% at 20% 0%, #4a2317, #0d0704)" },
  { key: "forest", name: "Хвоя", hint: "Глубокий зелёный с мятой", preview: "radial-gradient(120% 110% at 20% 0%, #17342a, #050b08)" },
  { key: "plum", name: "Слива", hint: "Бархатный фиолетовый", preview: "radial-gradient(120% 110% at 20% 0%, #33203a, #0b0710)" },
  { key: "midnight", name: "Полночь", hint: "Холодный синий, ясная ночь", preview: "radial-gradient(120% 110% at 20% 0%, #172a4a, #060911)" },
  { key: "abyss", name: "Бездна", hint: "Почти чёрный с бирюзой", preview: "radial-gradient(120% 110% at 20% 0%, #0f2b2e, #030607)" },
  { key: "carbon", name: "Карбон", hint: "Нейтральный графит", preview: "radial-gradient(120% 110% at 20% 0%, #26262a, #08080a)" },
  { key: "daylight", name: "День", hint: "Светлая тема для работы днём", preview: "radial-gradient(120% 110% at 20% 0%, #fff6e8, #efe7db)" },
];

export const ACCENTS: { key: AccentKey; name: string; color: string }[] = [
  { key: "amber", name: "Золото", color: "#ffca6b" },
  { key: "mint", name: "Мята", color: "#7beac1" },
  { key: "ice", name: "Лёд", color: "#9ac8ff" },
  { key: "violet", name: "Сирень", color: "#c5a1ed" },
  { key: "rose", name: "Роза", color: "#ff9fb1" },
  { key: "lime", name: "Лайм", color: "#c6ef7a" },
];

export const DEFAULT_THEME: ThemeKey = "graphite";
export const DEFAULT_ACCENT: AccentKey = "amber";

const LEGACY_THEME: Record<string, ThemeKey> = {
  "dark-neon": "graphite",
  dark: "graphite",
  light: "daylight",
  emerald: "forest",
  midnight: "midnight",
  sunset: "ember",
  ocean: "abyss",
  rose: "plum",
  black: "carbon",
  kitty: "plum",
  anime: "plum",
  rkn: "ember",
  matrix: "forest",
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

/** Светлая тема требует инвертированных теней и другого знака иконок. */
export function isLightTheme(t: ThemeKey): boolean {
  return t === "daylight";
}

export const THEME_STORAGE_KEY = "pulseshield.theme";
export const ACCENT_STORAGE_KEY = "pulseshield.accent";
