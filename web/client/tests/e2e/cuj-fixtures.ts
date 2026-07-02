import type { Page, Route } from "@playwright/test";

type Role = "player" | "league_support" | "league_admin" | "platform_operator";

export async function signInAs(page: Page, role: Role) {
  const principal = {
    subject: `local-${role}`,
    displayName: `CUJ ${role.replace("_", " ")}`,
    roles: [role],
  };
  await page.addInitScript((session) => {
    localStorage.setItem("kinetic.mockSession", JSON.stringify(session));
  }, principal);
  await page.evaluate((session) => {
    localStorage.setItem("kinetic.mockSession", JSON.stringify(session));
  }, principal).catch(() => {});
}

function json(route: Route, body: unknown, status = 200) {
  return route.fulfill({
    status,
    contentType: "application/json",
    body: JSON.stringify(body),
  });
}

export async function installCUJApi(page: Page) {
  let nextId = 100;
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
  const queueBans: Array<Record<string, unknown>> = [];
  const scrims = [
    {
      id: 9,
      queueId: 1,
      homeTeamId: 100,
      awayTeamId: 101,
      state: "started",
      createdAt: "2026-02-15T00:00:00Z",
      startedAt: "2026-02-15T00:03:00Z",
      endedAt: null,
    },
  ];
  const submissions = [
    {
      id: 22,
      contextType: "scrim",
      contextId: 9,
      submittedByTeamId: 100,
      homeTeamId: 100,
      awayTeamId: 101,
      winningTeamId: 100,
      losingTeamId: 101,
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
  const roleAssignments = [
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
  const ratings = [
    {
      id: 1,
      playerId: 1,
      contextKey: "scrim-3v3",
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
      contextKey: "scrim-3v3",
      previousRating: 980,
      newRating: 1000,
      previousUncertainty: 130,
      newUncertainty: 120,
      previousMatchesPlayed: 24,
      newMatchesPlayed: 25,
      reason: "scrim_result",
      createdAt: "2026-02-15T00:10:00Z",
    },
  ];
  const tickets = [
    {
      id: 11,
      category: "scheduling_conflict",
      contextType: "match",
      contextId: 42,
      reportedByTeamId: 7,
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

  await page.route("**/v1/platform-accounts?**", async (route) =>
    json(route, [
      {
        id: 1,
        subject: "local-player",
        provider: "steam",
        providerAccountId: "steam-123",
        providerAccountName: "Player One",
        isActive: true,
        linkedAt: "2026-02-15T00:00:00Z",
        unlinkedAt: null,
      },
    ]),
  );
  await page.route("**/v1/platform-accounts/link", async (route) => {
    const body = route.request().postDataJSON();
    return json(route, {
      id: nextId++,
      ...body,
      isActive: true,
      linkedAt: "2026-02-15T04:10:00Z",
      unlinkedAt: null,
    });
  });
  await page.route("**/v1/platform-accounts/unlink", async (route) => json(route, { id: 1, isActive: false }));
  await page.route("**/v1/eligibility?**", async (route) =>
    json(route, {
      subject: "local-player",
      points: 92,
      thresholdPoints: 40,
      decayPerWeek: 10,
      eligibleUntil: "2026-03-22T04:30:00Z",
      evaluatedAt: "2026-02-15T04:30:00Z",
      projection: [{ effectiveAt: "2026-02-15T04:30:00Z", points: 92, isEligible: true }],
    }),
  );
  await page.route("**/v1/queue-entries", async (route) => {
    if (route.request().method() === "GET") return json(route, queueEntries);
    if (route.request().method() === "DELETE") {
      queueEntries[0] = { ...queueEntries[0], isActive: false, leftAt: "2026-02-15T01:00:00Z" };
      return json(route, queueEntries[0]);
    }
    const body = route.request().postDataJSON();
    const entry = { id: nextId++, isActive: true, stage: 1, createdAt: "2026-02-15T00:00:00Z", stageAt: "2026-02-15T00:00:00Z", leftAt: null, ...body };
    queueEntries.unshift(entry);
    return json(route, entry, 201);
  });
  await page.route("**/v1/scrims", async (route) => json(route, scrims));
  await page.route("**/v1/result-submissions", async (route) => json(route, submissions));
  await page.route("**/v1/result-submission-ratifications", async (route) => {
    const body = route.request().postDataJSON();
    submissions[0] = { ...submissions[0], homeRatifiedAt: body.teamId === 100 ? "2026-02-15T01:00:00Z" : submissions[0].homeRatifiedAt };
    return json(route, submissions[0]);
  });
  await page.route("**/v1/result-submission-rejections", async (route) => {
    submissions[0] = { ...submissions[0], state: "rejected", rejectedAt: "2026-02-15T01:30:00Z" };
    return json(route, submissions[0]);
  });
  await page.route("**/v1/replay-evidence", async (route) =>
    json(route, {
      evidence: {
        id: nextId++,
        contextType: "scrim",
        contextId: 9,
        submittedByTeamId: 100,
        replaySha256: "sha256-demo",
        contentSizeBytes: 12,
        storageRef: "memory://replay",
        state: "parsed",
        createdAt: "2026-02-15T02:00:00Z",
      },
      parseRun: {
        id: nextId++,
        replayEvidenceId: nextId++,
        parserName: "kinetic-rl-parser",
        parserVersion: "v1",
        parserConfigDigest: "default",
        status: "parsed",
        outputJson: {},
        createdAt: "2026-02-15T02:00:00Z",
      },
      duplicate: false,
      linkedSubmissionId: 22,
    }),
  );

  await page.route("**/v1/seasons", async (route) => {
    if (route.request().method() === "GET") return json(route, [{ id: 1, name: "Season 1", slug: "season-1", isActive: true, createdAt: "2026-02-15T00:00:00Z" }]);
    return json(route, { id: nextId++, name: "Season 2", slug: "season-2", isActive: true, createdAt: "2026-02-15T03:00:00Z" }, 201);
  });
  await page.route("**/v1/schedule-groups", async (route) => {
    if (route.request().method() === "GET") return json(route, [{ id: 1, seasonId: 1, name: "Week 1", sequence: 1, isActive: true, createdAt: "2026-02-15T00:00:00Z" }]);
    return json(route, { id: nextId++, seasonId: 1, name: "Week 2", sequence: 2, isActive: true, createdAt: "2026-02-15T03:00:00Z" }, 201);
  });
  await page.route("**/v1/fixtures", async (route) => {
    if (route.request().method() === "GET") return json(route, [{ id: 1, scheduleGroupId: 1, homeClubId: 11, awayClubId: 12, isActive: true, createdAt: "2026-02-15T00:00:00Z" }]);
    return json(route, { id: nextId++, scheduleGroupId: 1, homeClubId: 11, awayClubId: 12, isActive: true, createdAt: "2026-02-15T03:00:00Z" }, 201);
  });
  await page.route("**/v1/matches", async (route) => {
    if (route.request().method() === "GET") return json(route, [{ id: 1, fixtureId: 1, homeTeamId: 101, awayTeamId: 102, state: "scheduled", scheduledFor: "2026-02-16T20:00:00Z", homeTimeRatifiedAt: null, awayTimeRatifiedAt: null, createdAt: "2026-02-15T00:00:00Z" }]);
    return json(route, { id: nextId++, fixtureId: 1, homeTeamId: 101, awayTeamId: 102, state: "scheduled", scheduledFor: "2026-02-16T20:00:00Z", homeTimeRatifiedAt: null, awayTimeRatifiedAt: null, createdAt: "2026-02-15T03:00:00Z" }, 201);
  });
  await page.route("**/v1/teams", async (route) => json(route, [{ id: 1, clubId: 11, name: "Team A", slug: "team-a", isActive: true, createdAt: "2026-02-15T00:00:00Z" }]));
  await page.route("**/v1/players", async (route) => json(route, [{ id: 1, displayName: "Player One", slug: "player-one", isActive: true, createdAt: "2026-02-15T00:00:00Z" }]));
  await page.route("**/v1/roster-memberships", async (route) => {
    if (route.request().method() === "GET") return json(route, [{ id: 1, playerId: 1, teamId: 1, isActive: true, createdAt: "2026-02-15T00:00:00Z" }]);
    return json(route, { id: nextId++, playerId: 1, teamId: 1, isActive: true, createdAt: "2026-02-15T03:00:00Z" }, 201);
  });
  await page.route("**/v1/role-assignments", async (route) => {
    if (route.request().method() === "GET") return json(route, roleAssignments);
    const body = route.request().postDataJSON();
    const assignment = { id: nextId++, playerId: body.targetPlayerId, role: body.role, franchiseId: null, clubId: null, teamId: body.teamId ?? null, assignedByActorPlayerId: body.actorPlayerId, assignReason: body.reason, isActive: true, assignedAt: "2026-02-15T03:10:00Z", revokedByActorPlayerId: null, revokeReason: null, revokedAt: null };
    roleAssignments.unshift(assignment);
    return json(route, assignment, 201);
  });
  await page.route("**/v1/role-assignments/revoke", async (route) => {
    roleAssignments[0] = { ...roleAssignments[0], isActive: false, revokedAt: "2026-02-15T03:20:00Z" };
    return json(route, roleAssignments[0]);
  });
  await page.route("**/v1/exception-actions", async (route) => json(route, [{ id: 1, ticketId: 11, actionType: "triaged", actor: "support-1", automated: false, notes: "reviewed", minutesSpent: 5, createdAt: "2026-02-15T00:30:00Z" }]));
  await page.route("**/v1/player-ratings", async (route) => json(route, ratings));
  await page.route("**/v1/rating-adjustments", async (route) => json(route, ratingAdjustments));
  await page.route("**/v1/player-ratings/adjust", async (route) => {
    const body = route.request().postDataJSON();
    const rating = { id: nextId++, playerId: body.targetPlayerId, contextKey: body.contextKey, rating: body.rating, uncertainty: body.uncertainty, matchesPlayed: body.matchesPlayed, lastCompetedAt: "2026-02-14T12:00:00Z", isActive: true, updatedAt: "2026-02-15T03:30:00Z" };
    ratings.unshift(rating);
    return json(route, rating);
  });
  await page.route("**/v1/result-overrides", async (route) => {
    if (route.request().method() === "GET") return json(route, []);
    return json(route, { ...submissions[0], state: "ratified" });
  });
  await page.route("**/v1/operator-inbox", async (route) => json(route, tickets));
  await page.route("**/v1/operator-inbox/triage", async (route) => {
    tickets[0] = { ...tickets[0], state: "triaged", triagedAt: "2026-02-15T01:00:00Z" };
    return json(route, tickets[0]);
  });
  await page.route("**/v1/operator-inbox/resolve", async (route) => {
    tickets[0] = { ...tickets[0], state: "resolved", resolutionCode: "resolved_manual", resolvedAt: "2026-02-15T02:00:00Z" };
    return json(route, tickets[0]);
  });
  await page.route("**/v1/queue-bans", async (route) => {
    if (route.request().method() === "GET") return json(route, queueBans);
    const body = route.request().postDataJSON();
    const ban = { id: nextId++, queueId: body.queueId, playerId: body.playerId, bannedByActor: body.actor, banReason: body.reason, isActive: true, bannedAt: "2026-02-15T04:00:00Z", unbannedByActor: null, unbanReason: null, unbannedAt: null };
    queueBans.unshift(ban);
    return json(route, ban, 201);
  });
  await page.route("**/v1/queue-bans/lift", async (route) => {
    queueBans[0] = { ...queueBans[0], isActive: false, unbannedAt: "2026-02-15T04:30:00Z" };
    return json(route, queueBans[0]);
  });
  await page.route("**/v1/exception-metrics", async (route) => json(route, { adminHoursPerWeek: 6.5, manualTouchesPerFixture: 1.4, zeroTouchFixtureRate: 0.62, timeToCloseHoursP50: 9.5 }));
}
