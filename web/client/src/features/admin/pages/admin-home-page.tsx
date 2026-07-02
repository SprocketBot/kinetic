import { useEffect, useMemo, useState } from "react";
import type { FormEvent } from "react";
import { useMutation, useQuery, useQueryClient } from "@tanstack/react-query";

import { AppShell } from "../../../components/layout/app-shell";
import { LoadingState } from "../../../components/feedback/loading-state";
import { getJson, postJson } from "../../../lib/api/client";
import { ApiError } from "../../../lib/api/errors";
import {
  adjustPlayerRatingInputSchema,
  assignRoleInputSchema,
  createFixtureInputSchema,
  createMatchInputSchema,
  createRosterMembershipInputSchema,
  createScheduleGroupInputSchema,
  createSeasonInputSchema,
  overrideResultSubmissionInputSchema,
  resultOverrideListSchema,
  resultSubmissionListSchema,
  ratingAdjustmentListSchema,
  playerRatingListSchema,
  exceptionActionListSchema,
  fixtureListSchema,
  fixtureSchema,
  matchListSchema,
  matchSchema,
  playerListSchema,
  rosterMembershipListSchema,
  rosterMembershipSchema,
  revokeRoleInputSchema,
  roleAssignmentListSchema,
  roleAssignmentSchema,
  scheduleGroupListSchema,
  scheduleGroupSchema,
  seasonListSchema,
  seasonSchema,
  teamListSchema,
  type ExceptionAction,
  type Fixture,
  type Match,
  type Player,
  type RosterMembership,
  type RoleAssignment,
  type ResultSubmission,
  type ResultOverride,
  type ScheduleGroup,
  type Season,
  type Team,
  type PlayerRating,
  type RatingAdjustment,
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

async function listPlayers() {
  return getJson("/v1/players", playerListSchema);
}

async function listTeams() {
  return getJson("/v1/teams", teamListSchema);
}

async function listRosterMemberships() {
  return getJson("/v1/roster-memberships", rosterMembershipListSchema);
}

async function createRosterMembership(input: { playerId: number; teamId: number }) {
  const payload = createRosterMembershipInputSchema.parse(input);
  return postJson("/v1/roster-memberships", payload, rosterMembershipSchema);
}

async function listRoleAssignments() {
  return getJson("/v1/role-assignments", roleAssignmentListSchema);
}

async function assignRole(input: {
  actorPlayerId: number;
  targetPlayerId: number;
  role: "fm" | "gm" | "agm" | "captain";
  franchiseId?: number;
  clubId?: number;
  teamId?: number;
  reason: string;
}) {
  const payload = assignRoleInputSchema.parse(input);
  return postJson("/v1/role-assignments", payload, roleAssignmentSchema);
}

async function revokeRole(input: { actorPlayerId: number; assignmentId: number; reason: string }) {
  const payload = revokeRoleInputSchema.parse(input);
  return postJson("/v1/role-assignments/revoke", payload, roleAssignmentSchema);
}

async function listExceptionActions() {
  return getJson("/v1/exception-actions", exceptionActionListSchema);
}

async function listPlayerRatings() {
  return getJson("/v1/player-ratings", playerRatingListSchema);
}

async function listResultSubmissions() {
  return getJson("/v1/result-submissions", resultSubmissionListSchema);
}

async function listResultOverrides() {
  return getJson("/v1/result-overrides", resultOverrideListSchema);
}

async function listRatingAdjustments() {
  return getJson("/v1/rating-adjustments", ratingAdjustmentListSchema);
}

async function adjustPlayerRating(input: {
  actorPlayerId: number;
  targetPlayerId: number;
  contextKey: string;
  rating: number;
  uncertainty: number;
  matchesPlayed: number;
  reason: string;
}) {
  const payload = adjustPlayerRatingInputSchema.parse(input);
  return postJson("/v1/player-ratings/adjust", payload, playerRatingListSchema.element);
}

async function overrideResultSubmission(input: {
  submissionId: number;
  actor: string;
  reason: string;
  winningTeamId: number;
  losingTeamId: number;
}) {
  const payload = overrideResultSubmissionInputSchema.parse(input);
  return postJson("/v1/result-overrides", payload, resultSubmissionListSchema.element);
}

export function AdminHomePage() {
  const queryClient = useQueryClient();

  const seasonsQuery = useQuery({ queryKey: ["seasons"], queryFn: listSeasons });
  const scheduleGroupsQuery = useQuery({ queryKey: ["schedule-groups"], queryFn: listScheduleGroups });
  const fixturesQuery = useQuery({ queryKey: ["fixtures"], queryFn: listFixtures });
  const matchesQuery = useQuery({ queryKey: ["matches"], queryFn: listMatches });

  const playersQuery = useQuery({ queryKey: ["players"], queryFn: listPlayers });
  const teamsQuery = useQuery({ queryKey: ["teams"], queryFn: listTeams });
  const rosterQuery = useQuery({ queryKey: ["roster-memberships"], queryFn: listRosterMemberships });
  const roleAssignmentsQuery = useQuery({ queryKey: ["role-assignments"], queryFn: listRoleAssignments });
  const actionsQuery = useQuery({ queryKey: ["exception-actions"], queryFn: listExceptionActions });
  const ratingsQuery = useQuery({ queryKey: ["player-ratings"], queryFn: listPlayerRatings });
  const resultSubmissionsQuery = useQuery({ queryKey: ["result-submissions"], queryFn: listResultSubmissions });
  const resultOverridesQuery = useQuery({ queryKey: ["result-overrides"], queryFn: listResultOverrides });
  const ratingAdjustmentsQuery = useQuery({ queryKey: ["rating-adjustments"], queryFn: listRatingAdjustments });

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

  const rosterMutation = useMutation({
    mutationFn: createRosterMembership,
    onSuccess: async () => {
      await queryClient.invalidateQueries({ queryKey: ["roster-memberships"] });
    },
  });

  const roleAssignMutation = useMutation({
    mutationFn: assignRole,
    onSuccess: async () => {
      await queryClient.invalidateQueries({ queryKey: ["role-assignments"] });
    },
  });

  const roleRevokeMutation = useMutation({
    mutationFn: revokeRole,
    onSuccess: async () => {
      await queryClient.invalidateQueries({ queryKey: ["role-assignments"] });
    },
  });

  const ratingAdjustmentMutation = useMutation({
    mutationFn: adjustPlayerRating,
    onSuccess: async () => {
      await Promise.all([
        queryClient.invalidateQueries({ queryKey: ["player-ratings"] }),
        queryClient.invalidateQueries({ queryKey: ["rating-adjustments"] }),
      ]);
    },
  });

  const overrideMutation = useMutation({
    mutationFn: overrideResultSubmission,
    onSuccess: async () => {
      await Promise.all([
        queryClient.invalidateQueries({ queryKey: ["result-submissions"] }),
        queryClient.invalidateQueries({ queryKey: ["result-overrides"] }),
      ]);
    },
  });

  return (
    <AppShell title="Kinetic Web Client">
      <h2>League Admin</h2>
      <p>Schedule lifecycle management and roster administration.</p>
      <p>Role delegation grants/revokes are now API-backed with scoped server enforcement.</p>

      <section className="layout-grid layout-grid--2">
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

      <section className="layout-grid layout-grid--2">
        <RosterPanel
          createError={rosterMutation.error}
          createMembership={rosterMutation.mutateAsync}
          createSuccess={rosterMutation.isSuccess}
          loading={playersQuery.isLoading || teamsQuery.isLoading || rosterQuery.isLoading}
          memberships={rosterQuery.data ?? []}
          players={playersQuery.data ?? []}
          teams={teamsQuery.data ?? []}
        />

        <RoleDelegationPanel
          actions={actionsQuery.data ?? []}
          assignError={roleAssignMutation.error}
          assignRole={roleAssignMutation.mutateAsync}
          assignSuccess={roleAssignMutation.isSuccess}
          assignments={roleAssignmentsQuery.data ?? []}
          loading={actionsQuery.isLoading || roleAssignmentsQuery.isLoading}
          revokeError={roleRevokeMutation.error}
          revokeRole={roleRevokeMutation.mutateAsync}
          revokeSuccess={roleRevokeMutation.isSuccess}
        />
      </section>

      <section className="layout-grid layout-grid--2">
        <ResultOverridePanel
          loading={resultSubmissionsQuery.isLoading || resultOverridesQuery.isLoading}
          overrideError={overrideMutation.error}
          overrideResult={overrideMutation.mutateAsync}
          overrideSuccess={overrideMutation.isSuccess}
          overrides={resultOverridesQuery.data ?? []}
          submissions={resultSubmissionsQuery.data ?? []}
        />
        <RatingAdminPanel
          adjustError={ratingAdjustmentMutation.error}
          adjustRating={ratingAdjustmentMutation.mutateAsync}
          adjustSuccess={ratingAdjustmentMutation.isSuccess}
          adjustments={ratingAdjustmentsQuery.data ?? []}
          loading={ratingsQuery.isLoading || ratingAdjustmentsQuery.isLoading}
          ratings={ratingsQuery.data ?? []}
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

function RosterPanel({
  players,
  teams,
  memberships,
  loading,
  createMembership,
  createSuccess,
  createError,
}: {
  players: Player[];
  teams: Team[];
  memberships: RosterMembership[];
  loading: boolean;
  createMembership: (input: { playerId: number; teamId: number }) => Promise<unknown>;
  createSuccess: boolean;
  createError: Error | null;
}) {
  const [playerId, setPlayerId] = useState(1);
  const [teamId, setTeamId] = useState(1);

  const playerNameById = new Map(players.map((player) => [player.id, player.displayName]));
  const teamNameById = new Map(teams.map((team) => [team.id, team.name]));

  async function submit(event: FormEvent<HTMLFormElement>) {
    event.preventDefault();
    await createMembership({ playerId, teamId });
  }

  return (
    <section>
      <h3>Roster Memberships</h3>
      <form onSubmit={submit}>
        <label>
          Player ID
          <input min={1} type="number" value={playerId} onChange={(event) => setPlayerId(Number(event.target.value))} />
        </label>
        <label>
          Team ID
          <input min={1} type="number" value={teamId} onChange={(event) => setTeamId(Number(event.target.value))} />
        </label>
        <button type="submit">Assign player to team</button>
      </form>
      {createSuccess && <p data-testid="admin-roster-success">Roster membership created.</p>}
      {createError && <ErrorView error={createError} label="Roster assignment failed" />}
      <DataList
        items={memberships.map(
          (membership) =>
            `#${membership.id} · ${playerNameById.get(membership.playerId) ?? membership.playerId} -> ${teamNameById.get(membership.teamId) ?? membership.teamId}`,
        )}
        loading={loading}
      />
    </section>
  );
}

function RoleDelegationPanel({
  actions,
  assignments,
  loading,
  assignRole,
  revokeRole,
  assignSuccess,
  revokeSuccess,
  assignError,
  revokeError,
}: {
  actions: ExceptionAction[];
  assignments: RoleAssignment[];
  loading: boolean;
  assignRole: (input: {
    actorPlayerId: number;
    targetPlayerId: number;
    role: "fm" | "gm" | "agm" | "captain";
    franchiseId?: number;
    clubId?: number;
    teamId?: number;
    reason: string;
  }) => Promise<unknown>;
  revokeRole: (input: { actorPlayerId: number; assignmentId: number; reason: string }) => Promise<unknown>;
  assignSuccess: boolean;
  revokeSuccess: boolean;
  assignError: Error | null;
  revokeError: Error | null;
}) {
  const [actorPlayerId, setActorPlayerId] = useState(1);
  const [targetPlayerId, setTargetPlayerId] = useState(2);
  const [scopeRole, setScopeRole] = useState<"fm" | "gm" | "agm" | "captain">("captain");
  const [franchiseId, setFranchiseId] = useState(1);
  const [clubId, setClubId] = useState(1);
  const [teamId, setTeamId] = useState(1);
  const [assignReason, setAssignReason] = useState("season staffing");
  const [revokeAssignmentId, setRevokeAssignmentId] = useState(1);
  const [revokeReason, setRevokeReason] = useState("scope realignment");

  async function submitAssign(event: FormEvent<HTMLFormElement>) {
    event.preventDefault();
    await assignRole({
      actorPlayerId,
      targetPlayerId,
      role: scopeRole,
      franchiseId: scopeRole === "fm" ? franchiseId : undefined,
      clubId: scopeRole === "gm" || scopeRole === "agm" ? clubId : undefined,
      teamId: scopeRole === "captain" ? teamId : undefined,
      reason: assignReason,
    });
  }

  async function submitRevoke(event: FormEvent<HTMLFormElement>) {
    event.preventDefault();
    await revokeRole({
      actorPlayerId,
      assignmentId: revokeAssignmentId,
      reason: revokeReason,
    });
  }

  function scopeLabel(assignment: RoleAssignment): string {
    if (assignment.franchiseId !== null && assignment.franchiseId !== undefined) {
      return `franchise ${assignment.franchiseId}`;
    }
    if (assignment.clubId !== null && assignment.clubId !== undefined) {
      return `club ${assignment.clubId}`;
    }
    if (assignment.teamId !== null && assignment.teamId !== undefined) {
      return `team ${assignment.teamId}`;
    }
    return "unknown scope";
  }

  return (
    <section>
      <h3>Role Delegation</h3>
      <form onSubmit={submitAssign}>
        <label>
          Actor player ID
          <input
            min={1}
            type="number"
            value={actorPlayerId}
            onChange={(event) => setActorPlayerId(Number(event.target.value))}
          />
        </label>
        <label>
          Target player ID
          <input
            min={1}
            type="number"
            value={targetPlayerId}
            onChange={(event) => setTargetPlayerId(Number(event.target.value))}
          />
        </label>
        <label>
          Role
          <select value={scopeRole} onChange={(event) => setScopeRole(event.target.value as "fm" | "gm" | "agm" | "captain")}>
            <option value="captain">Captain</option>
            <option value="agm">AGM</option>
            <option value="gm">GM</option>
            <option value="fm">FM</option>
          </select>
        </label>
        {scopeRole === "fm" && (
          <label>
            Franchise ID
            <input min={1} type="number" value={franchiseId} onChange={(event) => setFranchiseId(Number(event.target.value))} />
          </label>
        )}
        {(scopeRole === "gm" || scopeRole === "agm") && (
          <label>
            Club ID
            <input min={1} type="number" value={clubId} onChange={(event) => setClubId(Number(event.target.value))} />
          </label>
        )}
        {scopeRole === "captain" && (
          <label>
            Team ID
            <input min={1} type="number" value={teamId} onChange={(event) => setTeamId(Number(event.target.value))} />
          </label>
        )}
        <label>
          Reason
          <input value={assignReason} onChange={(event) => setAssignReason(event.target.value)} />
        </label>
        <button type="submit">Assign role</button>
      </form>

      <form onSubmit={submitRevoke} style={{ marginTop: "0.6rem" }}>
        <label>
          Revoke assignment ID
          <input
            min={1}
            type="number"
            value={revokeAssignmentId}
            onChange={(event) => setRevokeAssignmentId(Number(event.target.value))}
          />
        </label>
        <label>
          Revoke reason
          <input value={revokeReason} onChange={(event) => setRevokeReason(event.target.value)} />
        </label>
        <button type="submit">Revoke role</button>
      </form>

      {loading && <LoadingState label="Loading role delegation data..." />}
      {assignSuccess && <p data-testid="admin-role-assign-success">Role assignment created.</p>}
      {revokeSuccess && <p data-testid="admin-role-revoke-success">Role assignment revoked.</p>}
      {assignError && <ErrorView error={assignError} label="Role assignment failed" />}
      {revokeError && <ErrorView error={revokeError} label="Role revoke failed" />}

      <h4>Current Role Assignments</h4>
      {!loading && assignments.length === 0 && <p>No role assignments found.</p>}
      {!loading && assignments.length > 0 && (
        <ul style={{ margin: 0, paddingLeft: "1.2rem" }}>
          {assignments.map((assignment) => (
            <li key={assignment.id}>
              #{assignment.id} · player {assignment.playerId} · {assignment.role} · {scopeLabel(assignment)} ·{" "}
              {assignment.isActive ? "active" : "inactive"}
            </li>
          ))}
        </ul>
      )}

      <h4>Recent Operator Actions (Audit Feed)</h4>
      <DataList
        items={actions.map((action) => `#${action.id} · ${action.actionType} · ${action.actor} · ${action.createdAt}`)}
        loading={loading}
      />
    </section>
  );
}

function ResultOverridePanel({
  submissions,
  overrides,
  loading,
  overrideResult,
  overrideSuccess,
  overrideError,
}: {
  submissions: ResultSubmission[];
  overrides: ResultOverride[];
  loading: boolean;
  overrideResult: (input: {
    submissionId: number;
    actor: string;
    reason: string;
    winningTeamId: number;
    losingTeamId: number;
  }) => Promise<unknown>;
  overrideSuccess: boolean;
  overrideError: Error | null;
}) {
  const [submissionId, setSubmissionId] = useState(1);
  const [actor, setActor] = useState("league-admin");
  const [overrideReason, setOverrideReason] = useState("official review correction");
  const [winningTeamId, setWinningTeamId] = useState(1);
  const [losingTeamId, setLosingTeamId] = useState(2);
  const selectedSubmission = submissions.find((entry) => entry.id === submissionId) ?? submissions[0] ?? null;
  const selectedSubmissionId = selectedSubmission?.id ?? null;
  const selectedWinningTeamId = selectedSubmission?.winningTeamId ?? null;
  const selectedLosingTeamId = selectedSubmission?.losingTeamId ?? null;

  useEffect(() => {
    if (selectedSubmissionId === null || selectedWinningTeamId === null || selectedLosingTeamId === null) {
      return;
    }
    setSubmissionId(selectedSubmissionId);
    setWinningTeamId(selectedWinningTeamId);
    setLosingTeamId(selectedLosingTeamId);
  }, [selectedSubmissionId, selectedWinningTeamId, selectedLosingTeamId]);

  async function submit(event: FormEvent<HTMLFormElement>) {
    event.preventDefault();
    await overrideResult({ submissionId, actor, reason: overrideReason, winningTeamId, losingTeamId });
  }

  return (
    <section>
      <h3>Result Overrides (NCP)</h3>
      <p>Use this to apply an audited admin override for an existing submission.</p>
      <form onSubmit={submit}>
        <label>
          Submission ID
          <input
            min={1}
            type="number"
            value={submissionId}
            onChange={(event) => setSubmissionId(Number(event.target.value))}
          />
        </label>
        <label>
          Actor
          <input value={actor} onChange={(event) => setActor(event.target.value)} />
        </label>
        <label>
          Override reason
          <input value={overrideReason} onChange={(event) => setOverrideReason(event.target.value)} />
        </label>
        <label>
          Winning team ID
          <input
            min={1}
            type="number"
            value={winningTeamId}
            onChange={(event) => setWinningTeamId(Number(event.target.value))}
          />
        </label>
        <label>
          Losing team ID
          <input
            min={1}
            type="number"
            value={losingTeamId}
            onChange={(event) => setLosingTeamId(Number(event.target.value))}
          />
        </label>
        <button type="submit">Submit override</button>
      </form>
      {overrideSuccess && <p data-testid="admin-override-success">Override submitted.</p>}
      {overrideError && <ErrorView error={overrideError} label="Result override failed" />}
      <DataList
        items={[
          ...submissions.map(
            (submission) =>
              `Submission #${submission.id} · ${submission.homeTeamId}/${submission.awayTeamId} · winner ${submission.winningTeamId}`,
          ),
          ...overrides.map(
            (override) =>
              `Override #${override.id} · submission ${override.submissionId} · ${override.previousWinningTeamId} -> ${override.newWinningTeamId}`,
          ),
        ]}
        loading={loading}
      />
    </section>
  );
}

function RatingAdminPanel({
  ratings,
  adjustments,
  loading,
  adjustRating,
  adjustSuccess,
  adjustError,
}: {
  ratings: PlayerRating[];
  adjustments: RatingAdjustment[];
  loading: boolean;
  adjustRating: (input: {
    actorPlayerId: number;
    targetPlayerId: number;
    contextKey: string;
    rating: number;
    uncertainty: number;
    matchesPlayed: number;
    reason: string;
  }) => Promise<unknown>;
  adjustSuccess: boolean;
  adjustError: Error | null;
}) {
  const [actorPlayerId, setActorPlayerId] = useState(1);
  const [targetPlayerId, setTargetPlayerId] = useState(2);
  const [contextKey, setContextKey] = useState("scrim-3v3");
  const [rating, setRating] = useState(1000);
  const [uncertainty, setUncertainty] = useState(250);
  const [matchesPlayed, setMatchesPlayed] = useState(10);
  const [reason, setReason] = useState("admin correction");

  async function submit(event: FormEvent<HTMLFormElement>) {
    event.preventDefault();
    await adjustRating({
      actorPlayerId,
      targetPlayerId,
      contextKey,
      rating,
      uncertainty,
      matchesPlayed,
      reason,
    });
  }

  return (
    <section>
      <h3>Rating Administration</h3>
      <p>Admins can adjust other players&apos; ratings. Self-edits are blocked server-side.</p>
      <form onSubmit={submit}>
        <label>
          Actor player ID
          <input
            min={1}
            type="number"
            value={actorPlayerId}
            onChange={(event) => setActorPlayerId(Number(event.target.value))}
          />
        </label>
        <label>
          Target player ID
          <input
            min={1}
            type="number"
            value={targetPlayerId}
            onChange={(event) => setTargetPlayerId(Number(event.target.value))}
          />
        </label>
        <label>
          Context key
          <input value={contextKey} onChange={(event) => setContextKey(event.target.value)} />
        </label>
        <label>
          New rating
          <input min={0} type="number" value={rating} onChange={(event) => setRating(Number(event.target.value))} />
        </label>
        <label>
          Uncertainty
          <input
            min={0}
            type="number"
            value={uncertainty}
            onChange={(event) => setUncertainty(Number(event.target.value))}
          />
        </label>
        <label>
          Matches played
          <input
            min={0}
            type="number"
            value={matchesPlayed}
            onChange={(event) => setMatchesPlayed(Number(event.target.value))}
          />
        </label>
        <label>
          Reason
          <input value={reason} onChange={(event) => setReason(event.target.value)} />
        </label>
        <button type="submit">Apply rating change</button>
      </form>
      {adjustSuccess && <p data-testid="admin-rating-success">Rating adjusted.</p>}
      {adjustError && <ErrorView error={adjustError} label="Rating adjustment failed" />}
      <DataList
        items={[
          ...ratings.map((entry) => `Player ${entry.playerId} · ${entry.contextKey} · ${entry.rating}`),
          ...adjustments.map(
            (entry) =>
              `Audit #${entry.id} · actor ${entry.actorPlayerId} -> player ${entry.targetPlayerId} · ${entry.previousRating} -> ${entry.newRating}`,
          ),
        ]}
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
