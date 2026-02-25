import { useNavigate } from "react-router-dom";

import { setMockSession } from "../../../auth/auth-service";
import { useSession } from "../../../auth/session-context";
import { env } from "../../../lib/config/env";

const mockSupportPrincipal = {
  subject: "local-support",
  displayName: "Local Support Operator",
  roles: ["league_support" as const],
};

export function LoginPage() {
  const session = useSession();
  const navigate = useNavigate();

  async function handleMockLogin() {
    setMockSession(mockSupportPrincipal);
    await session.refresh();
    navigate("/app", { replace: true });
  }

  function handleApiLogin(role: "player" | "league_support" | "league_admin" | "platform_operator") {
    const labels: Record<typeof role, string> = {
      player: "Local Player",
      league_support: "Local Support",
      league_admin: "Local Admin",
      platform_operator: "Local Platform Operator",
    };
    const subject = `local-${role.replace("_", "-")}`;
    const redirect = `${window.location.origin}/app`;
    const query = new URLSearchParams({
      subject,
      displayName: labels[role],
      roles: role,
      redirect,
    });
    window.location.assign(`${env.apiBaseUrl}/v1/auth/login?${query.toString()}`);
  }

  return (
    <main className="login-page">
      <section className="login-page__card">
        <img alt="Sprocket" className="login-page__brand" src="/img/logo-horizontal.svg" />
        <h1>Sprocket Sign In</h1>
        {env.authMode === "mock" ? (
          <>
            <p>Mock login is enabled for local development.</p>
            <div className="login-page__actions">
              <button onClick={handleMockLogin} type="button">
                Continue with mock session
              </button>
            </div>
          </>
        ) : (
          <>
            <p>Sign in through the local OIDC shim.</p>
            <div className="login-page__actions">
              <button onClick={() => handleApiLogin("player")} type="button">
                Continue as player
              </button>
              <button onClick={() => handleApiLogin("league_support")} type="button">
                Continue as support
              </button>
              <button onClick={() => handleApiLogin("league_admin")} type="button">
                Continue as admin
              </button>
              <button onClick={() => handleApiLogin("platform_operator")} type="button">
                Continue as platform operator
              </button>
            </div>
          </>
        )}
      </section>
    </main>
  );
}
