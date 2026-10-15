"""
traffic2openapi-playwright: Capture Playwright HTTP traffic for OpenAPI generation.
"""

from .capture import CaptureOptions, PlaywrightCapture
from .types import IRRecord, Request, RequestMethod, Response
from .writer import GzipNDJSONWriter, NDJSONWriter

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
