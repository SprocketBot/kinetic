export const env = {
  apiBaseUrl: import.meta.env.VITE_API_BASE_URL ?? "http://localhost:8080",
  authMode: (import.meta.env.VITE_AUTH_MODE ?? "mock") as "mock" | "api",
  mockPrincipalJson: import.meta.env.VITE_MOCK_PRINCIPAL_JSON,
  evidenceBaseUrl: import.meta.env.VITE_EVIDENCE_BASE_URL ?? "https://evidence.sprocket.gg",
};
