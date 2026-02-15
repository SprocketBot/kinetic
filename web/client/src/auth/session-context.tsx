import { createContext, useContext, useEffect, useMemo, useState } from "react";
import type { ReactNode } from "react";

import { getSession, login, logout } from "./auth-service";
import type { SessionPrincipal } from "./types";

type SessionStatus = "loading" | "authenticated" | "anonymous";

type SessionContextValue = {
  status: SessionStatus;
  principal: SessionPrincipal | null;
  refresh: () => Promise<void>;
  login: () => Promise<void>;
  logout: () => Promise<void>;
};

const SessionContext = createContext<SessionContextValue | null>(null);

type SessionProviderProps = {
  children: ReactNode;
};

export function SessionProvider({ children }: SessionProviderProps) {
  const [status, setStatus] = useState<SessionStatus>("loading");
  const [principal, setPrincipal] = useState<SessionPrincipal | null>(null);

  async function refresh() {
    const session = await getSession();
    setPrincipal(session);
    setStatus(session ? "authenticated" : "anonymous");
  }

  useEffect(() => {
    void refresh();
  }, []);

  const value = useMemo(
    () => ({
      status,
      principal,
      refresh,
      login,
      logout,
    }),
    [status, principal],
  );

  return <SessionContext.Provider value={value}>{children}</SessionContext.Provider>;
}

export function useSession() {
  const context = useContext(SessionContext);

  if (context === null) {
    throw new Error("useSession must be used within SessionProvider");
  }

  return context;
}
