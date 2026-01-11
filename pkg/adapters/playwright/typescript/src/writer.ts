/**
 * NDJSON writers for IR records.
 */

import { createWriteStream, WriteStream } from 'fs';
import { createGzip, Gzip } from 'zlib';
import { pipeline } from 'stream/promises';
import { Writable } from 'stream';

import { IRRecord, toJSON } from './types.js';

/**
 * Writes IR records in NDJSON format (newline-delimited JSON).
 */
export class NDJSONWriter {
  private stream: WriteStream;
  private _count = 0;
  private _closed = false;

  /**
   * Create a new NDJSON writer.
   * @param output - File path to write to.
   */
  constructor(output: string) {
    this.stream = createWriteStream(output, { encoding: 'utf-8' });
  }

  /**
   * Write a single IR record.
   * @param record - The IR record to write.
   */
  write(record: IRRecord): void {
    if (this._closed) {
      throw new Error('Writer has been closed');
    }

    this.stream.write(toJSON(record) + '\n');
    this._count++;
  }

  /**
   * Flush any buffered data.
   */
  flush(): Promise<void> {
    return new Promise((resolve, reject) => {
      if (this._closed) {
        resolve();
        return;
      }
      this.stream.once('drain', resolve);
      this.stream.once('error', reject);
      // If buffer is not full, drain won't fire
      if (this.stream.writableLength === 0) {
        resolve();
      }
    });
  }

  /**
   * Close the writer.
   */
  close(): Promise<void> {
    return new Promise((resolve, reject) => {
      if (this._closed) {
        resolve();
        return;
      }
      this._closed = true;
      this.stream.end(() => resolve());
      this.stream.once('error', reject);
    });
  }

  /**
   * Number of records written.
   */
  get count(): number {
    return this._count;
  }
}

/**
 * Writes IR records in gzip-compressed NDJSON format.
 */
export class GzipNDJSONWriter {
  private fileStream: WriteStream;
  private gzip: Gzip;
  private _count = 0;
  private _closed = false;

  /**
   * Create a new gzip NDJSON writer.
   * @param output - File path to write to.
   * @param level - Compression level (1-9, default 9).
   */
  constructor(output: string, level = 9) {
    this.fileStream = createWriteStream(output);
    this.gzip = createGzip({ level });
    this.gzip.pipe(this.fileStream);
  }

  /**
   * Write a single IR record.
   * @param record - The IR record to write.
   */
  write(record: IRRecord): void {
    if (this._closed) {
      throw new Error('Writer has been closed');
    }

    this.gzip.write(toJSON(record) + '\n');
    this._count++;
  }

  /**
   * Flush any buffered data.
   */
  flush(): Promise<void> {
    return new Promise((resolve, reject) => {
      if (this._closed) {
        resolve();
        return;
      }
      this.gzip.flush(() => resolve());
    });
  }

  /**
   * Close the writer.
   */
  close(): Promise<void> {
    return new Promise((resolve, reject) => {
      if (this._closed) {
        resolve();
        return;
      }
      this._closed = true;
      this.gzip.end(() => {
        this.fileStream.end(() => resolve());
      });
      this.gzip.once('error', reject);
      this.fileStream.once('error', reject);
    });
  }

  /**
   * Number of records written.
   */
  get count(): number {
    return this._count;
  }
}
