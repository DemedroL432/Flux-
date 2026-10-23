import React from "react";

export interface CountryFlagProps {
  code?: string;
  name?: string;
  className?: string;
  style?: React.CSSProperties;
  width?: number | string;
  height?: number | string;
}

export function CountryFlag({
  code,
  name,
  className,
  style,
  width = 22,
  height = 15,
}: CountryFlagProps) {
  const rawId = React.useId();
  const gbClipId = `gb-clip-${rawId.replace(/[^a-zA-Z0-9_-]/g, "_")}`;

  const c = (code || "").trim().toUpperCase();
  const target = `${code || ""} ${name || ""}`.trim();

  const containerStyle: React.CSSProperties = {
    borderRadius: 3,
    overflow: "hidden",
    display: "inline-block",
    verticalAlign: "middle",
    boxShadow: "0 0 4px rgba(0, 0, 0, 0.5)",
    flexShrink: 0,
    ...style,
  };

  // Finland: White field with blue Scandinavian cross (checked before NL to prevent substring collision)
  if (
    c === "FI" ||
    c === "FIN" ||
    /\b(fi|finland|финляндия|хельсинки|helsinki)\b/i.test(target) ||
    /(?:^|[^\p{L}\p{N}])(fi|finland|финляндия|хельсинки|helsinki)(?:[^\p{L}\p{N}]|$)/iu.test(target)
  ) {
    return (
      <svg
        width={width}
        height={height}
        viewBox="0 0 24 16"
        className={className}
        style={containerStyle}
        aria-label="Флаг Финляндии"
      >
        <rect width="24" height="16" fill="#ffffff" />
        <rect x="7" y="0" width="4" height="16" fill="#003580" />
        <rect x="0" y="6" width="24" height="4" fill="#003580" />
      </svg>
    );
  }

  // Netherlands: Red, White, Blue tricolor
  if (
    c === "NL" ||
    c === "NLD" ||
    /\b(nl|netherlands|нидерланды|амстердам|amsterdam|holland)\b/i.test(target) ||
    /(?:^|[^\p{L}\p{N}])(nl|netherlands|нидерланды|амстердам|amsterdam|holland)(?:[^\p{L}\p{N}]|$)/iu.test(target)
  ) {
    return (
      <svg
        width={width}
        height={height}
        viewBox="0 0 24 16"
        className={className}
        style={containerStyle}
        aria-label="Флаг Нидерландов"
      >
        <rect width="24" height="5.333" y="0" fill="#ae1c28" />
        <rect width="24" height="5.333" y="5.333" fill="#ffffff" />
        <rect width="24" height="5.334" y="10.666" fill="#21468b" />
      </svg>
    );
  }

  // Russia: White, Blue, Red tricolor
  if (
    c === "RU" ||
    c === "RUS" ||
    /\b(ru|russia|россия|москва|moscow)\b/i.test(target) ||
    /(?:^|[^\p{L}\p{N}])(ru|russia|россия|москва|moscow)(?:[^\p{L}\p{N}]|$)/iu.test(target)
  ) {
    return (
      <svg
        width={width}
        height={height}
        viewBox="0 0 24 16"
        className={className}
        style={containerStyle}
        aria-label="Флаг России"
      >
        <rect width="24" height="5.333" y="0" fill="#ffffff" />
        <rect width="24" height="5.333" y="5.333" fill="#0039a6" />
        <rect width="24" height="5.334" y="10.666" fill="#d52b1e" />
      </svg>
    );
  }

  // Germany: Black, Red, Gold tricolor
  if (
    c === "DE" ||
    c === "DEU" ||
    /\b(de|germany|германия|франкфурт|frankfurt|berlin)\b/i.test(target) ||
    /(?:^|[^\p{L}\p{N}])(de|germany|германия|франкфурт|frankfurt|berlin)(?:[^\p{L}\p{N}]|$)/iu.test(target)
  ) {
    return (
      <svg
        width={width}
        height={height}
        viewBox="0 0 24 16"
        className={className}
        style={containerStyle}
        aria-label="Флаг Германии"
      >
        <rect width="24" height="5.333" y="0" fill="#000000" />
        <rect width="24" height="5.333" y="5.333" fill="#dd0000" />
        <rect width="24" height="5.334" y="10.666" fill="#ffce00" />
      </svg>
    );
  }

  // Great Britain: Union Jack
  if (
    c === "GB" ||
    c === "GBR" ||
    c === "UK" ||
    /\b(gb|uk|britain|великобритания|лондон|london|england|united kingdom)\b/i.test(target) ||
    /(?:^|[^\p{L}\p{N}])(gb|uk|britain|великобритания|лондон|london|england|united kingdom)(?:[^\p{L}\p{N}]|$)/iu.test(target)
  ) {
    return (
      <svg
        width={width}
        height={height}
        viewBox="0 0 60 30"
        className={className}
        style={containerStyle}
        aria-label="Флаг Великобритании"
      >
        <clipPath id={gbClipId}>
          <rect width="60" height="30" />
        </clipPath>
        <g clipPath={`url(#${gbClipId})`}>
          <rect width="60" height="30" fill="#012169" />
          <path d="M0 0 L60 30 M60 0 L0 30" stroke="#ffffff" strokeWidth="6" />
          <path d="M0 0 L60 30 M60 0 L0 30" stroke="#c8102e" strokeWidth="2" />
          <path d="M30 0 v30 M0 15 h60" stroke="#ffffff" strokeWidth="10" />
          <path d="M30 0 v30 M0 15 h60" stroke="#c8102e" strokeWidth="6" />
        </g>
      </svg>
    );
  }

  // United States: Stars and Stripes
  if (
    c === "US" ||
    c === "USA" ||
    /\b(us|usa|сша|united states|america|америка|new york|miami)\b/i.test(target) ||
    /(?:^|[^\p{L}\p{N}])(us|usa|сша|united states|america|америка|new york|miami)(?:[^\p{L}\p{N}]|$)/iu.test(target)
  ) {
    return (
      <svg
        width={width}
        height={height}
        viewBox="0 0 24 16"
        className={className}
        style={containerStyle}
        aria-label="Флаг США"
      >
        <rect width="24" height="16" fill="#b22234" />
        <rect width="24" height="1.23" y="1.23" fill="#ffffff" />
        <rect width="24" height="1.23" y="3.69" fill="#ffffff" />
        <rect width="24" height="1.23" y="6.15" fill="#ffffff" />
        <rect width="24" height="1.23" y="8.61" fill="#ffffff" />
        <rect width="24" height="1.23" y="11.07" fill="#ffffff" />
        <rect width="24" height="1.23" y="13.53" fill="#ffffff" />
        <rect width="10.5" height="8.62" fill="#3c3b6e" />
        {/* Star clusters */}
        <circle cx="2.2" cy="1.8" r="0.6" fill="#ffffff" />
        <circle cx="5.2" cy="1.8" r="0.6" fill="#ffffff" />
        <circle cx="8.2" cy="1.8" r="0.6" fill="#ffffff" />
        <circle cx="3.7" cy="3.5" r="0.6" fill="#ffffff" />
        <circle cx="6.7" cy="3.5" r="0.6" fill="#ffffff" />
        <circle cx="2.2" cy="5.2" r="0.6" fill="#ffffff" />
        <circle cx="5.2" cy="5.2" r="0.6" fill="#ffffff" />
        <circle cx="8.2" cy="5.2" r="0.6" fill="#ffffff" />
        <circle cx="3.7" cy="6.9" r="0.6" fill="#ffffff" />
        <circle cx="6.7" cy="6.9" r="0.6" fill="#ffffff" />
      </svg>
    );
  }

  // Hungary: Red, White, Green tricolor
  if (
    c === "HU" ||
    c === "HUN" ||
    /\b(hu|hungary|венгрия|будапешт|budapest)\b/i.test(target) ||
    /(?:^|[^\p{L}\p{N}])(hu|hungary|венгрия|будапешт|budapest)(?:[^\p{L}\p{N}]|$)/iu.test(target)
  ) {
    return (
      <svg
        width={width}
        height={height}
        viewBox="0 0 24 16"
        className={className}
        style={containerStyle}
        aria-label="Флаг Венгрии"
      >
        <rect width="24" height="5.333" y="0" fill="#ce2939" />
        <rect width="24" height="5.333" y="5.333" fill="#ffffff" />
        <rect width="24" height="5.334" y="10.666" fill="#477050" />
      </svg>
    );
  }

  // Fallback: World Globe SVG
  return (
    <svg
      width={width}
      height={height}
      viewBox="0 0 24 24"
      fill="none"
      stroke="#38bdf8"
      strokeWidth="1.8"
      className={className}
      style={{
        ...containerStyle,
        background: "rgba(56, 189, 248, 0.12)",
        padding: 2,
        boxSizing: "border-box",
      }}
      aria-label="Глобальный узел"
    >
      <circle cx="12" cy="12" r="10" />
      <path d="M2 12h20M12 2a15.3 15.3 0 0 1 4 10 15.3 15.3 0 0 1-4 10 15.3 15.3 0 0 1-4-10 15.3 15.3 0 0 1 4-10z" />
    </svg>
  );
}

export default CountryFlag;
