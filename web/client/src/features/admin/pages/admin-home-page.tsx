import { useMemo, useState } from "react";
import type { FormEvent } from "react";
import { useMutation, useQuery, useQueryClient } from "@tanstack/react-query";

import { AppShell } from "../../../components/layout/app-shell";
import { LoadingState } from "../../../components/feedback/loading-state";
import { getJson, postJson } from "../../../lib/api/client";
import { ApiError } from "../../../lib/api/errors";
import {
  createFixtureInputSchema,
  createMatchInputSchema,
  createScheduleGroupInputSchema,
  createSeasonInputSchema,
  fixtureListSchema,
  fixtureSchema,
  matchListSchema,
  matchSchema,
  scheduleGroupListSchema,
  scheduleGroupSchema,
  seasonListSchema,
  seasonSchema,
  type Fixture,
  type Match,
  type ScheduleGroup,
  type Season,
} from "../../../lib/api/schemas";

async function listSeasons() {
  return getJson("/v1/seasons", seasonListSchema);
}

async function createSeason(input: { name: string; slug: string }) {
  const payload = createSeasonInputSchema.parse(input);
  return postJson("/v1/seasons", payload, seasonSchema);
}

async function listScheduleGroups() {
  return getJson("/v1/schedule-groups", scheduleGroupListSchema);
}

async function createScheduleGroup(input: { seasonId: number; name: string; sequence: number }) {
  const payload = createScheduleGroupInputSchema.parse(input);
  return postJson("/v1/schedule-groups", payload, scheduleGroupSchema);
}

async function listFixtures() {
  return getJson("/v1/fixtures", fixtureListSchema);
}

async function createFixture(input: { scheduleGroupId: number; homeClubId: number; awayClubId: number }) {
  const payload = createFixtureInputSchema.parse(input);
  return postJson("/v1/fixtures", payload, fixtureSchema);
}

async function listMatches() {
  return getJson("/v1/matches", matchListSchema);
}

async function createMatch(input: {
  fixtureId: number;
  homeTeamId: number;
  awayTeamId: number;
  state: string;
  scheduledFor?: string;
}) {
  const payload = createMatchInputSchema.parse(input);
  return postJson("/v1/matches", payload, matchSchema);
}

export function AdminHomePage() {
  const queryClient = useQueryClient();

  const seasonsQuery = useQuery({ queryKey: ["seasons"], queryFn: listSeasons });
  const scheduleGroupsQuery = useQuery({ queryKey: ["schedule-groups"], queryFn: listScheduleGroups });
  const fixturesQuery = useQuery({ queryKey: ["fixtures"], queryFn: listFixtures });
  const matchesQuery = useQuery({ queryKey: ["matches"], queryFn: listMatches });

  const seasonMutation = useMutation({
    mutationFn: createSeason,
    onMutate: async (input) => {
      await queryClient.cancelQueries({ queryKey: ["seasons"] });
      const previous = queryClient.getQueryData<Season[]>(["seasons"]) ?? [];
      queryClient.setQueryData<Season[]>(["seasons"], [
        ...previous,
        { id: -Date.now(), name: input.name, slug: input.slug, isActive: true, createdAt: new Date().toISOString() },
      ]);
      return { previous };
    },
    onError: (_error, _input, context) => {
      if (context?.previous) {
        queryClient.setQueryData(["seasons"], context.previous);
      }
    },
    onSettled: async () => {
      await queryClient.invalidateQueries({ queryKey: ["seasons"] });
    },
  });

  const scheduleGroupMutation = useMutation({
    mutationFn: createScheduleGroup,
    onMutate: async (input) => {
      await queryClient.cancelQueries({ queryKey: ["schedule-groups"] });
      const previous = queryClient.getQueryData<ScheduleGroup[]>(["schedule-groups"]) ?? [];
      queryClient.setQueryData<ScheduleGroup[]>(["schedule-groups"], [
        ...previous,
        {
          id: -Date.now(),
          seasonId: input.seasonId,
          name: input.name,
          sequence: input.sequence,
          isActive: true,
          createdAt: new Date().toISOString(),
        },
      ]);
      return { previous };
    },
    onError: (_error, _input, context) => {
      if (context?.previous) {
        queryClient.setQueryData(["schedule-groups"], context.previous);
      }
    },
    onSettled: async () => {
      await queryClient.invalidateQueries({ queryKey: ["schedule-groups"] });
    },
  });

  const fixtureMutation = useMutation({
    mutationFn: createFixture,
    onMutate: async (input) => {
      await queryClient.cancelQueries({ queryKey: ["fixtures"] });
      const previous = queryClient.getQueryData<Fixture[]>(["fixtures"]) ?? [];
      queryClient.setQueryData<Fixture[]>(["fixtures"], [
        ...previous,
        {
          id: -Date.now(),
          scheduleGroupId: input.scheduleGroupId,
          homeClubId: input.homeClubId,
          awayClubId: input.awayClubId,
          isActive: true,
          createdAt: new Date().toISOString(),
        },
      ]);
      return { previous };
    },
    onError: (_error, _input, context) => {
      if (context?.previous) {
        queryClient.setQueryData(["fixtures"], context.previous);
      }
    },
    onSettled: async () => {
      await queryClient.invalidateQueries({ queryKey: ["fixtures"] });
    },
  });

  const matchMutation = useMutation({
    mutationFn: createMatch,
    onMutate: async (input) => {
      await queryClient.cancelQueries({ queryKey: ["matches"] });
      const previous = queryClient.getQueryData<Match[]>(["matches"]) ?? [];
      queryClient.setQueryData<Match[]>(["matches"], [
        ...previous,
        {
          id: -Date.now(),
          fixtureId: input.fixtureId,
          homeTeamId: input.homeTeamId,
          awayTeamId: input.awayTeamId,
          state: input.state,
          scheduledFor: input.scheduledFor ?? null,
          homeTimeRatifiedAt: null,
          awayTimeRatifiedAt: null,
          createdAt: new Date().toISOString(),
        },
      ]);
      return { previous };
    },
    onError: (_error, _input, context) => {
      if (context?.previous) {
        queryClient.setQueryData(["matches"], context.previous);
      }
    },
    onSettled: async () => {
      await queryClient.invalidateQueries({ queryKey: ["matches"] });
    },
  });

  return (
    <AppShell title="Sprocket Web Client">
      <h2>League Admin</h2>
      <p>Schedule lifecycle management for seasons, groups, fixtures, and matches.</p>
      <p>
        Update/delete workflows are disabled until API support is added. Current backend supports create/list for this
        slice.
      </p>

      <section style={{ display: "grid", gap: "1rem", gridTemplateColumns: "1fr 1fr" }}>
        <SeasonPanel
          createError={seasonMutation.error}
          createSeason={seasonMutation.mutateAsync}
          createSuccess={seasonMutation.isSuccess}
          loading={seasonsQuery.isLoading}
          seasons={seasonsQuery.data ?? []}
        />

        <ScheduleGroupPanel
          createError={scheduleGroupMutation.error}
          createScheduleGroup={scheduleGroupMutation.mutateAsync}
          createSuccess={scheduleGroupMutation.isSuccess}
          groups={scheduleGroupsQuery.data ?? []}
          loading={scheduleGroupsQuery.isLoading}
        />

        <FixturePanel
          createError={fixtureMutation.error}
          createFixture={fixtureMutation.mutateAsync}
          createSuccess={fixtureMutation.isSuccess}
          fixtures={fixturesQuery.data ?? []}
          loading={fixturesQuery.isLoading}
        />

        <MatchPanel
          createError={matchMutation.error}
          createMatch={matchMutation.mutateAsync}
          createSuccess={matchMutation.isSuccess}
          loading={matchesQuery.isLoading}
          matches={matchesQuery.data ?? []}
        />
      </section>
    </AppShell>
  );
}

function SeasonPanel({
  seasons,
  loading,
  createSeason,
  createSuccess,
  createError,
}: {
  seasons: Season[];
  loading: boolean;
  createSeason: (input: { name: string; slug: string }) => Promise<unknown>;
  createSuccess: boolean;
  createError: Error | null;
}) {
  const [name, setName] = useState("Season 1");
  const [slug, setSlug] = useState("season-1");

  async function submit(event: FormEvent<HTMLFormElement>) {
    event.preventDefault();
    await createSeason({ name, slug });
  }

  return (
    <section>
      <h3>Seasons</h3>
      <form onSubmit={submit}>
        <label>
          Name
          <input value={name} onChange={(event) => setName(event.target.value)} />
        </label>
        <label>
          Slug
          <input value={slug} onChange={(event) => setSlug(event.target.value)} />
        </label>
        <button type="submit">Create season</button>
      </form>
      {createSuccess && <p data-testid="admin-season-success">Season created.</p>}
      {createError && <ErrorView error={createError} label="Season create failed" />}
      <DataList items={seasons.map((season) => `#${season.id} · ${season.name} (${season.slug})`)} loading={loading} />
    </section>
  );
}

function ScheduleGroupPanel({
  groups,
  loading,
  createScheduleGroup,
  createSuccess,
  createError,
}: {
  groups: ScheduleGroup[];
  loading: boolean;
  createScheduleGroup: (input: { seasonId: number; name: string; sequence: number }) => Promise<unknown>;
  createSuccess: boolean;
  createError: Error | null;
}) {
  const [seasonId, setSeasonId] = useState(1);
  const [name, setName] = useState("Week 1");
  const [sequence, setSequence] = useState(1);

  async function submit(event: FormEvent<HTMLFormElement>) {
    event.preventDefault();
    await createScheduleGroup({ seasonId, name, sequence });
  }

  return (
    <section>
      <h3>Schedule Groups</h3>
      <form onSubmit={submit}>
        <label>
          Season ID
          <input min={1} type="number" value={seasonId} onChange={(event) => setSeasonId(Number(event.target.value))} />
        </label>
        <label>
          Name
          <input value={name} onChange={(event) => setName(event.target.value)} />
        </label>
        <label>
          Sequence
          <input min={0} type="number" value={sequence} onChange={(event) => setSequence(Number(event.target.value))} />
        </label>
        <button type="submit">Create schedule group</button>
      </form>
      {createSuccess && <p data-testid="admin-group-success">Schedule group created.</p>}
      {createError && <ErrorView error={createError} label="Schedule group create failed" />}
      <DataList
        items={groups.map((group) => `#${group.id} · season ${group.seasonId} · ${group.name} · seq ${group.sequence}`)}
        loading={loading}
      />
    </section>
  );
}

function FixturePanel({
  fixtures,
  loading,
  createFixture,
  createSuccess,
  createError,
}: {
  fixtures: Fixture[];
  loading: boolean;
  createFixture: (input: { scheduleGroupId: number; homeClubId: number; awayClubId: number }) => Promise<unknown>;
  createSuccess: boolean;
  createError: Error | null;
}) {
  const [scheduleGroupId, setScheduleGroupId] = useState(1);
  const [homeClubId, setHomeClubId] = useState(11);
  const [awayClubId, setAwayClubId] = useState(12);

  async function submit(event: FormEvent<HTMLFormElement>) {
    event.preventDefault();
    await createFixture({ scheduleGroupId, homeClubId, awayClubId });
  }

  return (
    <section>
      <h3>Fixtures</h3>
      <form onSubmit={submit}>
        <label>
          Schedule group ID
          <input
            min={1}
            type="number"
            value={scheduleGroupId}
            onChange={(event) => setScheduleGroupId(Number(event.target.value))}
          />
        </label>
        <label>
          Home club ID
          <input
            min={1}
            type="number"
            value={homeClubId}
            onChange={(event) => setHomeClubId(Number(event.target.value))}
          />
        </label>
        <label>
          Away club ID
          <input
            min={1}
            type="number"
            value={awayClubId}
            onChange={(event) => setAwayClubId(Number(event.target.value))}
          />
        </label>
        <button type="submit">Create fixture</button>
      </form>
      {createSuccess && <p data-testid="admin-fixture-success">Fixture created.</p>}
      {createError && <ErrorView error={createError} label="Fixture create failed" />}
      <DataList
        items={fixtures.map(
          (fixture) => `#${fixture.id} · group ${fixture.scheduleGroupId} · clubs ${fixture.homeClubId}/${fixture.awayClubId}`,
        )}
        loading={loading}
      />
    </section>
  );
}

function MatchPanel({
  matches,
  loading,
  createMatch,
  createSuccess,
  createError,
}: {
  matches: Match[];
  loading: boolean;
  createMatch: (input: {
    fixtureId: number;
    homeTeamId: number;
    awayTeamId: number;
    state: string;
    scheduledFor?: string;
  }) => Promise<unknown>;
  createSuccess: boolean;
  createError: Error | null;
}) {
  const [fixtureId, setFixtureId] = useState(1);
  const [homeTeamId, setHomeTeamId] = useState(101);
  const [awayTeamId, setAwayTeamId] = useState(102);
  const [state, setState] = useState("scheduled");
  const [scheduledFor, setScheduledFor] = useState("2026-02-16T20:00:00Z");

  async function submit(event: FormEvent<HTMLFormElement>) {
    event.preventDefault();
    await createMatch({ fixtureId, homeTeamId, awayTeamId, state, scheduledFor });
  }

  return (
    <section>
      <h3>Matches</h3>
      <form onSubmit={submit}>
        <label>
          Fixture ID
          <input min={1} type="number" value={fixtureId} onChange={(event) => setFixtureId(Number(event.target.value))} />
        </label>
        <label>
          Home team ID
          <input
            min={1}
            type="number"
            value={homeTeamId}
            onChange={(event) => setHomeTeamId(Number(event.target.value))}
          />
        </label>
        <label>
          Away team ID
          <input
            min={1}
            type="number"
            value={awayTeamId}
            onChange={(event) => setAwayTeamId(Number(event.target.value))}
          />
        </label>
        <label>
          State
          <input value={state} onChange={(event) => setState(event.target.value)} />
        </label>
        <label>
          Scheduled for (RFC3339)
          <input value={scheduledFor} onChange={(event) => setScheduledFor(event.target.value)} />
        </label>
        <button type="submit">Create match</button>
      </form>
      {createSuccess && <p data-testid="admin-match-success">Match created.</p>}
      {createError && <ErrorView error={createError} label="Match create failed" />}
      <DataList
        items={matches.map((match) => `#${match.id} · fixture ${match.fixtureId} · teams ${match.homeTeamId}/${match.awayTeamId}`)}
        loading={loading}
      />
    </section>
  );
}

function DataList({ items, loading }: { items: string[]; loading: boolean }) {
  const [filter, setFilter] = useState("");
  const [page, setPage] = useState(1);
  const pageSize = 5;

  const filteredItems = useMemo(() => {
    const normalized = filter.trim().toLowerCase();
    if (!normalized) {
      return items;
    }
    return items.filter((item) => item.toLowerCase().includes(normalized));
  }, [items, filter]);

  const pageCount = Math.max(1, Math.ceil(filteredItems.length / pageSize));
  const safePage = Math.min(page, pageCount);
  const pagedItems = filteredItems.slice((safePage - 1) * pageSize, safePage * pageSize);

  if (loading) {
    return <LoadingState />;
  }

  return (
    <div>
      <label>
        Filter list
        <input value={filter} onChange={(event) => setFilter(event.target.value)} />
      </label>
      <ul style={{ margin: 0, paddingLeft: "1.25rem" }}>
        {pagedItems.map((item) => (
          <li key={item}>{item}</li>
        ))}
      </ul>
      <div style={{ display: "flex", gap: "0.5rem" }}>
        <button disabled={safePage <= 1} onClick={() => setPage((value) => Math.max(1, value - 1))} type="button">
          Prev
        </button>
        <span>
          Page {safePage} / {pageCount}
        </span>
        <button
          disabled={safePage >= pageCount}
          onClick={() => setPage((value) => Math.min(pageCount, value + 1))}
          type="button"
        >
          Next
        </button>
      </div>
    </div>
  );
}

function ErrorView({ error, label }: { error: Error; label: string }) {
  if (error instanceof ApiError) {
    return (
      <p role="alert">
        {label} ({error.status}){error.code ? ` [${error.code}]` : ""}: {error.message}
      </p>
    );
  }

  return <p role="alert">{label}: {error.message}</p>;
}
