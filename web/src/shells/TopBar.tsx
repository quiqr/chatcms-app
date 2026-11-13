import type { ReactNode } from "react";

import { cn } from "../lib/cn";
import { PersonMenu } from "./PersonMenu";

export function TopBar({
  viewer,
  className,
  children,
}: {
  viewer?: { name: string; role?: string };
  className?: string;
  children?: ReactNode;
}) {
  return (
    <header
      className={cn(
        "flex min-h-14 flex-wrap items-center gap-3 border-b border-border bg-card px-6 py-3",
        className,
      )}
    >
      {children}
      {viewer && <PersonMenu name={viewer.name} role={viewer.role} className="ml-auto" />}
    </header>
  );
}
