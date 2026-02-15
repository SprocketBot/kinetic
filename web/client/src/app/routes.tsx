import { BrowserRouter, Navigate, Route, Routes } from "react-router-dom";
import type { ReactElement } from "react";

import { AuthGuard } from "../auth/auth-guard";
import { RoleGuard } from "../auth/role-guard";
import { AdminHomePage } from "../features/admin/pages/admin-home-page";
import { LoginPage } from "../features/admin/pages/login-page";
import { UnauthorizedPage } from "../features/admin/pages/unauthorized-page";
import { PlatformOpsPage } from "../features/platform/pages/platform-ops-page";
import { PlayerHomePage } from "../features/player/pages/player-home-page";
import { SupportDashboardPage } from "../features/support/pages/support-dashboard-page";

type AppRole = "player" | "league_admin" | "league_support" | "platform_operator";

function AppLandingRedirect() {
  return (
    <RoleGuard
      fallbacks={[
        { role: "league_support", to: "/app/support" },
        { role: "platform_operator", to: "/app/platform" },
        { role: "league_admin", to: "/app/admin" },
        { role: "player", to: "/app/player" },
      ]}
    />
  );
}

function ScopedRoute({ role, to, element }: { role: AppRole; to: string; element: ReactElement }) {
  return (
    <Route
      path={to}
      element={
        <AuthGuard>
          <RoleGuard roles={[role]}>{element}</RoleGuard>
        </AuthGuard>
      }
    />
  );
}

export function AppRoutes() {
  return (
    <BrowserRouter>
      <Routes>
        <Route path="/login" element={<LoginPage />} />
        <Route path="/unauthorized" element={<UnauthorizedPage />} />

        <Route
          path="/app"
          element={
            <AuthGuard>
              <AppLandingRedirect />
            </AuthGuard>
          }
        />

        {ScopedRoute({ role: "player", to: "/app/player", element: <PlayerHomePage /> })}
        {ScopedRoute({ role: "league_support", to: "/app/support", element: <SupportDashboardPage /> })}
        {ScopedRoute({ role: "league_admin", to: "/app/admin", element: <AdminHomePage /> })}
        {ScopedRoute({ role: "platform_operator", to: "/app/platform", element: <PlatformOpsPage /> })}

        <Route path="*" element={<Navigate to="/app" replace />} />
      </Routes>
    </BrowserRouter>
  );
}
