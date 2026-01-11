"""
Playwright traffic capture module.
"""

import json
import re
import time
from dataclasses import dataclass, field
from datetime import datetime
from pathlib import Path
from typing import Callable, Optional, Union
from urllib.parse import parse_qs, urlparse

from playwright.sync_api import BrowserContext, Page, Request as PWRequest, Response as PWResponse
from playwright.async_api import BrowserContext as AsyncBrowserContext

from .types import IRRecord, Request, Response, RequestMethod
from .writer import NDJSONWriter, GzipNDJSONWriter


# Default headers to exclude (security-sensitive)
DEFAULT_EXCLUDE_HEADERS = frozenset([
    "authorization",
    "cookie",
    "set-cookie",
    "x-api-key",
    "x-auth-token",
    "x-csrf-token",
    "proxy-authorization",
])


@dataclass
class CaptureOptions:
    """Configuration options for traffic capture."""

    # Output file path
    output: Union[str, Path]

    # Filter options
    filter_hosts: list[str] = field(default_factory=list)
    filter_methods: list[str] = field(default_factory=list)
    exclude_paths: list[str] = field(default_factory=list)
    exclude_path_patterns: list[re.Pattern] = field(default_factory=list)

    # Header filtering
    exclude_headers: set[str] = field(default_factory=lambda: set(DEFAULT_EXCLUDE_HEADERS))
    include_headers: bool = True

    # Body capture
    capture_request_body: bool = True
    capture_response_body: bool = True
    max_body_size: int = 1024 * 1024  # 1MB

    # Content type filtering for body capture
    capture_content_types: list[str] = field(default_factory=lambda: [
        "application/json",
        "application/xml",
        "text/xml",
        "text/plain",
        "text/html",
    ])

    # Compression
    gzip: bool = False
    compression_level: int = 9

    # Error handling
    on_error: Optional[Callable[[Exception], None]] = None


class PlaywrightCapture:
    """
    Captures HTTP traffic from Playwright and writes IR records.

    Usage (sync):
        with sync_playwright() as p:
            browser = p.chromium.launch()
            context = browser.new_context()

            capture = PlaywrightCapture("traffic.ndjson")
            capture.attach(context)

            page = context.new_page()
            page.goto("https://api.example.com")

            capture.close()
            browser.close()

    Usage (async):
        async with async_playwright() as p:
            browser = await p.chromium.launch()
            context = await browser.new_context()

            capture = PlaywrightCapture("traffic.ndjson")
            await capture.attach_async(context)

            page = await context.new_page()
            await page.goto("https://api.example.com")

            capture.close()
            await browser.close()
    """

    def __init__(
        self,
        output: Union[str, Path, CaptureOptions],
        **kwargs,
    ):
        """
        Initialize traffic capture.

        Args:
            output: Output file path or CaptureOptions.
            **kwargs: Additional options (if output is a path).
        """
        if isinstance(output, CaptureOptions):
            self.options = output
        else:
            self.options = CaptureOptions(output=output, **kwargs)

        # Create writer
        if self.options.gzip or str(self.options.output).endswith(".gz"):
            self._writer = GzipNDJSONWriter(
                self.options.output,
                compression_level=self.options.compression_level,
            )
        else:
            self._writer = NDJSONWriter(self.options.output)

        # Track pending requests (for timing)
        self._pending_requests: dict[str, tuple[PWRequest, float]] = {}

    def attach(self, context: BrowserContext) -> None:
        """
        Attach to a Playwright browser context (sync).

        Args:
            context: Playwright BrowserContext to capture traffic from.
        """
        context.on("request", self._on_request)
        context.on("response", self._on_response)

    async def attach_async(self, context: AsyncBrowserContext) -> None:
        """
        Attach to a Playwright browser context (async).

        Args:
            context: Playwright async BrowserContext to capture traffic from.
        """
        context.on("request", self._on_request)
        context.on("response", self._on_response_async)

    def _on_request(self, request: PWRequest) -> None:
        """Handle request event."""
        # Store request with timestamp for duration calculation
        self._pending_requests[request.url] = (request, time.time())

    def _on_response(self, response: PWResponse) -> None:
        """Handle response event (sync)."""
        try:
            record = self._create_record(response)
            if record:
                self._writer.write(record)
        except Exception as e:
            if self.options.on_error:
                self.options.on_error(e)

    async def _on_response_async(self, response: PWResponse) -> None:
        """Handle response event (async)."""
        try:
            record = await self._create_record_async(response)
            if record:
                self._writer.write(record)
        except Exception as e:
            if self.options.on_error:
                self.options.on_error(e)

    def _should_capture(self, request: PWRequest) -> bool:
        """Check if request should be captured based on filters."""
        url = urlparse(request.url)

        # Host filter
        if self.options.filter_hosts:
            if url.hostname not in self.options.filter_hosts:
                return False

        # Method filter
        if self.options.filter_methods:
            if request.method.upper() not in [m.upper() for m in self.options.filter_methods]:
                return False

        # Path exclusion
        if self.options.exclude_paths:
            if url.path in self.options.exclude_paths:
                return False

        # Path pattern exclusion
        for pattern in self.options.exclude_path_patterns:
            if pattern.match(url.path):
                return False

        return True

    def _filter_headers(self, headers: dict[str, str]) -> dict[str, str]:
        """Filter out excluded headers."""
        if not self.options.include_headers:
            return {}
        return {
            k.lower(): v
            for k, v in headers.items()
            if k.lower() not in self.options.exclude_headers
        }

    def _should_capture_body(self, content_type: Optional[str]) -> bool:
        """Check if body should be captured based on content type."""
        if not content_type:
            return False
        # Check if content type starts with any allowed type
        for allowed in self.options.capture_content_types:
            if content_type.startswith(allowed):
                return True
        return False

    def _parse_body(self, body: bytes, content_type: Optional[str]) -> Optional[any]:
        """Parse body based on content type."""
        if not body or len(body) > self.options.max_body_size:
            return None

        try:
            if content_type and "json" in content_type:
                return json.loads(body.decode("utf-8"))
            else:
                return body.decode("utf-8")
        except (json.JSONDecodeError, UnicodeDecodeError):
            return None

    def _create_record(self, response: PWResponse) -> Optional[IRRecord]:
        """Create IR record from Playwright response (sync)."""
        request = response.request

        if not self._should_capture(request):
            return None

        # Get timing info
        start_time = None
        if request.url in self._pending_requests:
            _, start_time = self._pending_requests.pop(request.url)

        url = urlparse(request.url)

        # Parse query parameters
        query_params = {}
        if url.query:
            parsed = parse_qs(url.query)
            query_params = {k: v[0] if len(v) == 1 else v for k, v in parsed.items()}

        # Get request body
        request_body = None
        if self.options.capture_request_body:
            try:
                post_data = request.post_data
                if post_data:
                    content_type = request.headers.get("content-type", "")
                    if self._should_capture_body(content_type):
                        if isinstance(post_data, str):
                            post_data = post_data.encode("utf-8")
                        request_body = self._parse_body(post_data, content_type)
            except Exception:
                pass

        # Get response body
        response_body = None
        if self.options.capture_response_body:
            try:
                content_type = response.headers.get("content-type", "")
                if self._should_capture_body(content_type):
                    body = response.body()
                    response_body = self._parse_body(body, content_type)
            except Exception:
                pass

        # Calculate duration
        duration_ms = None
        if start_time:
            duration_ms = (time.time() - start_time) * 1000

        return IRRecord(
            request=Request(
                method=RequestMethod(request.method.upper()),
                path=url.path or "/",
                scheme=url.scheme,
                host=url.hostname,
                query=query_params if query_params else None,
                headers=self._filter_headers(request.headers) or None,
                content_type=request.headers.get("content-type"),
                body=request_body,
            ),
            response=Response(
                status=response.status,
                headers=self._filter_headers(response.headers) or None,
                content_type=response.headers.get("content-type"),
                body=response_body,
            ),
            duration_ms=duration_ms,
        )

    async def _create_record_async(self, response: PWResponse) -> Optional[IRRecord]:
        """Create IR record from Playwright response (async)."""
        request = response.request

        if not self._should_capture(request):
            return None

        # Get timing info
        start_time = None
        if request.url in self._pending_requests:
            _, start_time = self._pending_requests.pop(request.url)

        url = urlparse(request.url)

        # Parse query parameters
        query_params = {}
        if url.query:
            parsed = parse_qs(url.query)
            query_params = {k: v[0] if len(v) == 1 else v for k, v in parsed.items()}

        # Get request body
        request_body = None
        if self.options.capture_request_body:
            try:
                post_data = request.post_data
                if post_data:
                    content_type = request.headers.get("content-type", "")
                    if self._should_capture_body(content_type):
                        if isinstance(post_data, str):
                            post_data = post_data.encode("utf-8")
                        request_body = self._parse_body(post_data, content_type)
            except Exception:
                pass

        # Get response body (async)
        response_body = None
        if self.options.capture_response_body:
            try:
                content_type = response.headers.get("content-type", "")
                if self._should_capture_body(content_type):
                    body = await response.body()
                    response_body = self._parse_body(body, content_type)
            except Exception:
                pass

        # Calculate duration
        duration_ms = None
        if start_time:
            duration_ms = (time.time() - start_time) * 1000

        return IRRecord(
            request=Request(
                method=RequestMethod(request.method.upper()),
                path=url.path or "/",
                scheme=url.scheme,
                host=url.hostname,
                query=query_params if query_params else None,
                headers=self._filter_headers(request.headers) or None,
                content_type=request.headers.get("content-type"),
                body=request_body,
            ),
            response=Response(
                status=response.status,
                headers=self._filter_headers(response.headers) or None,
                content_type=response.headers.get("content-type"),
                body=response_body,
            ),
            duration_ms=duration_ms,
        )

    def flush(self) -> None:
        """Flush any buffered records."""
        self._writer.flush()

    def close(self) -> None:
        """Close the capture and writer."""
        self._writer.close()

    @property
    def count(self) -> int:
        """Number of records captured."""
        return self._writer.count

    def __enter__(self) -> "PlaywrightCapture":
        return self

    def __exit__(self, exc_type, exc_val, exc_tb) -> None:
        self.close()
