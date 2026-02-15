import { http, HttpResponse } from "msw";

const baseTicket = {
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
};

export const handlers = [
  http.get("http://localhost:8080/v1/operator-inbox", () => {
    return HttpResponse.json([baseTicket]);
  }),
  http.get("http://localhost:8080/v1/scrims", () => {
    return HttpResponse.json([
      {
        id: 1,
        queueId: 1,
        homeTeamId: 101,
        awayTeamId: 102,
        state: "started",
        createdAt: "2026-02-15T00:00:00Z",
        startedAt: "2026-02-15T00:03:00Z",
        endedAt: null,
      },
    ]);
  }),
  http.get("http://localhost:8080/v1/result-submissions", () => {
    return HttpResponse.json([
      {
        id: 1,
        contextType: "scrim",
        contextId: 1,
        submittedByTeamId: 101,
        homeTeamId: 101,
        awayTeamId: 102,
        winningTeamId: 101,
        losingTeamId: 102,
        state: "pending_ratification",
        payloadJson: {},
        homeRatifiedAt: null,
        awayRatifiedAt: null,
        rejectedByTeamId: null,
        rejectionReason: null,
        rejectedAt: null,
        createdAt: "2026-02-15T00:10:00Z",
      },
    ]);
  }),
  http.get("http://localhost:8080/v1/exception-metrics", () => {
    return HttpResponse.json({
      adminHoursPerWeek: 6.5,
      manualTouchesPerFixture: 1.4,
      zeroTouchFixtureRate: 0.62,
      timeToCloseHoursP50: 9.5,
    });
  }),
  http.post("http://localhost:8080/v1/operator-inbox/triage", async ({ request }) => {
    const body = (await request.json()) as {
      reasonCode: string;
      severity: number;
      suggestedAction: string;
    };

    return HttpResponse.json({
      ...baseTicket,
      reasonCode: body.reasonCode,
      severity: body.severity,
      suggestedAction: body.suggestedAction,
      state: "triaged",
      triagedAt: "2026-02-15T01:00:00Z",
    });
  }),
  http.post("http://localhost:8080/v1/operator-inbox/resolve", async ({ request }) => {
    const body = (await request.json()) as { resolutionCode: string };

    return HttpResponse.json({
      ...baseTicket,
      state: "resolved",
      resolutionCode: body.resolutionCode,
      resolvedAt: "2026-02-15T02:00:00Z",
    });
  }),
];
