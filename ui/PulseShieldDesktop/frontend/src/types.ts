/* Навигация и ключи шторок интерфейса. */
export type NavKey = "shield" | "servers" | "routes" | "tuning" | "more";

export type SheetKey =
  | "server"
  | "rule"
  | "verify"
  | "tune-dpi"
  | "stats"
  | "logs"
  | "license"
  | "about"
  | "updates"
  | "guide"
  | "speed"
  | "dns"
  | "port"
  | "subscription"
  | "theme";

export type ThemeKey =
  | "graphite"
  | "daylight"
  | "ember"
  | "midnight"
  | "forest"
  | "plum"
  | "abyss"
  | "carbon";

export type AccentKey = "amber" | "mint" | "violet" | "rose" | "ice" | "lime";

export type LogSource = "core" | "dpi";
export type LogViewMode = "errors" | "filter" | "full";
export type ToastState = { text: string; variant?: "info" | "error"; openLogs?: boolean; dismissible?: boolean };
export type ChartPt = { x: number; y: number };

export const NAV_ITEMS: { key: NavKey; label: string; icon: string }[] = [
  { key: "shield", label: "Защита", icon: "shield" },
  { key: "servers", label: "Серверы", icon: "server" },
  { key: "routes", label: "Маршруты", icon: "routes" },
  { key: "tuning", label: "Настройка", icon: "sliders" },
  { key: "more", label: "Ещё", icon: "dots" },
];

/** Заголовки шторок. */
export const SHEET_TITLES: Record<SheetKey, string> = {
  server: "Серверы",
  rule: "Новое правило",
  verify: "Проверить сайт",
  "tune-dpi": "Обход DPI",
  stats: "Диагностика",
  logs: "Журнал",
  license: "Лицензия",
  about: "О программе",
  updates: "Обновления",
  guide: "Как это работает",
  speed: "Тест сервера",
  dns: "DNS",
  port: "Порт прокси",
  subscription: "Подписка и ключи",
  theme: "Оформление",
};

/** Куда уходит трафик по правилу. */
export type RouteTarget = "proxy" | "direct" | "dpi-bypass" | "block";

export const ROUTE_TARGET_LABEL: Record<RouteTarget, string> = {
  proxy: "Через прокси",
  direct: "Напрямую",
  "dpi-bypass": "Обход DPI",
  block: "Блокировать",
};
