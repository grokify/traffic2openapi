#!/usr/bin/env python3
"""
Async example of capturing Playwright traffic.

Usage:
    python async_example.py
"""

import asyncio

from playwright.async_api import async_playwright

from traffic2openapi_playwright import PlaywrightCapture


async def main():
    print("Async Playwright capture example")

    async with async_playwright() as p:
        browser = await p.chromium.launch(headless=True)
        context = await browser.new_context()

        # Start capturing
        capture = PlaywrightCapture("traffic-async.ndjson")
        await capture.attach_async(context)

        page = await context.new_page()

        # Make some API calls
        await page.goto("https://httpbin.org/get")
        await page.goto("https://httpbin.org/headers")

        # POST request via fetch
        await page.evaluate("""
            async () => {
                await fetch('https://httpbin.org/post', {
                    method: 'POST',
                    headers: {'Content-Type': 'application/json'},
                    body: JSON.stringify({message: 'Hello from async!'})
                });
            }
        """)

        # Wait for request to complete
        await page.wait_for_timeout(1000)

        # Close
        capture.close()
        await browser.close()

        print(f"Captured {capture.count} requests")
        print("Output: traffic-async.ndjson")


if __name__ == "__main__":
    asyncio.run(main())
