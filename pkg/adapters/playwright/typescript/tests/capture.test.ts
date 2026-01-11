/**
 * Tests for Playwright capture module.
 */

import { describe, it, expect, beforeEach, afterEach } from 'vitest';
import { mkdtempSync, rmSync, readFileSync, existsSync } from 'fs';
import { join } from 'path';
import { tmpdir } from 'os';

import {
  PlaywrightCapture,
  CaptureOptions,
  DEFAULT_EXCLUDE_HEADERS,
} from '../src/capture.js';

describe('DEFAULT_EXCLUDE_HEADERS', () => {
  it('should include security-sensitive headers', () => {
    expect(DEFAULT_EXCLUDE_HEADERS.has('authorization')).toBe(true);
    expect(DEFAULT_EXCLUDE_HEADERS.has('cookie')).toBe(true);
    expect(DEFAULT_EXCLUDE_HEADERS.has('set-cookie')).toBe(true);
    expect(DEFAULT_EXCLUDE_HEADERS.has('x-api-key')).toBe(true);
  });
});

describe('PlaywrightCapture', () => {
  let tmpDir: string;

  beforeEach(() => {
    tmpDir = mkdtempSync(join(tmpdir(), 'capture-test-'));
  });

  afterEach(() => {
    rmSync(tmpDir, { recursive: true, force: true });
  });

  describe('constructor', () => {
    it('should create with string output path', async () => {
      const output = join(tmpDir, 'test.ndjson');

      const capture = new PlaywrightCapture(output);
      await capture.close();

      expect(existsSync(output)).toBe(true);
    });

    it('should create with CaptureOptions', async () => {
      const output = join(tmpDir, 'test.ndjson');
      const options: CaptureOptions = {
        output,
        filterHosts: ['api.example.com'],
      };

      const capture = new PlaywrightCapture(options);
      await capture.close();

      expect(existsSync(output)).toBe(true);
    });

    it('should enable gzip based on file extension', async () => {
      const output = join(tmpDir, 'test.ndjson.gz');

      const capture = new PlaywrightCapture(output);
      await capture.close();

      expect(existsSync(output)).toBe(true);

      // Verify it's a gzip file by checking magic bytes
      const content = readFileSync(output);
      expect(content[0]).toBe(0x1f);
      expect(content[1]).toBe(0x8b);
    });
  });

  describe('count', () => {
    it('should start at zero', async () => {
      const output = join(tmpDir, 'test.ndjson');

      const capture = new PlaywrightCapture(output);
      expect(capture.count).toBe(0);
      await capture.close();
    });
  });

  // Note: Full integration tests with actual Playwright would require
  // browser automation. These are unit tests for the capture logic.
});

describe('CaptureOptions defaults', () => {
  it('should have expected defaults', () => {
    const options: CaptureOptions = {
      output: 'test.ndjson',
    };

    // Verify only output is required
    expect(options.output).toBe('test.ndjson');
    expect(options.filterHosts).toBeUndefined();
    expect(options.excludePaths).toBeUndefined();
    expect(options.gzip).toBeUndefined();
  });

  it('should accept all options', () => {
    const options: CaptureOptions = {
      output: 'test.ndjson.gz',
      filterHosts: ['api.example.com'],
      filterMethods: ['GET', 'POST'],
      excludePaths: ['/health'],
      excludePathPatterns: [/^\/_next/],
      excludeHeaders: new Set(['x-custom']),
      includeHeaders: true,
      captureRequestBody: true,
      captureResponseBody: true,
      maxBodySize: 1024 * 1024,
      captureContentTypes: ['application/json'],
      gzip: true,
      compressionLevel: 6,
      onError: (error) => console.error(error),
    };

    expect(options.filterHosts).toEqual(['api.example.com']);
    expect(options.filterMethods).toEqual(['GET', 'POST']);
    expect(options.gzip).toBe(true);
    expect(options.compressionLevel).toBe(6);
  });
});
