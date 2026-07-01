import { expect, test } from "@playwright/test";

import { installCUJApi, signInAs } from "./cuj-fixtures";

test("CUJ 7: player history links expose Evidence statistics surfaces", async ({ page }) => {
  await installCUJApi(page);
  await signInAs(page, "player");

  await page.goto("/app/player");
  await expect(page.getByRole("heading", { name: "Evidence Views" })).toBeVisible();
  await expect(page.locator("iframe[title='Evidence Standings']")).toHaveAttribute("src", /\/standings$/);

  await page.getByLabel("View").selectOption("ratings");
  await expect(page.locator("iframe[title='Evidence Ratings']")).toHaveAttribute("src", /\/ratings$/);
  await page.getByLabel("View").selectOption("eligibility");
  await expect(page.locator("iframe[title='Evidence Eligibility']")).toHaveAttribute("src", /\/eligibility$/);
});
