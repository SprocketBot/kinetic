import { expect, test } from "@playwright/test";

import { installCUJApi, signInAs } from "./cuj-fixtures";

test("CUJ 2: player can move through scrim queue, replay upload, and ratification actions", async ({ page }) => {
  await installCUJApi(page);
  await signInAs(page, "player");

  await page.goto("/app/player");
  await expect(page.getByText("Scrim #9 · queue 1 · started")).toBeVisible();

  await page.getByRole("button", { name: "Join queue" }).click();
  await expect(page.getByTestId("player-queue-success")).toBeVisible();

  await page.getByRole("button", { name: "Submit result" }).click();
  await expect(page.getByTestId("player-submission-success")).toBeVisible();

  await page.getByRole("button", { name: "Upload replay evidence" }).click();
  await expect(page.getByTestId("player-submission-success")).toBeVisible();

  await page.getByRole("button", { name: "Ratify result" }).click();
  await expect(page.getByTestId("player-submission-success")).toBeVisible();
});
