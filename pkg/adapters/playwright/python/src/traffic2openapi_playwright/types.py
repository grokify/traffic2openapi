"""
IR (Intermediate Representation) types for traffic2openapi.

These types match the JSON Schema at schemas/ir.v1.schema.json.
"""

from dataclasses import dataclass, field, asdict
from datetime import datetime
from enum import Enum
from typing import Any, Optional
import json
import uuid


class RequestMethod(str, Enum):
    """HTTP request methods."""
    GET = "GET"
    POST = "POST"
    PUT = "PUT"
    PATCH = "PATCH"
    DELETE = "DELETE"
    HEAD = "HEAD"
    OPTIONS = "OPTIONS"
    TRACE = "TRACE"
    CONNECT = "CONNECT"


@dataclass
class Request:
    """HTTP request details."""
    method: RequestMethod
    path: str
    scheme: Optional[str] = None
    host: Optional[str] = None
    path_template: Optional[str] = None
    path_params: Optional[dict[str, str]] = None
    query: Optional[dict[str, str]] = None
    headers: Optional[dict[str, str]] = None
    content_type: Optional[str] = None
    body: Optional[Any] = None

    def to_dict(self) -> dict:
        """Convert to dictionary for JSON serialization."""
        result = {
            "method": self.method.value if isinstance(self.method, RequestMethod) else self.method,
            "path": self.path,
        }
        if self.scheme:
            result["scheme"] = self.scheme
        if self.host:
            result["host"] = self.host
        if self.path_template:
            result["pathTemplate"] = self.path_template
        if self.path_params:
            result["pathParams"] = self.path_params
        if self.query:
            result["query"] = self.query
        if self.headers:
            result["headers"] = self.headers
        if self.content_type:
            result["contentType"] = self.content_type
        if self.body is not None:
            result["body"] = self.body
        return result


@dataclass
class Response:
    """HTTP response details."""
    status: int
    headers: Optional[dict[str, str]] = None
    content_type: Optional[str] = None
    body: Optional[Any] = None

    def to_dict(self) -> dict:
        """Convert to dictionary for JSON serialization."""
        result = {"status": self.status}
        if self.headers:
            result["headers"] = self.headers
        if self.content_type:
            result["contentType"] = self.content_type
        if self.body is not None:
            result["body"] = self.body
        return result


@dataclass
class IRRecord:
    """A single HTTP request/response capture."""
    request: Request
    response: Response
    id: Optional[str] = None
    timestamp: Optional[str] = None
    source: Optional[str] = None
    duration_ms: Optional[float] = None

    def __post_init__(self):
        """Set defaults after initialization."""
        if self.id is None:
            self.id = str(uuid.uuid4())
        if self.timestamp is None:
            self.timestamp = datetime.utcnow().isoformat() + "Z"
        if self.source is None:
            self.source = "playwright"

    def to_dict(self) -> dict:
        """Convert to dictionary for JSON serialization."""
        result = {
            "request": self.request.to_dict(),
            "response": self.response.to_dict(),
        }
        if self.id:
            result["id"] = self.id
        if self.timestamp:
            result["timestamp"] = self.timestamp
        if self.source:
            result["source"] = self.source
        if self.duration_ms is not None:
            result["durationMs"] = self.duration_ms
        return result

    def to_json(self) -> str:
        """Convert to JSON string."""
        return json.dumps(self.to_dict(), separators=(",", ":"))

    @classmethod
    def from_dict(cls, data: dict) -> "IRRecord":
        """Create an IRRecord from a dictionary."""
        request_data = data["request"]
        request = Request(
            method=RequestMethod(request_data["method"]),
            path=request_data["path"],
            scheme=request_data.get("scheme"),
            host=request_data.get("host"),
            path_template=request_data.get("pathTemplate"),
            path_params=request_data.get("pathParams"),
            query=request_data.get("query"),
            headers=request_data.get("headers"),
            content_type=request_data.get("contentType"),
            body=request_data.get("body"),
        )

        response_data = data["response"]
        response = Response(
            status=response_data["status"],
            headers=response_data.get("headers"),
            content_type=response_data.get("contentType"),
            body=response_data.get("body"),
        )

        return cls(
            request=request,
            response=response,
            id=data.get("id"),
            timestamp=data.get("timestamp"),
            source=data.get("source"),
            duration_ms=data.get("durationMs"),
        )
