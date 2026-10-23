import type { ReactNode } from "react";

type Props = { text: string; label?: string };

/** Компактная подсказка по hover/focus — текст под заголовками переносим сюда. */
export function PageTip({ text, label = "Справка" }: Props) {
  return (
    <span className="page-tip-wrap">
      <span className="page-tip-mark" tabIndex={0} role="img" aria-label={label + ": " + text} title={text}>
        i
      </span>
    </span>
  );
}

/** Заголовок секции + значок (i). */
export function SectionTitle({ children, tip }: { children: ReactNode; tip?: string }) {
  return (
    <div className="section-title-row">
      <h3>{children}</h3>
      {tip ? <PageTip text={tip} /> : null}
    </div>
  );
}
