export type Role = "player" | "league_admin" | "league_support" | "platform_operator";

export type SessionPrincipal = {
  subject: string;
  displayName: string;
  roles: Role[];
};
