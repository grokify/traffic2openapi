/**
 * Basic example of capturing Playwright traffic.
 *
 * Usage:
 *   npx ts-node examples/basic.ts
 *
 * This will:
 * 1. Launch a browser
 * 2. Navigate to httpbin.org and make some API calls
 * 3. Save captured traffic to traffic.ndjson
 * 4. Print the number of captured requests
 */

import { chromium } from 'playwright';
import { PlaywrightCapture } from '../src/index.js';

async function simpleExample() {
  console.log('Example 1: Simple capture');

  const browser = await chromium.launch({ headless: true });
  const context = await browser.newContext();

  // Start capturing
  const capture = new PlaywrightCapture('traffic.ndjson');
  capture.attach(context);

  const page = await context.newPage();

  // Make some API calls
  await page.goto('https://httpbin.org/get');
  await page.goto('https://httpbin.org/headers');
  await page.goto('https://httpbin.org/ip');

  // Close capture and browser
  await capture.close();
  await browser.close();

  console.log(`  Captured ${capture.count} requests`);
  console.log('  Output: traffic.ndjson');
}

async function optionsExample() {
  console.log('\nExample 2: Capture with options');

  const browser = await chromium.launch({ headless: true });
  const context = await browser.newContext();

  // Configure capture options
  const capture = new PlaywrightCapture({
    output: 'traffic-filtered.ndjson.gz',
    filterHosts: ['httpbin.org'],
    excludePaths: ['/favicon.ico'],
    gzip: true,
    onError: (err) => console.error('Capture error:', err),
  });

  capture.attach(context);

  const page = await context.newPage();

  // Make API calls
  await page.goto('https://httpbin.org/get?foo=bar');

  // POST request via fetch
  await page.evaluate(async () => {
    await fetch('https://httpbin.org/post', {
      method: 'POST',
      headers: { 'Content-Type': 'application/json' },
      body: JSON.stringify({ name: 'Alice', age: 30 }),
    });
  });

  // Wait for request to complete
  await page.waitForTimeout(1000);

  await capture.close();
  await browser.close();

  console.log(`  Captured ${capture.count} requests`);
  console.log('  Output: traffic-filtered.ndjson.gz');
}

async function main() {
  await simpleExample();
  await optionsExample();

  console.log(
    "\nDone! Run 'traffic2openapi generate -i traffic.ndjson -o openapi.yaml' to generate OpenAPI spec."
  );
}

main().catch(console.error);
