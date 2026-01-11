#!/usr/bin/env python3
"""
Basic example of capturing Playwright traffic.

Usage:
    python basic.py

This will:
1. Launch a browser
2. Navigate to httpbin.org and make some API calls
3. Save captured traffic to traffic.ndjson
4. Print the number of captured requests
"""

from playwright.sync_api import sync_playwright

from traffic2openapi_playwright import PlaywrightCapture, CaptureOptions


def main():
    # Example 1: Simple usage
    print("Example 1: Simple capture")
    with sync_playwright() as p:
        browser = p.chromium.launch(headless=True)
        context = browser.new_context()

        # Start capturing
        capture = PlaywrightCapture("traffic.ndjson")
        capture.attach(context)

        page = context.new_page()

        # Make some API calls
        page.goto("https://httpbin.org/get")
        page.goto("https://httpbin.org/headers")
        page.goto("https://httpbin.org/ip")

        # Close capture and browser
        capture.close()
        browser.close()

        print(f"  Captured {capture.count} requests")
        print(f"  Output: traffic.ndjson")

    # Example 2: With options
    print("\nExample 2: Capture with options")
    with sync_playwright() as p:
        browser = p.chromium.launch(headless=True)
        context = browser.new_context()

        # Configure capture options
        options = CaptureOptions(
            output="traffic-filtered.ndjson.gz",
            filter_hosts=["httpbin.org"],
            exclude_paths=["/favicon.ico"],
            gzip=True,
        )

        capture = PlaywrightCapture(options)
        capture.attach(context)

        page = context.new_page()

        # Make API calls
        page.goto("https://httpbin.org/get?foo=bar")
        page.goto("https://httpbin.org/post")

        # POST request
        page.evaluate("""
            fetch('https://httpbin.org/post', {
                method: 'POST',
                headers: {'Content-Type': 'application/json'},
                body: JSON.stringify({name: 'Alice', age: 30})
            })
        """)

        # Wait for request to complete
        page.wait_for_timeout(1000)

        capture.close()
        browser.close()

        print(f"  Captured {capture.count} requests")
        print(f"  Output: traffic-filtered.ndjson.gz")

    # Example 3: Using context manager
    print("\nExample 3: Context manager")
    with sync_playwright() as p:
        browser = p.chromium.launch(headless=True)
        context = browser.new_context()

        with PlaywrightCapture("traffic-context.ndjson") as capture:
            capture.attach(context)

            page = context.new_page()
            page.goto("https://httpbin.org/json")

            print(f"  Captured {capture.count} requests")
            print(f"  Output: traffic-context.ndjson")

        browser.close()

    print("\nDone! Run 'traffic2openapi generate -i traffic.ndjson -o openapi.yaml' to generate OpenAPI spec.")


if __name__ == "__main__":
    main()
