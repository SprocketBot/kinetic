import { z } from "zod";

export const exceptionTicketSchema = z.object({
  id: z.number(),
  category: z.string(),
  contextType: z.string(),
  contextId: z.number(),
  reportedByTeamId: z.number().nullable().optional(),
  state: z.string(),
  reasonCode: z.string(),
  severity: z.number(),
  suggestedAction: z.string(),
  detailsJson: z.unknown().optional(),
  resolutionCode: z.string().nullable().optional(),
  openedAt: z.string(),
  triagedAt: z.string().nullable().optional(),
  resolvedAt: z.string().nullable().optional(),
});

export const operatorInboxListSchema = z.array(exceptionTicketSchema);

export const triageExceptionInputSchema = z.object({
  ticketId: z.number().positive(),
  actor: z.string().min(1),
  reasonCode: z.string().min(1),
  severity: z.number().min(1).max(5),
  suggestedAction: z.string().min(1),
  minutesSpent: z.number().min(0),
});

export const resolveExceptionInputSchema = z.object({
  ticketId: z.number().positive(),
  actor: z.string().min(1),
  resolutionCode: z.string().min(1),
  notes: z.string(),
  automated: z.boolean(),
  minutesSpent: z.number().min(0),
});

export const queueEntrySchema = z.object({
  id: z.number(),
  queueId: z.number(),
  teamId: z.number(),
  isActive: z.boolean(),
  stage: z.number(),
  createdAt: z.string(),
  stageAt: z.string(),
  leftAt: z.string().nullable().optional(),
});

export const queueEntryListSchema = z.array(queueEntrySchema);

export const scrimSchema = z.object({
  id: z.number(),
  queueId: z.number(),
  homeTeamId: z.number(),
  awayTeamId: z.number(),
  state: z.string(),
  createdAt: z.string(),
  startedAt: z.string().nullable().optional(),
  endedAt: z.string().nullable().optional(),
});

export const scrimListSchema = z.array(scrimSchema);

export const resultSubmissionSchema = z.object({
  id: z.number(),
  contextType: z.string(),
  contextId: z.number(),
  submittedByTeamId: z.number(),
  homeTeamId: z.number(),
  awayTeamId: z.number(),
  winningTeamId: z.number(),
  losingTeamId: z.number(),
  state: z.string(),
  payloadJson: z.unknown().optional(),
  homeRatifiedAt: z.string().nullable().optional(),
  awayRatifiedAt: z.string().nullable().optional(),
  rejectedByTeamId: z.number().nullable().optional(),
  rejectionReason: z.string().nullable().optional(),
  rejectedAt: z.string().nullable().optional(),
  createdAt: z.string(),
});

export const resultSubmissionListSchema = z.array(resultSubmissionSchema);

export const exceptionMetricsSchema = z.object({
  adminHoursPerWeek: z.number(),
  manualTouchesPerFixture: z.number(),
  zeroTouchFixtureRate: z.number(),
  timeToCloseHoursP50: z.number(),
});

export const replayEvidenceSchema = z.object({
  id: z.number(),
  contextType: z.string(),
  contextId: z.number(),
  submittedByTeamId: z.number(),
  replaySha256: z.string(),
  contentSizeBytes: z.number(),
  storageRef: z.string(),
  state: z.string(),
  createdAt: z.string(),
});

export const replayParseRunSchema = z.object({
  id: z.number(),
  replayEvidenceId: z.number(),
  parserName: z.string(),
  parserVersion: z.string(),
  parserConfigDigest: z.string(),
  status: z.string(),
  outputJson: z.unknown().optional(),
  createdAt: z.string(),
});

export const replayIngestionResultSchema = z.object({
  evidence: replayEvidenceSchema,
  parseRun: replayParseRunSchema,
  duplicate: z.boolean(),
  linkedSubmissionId: z.number().nullable().optional(),
});

export const enqueueTeamInputSchema = z.object({
  queueId: z.number().positive(),
  teamId: z.number().positive(),
});

export const leaveQueueInputSchema = z.object({
  queueId: z.number().positive(),
  teamId: z.number().positive(),
});

export const ratifyResultSubmissionInputSchema = z.object({
  submissionId: z.number().positive(),
  teamId: z.number().positive(),
});

export const rejectResultSubmissionInputSchema = z.object({
  submissionId: z.number().positive(),
  teamId: z.number().positive(),
  reason: z.string().min(1),
});

export const ingestReplayEvidenceInputSchema = z.object({
  contextType: z.enum(["scrim", "match"]),
  contextId: z.number().positive(),
  submittedByTeamId: z.number().positive(),
  replayBody: z.string().min(1),
  parserName: z.string().min(1),
  parserVersion: z.string().min(1),
  parserConfigDigest: z.string().min(1),
  parseOutputJson: z.unknown().default({}),
  resultSubmissionId: z.number().positive().optional(),
});

export type ExceptionTicket = z.infer<typeof exceptionTicketSchema>;
export type TriageExceptionInput = z.infer<typeof triageExceptionInputSchema>;
export type ResolveExceptionInput = z.infer<typeof resolveExceptionInputSchema>;
export type QueueEntry = z.infer<typeof queueEntrySchema>;
export type Scrim = z.infer<typeof scrimSchema>;
export type ResultSubmission = z.infer<typeof resultSubmissionSchema>;
export type ExceptionMetrics = z.infer<typeof exceptionMetricsSchema>;
export type ReplayIngestionResult = z.infer<typeof replayIngestionResultSchema>;
export type EnqueueTeamInput = z.infer<typeof enqueueTeamInputSchema>;
export type LeaveQueueInput = z.infer<typeof leaveQueueInputSchema>;
export type RatifyResultSubmissionInput = z.infer<typeof ratifyResultSubmissionInputSchema>;
export type RejectResultSubmissionInput = z.infer<typeof rejectResultSubmissionInputSchema>;
export type IngestReplayEvidenceInput = z.infer<typeof ingestReplayEvidenceInputSchema>;
