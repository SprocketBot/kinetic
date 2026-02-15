import { QueryClient, QueryClientProvider } from "@tanstack/react-query";
import { cleanup, render, screen } from "@testing-library/react";
import userEvent from "@testing-library/user-event";
import { MemoryRouter } from "react-router-dom";
import { afterAll, afterEach, beforeAll, describe, expect, it } from "vitest";

import { setMockSession } from "../../../auth/auth-service";
import { SessionProvider } from "../../../auth/session-context";
import { server } from "../../../test/msw/server";
import { AdminHomePage } from "./admin-home-page";

beforeAll(() => server.listen());
afterEach(() => {
  server.resetHandlers();
  cleanup();
});
afterAll(() => server.close());

function renderPage() {
  const queryClient = new QueryClient();

  render(
    <MemoryRouter>
      <QueryClientProvider client={queryClient}>
        <SessionProvider>
          <AdminHomePage />
        </SessionProvider>
      </QueryClientProvider>
    </MemoryRouter>,
  );
}

describe("AdminHomePage", () => {
  it("renders scheduling entities and supports create actions", async () => {
    setMockSession({
      subject: "admin-1",
      displayName: "League Admin",
      roles: ["league_admin"],
    });

    renderPage();

    expect(await screen.findByRole("heading", { name: "Seasons" })).toBeInTheDocument();
    expect(screen.getByRole("heading", { name: "Schedule Groups" })).toBeInTheDocument();
    expect(screen.getByRole("heading", { name: "Fixtures" })).toBeInTheDocument();
    expect(screen.getByRole("heading", { name: "Matches" })).toBeInTheDocument();

    await userEvent.click(screen.getByRole("button", { name: "Create season" }));
    expect(await screen.findByTestId("admin-season-success")).toBeInTheDocument();

    await userEvent.click(screen.getByRole("button", { name: "Create schedule group" }));
    expect(await screen.findByTestId("admin-group-success")).toBeInTheDocument();

    await userEvent.click(screen.getByRole("button", { name: "Create fixture" }));
    expect(await screen.findByTestId("admin-fixture-success")).toBeInTheDocument();

    await userEvent.click(screen.getByRole("button", { name: "Create match" }));
    expect(await screen.findByTestId("admin-match-success")).toBeInTheDocument();
  });
});
