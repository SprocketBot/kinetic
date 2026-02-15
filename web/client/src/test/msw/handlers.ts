import { http, HttpResponse } from "msw";

const queueEntries = [
  {
    id: 1,
    queueId: 1,
    teamId: 101,
    isActive: true,
    stage: 1,
    createdAt: "2026-02-15T00:00:00Z",
    stageAt: "2026-02-15T00:00:00Z",
    leftAt: null,
  },
];

const scrims = [
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
];

const resultSubmissions = [
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
];

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
  http.get("http://localhost:8080/v1/queue-entries", () => HttpResponse.json(queueEntries)),
  http.post("http://localhost:8080/v1/queue-entries", async ({ request }) => {
    const body = (await request.json()) as { queueId: number; teamId: number };
    return HttpResponse.json({
      id: queueEntries.length + 1,
      queueId: body.queueId,
      teamId: body.teamId,
      isActive: true,
      stage: 1,
      createdAt: "2026-02-15T00:00:00Z",
      stageAt: "2026-02-15T00:00:00Z",
      leftAt: null,
    });
  }),
  http.delete("http://localhost:8080/v1/queue-entries", async ({ request }) => {
    const body = (await request.json()) as { queueId: number; teamId: number };
    const match = queueEntries.find((entry) => entry.queueId === body.queueId && entry.teamId === body.teamId);
    if (!match) {
      return HttpResponse.json({ code: "not_found", message: "queue entry not found" }, { status: 404 });
    }

    return HttpResponse.json({
      ...match,
      isActive: false,
      leftAt: "2026-02-15T01:00:00Z",
    });
  }),

  http.get("http://localhost:8080/v1/scrims", () => HttpResponse.json(scrims)),
  http.get("http://localhost:8080/v1/result-submissions", () => HttpResponse.json(resultSubmissions)),

  http.post("http://localhost:8080/v1/result-submission-ratifications", async ({ request }) => {
    const body = (await request.json()) as { submissionId: number; teamId: number };
    const match = resultSubmissions.find((submission) => submission.id === body.submissionId);
    if (!match) {
      return HttpResponse.json({ code: "not_found", message: "submission not found" }, { status: 404 });
    }

    return HttpResponse.json({
      ...match,
      homeRatifiedAt: match.homeTeamId === body.teamId ? "2026-02-15T01:00:00Z" : match.homeRatifiedAt,
      awayRatifiedAt: match.awayTeamId === body.teamId ? "2026-02-15T01:00:00Z" : match.awayRatifiedAt,
    });
  }),
  http.post("http://localhost:8080/v1/result-submission-rejections", async ({ request }) => {
    const body = (await request.json()) as { submissionId: number; teamId: number; reason: string };
    const match = resultSubmissions.find((submission) => submission.id === body.submissionId);
    if (!match) {
      return HttpResponse.json({ code: "not_found", message: "submission not found" }, { status: 404 });
    }

    return HttpResponse.json({
      ...match,
      state: "rejected",
      rejectedByTeamId: body.teamId,
      rejectionReason: body.reason,
      rejectedAt: "2026-02-15T01:30:00Z",
    });
  }),

  http.post("http://localhost:8080/v1/replay-evidence", async ({ request }) => {
    const body = (await request.json()) as {
      contextType: string;
      contextId: number;
      submittedByTeamId: number;
      parserName: string;
      parserVersion: string;
      parserConfigDigest: string;
      resultSubmissionId?: number;
    };

    return HttpResponse.json({
      evidence: {
        id: 10,
        contextType: body.contextType,
        contextId: body.contextId,
        submittedByTeamId: body.submittedByTeamId,
        replaySha256: "sha256-demo",
        contentSizeBytes: 12,
        storageRef: "memory://replay/10",
        state: "parsed",
        createdAt: "2026-02-15T02:00:00Z",
      },
      parseRun: {
        id: 99,
        replayEvidenceId: 10,
        parserName: body.parserName,
        parserVersion: body.parserVersion,
        parserConfigDigest: body.parserConfigDigest,
        status: "parsed",
        outputJson: {},
        createdAt: "2026-02-15T02:00:00Z",
      },
      duplicate: false,
      linkedSubmissionId: body.resultSubmissionId ?? null,
    });
  }),

  http.get("http://localhost:8080/v1/operator-inbox", () => {
    return HttpResponse.json([baseTicket]);
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
