/**
 * Tests for IR types.
 */

import { describe, it, expect } from 'vitest';
import {
  createIRRecord,
  toJSON,
  fromJSON,
  type Request,
  type Response,
  type IRRecord,
  type RequestMethod,
} from '../src/types.js';

describe('Request', () => {
  it('should create a minimal request', () => {
    const req: Request = {
      method: 'GET',
      path: '/users',
    };

    expect(req.method).toBe('GET');
    expect(req.path).toBe('/users');
    expect(req.host).toBeUndefined();
  });

  it('should create a full request', () => {
    const req: Request = {
      method: 'POST',
      path: '/users',
      scheme: 'https',
      host: 'api.example.com',
      query: { limit: '10' },
      headers: { 'content-type': 'application/json' },
      contentType: 'application/json',
      body: { name: 'Alice' },
    };

    expect(req.method).toBe('POST');
    expect(req.host).toBe('api.example.com');
    expect(req.query).toEqual({ limit: '10' });
    expect(req.body).toEqual({ name: 'Alice' });
  });
});

describe('Response', () => {
  it('should create a minimal response', () => {
    const resp: Response = {
      status: 200,
    };

    expect(resp.status).toBe(200);
    expect(resp.headers).toBeUndefined();
    expect(resp.body).toBeUndefined();
  });

  it('should create a full response', () => {
    const resp: Response = {
      status: 201,
      headers: { 'content-type': 'application/json' },
      contentType: 'application/json',
      body: { id: '123', name: 'Alice' },
    };

    expect(resp.status).toBe(201);
    expect(resp.body).toEqual({ id: '123', name: 'Alice' });
  });
});

describe('createIRRecord', () => {
  it('should auto-generate id, timestamp, and source', () => {
    const record = createIRRecord(
      { method: 'GET', path: '/test' },
      { status: 200 }
    );

    expect(record.id).toBeDefined();
    expect(record.id!.length).toBe(36); // UUID format
    expect(record.timestamp).toBeDefined();
    expect(record.timestamp!.endsWith('Z')).toBe(true);
    expect(record.source).toBe('playwright');
  });

  it('should use custom options', () => {
    const record = createIRRecord(
      { method: 'GET', path: '/test' },
      { status: 200 },
      {
        id: 'custom-id',
        timestamp: '2024-01-01T00:00:00Z',
        source: 'custom-source',
        durationMs: 45.5,
      }
    );

    expect(record.id).toBe('custom-id');
    expect(record.timestamp).toBe('2024-01-01T00:00:00Z');
    expect(record.source).toBe('custom-source');
    expect(record.durationMs).toBe(45.5);
  });

  it('should include request and response', () => {
    const record = createIRRecord(
      { method: 'POST', path: '/users', body: { name: 'Bob' } },
      { status: 201, body: { id: '123' } }
    );

    expect(record.request.method).toBe('POST');
    expect(record.request.path).toBe('/users');
    expect(record.request.body).toEqual({ name: 'Bob' });
    expect(record.response.status).toBe(201);
    expect(record.response.body).toEqual({ id: '123' });
  });
});

describe('toJSON', () => {
  it('should convert record to JSON string', () => {
    const record = createIRRecord(
      { method: 'GET', path: '/test' },
      { status: 200 },
      { id: 'test-id', timestamp: '2024-01-01T00:00:00Z' }
    );

    const json = toJSON(record);
    const parsed = JSON.parse(json);

    expect(parsed.id).toBe('test-id');
    expect(parsed.request.method).toBe('GET');
    expect(parsed.request.path).toBe('/test');
    expect(parsed.response.status).toBe(200);
  });

  it('should produce compact JSON', () => {
    const record = createIRRecord(
      { method: 'GET', path: '/test' },
      { status: 200 },
      { id: 'test-id' }
    );

    const json = toJSON(record);

    // Should not have newlines or extra spaces
    expect(json.includes('\n')).toBe(false);
  });
});

describe('fromJSON', () => {
  it('should parse JSON string to record', () => {
    const json = JSON.stringify({
      id: 'test-id',
      timestamp: '2024-01-01T00:00:00Z',
      source: 'playwright',
      request: {
        method: 'POST',
        path: '/users',
        body: { name: 'Alice' },
      },
      response: {
        status: 201,
        body: { id: '123' },
      },
      durationMs: 50.0,
    });

    const record = fromJSON(json);

    expect(record.id).toBe('test-id');
    expect(record.request.method).toBe('POST');
    expect(record.request.body).toEqual({ name: 'Alice' });
    expect(record.response.status).toBe(201);
    expect(record.durationMs).toBe(50.0);
  });
});

describe('RequestMethod', () => {
  it('should support all HTTP methods', () => {
    const methods: RequestMethod[] = [
      'GET',
      'POST',
      'PUT',
      'PATCH',
      'DELETE',
      'HEAD',
      'OPTIONS',
    ];

    expect(methods.length).toBeGreaterThanOrEqual(7);
  });
});
