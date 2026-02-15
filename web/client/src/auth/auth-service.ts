import { env } from "../lib/config/env";
import type { SessionPrincipal } from "./types";

const MOCK_SESSION_KEY = "sprocket.mockSession";

function parseMockSession(raw: string | null): SessionPrincipal | null {
  if (!raw) {
    return null;
  }

  try {
    return JSON.parse(raw) as SessionPrincipal;
  } catch {
    return null;
  }
}

function parseMockPrincipalFromEnv(): SessionPrincipal | null {
  if (!env.mockPrincipalJson) {
    return null;
  }

  try {
    return JSON.parse(env.mockPrincipalJson) as SessionPrincipal;
  } catch {
    return null;
  }
}

async function getMockSession(): Promise<SessionPrincipal | null> {
  const fromStorage = parseMockSession(window.localStorage.getItem(MOCK_SESSION_KEY));
  return fromStorage ?? parseMockPrincipalFromEnv();
}

async function getApiSession(): Promise<SessionPrincipal | null> {
  const response = await fetch(`${env.apiBaseUrl}/v1/session`, {
    credentials: "include",
    headers: { Accept: "application/json" },
  });

  if (response.status === 401) {
    return null;
  }

  if (!response.ok) {
    throw new Error(`session lookup failed: ${response.status}`);
  }

  return (await response.json()) as SessionPrincipal;
}

export async function getSession(): Promise<SessionPrincipal | null> {
  if (env.authMode === "mock") {
    return getMockSession();
  }

  return getApiSession();
}

export async function login(): Promise<void> {
  if (env.authMode === "mock") {
    return;
  }

  window.location.assign(`${env.apiBaseUrl}/v1/auth/login`);
}

export async function logout(): Promise<void> {
  if (env.authMode === "mock") {
    window.localStorage.removeItem(MOCK_SESSION_KEY);
    return;
  }

  await fetch(`${env.apiBaseUrl}/v1/auth/logout`, {
    credentials: "include",
    method: "POST",
  });
}

export function setMockSession(principal: SessionPrincipal | null) {
  if (principal === null) {
    window.localStorage.removeItem(MOCK_SESSION_KEY);
    return;
  }

  window.localStorage.setItem(MOCK_SESSION_KEY, JSON.stringify(principal));
}
