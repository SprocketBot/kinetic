import { defineConfig } from "@playwright/test";

const previewHost = process.env.PLAYWRIGHT_WEB_HOST ?? "127.0.0.1";
const previewPort = Number(process.env.PLAYWRIGHT_WEB_PORT ?? "4173");
const reuseExistingServer = !process.env.CI && process.env.RELEASE_EVIDENCE !== "1";

export default defineConfig({
  testDir: "./tests/e2e",
  use: {
    baseURL: `http://${previewHost}:${previewPort}`,
    headless: true,
  },
  webServer: {
    command: `npm run build && npm run preview -- --host ${previewHost} --port ${previewPort}`,
    port: previewPort,
    reuseExistingServer,
  },
});
