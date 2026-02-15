import { Link } from "react-router-dom";

import { useSession } from "../../auth/session-context";
import type { Role } from "../../auth/types";

type NavItem = {
  to: string;
  label: string;
  role: Role;
};

const navItems: NavItem[] = [
  { to: "/app/player", label: "Player", role: "player" },
  { to: "/app/support", label: "Support", role: "league_support" },
  { to: "/app/admin", label: "Admin", role: "league_admin" },
  { to: "/app/platform", label: "Platform", role: "platform_operator" },
];

export function AppNav() {
  const { principal } = useSession();
  const roles = principal?.roles ?? [];

  return (
    <nav>
      <div style={{ fontWeight: 600, marginBottom: "0.5rem" }}>Navigation</div>
      <ul style={{ listStyle: "none", margin: 0, padding: 0, display: "grid", gap: "0.4rem" }}>
        {navItems
          .filter((item) => roles.includes(item.role))
          .map((item) => (
            <li key={item.to}>
              <Link to={item.to}>{item.label}</Link>
            </li>
          ))}
      </ul>
    </nav>
  );
}
