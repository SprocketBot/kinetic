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

const queueBans: Array<{
  id: number;
  queueId: number;
  playerId: number;
  bannedByActor: string;
  banReason: string;
  isActive: boolean;
  bannedAt: string;
  unbannedByActor: string | null;
  unbanReason: string | null;
  unbannedAt: string | null;
}> = [];

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

const resultOverrides = [
  {
    id: 1,
    submissionId: 1,
    actor: "league-admin",
    reason: "video review correction",
    previousWinningTeamId: 102,
    previousLosingTeamId: 101,
    newWinningTeamId: 101,
    newLosingTeamId: 102,
    previousState: "pending_ratification",
    newState: "ratified",
    createdAt: "2026-02-15T00:20:00Z",
  },
];

const seasons = [
  {
    id: 1,
    name: "Season 1",
    slug: "season-1",
    isActive: true,
    createdAt: "2026-02-15T00:00:00Z",
  },
];

const scheduleGroups = [
  {
    id: 1,
    seasonId: 1,
    name: "Week 1",
    sequence: 1,
    isActive: true,
    createdAt: "2026-02-15T00:00:00Z",
  },
];

const fixtures = [
  {
    id: 1,
    scheduleGroupId: 1,
    homeClubId: 11,
    awayClubId: 12,
    isActive: true,
    createdAt: "2026-02-15T00:00:00Z",
  },
];

const matches = [
  {
    id: 1,
    fixtureId: 1,
    homeTeamId: 101,
    awayTeamId: 102,
    state: "scheduled",
    scheduledFor: "2026-02-16T20:00:00Z",
    homeTimeRatifiedAt: null,
    awayTimeRatifiedAt: null,
    createdAt: "2026-02-15T00:00:00Z",
  },
];

const teams = [
  {
    id: 1,
    clubId: 11,
    name: "Team A",
    slug: "team-a",
    isActive: true,
    createdAt: "2026-02-15T00:00:00Z",
  },
];

const players = [
  {
    id: 1,
    displayName: "Player One",
    slug: "player-one",
    isActive: true,
    createdAt: "2026-02-15T00:00:00Z",
  },
];

const rosterMemberships = [
  {
    id: 1,
    playerId: 1,
    teamId: 1,
    isActive: true,
    createdAt: "2026-02-15T00:00:00Z",
  },
];

const exceptionActions = [
  {
    id: 1,
    ticketId: 1,
    actionType: "triaged",
    actor: "support-1",
    automated: false,
    notes: "reviewed",
    minutesSpent: 5,
    createdAt: "2026-02-15T00:30:00Z",
  },
];

const playerRatings = [
  {
    id: 1,
    playerId: 1,
    contextKey: "default",
    rating: 1000,
    uncertainty: 120,
    matchesPlayed: 25,
    lastCompetedAt: "2026-02-14T12:00:00Z",
    isActive: true,
    updatedAt: "2026-02-15T00:00:00Z",
  },
];

const ratingAdjustments = [
  {
    id: 1,
    actorPlayerId: 2,
    targetPlayerId: 1,
    contextKey: "default",
    previousRating: 980,
    newRating: 1000,
    previousUncertainty: 130,
    newUncertainty: 120,
    previousMatchesPlayed: 24,
    newMatchesPlayed: 25,
    reason: "manual review",
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
  http.get("http://localhost:8080/v1/queue-bans", () => HttpResponse.json(queueBans)),
  http.post("http://localhost:8080/v1/queue-bans", async ({ request }) => {
    const body = (await request.json()) as { queueId: number; playerId: number; actor: string; reason: string };
    const existing = queueBans.find((ban) => ban.queueId === body.queueId && ban.playerId === body.playerId && ban.isActive);
    if (existing) {
      return HttpResponse.text("player already banned for queue", { status: 409 });
    }

    const ban = {
      id: queueBans.length + 1,
      queueId: body.queueId,
      playerId: body.playerId,
      bannedByActor: body.actor,
      banReason: body.reason,
      isActive: true,
      bannedAt: "2026-02-15T04:00:00Z",
      unbannedByActor: null,
      unbanReason: null,
      unbannedAt: null,
    };
    queueBans.unshift(ban);
    return HttpResponse.json(ban, { status: 201 });
  }),
  http.post("http://localhost:8080/v1/queue-bans/lift", async ({ request }) => {
    const body = (await request.json()) as { queueId: number; playerId: number; actor: string; reason: string };
    const existing = queueBans.find((ban) => ban.queueId === body.queueId && ban.playerId === body.playerId && ban.isActive);
    if (!existing) {
      return HttpResponse.text("active queue ban not found", { status: 409 });
    }

    Object.assign(existing, {
      isActive: false,
      unbannedByActor: body.actor,
      unbanReason: body.reason,
      unbannedAt: "2026-02-15T04:30:00Z",
    });
    return HttpResponse.json(existing);
  }),
  http.get("http://localhost:8080/v1/queue-entries", () => HttpResponse.json(queueEntries)),
  http.post("http://localhost:8080/v1/queue-entries", async ({ request }) => {
    const body = (await request.json()) as { queueId: number; teamId: number };

    const rosterPlayers = rosterMemberships
      .filter((membership) => membership.teamId === body.teamId && membership.isActive)
      .map((membership) => membership.playerId);
    const blockedBan = queueBans.find(
      (ban) => ban.queueId === body.queueId && ban.isActive && rosterPlayers.includes(ban.playerId),
    );
    if (blockedBan) {
      return HttpResponse.text(`team has player ${blockedBan.playerId} actively banned from this queue`, { status: 409 });
    }

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
  http.get("http://localhost:8080/v1/result-overrides", () => HttpResponse.json(resultOverrides)),
  http.post("http://localhost:8080/v1/result-overrides", async ({ request }) => {
    const body = (await request.json()) as {
      submissionId: number;
      actor: string;
      reason: string;
      winningTeamId: number;
      losingTeamId: number;
    };

    const submission = resultSubmissions.find((item) => item.id === body.submissionId);
    if (!submission) {
      return HttpResponse.text("submission not found", { status: 409 });
    }
    if (
      ![submission.homeTeamId, submission.awayTeamId].includes(body.winningTeamId) ||
      ![submission.homeTeamId, submission.awayTeamId].includes(body.losingTeamId)
    ) {
      return HttpResponse.text("winning and losing teams must match submission participants", { status: 409 });
    }

    resultOverrides.unshift({
      id: resultOverrides.length + 1,
      submissionId: submission.id,
      actor: body.actor,
      reason: body.reason,
      previousWinningTeamId: submission.winningTeamId,
      previousLosingTeamId: submission.losingTeamId,
      newWinningTeamId: body.winningTeamId,
      newLosingTeamId: body.losingTeamId,
      previousState: submission.state,
      newState: "ratified",
      createdAt: "2026-02-15T03:00:00Z",
    });

    Object.assign(submission, {
      winningTeamId: body.winningTeamId,
      losingTeamId: body.losingTeamId,
      state: "ratified",
      homeRatifiedAt: submission.homeRatifiedAt ?? "2026-02-15T03:00:00Z",
      awayRatifiedAt: submission.awayRatifiedAt ?? "2026-02-15T03:00:00Z",
      rejectedByTeamId: null,
      rejectionReason: null,
      rejectedAt: null,
    });

    return HttpResponse.json(submission);
  }),
  http.get("http://localhost:8080/v1/teams", () => HttpResponse.json(teams)),
  http.get("http://localhost:8080/v1/players", () => HttpResponse.json(players)),
  http.get("http://localhost:8080/v1/roster-memberships", () => HttpResponse.json(rosterMemberships)),
  http.post("http://localhost:8080/v1/roster-memberships", async ({ request }) => {
    const body = (await request.json()) as { playerId: number; teamId: number };
    return HttpResponse.json({
      id: rosterMemberships.length + 1,
      playerId: body.playerId,
      teamId: body.teamId,
      isActive: true,
      createdAt: "2026-02-15T03:00:00Z",
    });
  }),
  http.get("http://localhost:8080/v1/exception-actions", () => HttpResponse.json(exceptionActions)),
  http.get("http://localhost:8080/v1/player-ratings", () => HttpResponse.json(playerRatings)),
  http.get("http://localhost:8080/v1/rating-adjustments", () => HttpResponse.json(ratingAdjustments)),
  http.post("http://localhost:8080/v1/player-ratings/adjust", async ({ request }) => {
    const body = (await request.json()) as {
      actorPlayerId: number;
      targetPlayerId: number;
      contextKey: string;
      rating: number;
      uncertainty: number;
      matchesPlayed: number;
      reason: string;
    };

    if (body.actorPlayerId === body.targetPlayerId) {
      return HttpResponse.text("actorPlayerId cannot adjust own rating", { status: 409 });
    }

    const existing = playerRatings.find(
      (rating) => rating.playerId === body.targetPlayerId && rating.contextKey === body.contextKey,
    );
    const previousRating = existing?.rating ?? 1000;
    const previousUncertainty = existing?.uncertainty ?? 350;
    const previousMatchesPlayed = existing?.matchesPlayed ?? 0;

    const updated = {
      id: existing?.id ?? playerRatings.length + 1,
      playerId: body.targetPlayerId,
      contextKey: body.contextKey,
      rating: body.rating,
      uncertainty: body.uncertainty,
      matchesPlayed: body.matchesPlayed,
      lastCompetedAt: existing?.lastCompetedAt ?? null,
      isActive: true,
      updatedAt: "2026-02-15T03:30:00Z",
    };

    if (existing) {
      Object.assign(existing, updated);
    } else {
      playerRatings.push(updated);
    }

    ratingAdjustments.unshift({
      id: ratingAdjustments.length + 1,
      actorPlayerId: body.actorPlayerId,
      targetPlayerId: body.targetPlayerId,
      contextKey: body.contextKey,
      previousRating,
      newRating: body.rating,
      previousUncertainty,
      newUncertainty: body.uncertainty,
      previousMatchesPlayed,
      newMatchesPlayed: body.matchesPlayed,
      reason: body.reason,
      createdAt: "2026-02-15T03:30:00Z",
    });

    return HttpResponse.json(updated);
  }),
  http.get("http://localhost:8080/v1/seasons", () => HttpResponse.json(seasons)),
  http.post("http://localhost:8080/v1/seasons", async ({ request }) => {
    const body = (await request.json()) as { name: string; slug: string };
    return HttpResponse.json({
      id: seasons.length + 1,
      name: body.name,
      slug: body.slug,
      isActive: true,
      createdAt: "2026-02-15T03:00:00Z",
    });
  }),
  http.get("http://localhost:8080/v1/schedule-groups", () => HttpResponse.json(scheduleGroups)),
  http.post("http://localhost:8080/v1/schedule-groups", async ({ request }) => {
    const body = (await request.json()) as { seasonId: number; name: string; sequence: number };
    return HttpResponse.json({
      id: scheduleGroups.length + 1,
      seasonId: body.seasonId,
      name: body.name,
      sequence: body.sequence,
      isActive: true,
      createdAt: "2026-02-15T03:00:00Z",
    });
  }),
  http.get("http://localhost:8080/v1/fixtures", () => HttpResponse.json(fixtures)),
  http.post("http://localhost:8080/v1/fixtures", async ({ request }) => {
    const body = (await request.json()) as { scheduleGroupId: number; homeClubId: number; awayClubId: number };
    return HttpResponse.json({
      id: fixtures.length + 1,
      scheduleGroupId: body.scheduleGroupId,
      homeClubId: body.homeClubId,
      awayClubId: body.awayClubId,
      isActive: true,
      createdAt: "2026-02-15T03:00:00Z",
    });
  }),
  http.get("http://localhost:8080/v1/matches", () => HttpResponse.json(matches)),
  http.post("http://localhost:8080/v1/matches", async ({ request }) => {
    const body = (await request.json()) as {
      fixtureId: number;
      homeTeamId: number;
      awayTeamId: number;
      state: string;
      scheduledFor?: string;
    };
    return HttpResponse.json({
      id: matches.length + 1,
      fixtureId: body.fixtureId,
      homeTeamId: body.homeTeamId,
      awayTeamId: body.awayTeamId,
      state: body.state,
      scheduledFor: body.scheduledFor ?? null,
      homeTimeRatifiedAt: null,
      awayTimeRatifiedAt: null,
      createdAt: "2026-02-15T03:00:00Z",
    });
  }),

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
