"""Tests for Playwright capture module."""

import json
import re
import tempfile
from pathlib import Path
from unittest.mock import MagicMock, AsyncMock, patch

import pytest

from traffic2openapi_playwright.capture import (
    PlaywrightCapture,
    CaptureOptions,
    DEFAULT_EXCLUDE_HEADERS,
)
from traffic2openapi_playwright.types import RequestMethod


class MockRequest:
    """Mock Playwright Request."""

    def __init__(
        self,
        url: str = "https://api.example.com/users",
        method: str = "GET",
        headers: dict = None,
        post_data: str = None,
    ):
        self._url = url
        self._method = method
        self._headers = headers or {}
        self._post_data = post_data

    def url(self):
        return self._url

    @property
    def url(self):
        return self._url

    def method(self):
        return self._method

    @property
    def headers(self):
        return self._headers

    @property
    def post_data(self):
        return self._post_data


class MockResponse:
    """Mock Playwright Response."""

    def __init__(
        self,
        request: MockRequest,
        status: int = 200,
        headers: dict = None,
        body: bytes = None,
    ):
        self._request = request
        self._status = status
        self._headers = headers or {"content-type": "application/json"}
        self._body = body or b'{"data": []}'

    @property
    def request(self):
        return self._request

    def status(self):
        return self._status

    @property
    def headers(self):
        return self._headers

    def body(self):
        return self._body


class TestCaptureOptions:
    """Tests for CaptureOptions."""

    def test_default_options(self):
        """Test default option values."""
        opts = CaptureOptions(output="test.ndjson")

        assert opts.output == "test.ndjson"
        assert opts.filter_hosts == []
        assert opts.exclude_paths == []
        assert opts.capture_request_body is True
        assert opts.capture_response_body is True
        assert opts.max_body_size == 1024 * 1024
        assert opts.gzip is False

    def test_custom_options(self):
        """Test custom option values."""
        opts = CaptureOptions(
            output="test.ndjson.gz",
            filter_hosts=["api.example.com"],
            exclude_paths=["/health"],
            gzip=True,
            compression_level=6,
        )

        assert opts.filter_hosts == ["api.example.com"]
        assert opts.exclude_paths == ["/health"]
        assert opts.gzip is True
        assert opts.compression_level == 6

    def test_default_exclude_headers(self):
        """Test default excluded headers."""
        assert "authorization" in DEFAULT_EXCLUDE_HEADERS
        assert "cookie" in DEFAULT_EXCLUDE_HEADERS
        assert "set-cookie" in DEFAULT_EXCLUDE_HEADERS
        assert "x-api-key" in DEFAULT_EXCLUDE_HEADERS


class TestPlaywrightCapture:
    """Tests for PlaywrightCapture."""

    def test_create_with_string_output(self, tmp_path: Path):
        """Test creating capture with string output path."""
        output = tmp_path / "test.ndjson"

        capture = PlaywrightCapture(str(output))
        capture.close()

        assert output.exists()

    def test_create_with_options(self, tmp_path: Path):
        """Test creating capture with CaptureOptions."""
        output = tmp_path / "test.ndjson"
        opts = CaptureOptions(
            output=str(output),
            filter_hosts=["api.example.com"],
        )

        capture = PlaywrightCapture(opts)
        capture.close()

        assert output.exists()

    def test_gzip_from_extension(self, tmp_path: Path):
        """Test gzip is enabled based on file extension."""
        output = tmp_path / "test.ndjson.gz"

        capture = PlaywrightCapture(str(output))
        capture.close()

        assert output.exists()
        # Verify it's a gzip file by checking magic bytes
        with open(output, "rb") as f:
            magic = f.read(2)
        assert magic == b'\x1f\x8b'  # Gzip magic number

    def test_context_manager(self, tmp_path: Path):
        """Test using capture as context manager."""
        output = tmp_path / "test.ndjson"

        with PlaywrightCapture(str(output)) as capture:
            assert capture.count == 0

        assert output.exists()

    def test_should_capture_filter_hosts(self, tmp_path: Path):
        """Test host filtering."""
        output = tmp_path / "test.ndjson"
        opts = CaptureOptions(
            output=str(output),
            filter_hosts=["api.example.com"],
        )

        capture = PlaywrightCapture(opts)

        # Should capture
        req1 = MockRequest(url="https://api.example.com/users")
        assert capture._should_capture(req1) is True

        # Should not capture
        req2 = MockRequest(url="https://other.com/users")
        assert capture._should_capture(req2) is False

        capture.close()

    def test_should_capture_filter_methods(self, tmp_path: Path):
        """Test method filtering."""
        output = tmp_path / "test.ndjson"
        opts = CaptureOptions(
            output=str(output),
            filter_methods=["GET", "POST"],
        )

        capture = PlaywrightCapture(opts)

        # Should capture
        req1 = MockRequest(method="GET")
        assert capture._should_capture(req1) is True

        req2 = MockRequest(method="POST")
        assert capture._should_capture(req2) is True

        # Should not capture
        req3 = MockRequest(method="DELETE")
        assert capture._should_capture(req3) is False

        capture.close()

    def test_should_capture_exclude_paths(self, tmp_path: Path):
        """Test path exclusion."""
        output = tmp_path / "test.ndjson"
        opts = CaptureOptions(
            output=str(output),
            exclude_paths=["/health", "/metrics"],
        )

        capture = PlaywrightCapture(opts)

        # Should capture
        req1 = MockRequest(url="https://api.example.com/users")
        assert capture._should_capture(req1) is True

        # Should not capture
        req2 = MockRequest(url="https://api.example.com/health")
        assert capture._should_capture(req2) is False

        req3 = MockRequest(url="https://api.example.com/metrics")
        assert capture._should_capture(req3) is False

        capture.close()

    def test_should_capture_exclude_path_patterns(self, tmp_path: Path):
        """Test path pattern exclusion."""
        output = tmp_path / "test.ndjson"
        opts = CaptureOptions(
            output=str(output),
            exclude_path_patterns=[re.compile(r"^/_next"), re.compile(r"\.js$")],
        )

        capture = PlaywrightCapture(opts)

        # Should capture
        req1 = MockRequest(url="https://example.com/api/users")
        assert capture._should_capture(req1) is True

        # Should not capture
        req2 = MockRequest(url="https://example.com/_next/static/chunk.js")
        assert capture._should_capture(req2) is False

        req3 = MockRequest(url="https://example.com/bundle.js")
        assert capture._should_capture(req3) is False

        capture.close()

    def test_filter_headers(self, tmp_path: Path):
        """Test header filtering."""
        output = tmp_path / "test.ndjson"

        capture = PlaywrightCapture(str(output))

        headers = {
            "content-type": "application/json",
            "authorization": "Bearer token",
            "cookie": "session=abc",
            "x-custom": "value",
        }

        filtered = capture._filter_headers(headers)

        assert "content-type" in filtered
        assert "x-custom" in filtered
        assert "authorization" not in filtered
        assert "cookie" not in filtered

        capture.close()

    def test_filter_headers_disabled(self, tmp_path: Path):
        """Test disabling header inclusion."""
        output = tmp_path / "test.ndjson"
        opts = CaptureOptions(
            output=str(output),
            include_headers=False,
        )

        capture = PlaywrightCapture(opts)

        headers = {"content-type": "application/json"}
        filtered = capture._filter_headers(headers)

        assert filtered == {}

        capture.close()

    def test_should_capture_body(self, tmp_path: Path):
        """Test body capture content type check."""
        output = tmp_path / "test.ndjson"

        capture = PlaywrightCapture(str(output))

        assert capture._should_capture_body("application/json") is True
        assert capture._should_capture_body("application/json; charset=utf-8") is True
        assert capture._should_capture_body("text/plain") is True
        assert capture._should_capture_body("image/png") is False
        assert capture._should_capture_body(None) is False

        capture.close()

    def test_parse_body_json(self, tmp_path: Path):
        """Test parsing JSON body."""
        output = tmp_path / "test.ndjson"

        capture = PlaywrightCapture(str(output))

        body = b'{"name": "Alice", "age": 30}'
        parsed = capture._parse_body(body, "application/json")

        assert parsed == {"name": "Alice", "age": 30}

        capture.close()

    def test_parse_body_text(self, tmp_path: Path):
        """Test parsing text body."""
        output = tmp_path / "test.ndjson"

        capture = PlaywrightCapture(str(output))

        body = b"Hello, World!"
        parsed = capture._parse_body(body, "text/plain")

        assert parsed == "Hello, World!"

        capture.close()

    def test_parse_body_too_large(self, tmp_path: Path):
        """Test that large bodies are not parsed."""
        output = tmp_path / "test.ndjson"
        opts = CaptureOptions(
            output=str(output),
            max_body_size=10,
        )

        capture = PlaywrightCapture(opts)

        body = b"x" * 100
        parsed = capture._parse_body(body, "text/plain")

        assert parsed is None

        capture.close()

    def test_parse_body_invalid_json(self, tmp_path: Path):
        """Test handling invalid JSON."""
        output = tmp_path / "test.ndjson"

        capture = PlaywrightCapture(str(output))

        body = b"not valid json"
        parsed = capture._parse_body(body, "application/json")

        assert parsed is None

        capture.close()

    def test_count_property(self, tmp_path: Path):
        """Test count property."""
        output = tmp_path / "test.ndjson"

        capture = PlaywrightCapture(str(output))
        assert capture.count == 0
        capture.close()
