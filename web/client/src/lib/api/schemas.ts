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

export type ExceptionTicket = z.infer<typeof exceptionTicketSchema>;
export type TriageExceptionInput = z.infer<typeof triageExceptionInputSchema>;
export type ResolveExceptionInput = z.infer<typeof resolveExceptionInputSchema>;
