import { z } from "zod";

import { env } from "../config/env";
import { ApiError } from "./errors";

const apiErrorPayloadSchema = z
  .object({
    code: z.string().optional(),
    message: z.string().optional(),
    requestId: z.string().optional(),
  })
  .partial();

async function parseApiError(response: Response): Promise<ApiError> {
  let code: string | undefined;
  let message: string | undefined;
  let requestId: string | undefined;

  try {
    const parsed = apiErrorPayloadSchema.safeParse(await response.json());
    if (parsed.success) {
      code = parsed.data.code;
      message = parsed.data.message;
      requestId = parsed.data.requestId;
    }
  } catch {
    // Ignore parse failures and fallback to generic error.
  }

  return new ApiError(message ?? `request failed with status ${response.status}`, response.status, code, requestId);
}

async function requestJson<T>(
  method: "GET" | "POST" | "PATCH" | "DELETE",
  path: string,
  schema: z.ZodType<T>,
  body?: unknown,
): Promise<T> {
  const response = await fetch(`${env.apiBaseUrl}${path}`, {
    body: body === undefined ? undefined : JSON.stringify(body),
    credentials: "include",
    headers: {
      Accept: "application/json",
      "Content-Type": "application/json",
    },
    method,
  });

  if (!response.ok) {
    throw await parseApiError(response);
  }

  const payload = await response.json();
  return schema.parse(payload);
}

export async function getJson<T>(path: string, schema: z.ZodType<T>): Promise<T> {
  return requestJson("GET", path, schema);
}

export async function postJson<TResponse, TRequest>(
  path: string,
  body: TRequest,
  schema: z.ZodType<TResponse>,
): Promise<TResponse> {
  return requestJson("POST", path, schema, body);
}

export async function patchJson<TResponse, TRequest>(
  path: string,
  body: TRequest,
  schema: z.ZodType<TResponse>,
): Promise<TResponse> {
  return requestJson("PATCH", path, schema, body);
}

export async function deleteJson<TResponse, TRequest>(
  path: string,
  body: TRequest,
  schema: z.ZodType<TResponse>,
): Promise<TResponse> {
  return requestJson("DELETE", path, schema, body);
}
