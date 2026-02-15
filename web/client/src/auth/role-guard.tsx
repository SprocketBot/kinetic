import type { ReactNode } from "react";
import { Navigate } from "react-router-dom";

import { useSession } from "./session-context";
import type { Role } from "./types";

type RoleGuardProps = {
  children?: ReactNode;
  roles?: Role[];
  fallbacks?: Array<{ role: Role; to: string }>;
};

export function RoleGuard({ children, roles, fallbacks }: RoleGuardProps) {
  const { principal } = useSession();
  const assignedRoles = principal?.roles ?? [];

  if (fallbacks && fallbacks.length > 0) {
    const match = fallbacks.find(({ role }) => assignedRoles.includes(role));
    return <Navigate to={match?.to ?? "/unauthorized"} replace />;
  }

  if (!roles || roles.some((role) => assignedRoles.includes(role))) {
    return <>{children}</>;
  }

  return <Navigate to="/unauthorized" replace />;
}
