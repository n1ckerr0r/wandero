import type { SVGProps } from "react";

export type IconName =
  | "add"
  | "arrowUp"
  | "arrowDown"
  | "check"
  | "close"
  | "map"
  | "refresh"
  | "route"
  | "search"
  | "walk";

const paths: Record<IconName, string> = {
  add: "M12 5v14M5 12h14",
  arrowDown: "M12 5v14m-6-6 6 6 6-6",
  arrowUp: "M12 19V5m-6 6 6-6 6 6",
  check: "m5 12 4 4L19 6",
  close: "M6 6l12 12M18 6 6 18",
  map: "M9 5 3 7v12l6-2 6 2 6-2V7l-6 2-6-2Zm0 0v12m6-10v12",
  refresh: "M20 11a8 8 0 1 0 2 5.3M20 4v7h-7",
  route: "M6 18c4-5 8-8 12-12M6 6h.01M18 18h.01",
  search: "m21 21-4.3-4.3M11 18a7 7 0 1 1 0-14 7 7 0 0 1 0 14Z",
  walk: "M13 5a2 2 0 1 0 0-4 2 2 0 0 0 0 4Zm-1 4 3 2 3 1m-6-3-2 5 3 2-2 5m2-5 4-3 2 4",
};

export function Icon({ name, title, ...props }: { name: IconName; title?: string } & SVGProps<SVGSVGElement>) {
  return (
    <svg viewBox="0 0 24 24" aria-hidden={title ? undefined : true} role={title ? "img" : undefined} focusable="false" {...props}>
      {title ? <title>{title}</title> : null}
      <path d={paths[name]} fill="none" stroke="currentColor" strokeWidth="2" strokeLinecap="round" strokeLinejoin="round" />
    </svg>
  );
}
