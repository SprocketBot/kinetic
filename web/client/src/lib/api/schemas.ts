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

export const queueBanSchema = z.object({
  id: z.number(),
  queueId: z.number(),
  playerId: z.number(),
  bannedByActor: z.string(),
  banReason: z.string(),
  isActive: z.boolean(),
  bannedAt: z.string(),
  unbannedByActor: z.string().nullable().optional(),
  unbanReason: z.string().nullable().optional(),
  unbannedAt: z.string().nullable().optional(),
});

export const queueBanListSchema = z.array(queueBanSchema);

export const banPlayerFromQueueInputSchema = z.object({
  queueId: z.number().positive(),
  playerId: z.number().positive(),
  actor: z.string().min(1),
  reason: z.string().min(1),
});

export const unbanPlayerFromQueueInputSchema = z.object({
  queueId: z.number().positive(),
  playerId: z.number().positive(),
  actor: z.string().min(1),
  reason: z.string().min(1),
});

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

export const overrideResultSubmissionInputSchema = z.object({
  submissionId: z.number().positive(),
  actor: z.string().min(1),
  reason: z.string().min(1),
  winningTeamId: z.number().positive(),
  losingTeamId: z.number().positive(),
});

export const resultOverrideSchema = z.object({
  id: z.number(),
  submissionId: z.number(),
  actor: z.string(),
  reason: z.string(),
  previousWinningTeamId: z.number(),
  previousLosingTeamId: z.number(),
  newWinningTeamId: z.number(),
  newLosingTeamId: z.number(),
  previousState: z.string(),
  newState: z.string(),
  createdAt: z.string(),
});

export const resultOverrideListSchema = z.array(resultOverrideSchema);

export const exceptionMetricsSchema = z.object({
  adminHoursPerWeek: z.number(),
  manualTouchesPerFixture: z.number(),
  zeroTouchFixtureRate: z.number(),
  timeToCloseHoursP50: z.number(),
});

export const seasonSchema = z.object({
  id: z.number(),
  name: z.string(),
  slug: z.string(),
  isActive: z.boolean(),
  createdAt: z.string(),
});

export const seasonListSchema = z.array(seasonSchema);

export const createSeasonInputSchema = z.object({
  name: z.string().min(1),
  slug: z.string().min(1),
});

export const scheduleGroupSchema = z.object({
  id: z.number(),
  seasonId: z.number(),
  name: z.string(),
  sequence: z.number(),
  isActive: z.boolean(),
  createdAt: z.string(),
});

export const scheduleGroupListSchema = z.array(scheduleGroupSchema);

export const createScheduleGroupInputSchema = z.object({
  seasonId: z.number().positive(),
  name: z.string().min(1),
  sequence: z.number().min(0),
});

export const fixtureSchema = z.object({
  id: z.number(),
  scheduleGroupId: z.number(),
  homeClubId: z.number(),
  awayClubId: z.number(),
  isActive: z.boolean(),
  createdAt: z.string(),
});

export const fixtureListSchema = z.array(fixtureSchema);

export const createFixtureInputSchema = z.object({
  scheduleGroupId: z.number().positive(),
  homeClubId: z.number().positive(),
  awayClubId: z.number().positive(),
});

export const matchSchema = z.object({
  id: z.number(),
  fixtureId: z.number(),
  homeTeamId: z.number(),
  awayTeamId: z.number(),
  state: z.string(),
  scheduledFor: z.string().nullable().optional(),
  homeTimeRatifiedAt: z.string().nullable().optional(),
  awayTimeRatifiedAt: z.string().nullable().optional(),
  createdAt: z.string(),
});

export const matchListSchema = z.array(matchSchema);

export const teamSchema = z.object({
  id: z.number(),
  clubId: z.number(),
  name: z.string(),
  slug: z.string(),
  isActive: z.boolean(),
  createdAt: z.string(),
});

export const teamListSchema = z.array(teamSchema);

export const playerSchema = z.object({
  id: z.number(),
  displayName: z.string(),
  slug: z.string(),
  isActive: z.boolean(),
  createdAt: z.string(),
});

export const playerListSchema = z.array(playerSchema);

export const rosterMembershipSchema = z.object({
  id: z.number(),
  playerId: z.number(),
  teamId: z.number(),
  isActive: z.boolean(),
  createdAt: z.string(),
});

export const rosterMembershipListSchema = z.array(rosterMembershipSchema);

export const createRosterMembershipInputSchema = z.object({
  playerId: z.number().positive(),
  teamId: z.number().positive(),
});

export const exceptionActionSchema = z.object({
  id: z.number(),
  ticketId: z.number(),
  actionType: z.string(),
  actor: z.string(),
  automated: z.boolean(),
  notes: z.string(),
  minutesSpent: z.number(),
  createdAt: z.string(),
});

export const exceptionActionListSchema = z.array(exceptionActionSchema);

export const createMatchInputSchema = z.object({
  fixtureId: z.number().positive(),
  homeTeamId: z.number().positive(),
  awayTeamId: z.number().positive(),
  state: z.string().min(1),
  scheduledFor: z.string().optional(),
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

export const playerRatingSchema = z.object({
  id: z.number(),
  playerId: z.number(),
  contextKey: z.string(),
  rating: z.number(),
  uncertainty: z.number(),
  matchesPlayed: z.number(),
  lastCompetedAt: z.string().nullable().optional(),
  isActive: z.boolean(),
  updatedAt: z.string(),
});

export const playerRatingListSchema = z.array(playerRatingSchema);

export const adjustPlayerRatingInputSchema = z.object({
  actorPlayerId: z.number().positive(),
  targetPlayerId: z.number().positive(),
  contextKey: z.string().min(1),
  rating: z.number().min(0),
  uncertainty: z.number().min(0),
  matchesPlayed: z.number().min(0),
  reason: z.string().min(1),
});

export const ratingAdjustmentSchema = z.object({
  id: z.number(),
  actorPlayerId: z.number(),
  targetPlayerId: z.number(),
  contextKey: z.string(),
  previousRating: z.number(),
  newRating: z.number(),
  previousUncertainty: z.number(),
  newUncertainty: z.number(),
  previousMatchesPlayed: z.number(),
  newMatchesPlayed: z.number(),
  reason: z.string(),
  createdAt: z.string(),
});

export const ratingAdjustmentListSchema = z.array(ratingAdjustmentSchema);

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
export type QueueBan = z.infer<typeof queueBanSchema>;
export type Scrim = z.infer<typeof scrimSchema>;
export type ResultSubmission = z.infer<typeof resultSubmissionSchema>;
export type ResultOverride = z.infer<typeof resultOverrideSchema>;
export type ExceptionMetrics = z.infer<typeof exceptionMetricsSchema>;
export type ReplayIngestionResult = z.infer<typeof replayIngestionResultSchema>;
export type Season = z.infer<typeof seasonSchema>;
export type ScheduleGroup = z.infer<typeof scheduleGroupSchema>;
export type Fixture = z.infer<typeof fixtureSchema>;
export type Match = z.infer<typeof matchSchema>;
export type Team = z.infer<typeof teamSchema>;
export type Player = z.infer<typeof playerSchema>;
export type RosterMembership = z.infer<typeof rosterMembershipSchema>;
export type ExceptionAction = z.infer<typeof exceptionActionSchema>;
export type EnqueueTeamInput = z.infer<typeof enqueueTeamInputSchema>;
export type LeaveQueueInput = z.infer<typeof leaveQueueInputSchema>;
export type RatifyResultSubmissionInput = z.infer<typeof ratifyResultSubmissionInputSchema>;
export type RejectResultSubmissionInput = z.infer<typeof rejectResultSubmissionInputSchema>;
export type IngestReplayEvidenceInput = z.infer<typeof ingestReplayEvidenceInputSchema>;
export type PlayerRating = z.infer<typeof playerRatingSchema>;
export type AdjustPlayerRatingInput = z.infer<typeof adjustPlayerRatingInputSchema>;
export type RatingAdjustment = z.infer<typeof ratingAdjustmentSchema>;
export type OverrideResultSubmissionInput = z.infer<typeof overrideResultSubmissionInputSchema>;
export type BanPlayerFromQueueInput = z.infer<typeof banPlayerFromQueueInputSchema>;
export type UnbanPlayerFromQueueInput = z.infer<typeof unbanPlayerFromQueueInputSchema>;
