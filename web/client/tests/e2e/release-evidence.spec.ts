import { expect, test } from "@playwright/test";
import { mkdir, writeFile } from "node:fs/promises";
import { join } from "node:path";

type BrowserSession = {
  status: number;
  body: {
    subject?: string;
    displayName?: string;
    roles?: string[];
  };
};

const releaseEvidenceEnabled = process.env.RELEASE_EVIDENCE === "1";
const apiBaseUrl = process.env.VITE_API_BASE_URL ?? "http://127.0.0.1:8080";
const artifactDir = process.env.RELEASE_EVIDENCE_ARTIFACT_DIR;

test.skip(!releaseEvidenceEnabled, "release evidence tests require a live API started by tools/release-evidence.sh");

async function writeArtifact(name: string, payload: unknown) {
  if (!artifactDir) {
    return;
  }
  await mkdir(artifactDir, { recursive: true });
  await writeFile(join(artifactDir, name), `${JSON.stringify(payload, null, 2)}\n`, "utf8");
}

async function readBrowserSession(page: import("@playwright/test").Page): Promise<BrowserSession> {
  return page.evaluate(async (baseUrl) => {
    const response = await fetch(`${baseUrl}/v1/session`, {
      credentials: "include",
      headers: { Accept: "application/json" },
    });
    const body = response.ok ? await response.json() : {};
    return { status: response.status, body };
  }, apiBaseUrl);
}

test("release evidence: browser auth keeps actor identity isolated over credentialed CORS", async ({ page }) => {
  await page.goto("/login");
  await expect(page.getByRole("heading", { name: "Kinetic Sign In" })).toBeVisible();

  await page.getByRole("button", { name: "Continue as player" }).click();
  await expect(page).toHaveURL(/\/app\/player$/);
  await expect(page.getByRole("heading", { name: "Player" })).toBeVisible();
  await expect(page.getByText("Local Player")).toBeVisible();

  const playerSession = await readBrowserSession(page);
  await writeArtifact("browser-player-session.json", playerSession);
  expect(playerSession.status).toBe(200);
  expect(playerSession.body.subject).toBe("local-player");
  expect(playerSession.body.displayName).toBe("Local Player");
  expect(playerSession.body.roles).toEqual(["player"]);
  expect(playerSession.body.roles ?? []).not.toContain("admin");
  expect(playerSession.body.roles ?? []).not.toContain("league_admin");

  const forbiddenOverride = await page.evaluate(async (baseUrl) => {
    const response = await fetch(`${baseUrl}/v1/result-overrides`, {
      body: JSON.stringify({
        actor: "release-evidence",
        losingTeamId: 2,
        reason: "browser CORS negative control",
        submissionId: 1,
        winningTeamId: 1,
      }),
      credentials: "include",
      headers: {
        Accept: "application/json",
        "Content-Type": "application/json",
      },
      method: "POST",
    });
    return { body: await response.text(), status: response.status };
  }, apiBaseUrl);
  await writeArtifact("browser-player-forbidden-override.json", forbiddenOverride);
  expect(forbiddenOverride.status).toBe(403);

  if (artifactDir) {
    await page.screenshot({ fullPage: true, path: join(artifactDir, "browser-player-page.png") });
  }

  await page.goto("/app/admin");
  await expect(page.getByRole("heading", { name: "Unauthorized" })).toBeVisible();
  if (artifactDir) {
    await page.screenshot({ fullPage: true, path: join(artifactDir, "browser-player-admin-denied.png") });
  }

  await page.goto("/login");
  await page.getByRole("button", { name: "Continue as support" }).click();
  await expect(page).toHaveURL(/\/app\/support$/);
  await expect(page.getByRole("heading", { name: "League Support" })).toBeVisible();
  await expect(page.getByText("Local Support")).toBeVisible();

  const supportSession = await readBrowserSession(page);
  await writeArtifact("browser-support-session.json", supportSession);
  expect(supportSession.status).toBe(200);
  expect(supportSession.body.subject).toBe("local-league-support");
  expect(supportSession.body.displayName).toBe("Local Support");
  expect(supportSession.body.roles).toEqual(["league_support"]);
  expect(supportSession.body.subject).not.toBe(playerSession.body.subject);
});
