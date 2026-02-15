import { useQuery } from "@tanstack/react-query";

import { AppShell } from "../../../components/layout/app-shell";
import { LoadingState } from "../../../components/feedback/loading-state";
import { getJson } from "../../../lib/api/client";
import { ApiError } from "../../../lib/api/errors";
import { exceptionMetricsSchema } from "../../../lib/api/schemas";

const opsLinks = [
  { label: "Grafana", href: "https://grafana.example.internal" },
  { label: "GitHub Actions", href: "https://github.com/SprocketBot/sprocket-v3/actions" },
  { label: "GHCR Packages", href: "https://github.com/orgs/SprocketBot/packages" },
];

async function getExceptionMetrics() {
  return getJson("/v1/exception-metrics", exceptionMetricsSchema);
}

export function PlatformOpsPage() {
  const metricsQuery = useQuery({
    queryKey: ["exception-metrics"],
    queryFn: getExceptionMetrics,
  });

  return (
    <AppShell title="Sprocket Web Client">
      <h2>Platform Operations</h2>
      <p>Operational links and key exception metrics.</p>

      <section aria-label="operations links" style={{ display: "flex", gap: "1rem", marginBottom: "1rem" }}>
        {opsLinks.map((link) => (
          <a href={link.href} key={link.label} rel="noreferrer" target="_blank">
            {link.label}
          </a>
        ))}
      </section>

      {metricsQuery.isLoading && <LoadingState label="Loading exception metrics..." />}
      {metricsQuery.isError && <ErrorView error={metricsQuery.error} />}

      {metricsQuery.isSuccess && (
        <section style={{ display: "grid", gap: "1rem", gridTemplateColumns: "1fr 1fr" }}>
          <MetricCard label="Admin hours / week" value={metricsQuery.data.adminHoursPerWeek.toFixed(2)} />
          <MetricCard label="Manual touches / fixture" value={metricsQuery.data.manualTouchesPerFixture.toFixed(2)} />
          <MetricCard label="Zero touch fixture rate" value={`${(metricsQuery.data.zeroTouchFixtureRate * 100).toFixed(1)}%`} />
          <MetricCard label="Time to close p50 (hours)" value={metricsQuery.data.timeToCloseHoursP50.toFixed(2)} />
        </section>
      )}
    </AppShell>
  );
}

function MetricCard({ label, value }: { label: string; value: string }) {
  return (
    <article style={{ border: "1px solid #cbd5e1", borderRadius: "8px", padding: "0.8rem" }}>
      <h3 style={{ margin: 0 }}>{label}</h3>
      <p style={{ fontSize: "1.3rem", fontWeight: 700, marginBottom: 0 }}>{value}</p>
    </article>
  );
}

function ErrorView({ error }: { error: Error }) {
  if (error instanceof ApiError) {
    return (
      <p role="alert">
        Metrics request failed ({error.status}){error.code ? ` [${error.code}]` : ""}: {error.message}
      </p>
    );
  }

  return <p role="alert">Metrics request failed: {error.message}</p>;
}
