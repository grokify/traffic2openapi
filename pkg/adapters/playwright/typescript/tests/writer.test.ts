/**
 * Tests for NDJSON writers.
 */

import { describe, it, expect, beforeEach, afterEach } from 'vitest';
import { mkdtempSync, rmSync, readFileSync, createReadStream } from 'fs';
import { join } from 'path';
import { tmpdir } from 'os';
import { createGunzip } from 'zlib';
import { pipeline } from 'stream/promises';
import { Writable } from 'stream';

import { NDJSONWriter, GzipNDJSONWriter } from '../src/writer.js';
import { createIRRecord, type IRRecord } from '../src/types.js';

function createTestRecord(path = '/test', status = 200): IRRecord {
  return createIRRecord(
    { method: 'GET', path },
    { status },
    { id: 'test-id', timestamp: '2024-01-01T00:00:00Z' }
  );
}

describe('NDJSONWriter', () => {
  let tmpDir: string;

  beforeEach(() => {
    tmpDir = mkdtempSync(join(tmpdir(), 'ndjson-test-'));
  });

  afterEach(() => {
    rmSync(tmpDir, { recursive: true, force: true });
  });

  it('should write a single record', async () => {
    const output = join(tmpDir, 'test.ndjson');
    const writer = new NDJSONWriter(output);

    writer.write(createTestRecord());
    await writer.close();

    const content = readFileSync(output, 'utf-8');
    const lines = content.trim().split('\n');

    expect(lines.length).toBe(1);
    expect(writer.count).toBe(1);

    const parsed = JSON.parse(lines[0]);
    expect(parsed.id).toBe('test-id');
    expect(parsed.request.path).toBe('/test');
  });

  it('should write multiple records', async () => {
    const output = join(tmpDir, 'test.ndjson');
    const writer = new NDJSONWriter(output);

    writer.write(createTestRecord('/users', 200));
    writer.write(createTestRecord('/posts', 201));
    writer.write(createTestRecord('/comments', 204));
    await writer.close();

    const content = readFileSync(output, 'utf-8');
    const lines = content.trim().split('\n');

    expect(lines.length).toBe(3);
    expect(writer.count).toBe(3);

    // Verify each line is valid JSON
    for (const line of lines) {
      const parsed = JSON.parse(line);
      expect(parsed.request).toBeDefined();
      expect(parsed.response).toBeDefined();
    }
  });

  it('should throw when writing after close', async () => {
    const output = join(tmpDir, 'test.ndjson');
    const writer = new NDJSONWriter(output);

    writer.write(createTestRecord());
    await writer.close();

    expect(() => writer.write(createTestRecord())).toThrow('closed');
  });

  it('should flush buffered data', async () => {
    const output = join(tmpDir, 'test.ndjson');
    const writer = new NDJSONWriter(output);

    writer.write(createTestRecord());
    await writer.flush();

    // File should have content after flush
    const content = readFileSync(output, 'utf-8');
    expect(content.length).toBeGreaterThan(0);

    await writer.close();
  });

  it('should track count correctly', async () => {
    const output = join(tmpDir, 'test.ndjson');
    const writer = new NDJSONWriter(output);

    expect(writer.count).toBe(0);
    writer.write(createTestRecord());
    expect(writer.count).toBe(1);
    writer.write(createTestRecord());
    expect(writer.count).toBe(2);

    await writer.close();
  });
});

describe('GzipNDJSONWriter', () => {
  let tmpDir: string;

  beforeEach(() => {
    tmpDir = mkdtempSync(join(tmpdir(), 'gzip-test-'));
  });

  afterEach(() => {
    rmSync(tmpDir, { recursive: true, force: true });
  });

  async function readGzipFile(path: string): Promise<string> {
    const chunks: Buffer[] = [];
    const writable = new Writable({
      write(chunk, encoding, callback) {
        chunks.push(chunk);
        callback();
      },
    });

    await pipeline(createReadStream(path), createGunzip(), writable);

    return Buffer.concat(chunks).toString('utf-8');
  }

  it('should write gzip-compressed records', async () => {
    const output = join(tmpDir, 'test.ndjson.gz');
    const writer = new GzipNDJSONWriter(output);

    writer.write(createTestRecord('/users', 200));
    writer.write(createTestRecord('/posts', 201));
    await writer.close();

    // Verify file is gzip compressed
    const content = await readGzipFile(output);
    const lines = content.trim().split('\n');

    expect(lines.length).toBe(2);
    expect(writer.count).toBe(2);

    const parsed = JSON.parse(lines[0]);
    expect(parsed.request.path).toBe('/users');
  });

  it('should support different compression levels', async () => {
    for (const level of [1, 5, 9]) {
      const output = join(tmpDir, `test-level-${level}.ndjson.gz`);
      const writer = new GzipNDJSONWriter(output, level);

      writer.write(createTestRecord());
      await writer.close();

      // Verify file can be decompressed
      const content = await readGzipFile(output);
      expect(content.length).toBeGreaterThan(0);
    }
  });

  it('should throw when writing after close', async () => {
    const output = join(tmpDir, 'test.ndjson.gz');
    const writer = new GzipNDJSONWriter(output);

    writer.write(createTestRecord());
    await writer.close();

    expect(() => writer.write(createTestRecord())).toThrow('closed');
  });

  it('should produce smaller output than uncompressed', async () => {
    const uncompressed = join(tmpDir, 'test.ndjson');
    const compressed = join(tmpDir, 'test.ndjson.gz');

    // Create a record with repetitive data (compresses well)
    const record = createIRRecord(
      {
        method: 'GET',
        path: '/users',
        headers: Object.fromEntries(
          Array.from({ length: 20 }, (_, i) => [`header-${i}`, `value-${i}`])
        ),
      },
      {
        status: 200,
        body: { data: Array(100).fill('item') },
      },
      { id: 'test-id', timestamp: '2024-01-01T00:00:00Z' }
    );

    const plainWriter = new NDJSONWriter(uncompressed);
    for (let i = 0; i < 10; i++) {
      plainWriter.write(record);
    }
    await plainWriter.close();

    const gzipWriter = new GzipNDJSONWriter(compressed);
    for (let i = 0; i < 10; i++) {
      gzipWriter.write(record);
    }
    await gzipWriter.close();

    const plainSize = readFileSync(uncompressed).length;
    const gzipSize = readFileSync(compressed).length;

    expect(gzipSize).toBeLessThan(plainSize);
  });
});
