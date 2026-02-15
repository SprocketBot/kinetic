import type { ReactNode } from "react";
import { Navigate } from "react-router-dom";

import { LoadingState } from "../components/feedback/loading-state";
import { useSession } from "./session-context";

type AuthGuardProps = {
  children: ReactNode;
};

export function AuthGuard({ children }: AuthGuardProps) {
  const session = useSession();

  if (session.status === "loading") {
    return <LoadingState label="Checking session..." />;
  }

  if (session.status === "anonymous") {
    return <Navigate to="/login" replace />;
  }

  return <>{children}</>;
}
