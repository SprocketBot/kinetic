import { Link } from "react-router-dom";

export function UnauthorizedPage() {
  return (
    <main style={{ margin: "3rem auto", maxWidth: 480, fontFamily: "ui-sans-serif, system-ui" }}>
      <h1>Unauthorized</h1>
      <p>Your session does not have access to that route.</p>
      <Link to="/app">Return to app</Link>
    </main>
  );
}
