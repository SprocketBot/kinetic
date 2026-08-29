# Kinetic Web Client

The React/TypeScript client provides role-aware surfaces for players, league
support, league administrators, and platform operators. It uses Vite, TanStack
Query, Zod, Vitest, Testing Library, and Playwright.

## Run Locally

From this directory:

```bash
npm ci
npm run dev
```

The client uses mock authentication by default. Set `VITE_AUTH_MODE=api` to use
the backend session endpoints. The repository-level supported full-stack path
is:

```bash
./tools/start-dev.sh --with-web
```

## Verify

```bash
npm run lint
npm run typecheck
npm run test
npm run build
npm run test:e2e
```

From the repository root, `./tools/web-quality-gate.sh` runs the standard client
quality checks.

## Product Surfaces

- `/app/player`: queue, scrim, replay submission, ratings, and notifications
- `/app/support`: operator inbox, scrim support, and submission support
- `/app/admin`: scheduling, rosters, delegated roles, ratings, and overrides
- `/app/platform`: platform health and operations links

See `docs/interface-design.md` for the intended role surfaces and
`docs/plans/product-cutover-plan.md` for work that remains before cutover.
