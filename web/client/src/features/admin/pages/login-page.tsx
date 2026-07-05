import { useNavigate } from "react-router-dom";

import { setMockSession } from "../../../auth/auth-service";
import { useSession } from "../../../auth/session-context";
import type { Role } from "../../../auth/types";
import { env } from "../../../lib/config/env";

const localRoles: Array<{ role: Role; label: string; subject: string; buttonLabel: string }> = [
  { role: "player", label: "Local Player", subject: "local-player", buttonLabel: "Continue as player" },
  {
    role: "league_support",
    label: "Local Support",
    subject: "local-league-support",
    buttonLabel: "Continue as support",
  },
  { role: "league_admin", label: "Local Admin", subject: "local-league-admin", buttonLabel: "Continue as admin" },
  {
    role: "platform_operator",
    label: "Local Platform Operator",
    subject: "local-platform-operator",
    buttonLabel: "Continue as platform operator",
  },
];

export function LoginPage() {
  const session = useSession();
  const navigate = useNavigate();

  async function handleMockLogin(role: Role) {
    const localRole = localRoles.find((item) => item.role === role);

    if (!localRole) {
      return;
    }

    setMockSession({
      subject: localRole.subject,
      displayName: localRole.label,
      roles: [localRole.role],
    });
    await session.refresh();
    navigate("/app", { replace: true });
  }

  function handleApiLogin(role: Role) {
    const localRole = localRoles.find((item) => item.role === role);

    if (!localRole) {
      return;
    }

    const redirect = `${window.location.origin}/app`;
    const query = new URLSearchParams({
      subject: localRole.subject,
      displayName: localRole.label,
      roles: role,
      redirect,
    });
    window.location.assign(`${env.apiBaseUrl}/v1/auth/login?${query.toString()}`);
  }

  return (
    <main className="login-page">
      <section className="login-page__card">
        <img alt="Kinetic" className="login-page__brand" src="/img/logo-horizontal.svg" />
        <h1>Kinetic Sign In</h1>
        {env.authMode === "mock" ? (
          <>
            <p>Choose a local role for development testing.</p>
            <div className="login-page__actions">
              {localRoles.map((localRole) => (
                <button key={localRole.role} onClick={() => handleMockLogin(localRole.role)} type="button">
                  {localRole.buttonLabel}
                </button>
              ))}
            </div>
          </>
        ) : (
          <>
            <p>Sign in through the local OIDC shim.</p>
            <div className="login-page__actions">
              {localRoles.map((localRole) => (
                <button key={localRole.role} onClick={() => handleApiLogin(localRole.role)} type="button">
                  {localRole.buttonLabel}
                </button>
              ))}
            </div>
          </>
        )}
      </section>
    </main>
  );
}
