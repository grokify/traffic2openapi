/**
 * Playwright traffic capture module.
 */

import type { BrowserContext, Request as PWRequest, Response as PWResponse } from 'playwright';

import { IRRecord, Request, Response, RequestMethod, createIRRecord } from './types.js';
import { NDJSONWriter, GzipNDJSONWriter } from './writer.js';

/**
 * Default headers to exclude (security-sensitive).
 */
export const DEFAULT_EXCLUDE_HEADERS = new Set([
  'authorization',
  'cookie',
  'set-cookie',
  'x-api-key',
  'x-auth-token',
  'x-csrf-token',
  'proxy-authorization',
]);

/**
 * Configuration options for traffic capture.
 */
export interface CaptureOptions {
  /** Output file path. */
  output: string;

  /** Only capture requests to these hosts. */
  filterHosts?: string[];

  /** Only capture these HTTP methods. */
  filterMethods?: RequestMethod[];

  /** Skip requests matching these paths. */
  excludePaths?: string[];

  /** Skip requests matching these path patterns. */
  excludePathPatterns?: RegExp[];

  /** Headers to exclude from capture. */
  excludeHeaders?: Set<string>;

  /** Include headers in output (default: true). */
  includeHeaders?: boolean;

  /** Capture request bodies (default: true). */
  captureRequestBody?: boolean;

  /** Capture response bodies (default: true). */
  captureResponseBody?: boolean;

  /** Max body size in bytes (default: 1MB). */
  maxBodySize?: number;

  /** Content types to capture bodies for. */
  captureContentTypes?: string[];

  /** Use gzip compression (default: false). */
  gzip?: boolean;

  /** Gzip compression level 1-9 (default: 9). */
  compressionLevel?: number;

  /** Error handler callback. */
  onError?: (error: Error) => void;
}

/**
 * Default capture options.
 */
const DEFAULT_OPTIONS: Partial<CaptureOptions> = {
  excludeHeaders: DEFAULT_EXCLUDE_HEADERS,
  includeHeaders: true,
  captureRequestBody: true,
  captureResponseBody: true,
  maxBodySize: 1024 * 1024, // 1MB
  captureContentTypes: [
    'application/json',
    'application/xml',
    'text/xml',
    'text/plain',
    'text/html',
  ],
  gzip: false,
  compressionLevel: 9,
};

/**
 * Captures HTTP traffic from Playwright and writes IR records.
 *
 * @example
 * ```typescript
 * import { chromium } from 'playwright';
 * import { PlaywrightCapture } from 'traffic2openapi-playwright';
 *
 * const browser = await chromium.launch();
 * const context = await browser.newContext();
 *
 * const capture = new PlaywrightCapture({ output: 'traffic.ndjson' });
 * capture.attach(context);
 *
 * const page = await context.newPage();
 * await page.goto('https://api.example.com');
 *
 * await capture.close();
 * await browser.close();
 * ```
 */
export class PlaywrightCapture {
  private options: Required<CaptureOptions>;
  private writer: NDJSONWriter | GzipNDJSONWriter;
  private pendingRequests = new Map<string, { request: PWRequest; startTime: number }>();

  /**
   * Create a new traffic capture instance.
   * @param options - Capture options or output file path.
   */
  constructor(options: CaptureOptions | string) {
    if (typeof options === 'string') {
      options = { output: options };
    }

    this.options = { ...DEFAULT_OPTIONS, ...options } as Required<CaptureOptions>;

    // Create writer
    if (this.options.gzip || this.options.output.endsWith('.gz')) {
      this.writer = new GzipNDJSONWriter(this.options.output, this.options.compressionLevel);
    } else {
      this.writer = new NDJSONWriter(this.options.output);
    }
  }

  /**
   * Attach to a Playwright browser context.
   * @param context - Playwright BrowserContext to capture traffic from.
   */
  attach(context: BrowserContext): void {
    context.on('request', (request) => this.onRequest(request));
    context.on('response', (response) => this.onResponse(response));
  }

  private onRequest(request: PWRequest): void {
    // Store request with timestamp for duration calculation
    this.pendingRequests.set(request.url(), {
      request,
      startTime: Date.now(),
    });
  }

  private async onResponse(response: PWResponse): Promise<void> {
    try {
      const record = await this.createRecord(response);
      if (record) {
        this.writer.write(record);
      }
    } catch (error) {
      if (this.options.onError) {
        this.options.onError(error as Error);
      }
    }
  }

  private shouldCapture(request: PWRequest): boolean {
    const url = new URL(request.url());

    // Host filter
    if (this.options.filterHosts?.length) {
      if (!this.options.filterHosts.includes(url.hostname)) {
        return false;
      }
    }

    // Method filter
    if (this.options.filterMethods?.length) {
      if (!this.options.filterMethods.includes(request.method() as RequestMethod)) {
        return false;
      }
    }

    // Path exclusion
    if (this.options.excludePaths?.length) {
      if (this.options.excludePaths.includes(url.pathname)) {
        return false;
      }
    }

    // Path pattern exclusion
    if (this.options.excludePathPatterns?.length) {
      for (const pattern of this.options.excludePathPatterns) {
        if (pattern.test(url.pathname)) {
          return false;
        }
      }
    }

    return true;
  }

  private filterHeaders(headers: Record<string, string>): Record<string, string> | undefined {
    if (!this.options.includeHeaders) {
      return undefined;
    }

    const filtered: Record<string, string> = {};
    for (const [key, value] of Object.entries(headers)) {
      if (!this.options.excludeHeaders.has(key.toLowerCase())) {
        filtered[key.toLowerCase()] = value;
      }
    }
    return Object.keys(filtered).length > 0 ? filtered : undefined;
  }

  private shouldCaptureBody(contentType?: string): boolean {
    if (!contentType) return false;
    return this.options.captureContentTypes.some((type) => contentType.startsWith(type));
  }

  private parseBody(body: Buffer, contentType?: string): unknown {
    if (!body || body.length > this.options.maxBodySize) {
      return undefined;
    }

    try {
      const text = body.toString('utf-8');
      if (contentType?.includes('json')) {
        return JSON.parse(text);
      }
      return text;
    } catch {
      return undefined;
    }
  }

  private async createRecord(response: PWResponse): Promise<IRRecord | null> {
    const pwRequest = response.request();

    if (!this.shouldCapture(pwRequest)) {
      return null;
    }

    // Get timing info
    const pending = this.pendingRequests.get(pwRequest.url());
    if (pending) {
      this.pendingRequests.delete(pwRequest.url());
    }

    const url = new URL(pwRequest.url());

    // Parse query parameters
    const query: Record<string, string | string[]> = {};
    url.searchParams.forEach((value, key) => {
      const existing = query[key];
      if (existing) {
        if (Array.isArray(existing)) {
          existing.push(value);
        } else {
          query[key] = [existing, value];
        }
      } else {
        query[key] = value;
      }
    });

    // Get request body
    let requestBody: unknown;
    if (this.options.captureRequestBody) {
      try {
        const postData = pwRequest.postData();
        if (postData) {
          const contentType = pwRequest.headers()['content-type'];
          if (this.shouldCaptureBody(contentType)) {
            requestBody = this.parseBody(Buffer.from(postData), contentType);
          }
        }
      } catch {
        // Ignore
      }
    }

    // Get response body
    let responseBody: unknown;
    if (this.options.captureResponseBody) {
      try {
        const contentType = response.headers()['content-type'];
        if (this.shouldCaptureBody(contentType)) {
          const body = await response.body();
          responseBody = this.parseBody(body, contentType);
        }
      } catch {
        // Ignore
      }
    }

    // Calculate duration
    const durationMs = pending ? Date.now() - pending.startTime : undefined;

    const request: Request = {
      method: pwRequest.method() as RequestMethod,
      path: url.pathname || '/',
      scheme: url.protocol.replace(':', ''),
      host: url.hostname,
      query: Object.keys(query).length > 0 ? query : undefined,
      headers: this.filterHeaders(pwRequest.headers()),
      contentType: pwRequest.headers()['content-type'],
      body: requestBody,
    };

    const resp: Response = {
      status: response.status(),
      headers: this.filterHeaders(response.headers()),
      contentType: response.headers()['content-type'],
      body: responseBody,
    };

    return createIRRecord(request, resp, { durationMs });
  }

  /**
   * Flush any buffered records.
   */
  async flush(): Promise<void> {
    await this.writer.flush();
  }

  /**
   * Close the capture and writer.
   */
  async close(): Promise<void> {
    await this.writer.close();
  }

  /**
   * Number of records captured.
   */
  get count(): number {
    return this.writer.count;
  }
}
