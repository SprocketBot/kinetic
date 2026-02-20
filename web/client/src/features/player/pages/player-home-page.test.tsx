import { QueryClient, QueryClientProvider } from "@tanstack/react-query";
import { cleanup, render, screen, waitFor } from "@testing-library/react";
import userEvent from "@testing-library/user-event";
import { MemoryRouter } from "react-router-dom";
import { afterAll, afterEach, beforeAll, describe, expect, it } from "vitest";

import { setMockSession } from "../../../auth/auth-service";
import { SessionProvider } from "../../../auth/session-context";
import { server } from "../../../test/msw/server";
import { PlayerHomePage } from "./player-home-page";

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
          <PlayerHomePage />
        </SessionProvider>
      </QueryClientProvider>
    </MemoryRouter>,
  );
}

describe("PlayerHomePage", () => {
  it("renders player live counts", async () => {
    setMockSession({
      subject: "player-1",
      displayName: "Player One",
      roles: ["player"],
    });

    renderPage();

    await waitFor(() => expect(screen.getByTestId("player-queue-count")).toHaveTextContent("1"));
    await waitFor(() => expect(screen.getByTestId("player-scrim-count")).toHaveTextContent("1"));
    await waitFor(() => expect(screen.getByTestId("player-submission-count")).toHaveTextContent("1"));
    expect(screen.getByRole("heading", { name: "Ratings" })).toBeInTheDocument();
    expect(screen.getByRole("heading", { name: "Accounts and Eligibility" })).toBeInTheDocument();
  });

  it("submits queue and submission actions", async () => {
    setMockSession({
      subject: "player-1",
      displayName: "Player One",
      roles: ["player"],
    });

    renderPage();

    expect(await screen.findByRole("heading", { name: "Queue Actions" })).toBeInTheDocument();

    await userEvent.click(screen.getByRole("button", { name: "Join queue" }));
    expect(await screen.findByTestId("player-queue-success")).toBeInTheDocument();

    await userEvent.click(screen.getByRole("button", { name: "Ratify result" }));
    expect(await screen.findByTestId("player-submission-success")).toBeInTheDocument();

    await userEvent.click(screen.getByRole("button", { name: "Upload replay" }));
    expect(await screen.findByTestId("player-submission-success")).toBeInTheDocument();

    await userEvent.clear(screen.getByLabelText("Provider account ID"));
    await userEvent.type(screen.getByLabelText("Provider account ID"), "epic-222");
    await userEvent.type(screen.getByLabelText("Provider display name"), "Player One Epic");
    await userEvent.click(screen.getByRole("button", { name: "Link account" }));
    expect(await screen.findByTestId("player-platform-link-success")).toBeInTheDocument();

    await userEvent.click(screen.getAllByRole("button", { name: "Unlink" })[0]);
    expect(await screen.findByTestId("player-platform-link-success")).toBeInTheDocument();
  });
});
