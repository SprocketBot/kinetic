import { useEffect, useMemo, useState } from "react";
import type { FormEvent } from "react";
import { useMutation, useQuery, useQueryClient } from "@tanstack/react-query";

import { useSession } from "../../../auth/session-context";
import { AppShell } from "../../../components/layout/app-shell";
import { LoadingState } from "../../../components/feedback/loading-state";
import { deleteJson, getJson, postJson } from "../../../lib/api/client";
import { ApiError } from "../../../lib/api/errors";
import { env } from "../../../lib/config/env";
import {
  enqueueTeamInputSchema,
  eligibilityStatusSchema,
  createResultSubmissionInputSchema,
  ingestReplayEvidenceInputSchema,
  linkPlatformAccountInputSchema,
  leaveQueueInputSchema,
  platformAccountLinkListSchema,
  platformAccountLinkSchema,
  playerRatingListSchema,
  queueEntryListSchema,
  queueEntrySchema,
  ratifyResultSubmissionInputSchema,
  rejectResultSubmissionInputSchema,
  replayIngestionResultSchema,
  resultSubmissionListSchema,
  resultSubmissionSchema,
  scrimListSchema,
  unlinkPlatformAccountInputSchema,
  type EligibilityStatus,
  type CreateResultSubmissionInput,
  type IngestReplayEvidenceInput,
  type PlatformAccountLink,
  type QueueEntry,
  type ResultSubmission,
  type ReplayIngestionResult,
  type Scrim,
  type PlayerRating,
} from "../../../lib/api/schemas";

async function listQueueEntries() {
  return getJson("/v1/queue-entries", queueEntryListSchema);
}

async function listScrims() {
  return getJson("/v1/scrims", scrimListSchema);
}

async function listResultSubmissions() {
  return getJson("/v1/result-submissions", resultSubmissionListSchema);
}

async function listPlayerRatings() {
  return getJson("/v1/player-ratings", playerRatingListSchema);
}

async function listPlatformAccounts(subject: string) {
  return getJson(`/v1/platform-accounts?subject=${encodeURIComponent(subject)}`, platformAccountLinkListSchema);
}

async function getEligibilityStatus(subject: string) {
  return getJson(`/v1/eligibility?subject=${encodeURIComponent(subject)}`, eligibilityStatusSchema);
}

async function enqueueTeam(input: { queueId: number; teamId: number }) {
  const payload = enqueueTeamInputSchema.parse(input);
  return postJson("/v1/queue-entries", payload, queueEntrySchema);
}

async function leaveQueue(input: { queueId: number; teamId: number }) {
  const payload = leaveQueueInputSchema.parse(input);
  return deleteJson("/v1/queue-entries", payload, queueEntrySchema);
}

async function createResultSubmission(input: CreateResultSubmissionInput) {
  const payload = createResultSubmissionInputSchema.parse(input);
  return postJson("/v1/result-submissions", payload, resultSubmissionSchema);
}

async function ratifySubmission(input: { submissionId: number; teamId: number }) {
  const payload = ratifyResultSubmissionInputSchema.parse(input);
  return postJson("/v1/result-submission-ratifications", payload, resultSubmissionSchema);
}

async function rejectSubmission(input: { submissionId: number; teamId: number; reason: string }) {
  const payload = rejectResultSubmissionInputSchema.parse(input);
  return postJson("/v1/result-submission-rejections", payload, resultSubmissionSchema);
}

async function ingestReplay(input: IngestReplayEvidenceInput) {
  const payload = ingestReplayEvidenceInputSchema.parse(input);
  return postJson("/v1/replay-evidence", payload, replayIngestionResultSchema);
}

async function linkPlatformAccount(input: {
  subject: string;
  provider: "steam" | "xbox" | "psn" | "epic";
  providerAccountId: string;
  providerAccountName: string;
}) {
  const payload = linkPlatformAccountInputSchema.parse(input);
  return postJson("/v1/platform-accounts/link", payload, platformAccountLinkSchema);
}

async function unlinkPlatformAccount(input: {
  subject: string;
  provider: "steam" | "xbox" | "psn" | "epic";
  providerAccountId: string;
}) {
  const payload = unlinkPlatformAccountInputSchema.parse(input);
  return postJson("/v1/platform-accounts/unlink", payload, platformAccountLinkSchema);
}

function activeScrim(scrim: Scrim): boolean {
  return !["cancelled", "completed", "closed", "ended"].includes(scrim.state.toLowerCase());
}

function inProgressSubmission(submission: ResultSubmission): boolean {
  return !["accepted", "finalized", "rejected"].includes(submission.state.toLowerCase());
}

function activeQueueEntry(entry: QueueEntry): boolean {
  return entry.isActive;
}

const evidenceViews = [
  { id: "standings", title: "Standings", path: "/standings" },
  { id: "ratings", title: "Ratings", path: "/ratings" },
  { id: "eligibility", title: "Eligibility", path: "/eligibility" },
];

export function PlayerHomePage() {
  const session = useSession();
  const queryClient = useQueryClient();
  const [activeWorkspace, setActiveWorkspace] = useState<"overview" | "actions" | "profile" | "evidence">("overview");
  const [selectedEvidenceView, setSelectedEvidenceView] = useState(evidenceViews[0].id);
  const sessionSubject = session.principal?.subject ?? "";

  const queueQuery = useQuery({ queryKey: ["queue-entries"], queryFn: listQueueEntries });
  const scrimsQuery = useQuery({ queryKey: ["scrims"], queryFn: listScrims });
  const submissionsQuery = useQuery({ queryKey: ["result-submissions"], queryFn: listResultSubmissions });
  const ratingsQuery = useQuery({ queryKey: ["player-ratings"], queryFn: listPlayerRatings });
  const platformAccountsQuery = useQuery({
    queryKey: ["platform-accounts", sessionSubject],
    queryFn: () => listPlatformAccounts(sessionSubject),
    enabled: sessionSubject.length > 0,
  });
  const eligibilityQuery = useQuery({
    queryKey: ["eligibility", sessionSubject],
    queryFn: () => getEligibilityStatus(sessionSubject),
    enabled: sessionSubject.length > 0,
  });

  const activeQueueEntries = useMemo(() => (queueQuery.data ?? []).filter(activeQueueEntry), [queueQuery.data]);
  const activeScrims = useMemo(() => (scrimsQuery.data ?? []).filter(activeScrim), [scrimsQuery.data]);
  const inProcessSubmissions = useMemo(
    () => (submissionsQuery.data ?? []).filter(inProgressSubmission),
    [submissionsQuery.data],
  );

  const enqueueMutation = useMutation({
    mutationFn: enqueueTeam,
    onSuccess: async () => {
      await queryClient.invalidateQueries({ queryKey: ["queue-entries"] });
    },
  });

  const leaveQueueMutation = useMutation({
    mutationFn: leaveQueue,
    onSuccess: async () => {
      await queryClient.invalidateQueries({ queryKey: ["queue-entries"] });
    },
  });

  const createSubmissionMutation = useMutation({
    mutationFn: createResultSubmission,
    onSuccess: async () => {
      await queryClient.invalidateQueries({ queryKey: ["result-submissions"] });
    },
  });

  const ratifyMutation = useMutation({
    mutationFn: ratifySubmission,
    onSuccess: async () => {
      await queryClient.invalidateQueries({ queryKey: ["result-submissions"] });
    },
  });

  const rejectMutation = useMutation({
    mutationFn: rejectSubmission,
    onSuccess: async () => {
      await queryClient.invalidateQueries({ queryKey: ["result-submissions"] });
    },
  });

  const replayMutation = useMutation({
    mutationFn: ingestReplay,
    onSuccess: async () => {
      await queryClient.invalidateQueries({ queryKey: ["result-submissions"] });
    },
  });

  const linkPlatformMutation = useMutation({
    mutationFn: linkPlatformAccount,
    onSuccess: async () => {
      await queryClient.invalidateQueries({ queryKey: ["platform-accounts", sessionSubject] });
    },
  });

  const unlinkPlatformMutation = useMutation({
    mutationFn: unlinkPlatformAccount,
    onSuccess: async () => {
      await queryClient.invalidateQueries({ queryKey: ["platform-accounts", sessionSubject] });
    },
  });

  return (
    <AppShell title="Kinetic Web Client">
      <h2>Player</h2>
      <p>Queue, scrim, replay submission, and ratification workflows.</p>

      <nav aria-label="Player workspace" className="workspace-tabs">
        <button
          aria-pressed={activeWorkspace === "overview"}
          className={activeWorkspace === "overview" ? "workspace-tabs__tab active" : "workspace-tabs__tab"}
          onClick={() => setActiveWorkspace("overview")}
          type="button"
        >
          Overview
        </button>
        <button
          aria-pressed={activeWorkspace === "actions"}
          className={activeWorkspace === "actions" ? "workspace-tabs__tab active" : "workspace-tabs__tab"}
          onClick={() => setActiveWorkspace("actions")}
          type="button"
        >
          Actions
        </button>
        <button
          aria-pressed={activeWorkspace === "profile"}
          className={activeWorkspace === "profile" ? "workspace-tabs__tab active" : "workspace-tabs__tab"}
          onClick={() => setActiveWorkspace("profile")}
          type="button"
        >
          Profile
        </button>
        <button
          aria-pressed={activeWorkspace === "evidence"}
          className={activeWorkspace === "evidence" ? "workspace-tabs__tab active" : "workspace-tabs__tab"}
          onClick={() => setActiveWorkspace("evidence")}
          type="button"
        >
          Evidence
        </button>
      </nav>

      <section className="layout-grid layout-grid--3">
        <CountCard label="Queue entries" testId="player-queue-count" value={activeQueueEntries.length} />
        <CountCard label="Active scrims" testId="player-scrim-count" value={activeScrims.length} />
        <CountCard label="Submissions in progress" testId="player-submission-count" value={inProcessSubmissions.length} />
      </section>

      {activeWorkspace === "overview" && (
        <section className="layout-grid layout-grid--3 tab-panel">
          <DataList
            items={activeQueueEntries.map((entry) => `Entry #${entry.id} · queue ${entry.queueId} · team ${entry.teamId}`)}
            loading={queueQuery.isLoading}
            title="Active Queue Entries"
          />
          <DataList
            items={activeScrims.map((scrim) => `Scrim #${scrim.id} · queue ${scrim.queueId} · ${scrim.state}`)}
            loading={scrimsQuery.isLoading}
            title="Active Scrims"
          />
          <DataList
            items={inProcessSubmissions.map(
              (submission) =>
                `Submission #${submission.id} · ${submission.contextType}:${submission.contextId} · ${submission.state}`,
            )}
            loading={submissionsQuery.isLoading}
            title="Submissions In Progress"
          />
        </section>
      )}

      {activeWorkspace === "actions" && (
        <section className="layout-grid layout-grid--2 tab-panel">
          <QueueActions
            onEnqueue={enqueueMutation.mutateAsync}
            onLeave={leaveQueueMutation.mutateAsync}
            status={[enqueueMutation, leaveQueueMutation]}
          />
          <SubmissionActions
            defaultScrim={activeScrims[0] ?? null}
            onCreateSubmission={createSubmissionMutation.mutateAsync}
            onRatify={ratifyMutation.mutateAsync}
            onReject={rejectMutation.mutateAsync}
            onReplayUpload={replayMutation.mutateAsync}
            status={[createSubmissionMutation, ratifyMutation, rejectMutation, replayMutation]}
          />
        </section>
      )}

      {activeWorkspace === "profile" && (
        <section className="layout-grid layout-grid--2 tab-panel">
          <RatingsPanel loading={ratingsQuery.isLoading} ratings={ratingsQuery.data ?? []} />
          <AccountAndEligibilityPanel
            eligibility={eligibilityQuery.data ?? null}
            eligibilityError={eligibilityQuery.error}
            eligibilityLoading={session.status === "loading" || eligibilityQuery.isLoading}
            links={platformAccountsQuery.data ?? []}
            loading={session.status === "loading" || platformAccountsQuery.isLoading}
            onLink={linkPlatformMutation.mutateAsync}
            onUnlink={unlinkPlatformMutation.mutateAsync}
            queryError={platformAccountsQuery.error}
            status={[linkPlatformMutation, unlinkPlatformMutation]}
            subject={sessionSubject}
          />
        </section>
      )}

      {activeWorkspace === "evidence" && (
        <div className="tab-panel">
          <EvidenceSection selected={selectedEvidenceView} setSelected={setSelectedEvidenceView} />
        </div>
      )}

      {[queueQuery.error, scrimsQuery.error, submissionsQuery.error, ratingsQuery.error].map((error, index) =>
        error ? <ErrorView error={error} key={`query-error-${index}`} label="Player data request failed" /> : null,
      )}
    </AppShell>
  );
}

function CountCard({ label, value, testId }: { label: string; value: number; testId: string }) {
  return (
    <article>
      <h3 style={{ marginTop: 0 }}>{label}</h3>
      <p data-testid={testId}>
        {value}
      </p>
    </article>
  );
}

function QueueActions({
  onEnqueue,
  onLeave,
  status,
}: {
  onEnqueue: (input: { queueId: number; teamId: number }) => Promise<unknown>;
  onLeave: (input: { queueId: number; teamId: number }) => Promise<unknown>;
  status: Array<{ isPending: boolean; isSuccess: boolean; error: Error | null }>;
}) {
  const [queueId, setQueueId] = useState(1);
  const [teamId, setTeamId] = useState(1);

  async function submitEnqueue(event: FormEvent<HTMLFormElement>) {
    event.preventDefault();
    await onEnqueue({ queueId, teamId });
  }

  async function submitLeave() {
    await onLeave({ queueId, teamId });
  }

  return (
    <section>
      <h3>Queue Actions</h3>
      <p>Use the same queue and team values to join or leave the active queue.</p>
      <form onSubmit={submitEnqueue}>
        <label>
          Queue ID
          <input min={1} type="number" value={queueId} onChange={(event) => setQueueId(Number(event.target.value))} />
        </label>
        <label>
          Team ID
          <input min={1} type="number" value={teamId} onChange={(event) => setTeamId(Number(event.target.value))} />
        </label>
        <div className="form-actions">
          <button type="submit">Join queue</button>
          <button onClick={() => void submitLeave()} type="button">
            Leave queue
          </button>
        </div>
      </form>

      {status.some((item) => item.isPending) && <LoadingState label="Saving queue action..." />}
      {status.some((item) => item.isSuccess) && <p data-testid="player-queue-success">Queue action submitted.</p>}
      {status.map((item, index) =>
        item.error ? <ErrorView error={item.error} key={`queue-error-${index}`} label="Queue action failed" /> : null,
      )}
    </section>
  );
}

function scoreFromPayload(value: unknown): { home: number; away: number } | null {
  if (!value || typeof value !== "object") return null;
  const payload = value as { score?: unknown; homeScore?: unknown; awayScore?: unknown };
  if (payload.score && typeof payload.score === "object") {
    const score = payload.score as { home?: unknown; away?: unknown };
    if (typeof score.home === "number" && typeof score.away === "number") {
      return { home: score.home, away: score.away };
    }
  }
  if (typeof payload.score === "string") {
    const [home, away] = payload.score.split("-").map((part) => Number(part.trim()));
    if (Number.isFinite(home) && Number.isFinite(away)) return { home, away };
  }
  if (typeof payload.homeScore === "number" && typeof payload.awayScore === "number") {
    return { home: payload.homeScore, away: payload.awayScore };
  }
  return null;
}

function SubmissionActions({
  defaultScrim,
  onCreateSubmission,
  onRatify,
  onReject,
  onReplayUpload,
  status,
}: {
  defaultScrim: Scrim | null;
  onCreateSubmission: (input: CreateResultSubmissionInput) => Promise<unknown>;
  onRatify: (input: { submissionId: number; teamId: number }) => Promise<unknown>;
  onReject: (input: { submissionId: number; teamId: number; reason: string }) => Promise<unknown>;
  onReplayUpload: (input: IngestReplayEvidenceInput) => Promise<ReplayIngestionResult>;
  status: Array<{ isPending: boolean; isSuccess: boolean; error: Error | null }>;
}) {
  const [activeAction, setActiveAction] = useState<"submit" | "ratify" | "reject" | "replay">("submit");
  const [submissionId, setSubmissionId] = useState(1);
  const [teamId, setTeamId] = useState(1);
  const [rejectReason, setRejectReason] = useState("needs review");

  const [contextType, setContextType] = useState<"scrim" | "match">("scrim");
  const [contextId, setContextId] = useState(1);
  const [homeTeamId, setHomeTeamId] = useState(1);
  const [awayTeamId, setAwayTeamId] = useState(2);
  const [homeScore, setHomeScore] = useState(3);
  const [awayScore, setAwayScore] = useState(1);
  const [homeShots, setHomeShots] = useState(0);
  const [awayShots, setAwayShots] = useState(0);
  const [homeSaves, setHomeSaves] = useState(0);
  const [awaySaves, setAwaySaves] = useState(0);
  const [replayBody, setReplayBody] = useState("placeholder-replay");
  const [lastReplayConflict, setLastReplayConflict] = useState<unknown | null>(null);

  useEffect(() => {
    if (!defaultScrim) return;
    setContextType("scrim");
    setContextId(defaultScrim.id);
    setHomeTeamId(defaultScrim.homeTeamId);
    setAwayTeamId(defaultScrim.awayTeamId);
    setTeamId(defaultScrim.homeTeamId);
  }, [defaultScrim]);

  async function submitResult(event: FormEvent<HTMLFormElement>) {
    event.preventDefault();
    const winningTeamId = homeScore > awayScore ? homeTeamId : awayTeamId;
    const losingTeamId = homeScore > awayScore ? awayTeamId : homeTeamId;
    const payloadJson = {
      score: { home: homeScore, away: awayScore },
      summaryStats: {
        homeShots,
        awayShots,
        homeSaves,
        awaySaves,
      },
    };
    await onCreateSubmission({
      contextType,
      contextId,
      gameKey: "rocket_league",
      submittedByTeamId: teamId,
      winningTeamId,
      losingTeamId,
      payloadJson,
      provenanceJson: {
        fields: {
          contextType: defaultScrim ? "database_default" : "manual",
          contextId: defaultScrim ? "database_default" : "manual",
          homeTeamId: defaultScrim ? "database_default" : "manual",
          awayTeamId: defaultScrim ? "database_default" : "manual",
          score: "manual",
          summaryStats: "manual",
        },
        evidence: {
          replayRequired: false,
          replayAttached: false,
        },
      },
    });
  }

  async function submitRatify(event: FormEvent<HTMLFormElement>) {
    event.preventDefault();
    await onRatify({ submissionId, teamId });
  }

  async function submitReject(event: FormEvent<HTMLFormElement>) {
    event.preventDefault();
    await onReject({ submissionId, teamId, reason: rejectReason });
  }

  async function submitReplay(event: FormEvent<HTMLFormElement>) {
    event.preventDefault();
    const result = await onReplayUpload({
      contextType,
      contextId,
      submittedByTeamId: teamId,
      replayBody,
      parserName: "kinetic-rl-parser",
      parserVersion: "v1",
      parserConfigDigest: "default",
      parseOutputJson: {},
      resultSubmissionId: submissionId,
    });
    const score = scoreFromPayload(result.autofillJson);
    if (score) {
      setHomeScore(score.home);
      setAwayScore(score.away);
    }
    setLastReplayConflict(result.conflictJson ?? null);
  }

  return (
    <section>
      <h3>Submission Actions</h3>
      <div aria-label="Submission action type" className="segmented-control">
        {[
          ["submit", "Submit"],
          ["ratify", "Ratify"],
          ["reject", "Reject"],
          ["replay", "Replay"],
        ].map(([id, label]) => (
          <button
            aria-pressed={activeAction === id}
            className={activeAction === id ? "segmented-control__button active" : "segmented-control__button"}
            key={id}
            onClick={() => setActiveAction(id as "submit" | "ratify" | "reject" | "replay")}
            type="button"
          >
            {label}
          </button>
        ))}
      </div>

      {activeAction === "submit" && (
        <form onSubmit={submitResult}>
          <label>
            Context type
            <select value={contextType} onChange={(event) => setContextType(event.target.value as "scrim" | "match")}>
              <option value="scrim">scrim</option>
              <option value="match">match</option>
            </select>
          </label>
          <label>
            Context ID
            <input min={1} type="number" value={contextId} onChange={(event) => setContextId(Number(event.target.value))} />
          </label>
          <label>
            Reporting team ID
            <input min={1} type="number" value={teamId} onChange={(event) => setTeamId(Number(event.target.value))} />
          </label>
          <div className="layout-grid layout-grid--2 form-grid">
            <label>
              Home team ID
              <input min={1} type="number" value={homeTeamId} onChange={(event) => setHomeTeamId(Number(event.target.value))} />
            </label>
            <label>
              Away team ID
              <input min={1} type="number" value={awayTeamId} onChange={(event) => setAwayTeamId(Number(event.target.value))} />
            </label>
            <label>
              Home score
              <input min={0} type="number" value={homeScore} onChange={(event) => setHomeScore(Number(event.target.value))} />
            </label>
            <label>
              Away score
              <input min={0} type="number" value={awayScore} onChange={(event) => setAwayScore(Number(event.target.value))} />
            </label>
            <label>
              Home shots
              <input min={0} type="number" value={homeShots} onChange={(event) => setHomeShots(Number(event.target.value))} />
            </label>
            <label>
              Away shots
              <input min={0} type="number" value={awayShots} onChange={(event) => setAwayShots(Number(event.target.value))} />
            </label>
            <label>
              Home saves
              <input min={0} type="number" value={homeSaves} onChange={(event) => setHomeSaves(Number(event.target.value))} />
            </label>
            <label>
              Away saves
              <input min={0} type="number" value={awaySaves} onChange={(event) => setAwaySaves(Number(event.target.value))} />
            </label>
          </div>
          <div>
            <button type="submit">Submit result</button>
          </div>
        </form>
      )}

      {activeAction === "ratify" && (
        <form onSubmit={submitRatify}>
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
            Team ID
            <input min={1} type="number" value={teamId} onChange={(event) => setTeamId(Number(event.target.value))} />
          </label>
          <div>
            <button type="submit">Ratify result</button>
          </div>
        </form>
      )}

      {activeAction === "reject" && (
        <form onSubmit={submitReject}>
          <label>
            Reject reason
            <input value={rejectReason} onChange={(event) => setRejectReason(event.target.value)} />
          </label>
          <div>
            <button type="submit">Reject result</button>
          </div>
        </form>
      )}

      {activeAction === "replay" && (
        <form onSubmit={submitReplay}>
          <label>
            Replay body
            <input value={replayBody} onChange={(event) => setReplayBody(event.target.value)} />
          </label>
          <div>
            <button type="submit">Upload replay evidence</button>
          </div>
        </form>
      )}
      {lastReplayConflict ? <p data-testid="player-replay-conflict">Replay review needs attention.</p> : null}

      {status.some((item) => item.isPending) && <LoadingState label="Saving submission action..." />}
      {status.some((item) => item.isSuccess) && (
        <p data-testid="player-submission-success">Submission action submitted.</p>
      )}
      {status.map((item, index) =>
        item.error ? (
          <ErrorView error={item.error} key={`submission-error-${index}`} label="Submission action failed" />
        ) : null,
      )}
    </section>
  );
}

function DataList({ title, items, loading }: { title: string; items: string[]; loading: boolean }) {
  return (
    <section>
      <h3>{title}</h3>
      {loading && <LoadingState label={`Loading ${title.toLowerCase()}...`} />}
      {!loading && (
        <ul style={{ margin: 0, paddingLeft: "1.2rem" }}>
          {items.slice(0, 10).map((item) => (
            <li key={item}>{item}</li>
          ))}
        </ul>
      )}
    </section>
  );
}

function EvidenceSection({
  selected,
  setSelected,
}: {
  selected: string;
  setSelected: (value: string) => void;
}) {
  const selectedView = evidenceViews.find((view) => view.id === selected) ?? evidenceViews[0];

  return (
    <section>
      <h3>Evidence Views</h3>
      <p>Read-only data is embedded from Evidence to avoid duplicating reporting surfaces.</p>
      <label>
        View
        <select value={selected} onChange={(event) => setSelected(event.target.value)}>
          {evidenceViews.map((view) => (
            <option key={view.id} value={view.id}>
              {view.title}
            </option>
          ))}
        </select>
      </label>
      <div style={{ border: "1px solid rgba(207, 210, 211, 0.24)", borderRadius: "10px", height: "360px", marginTop: "0.75rem" }}>
        <iframe
          src={`${env.evidenceBaseUrl}${selectedView.path}`}
          style={{ border: 0, height: "100%", width: "100%" }}
          title={`Evidence ${selectedView.title}`}
        />
      </div>
    </section>
  );
}

function RatingsPanel({ ratings, loading }: { ratings: PlayerRating[]; loading: boolean }) {
  return (
    <section>
      <h3>Ratings</h3>
      {loading && <LoadingState />}
      {!loading && (
        <ul style={{ margin: 0, paddingLeft: "1.2rem" }}>
          {ratings.map((rating) => (
            <li key={rating.id}>
              Player {rating.playerId} · {rating.contextKey} · rating {rating.rating} ± {rating.uncertainty} · matches{" "}
              {rating.matchesPlayed}
            </li>
          ))}
        </ul>
      )}
      {!loading && ratings.length === 0 && <p>No ratings available for this account scope.</p>}
    </section>
  );
}

function AccountAndEligibilityPanel({
  subject,
  links,
  loading,
  eligibility,
  eligibilityError,
  eligibilityLoading,
  queryError,
  onLink,
  onUnlink,
  status,
}: {
  subject: string;
  links: PlatformAccountLink[];
  loading: boolean;
  eligibility: EligibilityStatus | null;
  eligibilityError: Error | null;
  eligibilityLoading: boolean;
  queryError: Error | null;
  onLink: (input: {
    subject: string;
    provider: "steam" | "xbox" | "psn" | "epic";
    providerAccountId: string;
    providerAccountName: string;
  }) => Promise<unknown>;
  onUnlink: (input: { subject: string; provider: "steam" | "xbox" | "psn" | "epic"; providerAccountId: string }) => Promise<unknown>;
  status: Array<{ isPending: boolean; isSuccess: boolean; error: Error | null }>;
}) {
  const [provider, setProvider] = useState<"steam" | "xbox" | "psn" | "epic">("steam");
  const [providerAccountId, setProviderAccountId] = useState("");
  const [providerAccountName, setProviderAccountName] = useState("");

  const activeLinks = links.filter((link) => link.isActive);

  async function submitLink(event: FormEvent<HTMLFormElement>) {
    event.preventDefault();
    await onLink({
      subject,
      provider,
      providerAccountId,
      providerAccountName,
    });
  }

  async function submitUnlink(link: PlatformAccountLink) {
    await onUnlink({
      subject: link.subject,
      provider: link.provider as "steam" | "xbox" | "psn" | "epic",
      providerAccountId: link.providerAccountId,
    });
  }

  return (
    <section>
      <h3>Accounts and Eligibility</h3>
      <p>Link and unlink verified platform identities for replay attribution and player ownership checks.</p>
      <form onSubmit={submitLink}>
        <label>
          Provider
          <select value={provider} onChange={(event) => setProvider(event.target.value as "steam" | "xbox" | "psn" | "epic")}>
            <option value="steam">Steam</option>
            <option value="xbox">Xbox</option>
            <option value="psn">PSN</option>
            <option value="epic">Epic</option>
          </select>
        </label>
        <label>
          Provider account ID
          <input value={providerAccountId} onChange={(event) => setProviderAccountId(event.target.value)} />
        </label>
        <label>
          Provider display name
          <input value={providerAccountName} onChange={(event) => setProviderAccountName(event.target.value)} />
        </label>
        <button disabled={subject.length === 0} type="submit">
          Link account
        </button>
      </form>
      {loading && <LoadingState label="Loading linked accounts..." />}
      {subject.length === 0 && <p>Session subject unavailable; refresh the page after login.</p>}
      {!loading && activeLinks.length === 0 && <p>No active platform accounts linked.</p>}
      {!loading && activeLinks.length > 0 && (
        <ul style={{ margin: "0.5rem 0", paddingLeft: "1.2rem" }}>
          {activeLinks.map((link) => (
            <li key={link.id}>
              {link.provider} · {link.providerAccountId}
              {link.providerAccountName ? ` (${link.providerAccountName})` : ""}
              <button onClick={() => void submitUnlink(link)} style={{ marginLeft: "0.5rem" }} type="button">
                Unlink
              </button>
            </li>
          ))}
        </ul>
      )}
      {status.some((item) => item.isPending) && <LoadingState label="Saving account link..." />}
      {status.some((item) => item.isSuccess) && <p data-testid="player-platform-link-success">Platform account updated.</p>}
      {status.map((item, index) =>
        item.error ? (
          <ErrorView error={item.error} key={`platform-account-error-${index}`} label="Platform account action failed" />
        ) : null,
      )}
      {queryError && <ErrorView error={queryError} label="Failed to load linked platform accounts" />}
      <h4>Eligibility</h4>
      {eligibilityLoading && <LoadingState label="Loading eligibility projection..." />}
      {!eligibilityLoading && eligibility !== null && (
        <>
          <p>
            {eligibility.points} points (threshold {eligibility.thresholdPoints}, decay {eligibility.decayPerWeek}/week)
          </p>
          <p>Eligible until: {new Date(eligibility.eligibleUntil).toLocaleString()}</p>
          <ul style={{ margin: "0.5rem 0", paddingLeft: "1.2rem" }}>
            {eligibility.projection.slice(0, 5).map((point) => (
              <li key={point.effectiveAt}>
                {new Date(point.effectiveAt).toLocaleDateString()} · {point.points} points ·{" "}
                {point.isEligible ? "eligible" : "ineligible"}
              </li>
            ))}
          </ul>
        </>
      )}
      {eligibilityError && <ErrorView error={eligibilityError} label="Failed to load eligibility status" />}
    </section>
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
