/**
 * traffic2openapi-playwright: Capture Playwright HTTP traffic for OpenAPI generation.
 */

export { IRRecord, Request, Response, RequestMethod } from './types.js';
export { PlaywrightCapture, CaptureOptions, DEFAULT_EXCLUDE_HEADERS } from './capture.js';
export { NDJSONWriter, GzipNDJSONWriter } from './writer.js';
