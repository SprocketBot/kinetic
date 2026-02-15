import type { Role } from "../../auth/types";

export function hasRole(roles: Role[], expected: Role): boolean {
  return roles.includes(expected);
}
