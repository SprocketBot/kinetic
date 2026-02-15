import { useEffect, useMemo, useState } from "react";
import type { FormEvent } from "react";
import { useMutation, useQuery, useQueryClient } from "@tanstack/react-query";

import { useSession } from "../../../auth/session-context";
import { AppShell } from "../../../components/layout/app-shell";
import { LoadingState } from "../../../components/feedback/loading-state";
import { getJson, postJson } from "../../../lib/api/client";
import { ApiError } from "../../../lib/api/errors";
import {
  exceptionTicketSchema,
  operatorInboxListSchema,
  resolveExceptionInputSchema,
  triageExceptionInputSchema,
  type ExceptionTicket,
  type ResolveExceptionInput,
  type TriageExceptionInput,
} from "../../../lib/api/schemas";

async function listOperatorInbox() {
  return getJson("/v1/operator-inbox", operatorInboxListSchema);
}

async function triageTicket(input: TriageExceptionInput) {
  const payload = triageExceptionInputSchema.parse(input);
  return postJson("/v1/operator-inbox/triage", payload, exceptionTicketSchema);
}

async function resolveTicket(input: ResolveExceptionInput) {
  const payload = resolveExceptionInputSchema.parse(input);
  return postJson("/v1/operator-inbox/resolve", payload, exceptionTicketSchema);
}

export function SupportDashboardPage() {
  const session = useSession();
  const queryClient = useQueryClient();
  const [severityFilter, setSeverityFilter] = useState<string>("all");
  const [stateFilter, setStateFilter] = useState<string>("all");
  const [selectedTicketId, setSelectedTicketId] = useState<number | null>(null);

  const query = useQuery({
    queryKey: ["operator-inbox"],
    queryFn: listOperatorInbox,
  });

  const actorDefault = session.principal?.displayName || session.principal?.subject || "support-operator";

  const filteredTickets = useMemo(() => {
    if (!query.data) {
      return [];
    }

    return query.data.filter((ticket) => {
      const severityMatches = severityFilter === "all" || String(ticket.severity) === severityFilter;
      const stateMatches = stateFilter === "all" || ticket.state === stateFilter;
      return severityMatches && stateMatches;
    });
  }, [query.data, severityFilter, stateFilter]);

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

  return (
    <AppShell title="Sprocket Web Client">
      <h2>League Support</h2>
      <p>Inbox triage workspace for active exception operations.</p>

      <section aria-label="support filters" style={{ display: "flex", gap: "0.75rem", marginBottom: "1rem" }}>
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

      {query.isLoading && <LoadingState label="Loading operator inbox..." />}
      {query.isError && <ErrorView error={query.error} label="Inbox request failed" />}

      {query.isSuccess && (
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
    </AppShell>
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
                background: ticket.id === selectedTicketId ? "#dbeafe" : "#f8fafc",
                border: "1px solid #cbd5e1",
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
