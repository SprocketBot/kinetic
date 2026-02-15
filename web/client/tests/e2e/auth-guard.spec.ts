import { expect, test } from "@playwright/test";

test("redirects unauthenticated users to login", async ({ page }) => {
  await page.goto("/app/player");
  await expect(page.getByRole("heading", { name: "Sprocket Sign In" })).toBeVisible();
});

test("allows authorized mock users into scoped route", async ({ page }) => {
  await page.addInitScript(() => {
    localStorage.setItem(
      "sprocket.mockSession",
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
      "sprocket.mockSession",
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
