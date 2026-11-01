/** Определение страны узла по названию и хосту, с отрисовкой флага из эмодзи-пары. */

type Country = { code: string; name: string; flag: string };

const COUNTRIES: Record<string, Country> = {
  nl: { code: "nl", name: "Нидерланды", flag: "🇳🇱" },
  de: { code: "de", name: "Германия", flag: "🇩🇪" },
  fi: { code: "fi", name: "Финляндия", flag: "🇫🇮" },
  se: { code: "se", name: "Швеция", flag: "🇸🇪" },
  no: { code: "no", name: "Норвегия", flag: "🇳🇴" },
  dk: { code: "dk", name: "Дания", flag: "🇩🇰" },
  gb: { code: "gb", name: "Великобритания", flag: "🇬🇧" },
  uk: { code: "gb", name: "Великобритания", flag: "🇬🇧" },
  fr: { code: "fr", name: "Франция", flag: "🇫🇷" },
  it: { code: "it", name: "Италия", flag: "🇮🇹" },
  es: { code: "es", name: "Испания", flag: "🇪🇸" },
  pt: { code: "pt", name: "Португалия", flag: "🇵🇹" },
  pl: { code: "pl", name: "Польша", flag: "🇵🇱" },
  cz: { code: "cz", name: "Чехия", flag: "🇨🇿" },
  at: { code: "at", name: "Австрия", flag: "🇦🇹" },
  ch: { code: "ch", name: "Швейцария", flag: "🇨🇭" },
  be: { code: "be", name: "Бельгия", flag: "🇧🇪" },
  ro: { code: "ro", name: "Румыния", flag: "🇷🇴" },
  bg: { code: "bg", name: "Болгария", flag: "🇧🇬" },
  hu: { code: "hu", name: "Венгрия", flag: "🇭🇺" },
  lv: { code: "lv", name: "Латвия", flag: "🇱🇻" },
  lt: { code: "lt", name: "Литва", flag: "🇱🇹" },
  ee: { code: "ee", name: "Эстония", flag: "🇪🇪" },
  ua: { code: "ua", name: "Украина", flag: "🇺🇦" },
  tr: { code: "tr", name: "Турция", flag: "🇹🇷" },
  ru: { code: "ru", name: "Россия", flag: "🇷🇺" },
  kz: { code: "kz", name: "Казахстан", flag: "🇰🇿" },
  ge: { code: "ge", name: "Грузия", flag: "🇬🇪" },
  am: { code: "am", name: "Армения", flag: "🇦🇲" },
  ae: { code: "ae", name: "ОАЭ", flag: "🇦🇪" },
  il: { code: "il", name: "Израиль", flag: "🇮🇱" },
  us: { code: "us", name: "США", flag: "🇺🇸" },
  ca: { code: "ca", name: "Канада", flag: "🇨🇦" },
  mx: { code: "mx", name: "Мексика", flag: "🇲🇽" },
  br: { code: "br", name: "Бразилия", flag: "🇧🇷" },
  jp: { code: "jp", name: "Япония", flag: "🇯🇵" },
  kr: { code: "kr", name: "Корея", flag: "🇰🇷" },
  sg: { code: "sg", name: "Сингапур", flag: "🇸🇬" },
  hk: { code: "hk", name: "Гонконг", flag: "🇭🇰" },
  tw: { code: "tw", name: "Тайвань", flag: "🇹🇼" },
  in: { code: "in", name: "Индия", flag: "🇮🇳" },
  id: { code: "id", name: "Индонезия", flag: "🇮🇩" },
  th: { code: "th", name: "Таиланд", flag: "🇹🇭" },
  vn: { code: "vn", name: "Вьетнам", flag: "🇻🇳" },
  au: { code: "au", name: "Австралия", flag: "🇦🇺" },
  za: { code: "za", name: "ЮАР", flag: "🇿🇦" },
  md: { code: "md", name: "Молдова", flag: "🇲🇩" },
  rs: { code: "rs", name: "Сербия", flag: "🇷🇸" },
  gr: { code: "gr", name: "Греция", flag: "🇬🇷" },
  hr: { code: "hr", name: "Хорватия", flag: "🇭🇷" },
  ie: { code: "ie", name: "Ирландия", flag: "🇮🇪" },
  is: { code: "is", name: "Исландия", flag: "🇮🇸" },
  lu: { code: "lu", name: "Люксембург", flag: "🇱🇺" },
  sk: { code: "sk", name: "Словакия", flag: "🇸🇰" },
  si: { code: "si", name: "Словения", flag: "🇸🇮" },
};

/** Слова, которые панели кладут в имя узла вместо кода страны. */
const NAME_HINTS: { words: string[]; code: string }[] = [
  { words: ["амстердам", "amsterdam", "нидерланд", "netherlands", "holland", "голланд"], code: "nl" },
  { words: ["франкфурт", "frankfurt", "герман", "germany", "berlin", "берлин", "nuremberg"], code: "de" },
  { words: ["хельсинки", "helsinki", "финлянд", "finland"], code: "fi" },
  { words: ["стокгольм", "stockholm", "швец", "sweden"], code: "se" },
  { words: ["осло", "oslo", "норвег", "norway"], code: "no" },
  { words: ["копенгаген", "copenhagen", "дани", "denmark"], code: "dk" },
  { words: ["лондон", "london", "англи", "britain", "uk", "england"], code: "gb" },
  { words: ["париж", "paris", "франц", "france"], code: "fr" },
  { words: ["милан", "milan", "итали", "italy", "rome", "рим"], code: "it" },
  { words: ["мадрид", "madrid", "испан", "spain", "barcelona"], code: "es" },
  { words: ["варшав", "warsaw", "польш", "poland"], code: "pl" },
  { words: ["прага", "prague", "чехи", "czech"], code: "cz" },
  { words: ["вена", "vienna", "австри", "austria"], code: "at" },
  { words: ["цюрих", "zurich", "швейцар", "switzerland", "geneva"], code: "ch" },
  { words: ["стамбул", "istanbul", "турци", "turkey", "turkiye"], code: "tr" },
  { words: ["токио", "tokyo", "япони", "japan", "osaka"], code: "jp" },
  { words: ["сингапур", "singapore"], code: "sg" },
  { words: ["гонконг", "hongkong", "hong kong"], code: "hk" },
  { words: ["сеул", "seoul", "коре", "korea"], code: "kr" },
  { words: ["нью-йорк", "new york", "сша", "usa", "united states", "dallas", "seattle", "miami"], code: "us" },
  { words: ["торонто", "toronto", "канад", "canada", "montreal"], code: "ca" },
  { words: ["сидней", "sydney", "австрал", "australia"], code: "au" },
  { words: ["дубай", "dubai", "оаэ", "emirates"], code: "ae" },
  { words: ["тель-авив", "tel aviv", "израил", "israel"], code: "il" },
  { words: ["киев", "kyiv", "kiev", "украин", "ukraine"], code: "ua" },
  { words: ["москва", "moscow", "росси", "russia", "спб", "saint petersburg"], code: "ru" },
  { words: ["алматы", "almaty", "казах", "kazakhstan"], code: "kz" },
  { words: ["тбилиси", "tbilisi", "грузи", "georgia"], code: "ge" },
  { words: ["ереван", "yerevan", "армени", "armenia"], code: "am" },
  { words: ["кишинёв", "chisinau", "молдов", "moldova"], code: "md" },
  { words: ["белград", "belgrade", "серби", "serbia"], code: "rs" },
  { words: ["афины", "athens", "греци", "greece"], code: "gr" },
  { words: ["рига", "riga", "латви", "latvia"], code: "lv" },
  { words: ["вильнюс", "vilnius", "литв", "lithuania"], code: "lt" },
  { words: ["таллин", "tallinn", "эстони", "estonia"], code: "ee" },
  { words: ["будапешт", "budapest", "венгри", "hungary"], code: "hu" },
  { words: ["бухарест", "bucharest", "румын", "romania"], code: "ro" },
  { words: ["софия", "sofia", "болгари", "bulgaria"], code: "bg" },
];

/**
 * Определяет страну узла.
 *
 * Порядок: сначала код в начале хоста (nl1.flux.net → NL), затем домен первого
 * уровня (flux.de → DE), затем слова в имени и хосте. Возвращает null, если
 * страну определить не удалось — тогда показываем нейтральный значок.
 */
export function detectCountry(name?: string, host?: string): Country | null {
  const n = (name || "").toLowerCase();
  const h = (host || "").toLowerCase();

  // 1. Код страны в хосте: "nl1.example.com", "de1.flux.net", "geo-us.example.org".
  const hostParts = h.split(/[.\-_]/).filter(Boolean);
  if (hostParts.length > 0) {
    const first = hostParts[0].replace(/^\d+|\d+$/g, "");
    if (first && COUNTRIES[first]) return COUNTRIES[first];
    // Префиксы вроде "geo", "node", "vpn" — смотрим следующий сегмент.
    if (hostParts.length > 1) {
      const second = hostParts[1].replace(/\d/g, "");
      if (second && COUNTRIES[second]) return COUNTRIES[second];
    }
  }

  // 2. Домен первого уровня: example.de, node.fr.
  const m = /\.([a-z]{2})$/.exec(h);
  if (m && COUNTRIES[m[1]]) return COUNTRIES[m[1]];

  // 3. Слова в имени узла и хосте.
  const hay = `${n} ${h}`;
  for (const hint of NAME_HINTS) {
    for (const w of hint.words) {
      if (hay.includes(w)) return COUNTRIES[hint.code] || null;
    }
  }

  return null;
}

/** Понятное имя страны или пустая строка. */
export function countryName(name?: string, host?: string): string {
  return detectCountry(name, host)?.name || "";
}
