import { expect, test } from "@playwright/test";

import { installCUJApi, signInAs } from "./cuj-fixtures";

test("CUJ 6: player ratings are visible and admins can apply audited changes", async ({ page }) => {
  await installCUJApi(page);
  await signInAs(page, "player");

  await page.goto("/app/player");
  await expect(page.getByText(/Player 1 · scrim-3v3 · rating 1000/)).toBeVisible();

  await signInAs(page, "league_admin");
  await page.goto("/app/admin");
  await page.getByRole("button", { name: "Apply rating change" }).click();
  await expect(page.getByTestId("admin-rating-success")).toBeVisible();
});
