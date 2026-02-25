import { NavLink } from "react-router-dom";

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
    <nav className="app-nav">
      <p className="app-nav__title">Navigation</p>
      <ul className="app-nav__list">
        {navItems
          .filter((item) => roles.includes(item.role))
          .map((item) => (
            <li key={item.to}>
              <NavLink className={({ isActive }) => `app-nav__link${isActive ? " active" : ""}`} to={item.to}>
                {item.label}
              </NavLink>
            </li>
          ))}
      </ul>
    </nav>
  );
}
