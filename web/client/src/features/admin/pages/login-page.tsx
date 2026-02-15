import { useNavigate } from "react-router-dom";

import { setMockSession } from "../../../auth/auth-service";
import { useSession } from "../../../auth/session-context";

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

  return (
    <main style={{ margin: "3rem auto", maxWidth: 420, fontFamily: "ui-sans-serif, system-ui" }}>
      <h1>Sprocket Sign In</h1>
      <p>OAuth/OIDC provider wiring lands with API-WEB-01. Mock login is available for local Phase 0 work.</p>
      <button onClick={handleMockLogin} type="button">
        Continue with mock session
      </button>
    </main>
  );
}
