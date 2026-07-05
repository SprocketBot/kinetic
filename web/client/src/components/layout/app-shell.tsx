import type { ReactNode } from "react";

import { useSession } from "../../auth/session-context";
import { AppNav } from "./nav";

type AppShellProps = {
  title: string;
  children: ReactNode;
};

export function AppShell({ title, children }: AppShellProps) {
  const { principal } = useSession();
  const displayName = principal?.displayName ?? principal?.subject ?? "Local session";
  const roles = principal?.roles ?? [];

  function roleLabel(role: string): string {
    if (role === "league_support") {
      return "Support";
    }
    if (role === "league_admin") {
      return "Admin";
    }
    if (role === "platform_operator") {
      return "Platform";
    }
    if (role === "player") {
      return "Player";
    }
    return role;
  }

  return (
    <div className="app-shell">
      <header className="app-shell__header">
        <div className="app-shell__brand">
          <img alt="Kinetic" className="app-shell__logo" src="/img/logo-horizontal.svg" />
          <div className="app-shell__title-wrap">
            <p className="app-shell__eyebrow">Kinetic Console</p>
            <h1 className="app-shell__title">{title}</h1>
          </div>
        </div>

        <div className="app-shell__user">
          <p className="app-shell__user-label">{displayName}</p>
          <div className="app-shell__role-list">
            {roles.length === 0 && <span className="app-shell__role-chip">No role</span>}
            {roles.map((role) => (
              <span className="app-shell__role-chip" key={role}>
                {roleLabel(role)}
              </span>
            ))}
          </div>
        </div>
      </header>
      <div className="app-shell__body">
        <aside className="app-shell__sidebar">
          <AppNav />
        </aside>
        <main className="app-shell__main">
          <div className="app-shell__content">{children}</div>
        </main>
      </div>
    </div>
  );
}
