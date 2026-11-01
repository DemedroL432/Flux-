/** Форматирование чисел, времени и адресов для интерфейса. */

/** Скорость: биты/с → читаемая строка в Мбит/с. */
export function formatMbps(bps: number): string {
  const n = Number(bps) || 0;
  if (n <= 0) return "0.0";
  const mbit = (n * 8) / 1_000_000;
  if (mbit < 0.1) return mbit.toFixed(2);
  return mbit.toFixed(1);
}

/** Байты → человекочитаемый размер. */
export function formatBytes(n: number): string {
  const v = Number(n) || 0;
  if (v <= 0) return "0 Б";
  const units = ["Б", "КБ", "МБ", "ГБ", "ТБ"];
  let x = v;
  let i = 0;
  while (x >= 1024 && i < units.length - 1) {
    x /= 1024;
    i++;
  }
  return `${x < 10 && i > 0 ? x.toFixed(1) : Math.round(x)} ${units[i]}`;
}

/** Секунды → ЧЧ:ММ:СС. */
export function formatClock(totalSec: number): string {
  const s = Math.max(0, Math.floor(Number(totalSec) || 0));
  const h = Math.floor(s / 3600);
  const m = Math.floor((s % 3600) / 60);
  const sec = s % 60;
  const p = (v: number) => String(v).padStart(2, "0");
  return `${p(h)}:${p(m)}:${p(sec)}`;
}

/** Задержка в мс → подпись. */
export function formatLatency(ms: number | undefined | null): string {
  const v = Number(ms) || 0;
  if (v <= 0) return "не измерен";
  return `${Math.round(v)} мс`;
}

/** Класс тона по задержке. */
export function latencyTone(ms: number | undefined | null): "good" | "mid" | "bad" | "" {
  const v = Number(ms) || 0;
  if (v <= 0) return "";
  if (v < 80) return "good";
  if (v < 200) return "mid";
  return "bad";
}

/** Unix-секунды → «дд.мм.гггг». */
export function formatDate(unix: number | undefined | null): string {
  const v = Number(unix) || 0;
  if (v <= 0) return "—";
  return new Date(v * 1000).toLocaleDateString("ru-RU");
}

/** Unix-секунды → «осталось N дн.». */
export function formatRemaining(unix: number | undefined | null): string {
  const v = Number(unix) || 0;
  if (v <= 0) return "без срока";
  const days = Math.round((v * 1000 - Date.now()) / 86_400_000);
  if (days < 0) return "истёк";
  if (days === 0) return "истекает сегодня";
  if (days < 31) return `${days} дн.`;
  const months = Math.round(days / 30);
  return `${months} мес.`;
}

/** Убирает схему и путь, оставляя хост. */
export function hostOf(raw: string): string {
  let s = (raw || "").trim().toLowerCase();
  s = s.replace(/^[a-z]+:\/\//, "");
  s = s.replace(/^\/\//, "");
  const cut = s.search(/[/?#]/);
  if (cut >= 0) s = s.slice(0, cut);
  return s.replace(/:\d+$/, "");
}

/** Проверяет, выглядит ли строка как домен или IP. */
export function looksLikeHost(s: string): boolean {
  const v = (s || "").trim();
  if (!v) return false;
  if (/^\d{1,3}(\.\d{1,3}){3}(\/\d{1,2})?$/.test(v)) return true;
  return /^[a-z0-9]([a-z0-9-]*[a-z0-9])?(\.[a-z0-9]([a-z0-9-]*[a-z0-9])?)+$/i.test(v);
}
