import { useEffect, useMemo, useState } from "react";
import type { FormEvent } from "react";
import { useMutation, useQuery, useQueryClient } from "@tanstack/react-query";

import { useSession } from "../../../auth/session-context";
import { AppShell } from "../../../components/layout/app-shell";
import { LoadingState } from "../../../components/feedback/loading-state";
import { getJson, postJson } from "../../../lib/api/client";
import { ApiError } from "../../../lib/api/errors";
import {
  banPlayerFromQueueInputSchema,
  exceptionTicketSchema,
  operatorInboxListSchema,
  queueBanListSchema,
  queueBanSchema,
  resolveExceptionInputSchema,
  resultSubmissionListSchema,
  scrimListSchema,
  triageExceptionInputSchema,
  unbanPlayerFromQueueInputSchema,
  type QueueBan,
  type ExceptionTicket,
  type BanPlayerFromQueueInput,
  type ResolveExceptionInput,
  type ResultSubmission,
  type Scrim,
  type TriageExceptionInput,
  type UnbanPlayerFromQueueInput,
} from "../../../lib/api/schemas";

async function listOperatorInbox() {
  return getJson("/v1/operator-inbox", operatorInboxListSchema);
}

async function listScrims() {
  return getJson("/v1/scrims", scrimListSchema);
}

async function listResultSubmissions() {
  return getJson("/v1/result-submissions", resultSubmissionListSchema);
}

async function listQueueBans() {
  return getJson("/v1/queue-bans", queueBanListSchema);
}

async function triageTicket(input: TriageExceptionInput) {
  const payload = triageExceptionInputSchema.parse(input);
  return postJson("/v1/operator-inbox/triage", payload, exceptionTicketSchema);
}

async function resolveTicket(input: ResolveExceptionInput) {
  const payload = resolveExceptionInputSchema.parse(input);
  return postJson("/v1/operator-inbox/resolve", payload, exceptionTicketSchema);
}

async function banPlayerFromQueue(input: BanPlayerFromQueueInput) {
  const payload = banPlayerFromQueueInputSchema.parse(input);
  return postJson("/v1/queue-bans", payload, queueBanSchema);
}

async function unbanPlayerFromQueue(input: UnbanPlayerFromQueueInput) {
  const payload = unbanPlayerFromQueueInputSchema.parse(input);
  return postJson("/v1/queue-bans/lift", payload, queueBanSchema);
}

function isActiveScrim(scrim: Scrim): boolean {
  return !["cancelled", "completed", "closed", "ended"].includes(scrim.state.toLowerCase());
}

function isInProcessSubmission(submission: ResultSubmission): boolean {
  return !["finalized", "accepted", "rejected"].includes(submission.state.toLowerCase());
}

export function SupportDashboardPage() {
  const session = useSession();
  const queryClient = useQueryClient();
  const [severityFilter, setSeverityFilter] = useState<string>("all");
  const [stateFilter, setStateFilter] = useState<string>("all");
  const [selectedTicketId, setSelectedTicketId] = useState<number | null>(null);

  const inboxQuery = useQuery({
    queryKey: ["operator-inbox"],
    queryFn: listOperatorInbox,
  });

  const scrimsQuery = useQuery({
    queryKey: ["scrims"],
    queryFn: listScrims,
  });

  const submissionsQuery = useQuery({
    queryKey: ["result-submissions"],
    queryFn: listResultSubmissions,
  });

  const queueBansQuery = useQuery({
    queryKey: ["queue-bans"],
    queryFn: listQueueBans,
  });

  const actorDefault = session.principal?.displayName || session.principal?.subject || "support-operator";

  const filteredTickets = useMemo(() => {
    if (!inboxQuery.data) {
      return [];
    }

    return inboxQuery.data.filter((ticket) => {
      const severityMatches = severityFilter === "all" || String(ticket.severity) === severityFilter;
      const stateMatches = stateFilter === "all" || ticket.state === stateFilter;
      return severityMatches && stateMatches;
    });
  }, [inboxQuery.data, severityFilter, stateFilter]);

  const activeScrims = useMemo(() => {
    if (!scrimsQuery.data) {
      return [];
    }

    return scrimsQuery.data.filter(isActiveScrim);
  }, [scrimsQuery.data]);

  const inProcessSubmissions = useMemo(() => {
    if (!submissionsQuery.data) {
      return [];
    }

    return submissionsQuery.data.filter(isInProcessSubmission);
  }, [submissionsQuery.data]);

  useEffect(() => {
    if (filteredTickets.length === 0) {
      setSelectedTicketId(null);
      return;
    }

    const selectedExists = filteredTickets.some((ticket) => ticket.id === selectedTicketId);
    if (!selectedExists) {
      setSelectedTicketId(filteredTickets[0].id);
    }
  }, [filteredTickets, selectedTicketId]);

  const selectedTicket = filteredTickets.find((ticket) => ticket.id === selectedTicketId) ?? null;

  const triageMutation = useMutation({
    mutationFn: triageTicket,
    onSuccess: async (ticket) => {
      setSelectedTicketId(ticket.id);
      await queryClient.invalidateQueries({ queryKey: ["operator-inbox"] });
    },
  });

  const resolveMutation = useMutation({
    mutationFn: resolveTicket,
    onSuccess: async (ticket) => {
      setSelectedTicketId(ticket.id);
      await queryClient.invalidateQueries({ queryKey: ["operator-inbox"] });
    },
  });

  const banMutation = useMutation({
    mutationFn: banPlayerFromQueue,
    onSuccess: async () => {
      await queryClient.invalidateQueries({ queryKey: ["queue-bans"] });
    },
  });

  const unbanMutation = useMutation({
    mutationFn: unbanPlayerFromQueue,
    onSuccess: async () => {
      await queryClient.invalidateQueries({ queryKey: ["queue-bans"] });
    },
  });

  return (
    <AppShell title="Sprocket Web Client">
      <h2>League Support</h2>
      <p>Inbox triage workspace for active exception operations.</p>

      <section aria-label="support live snapshots" className="layout-grid layout-grid--2">
        <LiveCard
          count={activeScrims.length}
          label="Active scrims"
          loading={scrimsQuery.isLoading}
          testId="active-scrims-count"
        />
        <LiveCard
          count={inProcessSubmissions.length}
          label="Submissions in process"
          loading={submissionsQuery.isLoading}
          testId="submissions-in-process-count"
        />
      </section>

      <section aria-label="support filters" className="layout-grid layout-grid--2">
        <label>
          Filter severity
          <select value={severityFilter} onChange={(event) => setSeverityFilter(event.target.value)}>
            <option value="all">All</option>
            <option value="1">1</option>
            <option value="2">2</option>
            <option value="3">3</option>
            <option value="4">4</option>
            <option value="5">5</option>
          </select>
        </label>

        <label>
          Filter state
          <select value={stateFilter} onChange={(event) => setStateFilter(event.target.value)}>
            <option value="all">All</option>
            <option value="open">open</option>
            <option value="triaged">triaged</option>
            <option value="resolved">resolved</option>
          </select>
        </label>
      </section>

      {inboxQuery.isLoading && <LoadingState label="Loading operator inbox..." />}
      {inboxQuery.isError && <ErrorView error={inboxQuery.error} label="Inbox request failed" />}

      {inboxQuery.isSuccess && (
        <section style={{ display: "grid", gap: "1rem", gridTemplateColumns: "minmax(280px, 360px) 1fr" }}>
          <InboxList
            onSelect={setSelectedTicketId}
            selectedTicketId={selectedTicketId}
            tickets={filteredTickets}
          />

          {selectedTicket ? (
            <TicketDetail
              actorDefault={actorDefault}
              onResolve={resolveMutation.mutateAsync}
              onTriage={triageMutation.mutateAsync}
              resolveError={resolveMutation.error}
              resolveSuccess={resolveMutation.isSuccess}
              resolving={resolveMutation.isPending}
              ticket={selectedTicket}
              triageError={triageMutation.error}
              triageSuccess={triageMutation.isSuccess}
              triaging={triageMutation.isPending}
            />
          ) : (
            <p data-testid="ticket-detail-empty">No ticket selected for current filters.</p>
          )}
        </section>
      )}

      <section className="layout-grid layout-grid--2">
        <ScrimList scrims={activeScrims} />
        <SubmissionList submissions={inProcessSubmissions} />
      </section>

      <section>
        <QueueModerationPanel
          actorDefault={actorDefault}
          banError={banMutation.error}
          banPlayer={banMutation.mutateAsync}
          banSuccess={banMutation.isSuccess}
          bans={queueBansQuery.data ?? []}
          loading={queueBansQuery.isLoading}
          unbanError={unbanMutation.error}
          unbanPlayer={unbanMutation.mutateAsync}
          unbanSuccess={unbanMutation.isSuccess}
        />
      </section>
    </AppShell>
  );
}

function LiveCard({
  label,
  count,
  loading,
  testId,
}: {
  label: string;
  count: number;
  loading: boolean;
  testId: string;
}) {
  return (
    <article>
      <h3 style={{ marginTop: 0 }}>{label}</h3>
      <p data-testid={testId}>
        {loading ? "..." : count}
      </p>
    </article>
  );
}

function InboxList({
  tickets,
  selectedTicketId,
  onSelect,
}: {
  tickets: ExceptionTicket[];
  selectedTicketId: number | null;
  onSelect: (ticketId: number) => void;
}) {
  return (
    <div>
      <h3>Inbox</h3>
      <p data-testid="operator-inbox-count">Visible tickets: {tickets.length}</p>
      <ul style={{ listStyle: "none", margin: 0, padding: 0, display: "grid", gap: "0.5rem" }}>
        {tickets.map((ticket) => (
          <li key={ticket.id}>
            <button
              onClick={() => onSelect(ticket.id)}
              style={{
                background: ticket.id === selectedTicketId ? "rgba(254, 191, 43, 0.2)" : "rgba(13, 14, 14, 0.5)",
                border: "1px solid rgba(207, 210, 211, 0.25)",
                borderRadius: "10px",
                color: "rgba(247, 248, 248, 0.94)",
                cursor: "pointer",
                padding: "0.6rem",
                textAlign: "left",
                width: "100%",
              }}
              type="button"
            >
              <div style={{ fontWeight: 600 }}>#{ticket.id} · {ticket.category}</div>
              <div>{ticket.reasonCode} · severity {ticket.severity} · {ticket.state}</div>
            </button>
          </li>
        ))}
      </ul>
    </div>
  );
}

function TicketDetail({
  ticket,
  actorDefault,
  onTriage,
  onResolve,
  triaging,
  resolving,
  triageSuccess,
  resolveSuccess,
  triageError,
  resolveError,
}: {
  ticket: ExceptionTicket;
  actorDefault: string;
  onTriage: (input: TriageExceptionInput) => Promise<unknown>;
  onResolve: (input: ResolveExceptionInput) => Promise<unknown>;
  triaging: boolean;
  resolving: boolean;
  triageSuccess: boolean;
  resolveSuccess: boolean;
  triageError: Error | null;
  resolveError: Error | null;
}) {
  const [triageReasonCode, setTriageReasonCode] = useState(ticket.reasonCode);
  const [triageSeverity, setTriageSeverity] = useState(ticket.severity);
  const [triageSuggestedAction, setTriageSuggestedAction] = useState(ticket.suggestedAction);
  const [triageActor, setTriageActor] = useState(actorDefault);
  const [triageMinutesSpent, setTriageMinutesSpent] = useState(5);

  const [resolveCode, setResolveCode] = useState("resolved_manual");
  const [resolveNotes, setResolveNotes] = useState("");
  const [resolveAutomated, setResolveAutomated] = useState(false);
  const [resolveActor, setResolveActor] = useState(actorDefault);
  const [resolveMinutesSpent, setResolveMinutesSpent] = useState(5);

  useEffect(() => {
    setTriageReasonCode(ticket.reasonCode);
    setTriageSeverity(ticket.severity);
    setTriageSuggestedAction(ticket.suggestedAction);
  }, [ticket.id, ticket.reasonCode, ticket.severity, ticket.suggestedAction]);

  async function handleTriageSubmit(event: FormEvent<HTMLFormElement>) {
    event.preventDefault();
    await onTriage({
      ticketId: ticket.id,
      actor: triageActor,
      reasonCode: triageReasonCode,
      severity: triageSeverity,
      suggestedAction: triageSuggestedAction,
      minutesSpent: triageMinutesSpent,
    });
  }

  async function handleResolveSubmit(event: FormEvent<HTMLFormElement>) {
    event.preventDefault();
    await onResolve({
      ticketId: ticket.id,
      actor: resolveActor,
      resolutionCode: resolveCode,
      notes: resolveNotes,
      automated: resolveAutomated,
      minutesSpent: resolveMinutesSpent,
    });
  }

  return (
    <div data-testid="ticket-detail">
      <h3>Ticket #{ticket.id}</h3>
      <p>
        Context: {ticket.contextType}:{ticket.contextId}
      </p>
      <p>Current state: {ticket.state}</p>

      <section style={{ borderTop: "1px solid #e5e7eb", marginTop: "1rem", paddingTop: "1rem" }}>
        <h4>Triage</h4>
        <form onSubmit={handleTriageSubmit}>
          <label>
            Actor
            <input value={triageActor} onChange={(event) => setTriageActor(event.target.value)} />
          </label>
          <label>
            Reason code
            <input value={triageReasonCode} onChange={(event) => setTriageReasonCode(event.target.value)} />
          </label>
          <label>
            Severity
            <input
              max={5}
              min={1}
              type="number"
              value={triageSeverity}
              onChange={(event) => setTriageSeverity(Number(event.target.value))}
            />
          </label>
          <label>
            Suggested action
            <input value={triageSuggestedAction} onChange={(event) => setTriageSuggestedAction(event.target.value)} />
          </label>
          <label>
            Minutes spent
            <input
              min={0}
              type="number"
              value={triageMinutesSpent}
              onChange={(event) => setTriageMinutesSpent(Number(event.target.value))}
            />
          </label>
          <div>
            <button disabled={triaging} type="submit">
              {triaging ? "Submitting..." : "Submit triage"}
            </button>
          </div>
        </form>
        {triageSuccess && <p data-testid="triage-success">Triage submitted.</p>}
        {triageError && <ErrorView error={triageError} label="Triage failed" />}
      </section>

      <section style={{ borderTop: "1px solid #e5e7eb", marginTop: "1rem", paddingTop: "1rem" }}>
        <h4>Resolve</h4>
        <form onSubmit={handleResolveSubmit}>
          <label>
            Actor
            <input value={resolveActor} onChange={(event) => setResolveActor(event.target.value)} />
          </label>
          <label>
            Resolution code
            <input value={resolveCode} onChange={(event) => setResolveCode(event.target.value)} />
          </label>
          <label>
            Notes
            <input value={resolveNotes} onChange={(event) => setResolveNotes(event.target.value)} />
          </label>
          <label>
            Minutes spent
            <input
              min={0}
              type="number"
              value={resolveMinutesSpent}
              onChange={(event) => setResolveMinutesSpent(Number(event.target.value))}
            />
          </label>
          <label>
            Automated
            <input
              checked={resolveAutomated}
              type="checkbox"
              onChange={(event) => setResolveAutomated(event.target.checked)}
            />
          </label>
          <div>
            <button disabled={resolving} type="submit">
              {resolving ? "Submitting..." : "Submit resolution"}
            </button>
          </div>
        </form>
        {resolveSuccess && <p data-testid="resolve-success">Resolution submitted.</p>}
        {resolveError && <ErrorView error={resolveError} label="Resolve failed" />}
      </section>
    </div>
  );
}

function ScrimList({ scrims }: { scrims: Scrim[] }) {
  return (
    <section>
      <h3>Active Scrims</h3>
      <ul style={{ margin: 0, paddingLeft: "1.25rem" }}>
        {scrims.slice(0, 8).map((scrim) => (
          <li key={scrim.id}>Scrim #{scrim.id} · queue {scrim.queueId} · {scrim.state}</li>
        ))}
      </ul>
    </section>
  );
}

function SubmissionList({ submissions }: { submissions: ResultSubmission[] }) {
  return (
    <section>
      <h3>Submissions In Process</h3>
      <ul style={{ margin: 0, paddingLeft: "1.25rem" }}>
        {submissions.slice(0, 8).map((submission) => (
          <li key={submission.id}>
            Submission #{submission.id} · {submission.contextType}:{submission.contextId} · {submission.state}
          </li>
        ))}
      </ul>
    </section>
  );
}

function QueueModerationPanel({
  actorDefault,
  bans,
  loading,
  banPlayer,
  unbanPlayer,
  banSuccess,
  unbanSuccess,
  banError,
  unbanError,
}: {
  actorDefault: string;
  bans: QueueBan[];
  loading: boolean;
  banPlayer: (input: BanPlayerFromQueueInput) => Promise<unknown>;
  unbanPlayer: (input: UnbanPlayerFromQueueInput) => Promise<unknown>;
  banSuccess: boolean;
  unbanSuccess: boolean;
  banError: Error | null;
  unbanError: Error | null;
}) {
  const [queueId, setQueueId] = useState(1);
  const [playerId, setPlayerId] = useState(1);
  const [actor, setActor] = useState(actorDefault);
  const [banReason, setBanReason] = useState("support moderation action");
  const [unbanReason, setUnbanReason] = useState("appeal accepted");

  async function submitBan(event: FormEvent<HTMLFormElement>) {
    event.preventDefault();
    await banPlayer({ queueId, playerId, actor, reason: banReason });
  }

  async function submitUnban(event: FormEvent<HTMLFormElement>) {
    event.preventDefault();
    await unbanPlayer({ queueId, playerId, actor, reason: unbanReason });
  }

  return (
    <section>
      <h3>Queue Moderation</h3>
      <div className="layout-grid layout-grid--2">
        <form onSubmit={submitBan}>
          <h4>Ban Player</h4>
          <label>
            Queue ID
            <input min={1} type="number" value={queueId} onChange={(event) => setQueueId(Number(event.target.value))} />
          </label>
          <label>
            Player ID
            <input min={1} type="number" value={playerId} onChange={(event) => setPlayerId(Number(event.target.value))} />
          </label>
          <label>
            Actor
            <input value={actor} onChange={(event) => setActor(event.target.value)} />
          </label>
          <label>
            Reason
            <input value={banReason} onChange={(event) => setBanReason(event.target.value)} />
          </label>
          <button type="submit">Ban player from queue</button>
          {banSuccess && <p data-testid="queue-ban-success">Queue ban submitted.</p>}
          {banError && <ErrorView error={banError} label="Queue ban failed" />}
        </form>

        <form onSubmit={submitUnban}>
          <h4>Lift Queue Ban</h4>
          <label>
            Queue ID
            <input min={1} type="number" value={queueId} onChange={(event) => setQueueId(Number(event.target.value))} />
          </label>
          <label>
            Player ID
            <input min={1} type="number" value={playerId} onChange={(event) => setPlayerId(Number(event.target.value))} />
          </label>
          <label>
            Actor
            <input value={actor} onChange={(event) => setActor(event.target.value)} />
          </label>
          <label>
            Reason
            <input value={unbanReason} onChange={(event) => setUnbanReason(event.target.value)} />
          </label>
          <button type="submit">Lift queue ban</button>
          {unbanSuccess && <p data-testid="queue-unban-success">Queue ban lifted.</p>}
          {unbanError && <ErrorView error={unbanError} label="Queue unban failed" />}
        </form>
      </div>

      {loading ? (
        <LoadingState label="Loading queue bans..." />
      ) : (
        <ul style={{ marginBottom: 0, marginTop: "1rem", paddingLeft: "1.25rem" }}>
          {bans.map((ban) => (
            <li key={ban.id}>
              Queue {ban.queueId} · player {ban.playerId} · {ban.isActive ? "active" : "lifted"} · {ban.banReason}
            </li>
          ))}
        </ul>
      )}
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
