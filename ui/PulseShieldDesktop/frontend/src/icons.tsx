import type { CSSProperties, ReactNode } from "react";

/** Набор линейных иконок интерфейса. Совпадает по именам с дизайн-системой. */
const BASE: CSSProperties & Record<string, unknown> = {
  width: 18,
  height: 18,
  viewBox: "0 0 24 24",
  fill: "none",
  stroke: "currentColor",
  strokeWidth: 1.8,
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
          <circle cx="6" cy="6" r="2.5" />
          <circle cx="18" cy="18" r="2.5" />
          <path d="M8.5 6H14a4 4 0 0 1 0 8h-4a4 4 0 0 0 0 8h.5" />
        </>
      );
    case "sliders":
      return svg(
        <>
          <path d="M4 6h10M18 6h2M4 12h4M12 12h8M4 18h12M20 18h0" />
          <circle cx="16" cy="6" r="2" />
          <circle cx="10" cy="12" r="2" />
          <circle cx="18" cy="18" r="2" />
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
      return svg(<path d="M13 2 4 14h6l-1 8 9-12h-6l1-8Z" />);
    case "server":
      return svg(
        <>
          <rect x="3" y="4" width="18" height="6" rx="2" />
          <rect x="3" y="14" width="18" height="6" rx="2" />
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
      return svg(
        <>
          <path d="M4 7h16M9 7V5h6v2M6 7l1 13h10l1-13" />
        </>
      );
    case "copy":
      return svg(
        <>
          <rect x="9" y="9" width="11" height="11" rx="2" />
          <path d="M5 15V5h10" />
        </>
      );
    case "folder":
      return svg(<path d="M3 7a2 2 0 0 1 2-2h4l2 2h8a2 2 0 0 1 2 2v8a2 2 0 0 1-2 2H5a2 2 0 0 1-2-2V7Z" />);
    case "palette":
      return svg(
        <>
          <circle cx="12" cy="12" r="9" />
          <circle cx="9" cy="9" r="1.2" fill="currentColor" stroke="none" />
          <circle cx="15" cy="9" r="1.2" fill="currentColor" stroke="none" />
          <circle cx="9.5" cy="15" r="1.2" fill="currentColor" stroke="none" />
        </>
      );
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
