import { QueryClient, QueryClientProvider } from "@tanstack/react-query";
import { cleanup, render, screen, within } from "@testing-library/react";
import userEvent from "@testing-library/user-event";
import { http, HttpResponse } from "msw";
import { MemoryRouter } from "react-router-dom";
import { afterAll, afterEach, beforeAll, describe, expect, it } from "vitest";

import { setMockSession } from "../../../auth/auth-service";
import { SessionProvider } from "../../../auth/session-context";
import { server } from "../../../test/msw/server";
import { SupportDashboardPage } from "./support-dashboard-page";

type Ticket = {
  id: number;
  category: string;
  contextType: string;
  contextId: number;
  reportedByTeamId: number;
  state: string;
  reasonCode: string;
  severity: number;
  suggestedAction: string;
  detailsJson: Record<string, unknown>;
  resolutionCode: string | null;
  openedAt: string;
  triagedAt: string | null;
  resolvedAt: string | null;
};

beforeAll(() => server.listen());
afterEach(() => {
  server.resetHandlers();
  cleanup();
});
afterAll(() => server.close());

function renderPage() {
  const queryClient = new QueryClient();

  render(
    <MemoryRouter>
      <QueryClientProvider client={queryClient}>
        <SessionProvider>
          <SupportDashboardPage />
        </SessionProvider>
      </QueryClientProvider>
    </MemoryRouter>,
  );
}

describe("SupportDashboardPage", () => {
  it("renders and filters operator inbox", async () => {
    const tickets: Ticket[] = [
      {
        id: 1,
        category: "scheduling_conflict",
        contextType: "match",
        contextId: 100,
        reportedByTeamId: 10,
        state: "open",
        reasonCode: "availability_conflict",
        severity: 3,
        suggestedAction: "request_reschedule",
        detailsJson: {},
        resolutionCode: null,
        openedAt: "2026-02-15T00:00:00Z",
        triagedAt: null,
        resolvedAt: null,
      },
      {
        id: 2,
        category: "result_dispute",
        contextType: "match",
        contextId: 101,
        reportedByTeamId: 11,
        state: "triaged",
        reasonCode: "dispute_opened",
        severity: 5,
        suggestedAction: "review_replay",
        detailsJson: {},
        resolutionCode: null,
        openedAt: "2026-02-15T00:00:00Z",
        triagedAt: "2026-02-15T01:00:00Z",
        resolvedAt: null,
      },
    ];

    server.use(
      http.get("http://localhost:8080/v1/operator-inbox", () => HttpResponse.json(tickets)),
    );

    setMockSession({
      subject: "support-1",
      displayName: "Support",
      roles: ["league_support"],
    });

    renderPage();

    expect(await screen.findByTestId("operator-inbox-count")).toHaveTextContent("Visible tickets: 2");
    expect(screen.getByTestId("active-scrims-count")).toHaveTextContent("1");
    expect(screen.getByTestId("submissions-in-process-count")).toHaveTextContent("1");

    await userEvent.selectOptions(screen.getByLabelText("Filter severity"), "5");

    expect(await screen.findByTestId("operator-inbox-count")).toHaveTextContent("Visible tickets: 1");
    expect(screen.getByRole("heading", { name: "Ticket #2" })).toBeInTheDocument();
  });

  it("submits triage and resolve actions", async () => {
    const tickets: Ticket[] = [
      {
        id: 1,
        category: "scheduling_conflict",
        contextType: "match",
        contextId: 100,
        reportedByTeamId: 10,
        state: "open",
        reasonCode: "availability_conflict",
        severity: 3,
        suggestedAction: "request_reschedule",
        detailsJson: {},
        resolutionCode: null,
        openedAt: "2026-02-15T00:00:00Z",
        triagedAt: null,
        resolvedAt: null,
      },
    ];

    server.use(
      http.get("http://localhost:8080/v1/operator-inbox", () => HttpResponse.json(tickets)),
      http.post("http://localhost:8080/v1/operator-inbox/triage", async ({ request }) => {
        const body = (await request.json()) as {
          reasonCode: string;
          severity: number;
          suggestedAction: string;
        };

        tickets[0] = {
          ...tickets[0],
          reasonCode: body.reasonCode,
          severity: body.severity,
          suggestedAction: body.suggestedAction,
          state: "triaged",
          triagedAt: "2026-02-15T01:00:00Z",
        };

        return HttpResponse.json(tickets[0]);
      }),
      http.post("http://localhost:8080/v1/operator-inbox/resolve", async ({ request }) => {
        const body = (await request.json()) as { resolutionCode: string };

        tickets[0] = {
          ...tickets[0],
          state: "resolved",
          resolutionCode: body.resolutionCode,
          resolvedAt: "2026-02-15T02:00:00Z",
        };

        return HttpResponse.json(tickets[0]);
      }),
    );

    setMockSession({
      subject: "support-1",
      displayName: "Support",
      roles: ["league_support"],
    });

    renderPage();

    expect(await screen.findByRole("heading", { name: "Ticket #1" })).toBeInTheDocument();
    expect(screen.getByTestId("active-scrims-count")).toHaveTextContent("1");

    await userEvent.click(screen.getByRole("button", { name: "Submit triage" }));
    expect(await screen.findByTestId("triage-success")).toBeInTheDocument();
    expect(await screen.findByText("Current state: triaged")).toBeInTheDocument();

    await userEvent.click(within(screen.getByTestId("ticket-detail")).getByRole("button", { name: "Submit resolution" }));
    expect(await screen.findByTestId("resolve-success")).toBeInTheDocument();
    expect(await screen.findByText("Current state: resolved")).toBeInTheDocument();
  });
});
