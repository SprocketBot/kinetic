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

const platformAccountLinks: Array<{
  id: number;
  subject: string;
  playerId: number | null;
  provider: string;
  providerAccountId: string;
  providerAccountName: string;
  isActive: boolean;
  linkedAt: string;
  unlinkedAt: string | null;
}> = [
  {
    id: 1,
    subject: "player-1",
    playerId: 1,
    provider: "steam",
    providerAccountId: "steam-123",
    providerAccountName: "Player One",
    isActive: true,
    linkedAt: "2026-02-15T00:00:00Z",
    unlinkedAt: null,
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

type MockResultSubmission = {
  id: number;
  contextType: string;
  contextId: number;
  gameKey: string;
  submittedByTeamId: number;
  submittedBySubject: string;
  submittedByDisplayName: string;
  homeTeamId: number;
  awayTeamId: number;
  winningTeamId: number;
  losingTeamId: number;
  state: string;
  payloadJson: unknown;
  payloadDigest: string;
  provenanceJson: unknown;
  homeRatifiedAt: string | null;
  homeRatifiedBySubject: string | null;
  homeRatifiedByDisplayName: string | null;
  awayRatifiedAt: string | null;
  awayRatifiedBySubject: string | null;
  awayRatifiedByDisplayName: string | null;
  rejectedByTeamId: number | null;
  rejectionReason: string | null;
  rejectedAt: string | null;
  createdAt: string;
};

const resultSubmissions: MockResultSubmission[] = [
  {
    id: 1,
    contextType: "scrim",
    contextId: 1,
    submittedByTeamId: 101,
    submittedBySubject: "player-1",
    submittedByDisplayName: "Player One",
    gameKey: "rocket_league",
    homeTeamId: 101,
    awayTeamId: 102,
    winningTeamId: 101,
    losingTeamId: 102,
    state: "pending_ratification",
    payloadJson: {},
    payloadDigest: "fixture",
    provenanceJson: {},
    homeRatifiedAt: null,
    homeRatifiedBySubject: "player-1",
    homeRatifiedByDisplayName: "Player One",
    awayRatifiedAt: null,
    awayRatifiedBySubject: null,
    awayRatifiedByDisplayName: null,
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

const roleAssignments: Array<{
  id: number;
  playerId: number;
  role: "fm" | "gm" | "agm" | "captain";
  franchiseId: number | null;
  clubId: number | null;
  teamId: number | null;
  assignedByActorPlayerId: number;
  assignReason: string;
  isActive: boolean;
  assignedAt: string;
  revokedByActorPlayerId: number | null;
  revokeReason: string | null;
  revokedAt: string | null;
}> = [
  {
    id: 1,
    playerId: 1,
    role: "captain",
    franchiseId: null,
    clubId: null,
    teamId: 1,
    assignedByActorPlayerId: 2,
    assignReason: "initial assignment",
    isActive: true,
    assignedAt: "2026-02-15T00:00:00Z",
    revokedByActorPlayerId: null,
    revokeReason: null,
    revokedAt: null,
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
  http.get("http://localhost:8080/v1/me/players", () => HttpResponse.json([{
    userId: 1, playerId: 1, gameId: 1, createdAt: "2026-02-15T00:00:00Z",
    player: { id: 1, displayName: "Player One", slug: "player-one", isActive: true, createdAt: "2026-02-15T00:00:00Z" },
    game: { id: 1, name: "Rocket League", slug: "rocket-league", isActive: true, createdAt: "2026-02-15T00:00:00Z" },
  }])),
  http.get("http://localhost:8080/v1/platform-accounts", ({ request }) => {
    const url = new URL(request.url);
    const playerId = Number(url.searchParams.get("playerId"));
    if (!playerId) {
      return HttpResponse.text("playerId is required", { status: 400 });
    }
    return HttpResponse.json(platformAccountLinks.filter((link) => link.playerId === playerId));
  }),
  http.post("http://localhost:8080/v1/platform-accounts/link", async ({ request }) => {
    const body = (await request.json()) as {
      playerId: number;
      provider: string;
      providerAccountId: string;
      providerAccountName?: string;
    };
    if (!["steam", "xbox", "psn", "epic"].includes(body.provider)) {
      return HttpResponse.text("provider must be steam, xbox, psn, or epic", { status: 400 });
    }

    const existing = platformAccountLinks.find(
      (link) => link.provider === body.provider && link.providerAccountId === body.providerAccountId && link.isActive,
    );
    if (existing) {
      return HttpResponse.text("provider account already linked", { status: 409 });
    }

    const link = {
      id: platformAccountLinks.length + 1,
      subject: "player-1",
      playerId: body.playerId,
      provider: body.provider,
      providerAccountId: body.providerAccountId,
      providerAccountName: body.providerAccountName ?? "",
      isActive: true,
      linkedAt: "2026-02-15T04:10:00Z",
      unlinkedAt: null,
    };
    platformAccountLinks.unshift(link);
    return HttpResponse.json(link, { status: 201 });
  }),
  http.post("http://localhost:8080/v1/platform-accounts/unlink", async ({ request }) => {
    const body = (await request.json()) as {
      playerId: number;
      provider: string;
      providerAccountId: string;
    };
    const existing = platformAccountLinks.find(
      (link) =>
        link.playerId === body.playerId &&
        link.provider === body.provider &&
        link.providerAccountId === body.providerAccountId &&
        link.isActive,
    );
    if (!existing) {
      return HttpResponse.text("active platform account link not found", { status: 409 });
    }

    Object.assign(existing, {
      isActive: false,
      unlinkedAt: "2026-02-15T04:20:00Z",
    });
    return HttpResponse.json(existing);
  }),
  http.get("http://localhost:8080/v1/eligibility", ({ request }) => {
    const url = new URL(request.url);
    const subject = url.searchParams.get("subject");
    if (!subject) {
      return HttpResponse.text("subject is required", { status: 400 });
    }

    return HttpResponse.json({
      subject,
      points: 92,
      thresholdPoints: 40,
      decayPerWeek: 10,
      eligibleUntil: "2026-03-22T04:30:00Z",
      evaluatedAt: "2026-02-15T04:30:00Z",
      projection: [
        { effectiveAt: "2026-02-15T04:30:00Z", points: 92, isEligible: true },
        { effectiveAt: "2026-02-22T04:30:00Z", points: 82, isEligible: true },
        { effectiveAt: "2026-03-01T04:30:00Z", points: 72, isEligible: true },
        { effectiveAt: "2026-03-08T04:30:00Z", points: 62, isEligible: true },
        { effectiveAt: "2026-03-15T04:30:00Z", points: 52, isEligible: true },
      ],
    });
  }),
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
    const existing = queueEntries.find((entry) => entry.queueId === body.queueId && entry.teamId === body.teamId && entry.isActive);
    if (existing) {
      return HttpResponse.text("conflict: uq_queue_entries_active_queue_team", { status: 409 });
    }

    const rosterPlayers = rosterMemberships
      .filter((membership) => membership.teamId === body.teamId && membership.isActive)
      .map((membership) => membership.playerId);
    const blockedBan = queueBans.find(
      (ban) => ban.queueId === body.queueId && ban.isActive && rosterPlayers.includes(ban.playerId),
    );
    if (blockedBan) {
      return HttpResponse.text(`team has player ${blockedBan.playerId} actively banned from this queue`, { status: 409 });
    }

    const entry = {
      id: queueEntries.length + 1,
      queueId: body.queueId,
      teamId: body.teamId,
      isActive: true,
      stage: 1,
      createdAt: "2026-02-15T00:00:00Z",
      stageAt: "2026-02-15T00:00:00Z",
      leftAt: null,
    };
    queueEntries.unshift(entry);
    return HttpResponse.json(entry);
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
  http.post("http://localhost:8080/v1/result-submissions", async ({ request }) => {
    const body = (await request.json()) as {
      contextType: string;
      contextId: number;
      gameKey?: string;
      submittedByTeamId: number;
      winningTeamId: number;
      losingTeamId: number;
      payloadJson: unknown;
      provenanceJson?: unknown;
    };
    const scrim = scrims.find((item) => item.id === body.contextId);
    const submission = {
      id: resultSubmissions.length + 1,
      contextType: body.contextType,
      contextId: body.contextId,
      gameKey: body.gameKey ?? "rocket_league",
      submittedByTeamId: body.submittedByTeamId,
      submittedBySubject: "player-1",
      submittedByDisplayName: "Player One",
      homeTeamId: scrim?.homeTeamId ?? 101,
      awayTeamId: scrim?.awayTeamId ?? 102,
      winningTeamId: body.winningTeamId,
      losingTeamId: body.losingTeamId,
      state: "pending_ratification",
      payloadJson: body.payloadJson,
      payloadDigest: "mock-digest",
      provenanceJson: body.provenanceJson ?? {},
      homeRatifiedAt: body.submittedByTeamId === (scrim?.homeTeamId ?? 101) ? "2026-02-15T01:00:00Z" : null,
      homeRatifiedBySubject: body.submittedByTeamId === (scrim?.homeTeamId ?? 101) ? "player-1" : null,
      homeRatifiedByDisplayName: body.submittedByTeamId === (scrim?.homeTeamId ?? 101) ? "Player One" : null,
      awayRatifiedAt: body.submittedByTeamId === (scrim?.awayTeamId ?? 102) ? "2026-02-15T01:00:00Z" : null,
      awayRatifiedBySubject: body.submittedByTeamId === (scrim?.awayTeamId ?? 102) ? "player-1" : null,
      awayRatifiedByDisplayName: body.submittedByTeamId === (scrim?.awayTeamId ?? 102) ? "Player One" : null,
      rejectedByTeamId: null,
      rejectionReason: null,
      rejectedAt: null,
      createdAt: "2026-02-15T01:00:00Z",
    };
    resultSubmissions.unshift(submission);
    return HttpResponse.json(submission, { status: 201 });
  }),
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
  http.get("http://localhost:8080/v1/role-assignments", () => HttpResponse.json(roleAssignments)),
  http.post("http://localhost:8080/v1/role-assignments", async ({ request }) => {
    const body = (await request.json()) as {
      actorPlayerId: number;
      targetPlayerId: number;
      role: "fm" | "gm" | "agm" | "captain";
      franchiseId?: number;
      clubId?: number;
      teamId?: number;
      reason: string;
    };

    if (body.role === "fm" && !body.franchiseId) {
      return HttpResponse.text("franchiseId required for fm role", { status: 400 });
    }
    if ((body.role === "gm" || body.role === "agm") && !body.clubId) {
      return HttpResponse.text("clubId required for gm/agm role", { status: 400 });
    }
    if (body.role === "captain" && !body.teamId) {
      return HttpResponse.text("teamId required for captain role", { status: 400 });
    }

    const duplicate = roleAssignments.find(
      (assignment) =>
        assignment.isActive &&
        assignment.playerId === body.targetPlayerId &&
        assignment.role === body.role &&
        (assignment.franchiseId ?? null) === (body.franchiseId ?? null) &&
        (assignment.clubId ?? null) === (body.clubId ?? null) &&
        (assignment.teamId ?? null) === (body.teamId ?? null),
    );
    if (duplicate) {
      return HttpResponse.text("role assignment already active", { status: 409 });
    }

    const assignment = {
      id: roleAssignments.length + 1,
      playerId: body.targetPlayerId,
      role: body.role,
      franchiseId: body.role === "fm" ? (body.franchiseId ?? null) : null,
      clubId: body.role === "gm" || body.role === "agm" ? (body.clubId ?? null) : null,
      teamId: body.role === "captain" ? (body.teamId ?? null) : null,
      assignedByActorPlayerId: body.actorPlayerId,
      assignReason: body.reason,
      isActive: true,
      assignedAt: "2026-02-15T03:10:00Z",
      revokedByActorPlayerId: null,
      revokeReason: null,
      revokedAt: null,
    };
    roleAssignments.unshift(assignment);
    return HttpResponse.json(assignment, { status: 201 });
  }),
  http.post("http://localhost:8080/v1/role-assignments/revoke", async ({ request }) => {
    const body = (await request.json()) as { actorPlayerId: number; assignmentId: number; reason: string };
    const assignment = roleAssignments.find((entry) => entry.id === body.assignmentId && entry.isActive);
    if (!assignment) {
      return HttpResponse.text("active role assignment not found", { status: 409 });
    }

    Object.assign(assignment, {
      isActive: false,
      revokedByActorPlayerId: body.actorPlayerId,
      revokeReason: body.reason,
      revokedAt: "2026-02-15T03:20:00Z",
    });
    return HttpResponse.json(assignment);
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
      lastCompetedAt: existing?.lastCompetedAt ?? "2026-02-14T12:00:00Z",
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
      autofillJson: { score: { home: 3, away: 1 } },
      conflictJson: null,
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
