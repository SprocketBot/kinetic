import type { ReactNode } from "react";

import { AppNav } from "./nav";

type AppShellProps = {
  title: string;
  children: ReactNode;
};

export function AppShell({ title, children }: AppShellProps) {
  return (
    <div style={{ fontFamily: "ui-sans-serif, system-ui", minHeight: "100vh" }}>
      <header style={{ borderBottom: "1px solid #d1d5db", padding: "0.75rem 1rem" }}>
        <h1 style={{ fontSize: "1.1rem", margin: 0 }}>{title}</h1>
      </header>
      <div style={{ display: "grid", gridTemplateColumns: "220px 1fr", minHeight: "calc(100vh - 56px)" }}>
        <aside style={{ borderRight: "1px solid #d1d5db", padding: "1rem" }}>
          <AppNav />
        </aside>
        <main style={{ padding: "1rem" }}>{children}</main>
      </div>
    </div>
  );
}
