// Navigation
export type NavKey = "dashboard" | "servers" | "network" | "dpi" | "logs" | "guide" | "diagnostics" | "support" | "settings";
export type ThemeKey = "dark-neon" | "graphite" | "light" | "emerald" | "midnight" | "sunset" | "ocean" | "rose" | "black" | "kitty" | "anime" | "rkn" | "matrix";
export type ServerPoolTab = "manual" | "subscription";
export type LogSource = "core" | "dpi";
export type LogViewMode = "errors" | "filter" | "full";
export type ToastState = { text: string; variant?: "info" | "error"; openLogs?: boolean; dismissible?: boolean };
export type ChartPt = { x: number; y: number };
