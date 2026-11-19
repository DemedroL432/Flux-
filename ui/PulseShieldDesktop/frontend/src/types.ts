/* Навигация и основные типы интерфейса. */
export type NavKey = "shield" | "servers" | "routes" | "dpi" | "net";

export type DpiSubTab = "profiles" | "lists";

export type ThemeKey = "monolith" | "midnight" | "daylight";

export type AccentKey = "amber" | "mint" | "violet" | "rose" | "ice" | "lime";

export type LogSource = "core" | "dpi";
export type LogViewMode = "errors" | "filter" | "full";
export type ToastState = { text: string; variant?: "info" | "error"; openLogs?: boolean; dismissible?: boolean };
export type ChartPt = { x: number; y: number };

export const NAV_ITEMS: { key: NavKey; label: string; icon: string }[] = [
  { key: "shield", label: "Обзор", icon: "overview" },
  { key: "servers", label: "Серверы", icon: "server" },
  { key: "routes", label: "Маршруты", icon: "routes" },
  { key: "dpi", label: "Обход DPI", icon: "dpi" },
  { key: "net", label: "Сеть и логи", icon: "sliders" },
];

/** Куда уходит трафик по правилу. */
export type RouteTarget = "proxy" | "direct" | "dpi-bypass" | "block";

export const ROUTE_TARGET_LABEL: Record<RouteTarget, string> = {
  proxy: "Через прокси",
  direct: "Напрямую",
  "dpi-bypass": "Обход DPI",
  block: "Блокировать",
};
