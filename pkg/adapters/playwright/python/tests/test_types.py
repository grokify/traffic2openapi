"""Tests for IR types."""

import json
import pytest

from traffic2openapi_playwright.types import (
    IRRecord,
    Request,
    Response,
    RequestMethod,
)


class TestRequest:
    """Tests for Request dataclass."""

    def test_minimal_request(self):
        """Test creating a request with only required fields."""
        req = Request(method=RequestMethod.GET, path="/users")

        assert req.method == RequestMethod.GET
        assert req.path == "/users"
        assert req.host is None
        assert req.headers is None

    def test_full_request(self):
        """Test creating a request with all fields."""
        req = Request(
            method=RequestMethod.POST,
            path="/users",
            scheme="https",
            host="api.example.com",
            query={"limit": "10"},
            headers={"content-type": "application/json"},
            content_type="application/json",
            body={"name": "Alice"},
        )

        assert req.method == RequestMethod.POST
        assert req.host == "api.example.com"
        assert req.query == {"limit": "10"}
        assert req.body == {"name": "Alice"}

    def test_to_dict(self):
        """Test converting request to dictionary."""
        req = Request(
            method=RequestMethod.GET,
            path="/users",
            host="api.example.com",
        )

        d = req.to_dict()

        assert d["method"] == "GET"
        assert d["path"] == "/users"
        assert d["host"] == "api.example.com"
        assert "headers" not in d  # None fields excluded

    def test_to_dict_with_body(self):
        """Test converting request with body to dictionary."""
        req = Request(
            method=RequestMethod.POST,
            path="/users",
            body={"name": "Bob"},
        )

        d = req.to_dict()

        assert d["body"] == {"name": "Bob"}


class TestResponse:
    """Tests for Response dataclass."""

    def test_minimal_response(self):
        """Test creating a response with only required fields."""
        resp = Response(status=200)

        assert resp.status == 200
        assert resp.headers is None
        assert resp.body is None

    def test_full_response(self):
        """Test creating a response with all fields."""
        resp = Response(
            status=201,
            headers={"content-type": "application/json"},
            content_type="application/json",
            body={"id": "123", "name": "Alice"},
        )

        assert resp.status == 201
        assert resp.body == {"id": "123", "name": "Alice"}

    def test_to_dict(self):
        """Test converting response to dictionary."""
        resp = Response(
            status=200,
            body={"users": []},
        )

        d = resp.to_dict()

        assert d["status"] == 200
        assert d["body"] == {"users": []}


class TestIRRecord:
    """Tests for IRRecord dataclass."""

    def test_auto_generated_fields(self):
        """Test that id, timestamp, and source are auto-generated."""
        record = IRRecord(
            request=Request(method=RequestMethod.GET, path="/test"),
            response=Response(status=200),
        )

        assert record.id is not None
        assert len(record.id) == 36  # UUID format
        assert record.timestamp is not None
        assert record.timestamp.endswith("Z")
        assert record.source == "playwright"

    def test_custom_fields(self):
        """Test creating record with custom id and timestamp."""
        record = IRRecord(
            request=Request(method=RequestMethod.GET, path="/test"),
            response=Response(status=200),
            id="custom-id",
            timestamp="2024-01-01T00:00:00Z",
            source="custom-source",
            duration_ms=45.5,
        )

        assert record.id == "custom-id"
        assert record.timestamp == "2024-01-01T00:00:00Z"
        assert record.source == "custom-source"
        assert record.duration_ms == 45.5

    def test_to_dict(self):
        """Test converting record to dictionary."""
        record = IRRecord(
            request=Request(method=RequestMethod.GET, path="/users"),
            response=Response(status=200, body={"users": []}),
            id="test-id",
            timestamp="2024-01-01T00:00:00Z",
        )

        d = record.to_dict()

        assert d["id"] == "test-id"
        assert d["timestamp"] == "2024-01-01T00:00:00Z"
        assert d["request"]["method"] == "GET"
        assert d["request"]["path"] == "/users"
        assert d["response"]["status"] == 200
        assert d["response"]["body"] == {"users": []}

    def test_to_json(self):
        """Test converting record to JSON string."""
        record = IRRecord(
            request=Request(method=RequestMethod.GET, path="/test"),
            response=Response(status=200),
            id="test-id",
            timestamp="2024-01-01T00:00:00Z",
        )

        json_str = record.to_json()
        parsed = json.loads(json_str)

        assert parsed["id"] == "test-id"
        assert parsed["request"]["method"] == "GET"

    def test_from_dict(self):
        """Test creating record from dictionary."""
        data = {
            "id": "test-id",
            "timestamp": "2024-01-01T00:00:00Z",
            "source": "playwright",
            "request": {
                "method": "POST",
                "path": "/users",
                "host": "api.example.com",
                "body": {"name": "Alice"},
            },
            "response": {
                "status": 201,
                "body": {"id": "123"},
            },
            "durationMs": 50.0,
        }

        record = IRRecord.from_dict(data)

        assert record.id == "test-id"
        assert record.request.method == RequestMethod.POST
        assert record.request.path == "/users"
        assert record.request.body == {"name": "Alice"}
        assert record.response.status == 201
        assert record.duration_ms == 50.0


class TestRequestMethod:
    """Tests for RequestMethod enum."""

    def test_all_methods(self):
        """Test all HTTP methods are defined."""
        methods = [
            RequestMethod.GET,
            RequestMethod.POST,
            RequestMethod.PUT,
            RequestMethod.PATCH,
            RequestMethod.DELETE,
            RequestMethod.HEAD,
            RequestMethod.OPTIONS,
        ]

        assert len(methods) >= 7

    def test_method_values(self):
        """Test method string values."""
        assert RequestMethod.GET.value == "GET"
        assert RequestMethod.POST.value == "POST"
        assert RequestMethod.DELETE.value == "DELETE"
