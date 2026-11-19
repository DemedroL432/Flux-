import type { CSSProperties, ReactNode } from "react";

const BASE: CSSProperties & Record<string, unknown> = {
  width: 16,
  height: 16,
  viewBox: "0 0 24 24",
  fill: "none",
  stroke: "currentColor",
  strokeWidth: 1.75,
  strokeLinecap: "round",
  strokeLinejoin: "round",
};

type Props = { name: string; style?: CSSProperties; size?: number };

export function Icon({ name, style, size }: Props) {
  const p: CSSProperties & Record<string, unknown> = { ...BASE, ...(style || {}) };
  if (size) {
    p.width = size;
    p.height = size;
  }
  const svg = (children: ReactNode) => (
    <svg {...(p as React.SVGProps<SVGSVGElement>)}>{children}</svg>
  );

  switch (name) {
    case "overview":
      return svg(
        <>
          <rect x="3" y="3" width="8" height="8" rx="1.5" />
          <rect x="13" y="3" width="8" height="5" rx="1.5" />
          <rect x="13" y="10" width="8" height="11" rx="1.5" />
          <rect x="3" y="13" width="8" height="8" rx="1.5" />
        </>
      );
    case "back":
      return svg(<path d="M15 6l-6 6 6 6" />);
    case "bolt":
      return svg(<path d="M13 2 4 14h6l-1 8 9-12h-6l1-8Z" />);
    case "globe":
      return svg(
        <>
          <circle cx="12" cy="12" r="9" />
          <path d="M3 12h18M12 3c3 3.6 3 14.4 0 18M12 3c-3 3.6-3 14.4 0 18" />
        </>
      );
    case "pulse":
      return svg(<path d="M2 12h3l2-6 3 12 3-9 2 3h4" />);
    case "link":
      return svg(
        <path d="M10 13a5 5 0 0 0 7.54.54l3-3a5 5 0 0 0-7.07-7.07l-1.72 1.71M14 11a5 5 0 0 0-7.54-.54l-3 3a5 5 0 0 0 7.07 7.07l1.71-1.71" />
      );
    case "log":
      return svg(
        <>
          <rect x="4" y="3" width="16" height="18" rx="2" />
          <path d="M8 7h8M8 11h8M8 15h5" />
        </>
      );
    case "shield":
      return svg(<path d="M12 3 5 6v5c0 4.5 3 7.5 7 9 4-1.5 7-4.5 7-9V6l-7-3Z" />);
    case "routes":
      return svg(
        <>
          <circle cx="6" cy="18" r="2.5" />
          <circle cx="18" cy="6" r="2.5" />
          <path d="M8.5 18H13a3 3 0 0 0 3-3V8.5" />
        </>
      );
    case "sliders":
      return svg(
        <>
          <path d="M4 7h8M16 7h4M4 17h4M12 17h8" />
          <circle cx="14" cy="7" r="2" />
          <circle cx="10" cy="17" r="2" />
        </>
      );
    case "dots":
      return svg(
        <>
          <circle cx="5" cy="12" r="1.4" fill="currentColor" stroke="none" />
          <circle cx="12" cy="12" r="1.4" fill="currentColor" stroke="none" />
          <circle cx="19" cy="12" r="1.4" fill="currentColor" stroke="none" />
        </>
      );
    case "power":
      return svg(
        <>
          <path d="M12 3v9" />
          <path d="M6.5 6.5a8 8 0 1 0 11 0" />
        </>
      );
    case "vpn":
      return svg(
        <>
          <path d="M12 3 5 6v5c0 4.5 3 7.5 7 9 4-1.5 7-4.5 7-9V6l-7-3Z" />
          <path d="M9.5 12l1.8 1.8L15 10" />
        </>
      );
    case "dpi":
      return svg(<path d="M12 21s7-3.5 7-9V5.5L12 3 5 5.5V12c0 5.5 7 9 7 9z" />);
    case "server":
      return svg(
        <>
          <rect x="3" y="4" width="18" height="6" rx="1.5" />
          <rect x="3" y="14" width="18" height="6" rx="1.5" />
          <circle cx="7" cy="7" r="1" fill="currentColor" stroke="none" />
          <circle cx="7" cy="17" r="1" fill="currentColor" stroke="none" />
        </>
      );
    case "gauge":
      return svg(
        <>
          <path d="M4 16a8 8 0 1 1 16 0" />
          <path d="M12 16l3-4" />
          <circle cx="12" cy="16" r="1.4" fill="currentColor" stroke="none" />
        </>
      );
    case "download":
      return svg(
        <>
          <path d="M12 4v10M8 11l4 4 4-4" />
          <path d="M5 19h14" />
        </>
      );
    case "wave":
      return svg(<path d="M2 12h3l2-6 3 12 3-9 2 3h4" />);
    case "refresh":
      return svg(
        <>
          <path d="M20 11a8 8 0 1 0-1.5 5" />
          <path d="M20 5v6h-6" />
        </>
      );
    case "info":
      return svg(
        <>
          <circle cx="12" cy="12" r="9" />
          <path d="M12 11v5M12 8h.01" />
        </>
      );
    case "chev":
      return svg(<path d="M9 6l6 6-6 6" />);
    case "close":
      return svg(<path d="M6 6l12 12M18 6L6 18" />);
    case "trash":
      return svg(<path d="M4 7h16M9 7V5h6v2M6 7l1 13h10l1-13" />);
    case "copy":
      return svg(
        <>
          <rect x="9" y="9" width="11" height="11" rx="2" />
          <path d="M5 15V5h10" />
        </>
      );
    case "folder":
      return svg(<path d="M3 7a2 2 0 0 1 2-2h4l2 2h8a2 2 0 0 1 2 2v8a2 2 0 0 1-2 2H5a2 2 0 0 1-2-2V7Z" />);
    case "plus":
      return svg(<path d="M12 5v14M5 12h14" />);
    case "check":
      return svg(<path d="M4 12l5 5L20 6" />);
    case "warn":
      return svg(
        <>
          <path d="M12 3 2 20h20L12 3Z" />
          <path d="M12 10v4M12 17h.01" />
        </>
      );
    default:
      return svg(<circle cx="12" cy="12" r="7" />);
  }
}

const H_STRIPES: Record<string, string[]> = {
  nl: ["#AE1C28", "#FFFFFF", "#21468B"],
  de: ["#181818", "#DD0000", "#FFCE00"],
  ru: ["#FFFFFF", "#0039A6", "#D52B1E"],
  pl: ["#FFFFFF", "#DC143C"],
  ua: ["#005BBB", "#FFD500"],
  at: ["#ED2939", "#FFFFFF", "#ED2939"],
  lt: ["#FDB913", "#006A44", "#C1272D"],
  ee: ["#0072CE", "#181818", "#FFFFFF"],
  lv: ["#9E3039", "#9E3039", "#FFFFFF", "#9E3039", "#9E3039"],
  hu: ["#CD2A3E", "#FFFFFF", "#436F4D"],
  bg: ["#FFFFFF", "#00966E", "#D62612"],
  lu: ["#EA141D", "#FFFFFF", "#00A1DE"],
  rs: ["#C6363C", "#0C4076", "#FFFFFF"],
  es: ["#AA151B", "#F1BF00", "#F1BF00", "#AA151B"],
  sg: ["#EF3340", "#FFFFFF"],
  id: ["#CE1126", "#FFFFFF"],
  hk: ["#DE2910"],
  am: ["#D90012", "#0033A0", "#F2A800"],
  si: ["#FFFFFF", "#005DA4", "#ED1C24"],
  sk: ["#FFFFFF", "#0B4EA2", "#EE1C25"],
};

const V_STRIPES: Record<string, string[]> = {
  fr: ["#002395", "#FFFFFF", "#ED2939"],
  it: ["#009246", "#FFFFFF", "#CE2B37"],
  be: ["#181818", "#FAE042", "#ED2939"],
  ie: ["#169B62", "#FFFFFF", "#FF883E"],
  ro: ["#002B7F", "#FCD116", "#CE1126"],
  md: ["#0046AE", "#FFD200", "#CC092F"],
  ca: ["#D80621", "#FFFFFF", "#D80621"],
};

const NORDIC: Record<string, { bg: string; fg: string; inner?: string }> = {
  fi: { bg: "#FFFFFF", fg: "#003580" },
  se: { bg: "#006AA7", fg: "#FECC02" },
  dk: { bg: "#C60C30", fg: "#FFFFFF" },
  no: { bg: "#BA0C2F", fg: "#FFFFFF", inner: "#00205B" },
  is: { bg: "#02529C", fg: "#FFFFFF", inner: "#DC1E35" },
};

/** Чистый векторный флаг страны (не зависит от системного эмодзи-шрифта Windows). */
export function FlagBadge({ code, name }: { code?: string; name?: string }) {
  const c = (code || "").toLowerCase();
  const title = name || (c ? c.toUpperCase() : "Узел");

  const wrap = (inner: ReactNode) => (
    <svg
      width="20"
      height="14"
      viewBox="0 0 20 14"
      style={{ borderRadius: 2, flex: "0 0 20px", boxShadow: "inset 0 0 0 1px rgba(255,255,255,0.12)" }}
      aria-label={title}
    >
      <title>{title}</title>
      {inner}
    </svg>
  );

  if (H_STRIPES[c]) {
    const arr = H_STRIPES[c];
    const h = 14 / arr.length;
    return wrap(arr.map((fill, i) => <rect key={i} x="0" y={i * h} width="20" height={h + 0.1} fill={fill} />));
  }

  if (V_STRIPES[c]) {
    const arr = V_STRIPES[c];
    const w = 20 / arr.length;
    return wrap(arr.map((fill, i) => <rect key={i} x={i * w} y="0" width={w + 0.1} height="14" fill={fill} />));
  }

  if (NORDIC[c]) {
    const { bg, fg, inner } = NORDIC[c];
    return wrap(
      <>
        <rect width="20" height="14" fill={bg} />
        <rect x="5" y="0" width="3.6" height="14" fill={fg} />
        <rect x="0" y="5.2" width="20" height="3.6" fill={fg} />
        {inner ? (
          <>
            <rect x="6" y="0" width="1.6" height="14" fill={inner} />
            <rect x="0" y="6.2" width="20" height="1.6" fill={inner} />
          </>
        ) : null}
      </>
    );
  }

  if (c === "us") {
    return wrap(
      <>
        <rect width="20" height="14" fill="#B22234" />
        <rect y="2" width="20" height="2" fill="#FFFFFF" />
        <rect y="6" width="20" height="2" fill="#FFFFFF" />
        <rect y="10" width="20" height="2" fill="#FFFFFF" />
        <rect width="9" height="8" fill="#3C3B6E" />
      </>
    );
  }

  if (c === "gb") {
    return wrap(
      <>
        <rect width="20" height="14" fill="#012169" />
        <path d="M0 0L20 14M20 0L0 14" stroke="#FFFFFF" strokeWidth="2.4" />
        <path d="M0 0L20 14M20 0L0 14" stroke="#C8102E" strokeWidth="1.2" />
        <rect x="8.2" y="0" width="3.6" height="14" fill="#FFFFFF" />
        <rect x="0" y="5.2" width="20" height="3.6" fill="#FFFFFF" />
        <rect x="9" y="0" width="2" height="14" fill="#C8102E" />
        <rect x="0" y="6" width="20" height="2" fill="#C8102E" />
      </>
    );
  }

  if (c === "kz") {
    return wrap(
      <>
        <rect width="20" height="14" fill="#00AFCA" />
        <circle cx="10" cy="7" r="3.1" fill="#FEC50C" />
      </>
    );
  }

  if (c === "tr") {
    return wrap(
      <>
        <rect width="20" height="14" fill="#E30A17" />
        <circle cx="8.5" cy="7" r="3.4" fill="#FFFFFF" />
        <circle cx="9.6" cy="7" r="2.7" fill="#E30A17" />
      </>
    );
  }

  if (c === "jp") {
    return wrap(
      <>
        <rect width="20" height="14" fill="#FFFFFF" />
        <circle cx="10" cy="7" r="3.8" fill="#BC002D" />
      </>
    );
  }

  if (c === "ch") {
    return wrap(
      <>
        <rect width="20" height="14" fill="#D52B1E" />
        <rect x="8.5" y="3" width="3" height="8" fill="#FFFFFF" />
        <rect x="6" y="5.5" width="8" height="3" fill="#FFFFFF" />
      </>
    );
  }

  if (c === "cz") {
    return wrap(
      <>
        <rect width="20" height="7" fill="#FFFFFF" />
        <rect y="7" width="20" height="7" fill="#D7141A" />
        <polygon points="0,0 10,7 0,14" fill="#11457E" />
      </>
    );
  }

  if (c === "ae") {
    return wrap(
      <>
        <rect x="5" y="0" width="15" height="4.7" fill="#00732F" />
        <rect x="5" y="4.7" width="15" height="4.6" fill="#FFFFFF" />
        <rect x="5" y="9.3" width="15" height="4.7" fill="#181818" />
        <rect x="0" y="0" width="5.5" height="14" fill="#FF0000" />
      </>
    );
  }

  if (c === "il") {
    return wrap(
      <>
        <rect width="20" height="14" fill="#FFFFFF" />
        <rect y="1.8" width="20" height="2" fill="#0038B8" />
        <rect y="10.2" width="20" height="2" fill="#0038B8" />
        <circle cx="10" cy="7" r="2.1" stroke="#0038B8" strokeWidth="1" fill="none" />
      </>
    );
  }

  if (c) {
    return (
      <span className="fx-flag-code fx-mono" title={title}>
        {c.slice(0, 2).toUpperCase()}
      </span>
    );
  }

  return <Icon name="globe" size={16} style={{ color: "var(--text-muted)" }} />;
}
