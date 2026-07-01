import { expect, test } from "@playwright/test";

import { installCUJApi, signInAs } from "./cuj-fixtures";

test("CUJ 4: admin can view and mutate rosters and scoped staff roles", async ({ page }) => {
  await installCUJApi(page);
  await signInAs(page, "league_admin");

  await page.goto("/app/admin");
  await expect(page.getByRole("heading", { name: "Roster Memberships" })).toBeVisible();
  await expect(page.getByText("Player One -> Team A")).toBeVisible();

  await page.getByRole("button", { name: "Assign player to team" }).click();
  await expect(page.getByTestId("admin-roster-success")).toBeVisible();

  await page.getByRole("button", { name: "Assign role" }).click();
  await expect(page.getByTestId("admin-role-assign-success")).toBeVisible();
  await page.getByRole("button", { name: "Revoke role" }).click();
  await expect(page.getByTestId("admin-role-revoke-success")).toBeVisible();
});
