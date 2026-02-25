import { Link } from "react-router-dom";

export function UnauthorizedPage() {
  return (
    <main className="unauthorized-page">
      <section className="unauthorized-page__card">
        <h1>Unauthorized</h1>
        <p>Your session does not have access to that route.</p>
        <Link to="/app">Return to app</Link>
      </section>
    </main>
  );
}
