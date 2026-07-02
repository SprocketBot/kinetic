import { expect, test } from "@playwright/test";

async function mockPlayerEndpoints(page: import("@playwright/test").Page) {
  await page.route("**/v1/queue-entries", async (route) => {
    if (route.request().method() === "GET") {
      await route.fulfill({
        contentType: "application/json",
        body: JSON.stringify([
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
        ]),
      });
      return;
    }

    await route.fulfill({
      contentType: "application/json",
      body: JSON.stringify({
        id: 2,
        queueId: 1,
        teamId: 101,
        isActive: route.request().method() !== "DELETE",
        stage: 1,
        createdAt: "2026-02-15T00:00:00Z",
        stageAt: "2026-02-15T00:00:00Z",
        leftAt: route.request().method() === "DELETE" ? "2026-02-15T01:00:00Z" : null,
      }),
    });
  });

  await page.route("**/v1/scrims", async (route) => {
    await route.fulfill({
      contentType: "application/json",
      body: JSON.stringify([
        {
          id: 9,
          queueId: 3,
          homeTeamId: 100,
          awayTeamId: 101,
          state: "started",
          createdAt: "2026-02-15T00:00:00Z",
          startedAt: "2026-02-15T00:03:00Z",
          endedAt: null,
        },
      ]),
    });
  });

  await page.route("**/v1/result-submissions", async (route) => {
    await route.fulfill({
      contentType: "application/json",
      body: JSON.stringify([
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
      ]),
    });
  });
  await page.route("**/v1/player-ratings", async (route) => {
    await route.fulfill({
      contentType: "application/json",
      body: JSON.stringify([
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
      ]),
    });
  });

  await page.route("**/v1/result-submission-ratifications", async (route) => {
    await route.fulfill({
      contentType: "application/json",
      body: JSON.stringify({
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
        homeRatifiedAt: "2026-02-15T01:00:00Z",
        awayRatifiedAt: null,
        rejectedByTeamId: null,
        rejectionReason: null,
        rejectedAt: null,
        createdAt: "2026-02-15T00:10:00Z",
      }),
    });
  });

  await page.route("**/v1/result-submission-rejections", async (route) => {
    await route.fulfill({
      contentType: "application/json",
      body: JSON.stringify({
        id: 22,
        contextType: "scrim",
        contextId: 9,
        submittedByTeamId: 100,
        homeTeamId: 100,
        awayTeamId: 101,
        winningTeamId: 100,
        losingTeamId: 101,
        state: "rejected",
        payloadJson: {},
        homeRatifiedAt: null,
        awayRatifiedAt: null,
        rejectedByTeamId: 100,
        rejectionReason: "needs review",
        rejectedAt: "2026-02-15T02:00:00Z",
        createdAt: "2026-02-15T00:10:00Z",
      }),
    });
  });

  await page.route("**/v1/replay-evidence", async (route) => {
    await route.fulfill({
      contentType: "application/json",
      body: JSON.stringify({
        evidence: {
          id: 10,
          contextType: "scrim",
          contextId: 9,
          submittedByTeamId: 100,
          replaySha256: "sha256-demo",
          contentSizeBytes: 12,
          storageRef: "memory://replay/10",
          state: "parsed",
          createdAt: "2026-02-15T02:00:00Z",
        },
        parseRun: {
          id: 99,
          replayEvidenceId: 10,
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
    });
  });
}

async function mockAdminEndpoints(page: import("@playwright/test").Page) {
  await page.route("**/v1/seasons", async (route) => {
    if (route.request().method() === "GET") {
      await route.fulfill({
        contentType: "application/json",
        body: JSON.stringify([
          {
            id: 1,
            name: "Season 1",
            slug: "season-1",
            isActive: true,
            createdAt: "2026-02-15T00:00:00Z",
          },
        ]),
      });
      return;
    }
    await route.fulfill({
      contentType: "application/json",
      body: JSON.stringify({
        id: 2,
        name: "Season 2",
        slug: "season-2",
        isActive: true,
        createdAt: "2026-02-15T03:00:00Z",
      }),
    });
  });

  await page.route("**/v1/schedule-groups", async (route) => {
    if (route.request().method() === "GET") {
      await route.fulfill({
        contentType: "application/json",
        body: JSON.stringify([
          {
            id: 1,
            seasonId: 1,
            name: "Week 1",
            sequence: 1,
            isActive: true,
            createdAt: "2026-02-15T00:00:00Z",
          },
        ]),
      });
      return;
    }
    await route.fulfill({
      contentType: "application/json",
      body: JSON.stringify({
        id: 2,
        seasonId: 1,
        name: "Week 2",
        sequence: 2,
        isActive: true,
        createdAt: "2026-02-15T03:00:00Z",
      }),
    });
  });

  await page.route("**/v1/fixtures", async (route) => {
    if (route.request().method() === "GET") {
      await route.fulfill({
        contentType: "application/json",
        body: JSON.stringify([
          {
            id: 1,
            scheduleGroupId: 1,
            homeClubId: 11,
            awayClubId: 12,
            isActive: true,
            createdAt: "2026-02-15T00:00:00Z",
          },
        ]),
      });
      return;
    }
    await route.fulfill({
      contentType: "application/json",
      body: JSON.stringify({
        id: 2,
        scheduleGroupId: 1,
        homeClubId: 13,
        awayClubId: 14,
        isActive: true,
        createdAt: "2026-02-15T03:00:00Z",
      }),
    });
  });

  await page.route("**/v1/matches", async (route) => {
    if (route.request().method() === "GET") {
      await route.fulfill({
        contentType: "application/json",
        body: JSON.stringify([
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
        ]),
      });
      return;
    }
    await route.fulfill({
      contentType: "application/json",
      body: JSON.stringify({
        id: 2,
        fixtureId: 1,
        homeTeamId: 101,
        awayTeamId: 102,
        state: "scheduled",
        scheduledFor: "2026-02-16T20:00:00Z",
        homeTimeRatifiedAt: null,
        awayTimeRatifiedAt: null,
        createdAt: "2026-02-15T03:00:00Z",
      }),
    });
  });

  await page.route("**/v1/teams", async (route) => {
    await route.fulfill({
      contentType: "application/json",
      body: JSON.stringify([
        {
          id: 1,
          clubId: 11,
          name: "Team A",
          slug: "team-a",
          isActive: true,
          createdAt: "2026-02-15T00:00:00Z",
        },
      ]),
    });
  });

  await page.route("**/v1/players", async (route) => {
    await route.fulfill({
      contentType: "application/json",
      body: JSON.stringify([
        {
          id: 1,
          displayName: "Player One",
          slug: "player-one",
          isActive: true,
          createdAt: "2026-02-15T00:00:00Z",
        },
      ]),
    });
  });

  await page.route("**/v1/roster-memberships", async (route) => {
    if (route.request().method() === "GET") {
      await route.fulfill({
        contentType: "application/json",
        body: JSON.stringify([
          {
            id: 1,
            playerId: 1,
            teamId: 1,
            isActive: true,
            createdAt: "2026-02-15T00:00:00Z",
          },
        ]),
      });
      return;
    }
    await route.fulfill({
      contentType: "application/json",
      body: JSON.stringify({
        id: 2,
        playerId: 1,
        teamId: 1,
        isActive: true,
        createdAt: "2026-02-15T03:00:00Z",
      }),
    });
  });

  await page.route("**/v1/exception-actions", async (route) => {
    await route.fulfill({
      contentType: "application/json",
      body: JSON.stringify([
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
      ]),
    });
  });
  await page.route("**/v1/player-ratings", async (route) => {
    await route.fulfill({
      contentType: "application/json",
      body: JSON.stringify([
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
      ]),
    });
  });
}

test("redirects unauthenticated users to login", async ({ page }) => {
  await page.goto("/app/player");
  await expect(page.getByRole("heading", { name: "Kinetic Sign In" })).toBeVisible();
});

test("allows authorized mock users into scoped route", async ({ page }) => {
  await mockPlayerEndpoints(page);
  await page.addInitScript(() => {
    localStorage.setItem(
      "kinetic.mockSession",
      JSON.stringify({
        subject: "player-1",
        displayName: "Player One",
        roles: ["player"],
      }),
    );
  });

  await page.goto("/app/player");
  await expect(page.getByRole("heading", { name: "Player" })).toBeVisible();
});

test("runs player queue and submission actions", async ({ page }) => {
  await mockPlayerEndpoints(page);
  await page.addInitScript(() => {
    localStorage.setItem(
      "kinetic.mockSession",
      JSON.stringify({
        subject: "player-1",
        displayName: "Player One",
        roles: ["player"],
      }),
    );
  });

  await page.goto("/app/player");
  await page.getByRole("button", { name: "Join queue" }).click();
  await expect(page.getByTestId("player-queue-success")).toBeVisible();

  await page.getByRole("button", { name: "Ratify result" }).click();
  await expect(page.getByTestId("player-submission-success")).toBeVisible();

  await page.getByRole("button", { name: "Upload replay" }).click();
  await expect(page.getByTestId("player-submission-success")).toBeVisible();
});

test("renders admin scheduling workspace and creates entities", async ({ page }) => {
  await mockAdminEndpoints(page);
  await page.addInitScript(() => {
    localStorage.setItem(
      "kinetic.mockSession",
      JSON.stringify({
        subject: "admin-1",
        displayName: "League Admin",
        roles: ["league_admin"],
      }),
    );
  });

  await page.goto("/app/admin");
  await expect(page.getByRole("heading", { name: "League Admin" })).toBeVisible();

  await page.getByRole("button", { name: "Create season" }).click();
  await expect(page.getByTestId("admin-season-success")).toBeVisible();
  await page.getByRole("button", { name: "Assign player to team" }).click();
  await expect(page.getByTestId("admin-roster-success")).toBeVisible();
});

test("renders support inbox workspace and submits triage", async ({ page }) => {
  await page.route("**/v1/operator-inbox", async (route) => {
    await route.fulfill({
      contentType: "application/json",
      body: JSON.stringify([
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
      ]),
    });
  });
  await page.route("**/v1/scrims", async (route) => {
    await route.fulfill({
      contentType: "application/json",
      body: JSON.stringify([
        {
          id: 9,
          queueId: 3,
          homeTeamId: 100,
          awayTeamId: 101,
          state: "started",
          createdAt: "2026-02-15T00:00:00Z",
          startedAt: "2026-02-15T00:03:00Z",
          endedAt: null,
        },
      ]),
    });
  });
  await page.route("**/v1/result-submissions", async (route) => {
    await route.fulfill({
      contentType: "application/json",
      body: JSON.stringify([
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
      ]),
    });
  });
  await page.route("**/v1/operator-inbox/triage", async (route) => {
    const body = route.request().postDataJSON() as {
      reasonCode: string;
      severity: number;
      suggestedAction: string;
    };
    await route.fulfill({
      contentType: "application/json",
      body: JSON.stringify({
        id: 11,
        category: "scheduling_conflict",
        contextType: "match",
        contextId: 42,
        reportedByTeamId: 7,
        state: "triaged",
        reasonCode: body.reasonCode,
        severity: body.severity,
        suggestedAction: body.suggestedAction,
        detailsJson: {},
        resolutionCode: null,
        openedAt: "2026-02-15T00:00:00Z",
        triagedAt: "2026-02-15T01:00:00Z",
        resolvedAt: null,
      }),
    });
  });
  await page.route("**/v1/operator-inbox/resolve", async (route) => {
    const body = route.request().postDataJSON() as {
      resolutionCode: string;
    };
    await route.fulfill({
      contentType: "application/json",
      body: JSON.stringify({
        id: 11,
        category: "scheduling_conflict",
        contextType: "match",
        contextId: 42,
        reportedByTeamId: 7,
        state: "resolved",
        reasonCode: "availability_conflict",
        severity: 3,
        suggestedAction: "request_reschedule",
        detailsJson: {},
        resolutionCode: body.resolutionCode,
        openedAt: "2026-02-15T00:00:00Z",
        triagedAt: "2026-02-15T01:00:00Z",
        resolvedAt: "2026-02-15T02:00:00Z",
      }),
    });
  });

  await page.addInitScript(() => {
    localStorage.setItem(
      "kinetic.mockSession",
      JSON.stringify({
        subject: "support-1",
        displayName: "Support One",
        roles: ["league_support"],
      }),
    );
  });

  await page.goto("/app/support");
  await expect(page.getByRole("heading", { name: "League Support" })).toBeVisible();
  await expect(page.getByTestId("active-scrims-count")).toContainText("1");
  await expect(page.getByTestId("submissions-in-process-count")).toContainText("1");

  await page.getByRole("button", { name: "Submit triage" }).click();
  await expect(page.getByTestId("triage-success")).toBeVisible();
});
