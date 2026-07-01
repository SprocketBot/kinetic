import { expect, test } from "@playwright/test";

import { installCUJApi, signInAs } from "./cuj-fixtures";

test("CUJ 3: league admin can see schedule entities and create the next match state", async ({ page }) => {
  await installCUJApi(page);
  await signInAs(page, "league_admin");

  await page.goto("/app/admin");
  await expect(page.getByRole("heading", { name: "League Admin" })).toBeVisible();
  await expect(page.getByRole("heading", { name: "Seasons" })).toBeVisible();
  await expect(page.getByRole("heading", { name: "Schedule Groups" })).toBeVisible();
  await expect(page.getByRole("heading", { name: "Fixtures" })).toBeVisible();
  await expect(page.getByRole("heading", { name: "Matches" })).toBeVisible();

  await page.getByRole("button", { name: "Create season" }).click();
  await expect(page.getByTestId("admin-season-success")).toBeVisible();
  await page.getByRole("button", { name: "Create schedule group" }).click();
  await expect(page.getByTestId("admin-group-success")).toBeVisible();
  await page.getByRole("button", { name: "Create fixture" }).click();
  await expect(page.getByTestId("admin-fixture-success")).toBeVisible();
  await page.getByRole("button", { name: "Create match" }).click();
  await expect(page.getByTestId("admin-match-success")).toBeVisible();
});
