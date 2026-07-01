import { expect, test } from "@playwright/test";

import { installCUJApi, signInAs } from "./cuj-fixtures";

test("CUJ 1: login redirects anonymous users and player landing shows scrim status", async ({ page }) => {
  await installCUJApi(page);

  await page.goto("/app/player");
  await expect(page.getByRole("heading", { name: "Sprocket Sign In" })).toBeVisible();

  await signInAs(page, "player");
  await page.goto("/app/player");

  await expect(page.getByRole("heading", { name: "Player" })).toBeVisible();
  await expect(page.getByTestId("player-queue-count")).toHaveText("1");
  await expect(page.getByTestId("player-scrim-count")).toHaveText("1");
  await expect(page.getByTestId("player-submission-count")).toHaveText("1");
  await expect(page.getByText(/92 points \(threshold 40/)).toBeVisible();
});
