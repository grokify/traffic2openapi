"""
traffic2openapi-playwright: Capture Playwright HTTP traffic for OpenAPI generation.
"""

from .types import IRRecord, Request, Response, RequestMethod
from .capture import PlaywrightCapture, CaptureOptions
from .writer import NDJSONWriter, GzipNDJSONWriter

__version__ = "0.1.0"

__all__ = [
    # Types
    "IRRecord",
    "Request",
    "Response",
    "RequestMethod",
    # Capture
    "PlaywrightCapture",
    "CaptureOptions",
    # Writers
    "NDJSONWriter",
    "GzipNDJSONWriter",
]
