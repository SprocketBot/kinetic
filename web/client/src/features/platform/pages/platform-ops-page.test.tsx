import { QueryClient, QueryClientProvider } from "@tanstack/react-query";
import { cleanup, render, screen } from "@testing-library/react";
import { MemoryRouter } from "react-router-dom";
import { afterAll, afterEach, beforeAll, describe, expect, it } from "vitest";

import { setMockSession } from "../../../auth/auth-service";
import { SessionProvider } from "../../../auth/session-context";
import { server } from "../../../test/msw/server";
import { PlatformOpsPage } from "./platform-ops-page";

beforeAll(() => server.listen());
afterEach(() => {
  server.resetHandlers();
  cleanup();
});
afterAll(() => server.close());

describe("PlatformOpsPage", () => {
  it("renders metrics cards", async () => {
    setMockSession({
      subject: "platform-1",
      displayName: "Platform Operator",
      roles: ["platform_operator"],
    });

    const queryClient = new QueryClient();

    render(
      <MemoryRouter>
        <QueryClientProvider client={queryClient}>
          <SessionProvider>
            <PlatformOpsPage />
          </SessionProvider>
        </QueryClientProvider>
      </MemoryRouter>,
    );

    expect(await screen.findByText("Admin hours / week")).toBeInTheDocument();
    expect(screen.getByText("6.50")).toBeInTheDocument();
    expect(screen.getByText("62.0%")).toBeInTheDocument();
  });
});
