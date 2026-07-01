import { expect, test } from "@playwright/test";

import { installCUJApi, signInAs } from "./cuj-fixtures";

test("CUJ 5: support can triage, resolve, and moderate queue access", async ({ page }) => {
  await installCUJApi(page);
  await signInAs(page, "league_support");

  await page.goto("/app/support");
  await expect(page.getByRole("heading", { name: "League Support" })).toBeVisible();
  await expect(page.getByTestId("active-scrims-count")).toHaveText("1");
  await expect(page.getByTestId("submissions-in-process-count")).toHaveText("1");

  await page.getByRole("button", { name: "Submit triage" }).click();
  await expect(page.getByTestId("triage-success")).toBeVisible();
  await page.getByRole("button", { name: "Submit resolution" }).click();
  await expect(page.getByTestId("resolve-success")).toBeVisible();

  await page.getByRole("button", { name: "Ban player from queue" }).click();
  await expect(page.getByTestId("queue-ban-success")).toBeVisible();
  await page.getByRole("button", { name: "Lift queue ban" }).click();
  await expect(page.getByTestId("queue-unban-success")).toBeVisible();
});
