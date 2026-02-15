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
