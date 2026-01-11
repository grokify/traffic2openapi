/**
 * IR (Intermediate Representation) types for traffic2openapi.
 *
 * These types match the JSON Schema at schemas/ir.v1.schema.json.
 */

import { randomUUID } from 'crypto';

/**
 * HTTP request methods.
 */
export type RequestMethod =
  | 'GET'
  | 'POST'
  | 'PUT'
  | 'PATCH'
  | 'DELETE'
  | 'HEAD'
  | 'OPTIONS'
  | 'TRACE'
  | 'CONNECT';

/**
 * HTTP request details.
 */
export interface Request {
  method: RequestMethod;
  path: string;
  scheme?: string;
  host?: string;
  pathTemplate?: string;
  pathParams?: Record<string, string>;
  query?: Record<string, string | string[]>;
  headers?: Record<string, string>;
  contentType?: string;
  body?: unknown;
}

/**
 * HTTP response details.
 */
export interface Response {
  status: number;
  headers?: Record<string, string>;
  contentType?: string;
  body?: unknown;
}

/**
 * A single HTTP request/response capture.
 */
export interface IRRecord {
  id?: string;
  timestamp?: string;
  source?: string;
  request: Request;
  response: Response;
  durationMs?: number;
}

/**
 * Create a new IR record with defaults.
 */
export function createIRRecord(
  request: Request,
  response: Response,
  options?: {
    id?: string;
    timestamp?: string;
    source?: string;
    durationMs?: number;
  }
): IRRecord {
  return {
    id: options?.id ?? randomUUID(),
    timestamp: options?.timestamp ?? new Date().toISOString(),
    source: options?.source ?? 'playwright',
    request,
    response,
    durationMs: options?.durationMs,
  };
}

/**
 * Convert IR record to JSON string.
 */
export function toJSON(record: IRRecord): string {
  return JSON.stringify(record);
}

/**
 * Parse IR record from JSON string.
 */
export function fromJSON(json: string): IRRecord {
  return JSON.parse(json) as IRRecord;
}
