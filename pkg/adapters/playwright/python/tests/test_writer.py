"""Tests for NDJSON writers."""

import gzip
import json
import tempfile
from pathlib import Path

import pytest

from traffic2openapi_playwright.types import IRRecord, Request, Response, RequestMethod
from traffic2openapi_playwright.writer import NDJSONWriter, GzipNDJSONWriter


def create_test_record(path: str = "/test", status: int = 200) -> IRRecord:
    """Create a test IR record."""
    return IRRecord(
        request=Request(method=RequestMethod.GET, path=path),
        response=Response(status=status),
        id="test-id",
        timestamp="2024-01-01T00:00:00Z",
    )


class TestNDJSONWriter:
    """Tests for NDJSONWriter."""

    def test_write_single_record(self, tmp_path: Path):
        """Test writing a single record."""
        output = tmp_path / "test.ndjson"

        with NDJSONWriter(str(output)) as writer:
            writer.write(create_test_record())

        content = output.read_text()
        lines = content.strip().split("\n")

        assert len(lines) == 1
        assert writer.count == 1

        parsed = json.loads(lines[0])
        assert parsed["id"] == "test-id"
        assert parsed["request"]["path"] == "/test"

    def test_write_multiple_records(self, tmp_path: Path):
        """Test writing multiple records."""
        output = tmp_path / "test.ndjson"

        with NDJSONWriter(str(output)) as writer:
            writer.write(create_test_record("/users", 200))
            writer.write(create_test_record("/posts", 201))
            writer.write(create_test_record("/comments", 204))

        content = output.read_text()
        lines = content.strip().split("\n")

        assert len(lines) == 3
        assert writer.count == 3

        # Verify each line is valid JSON
        for line in lines:
            parsed = json.loads(line)
            assert "request" in parsed
            assert "response" in parsed

    def test_write_after_close_raises(self, tmp_path: Path):
        """Test that writing after close raises an error."""
        output = tmp_path / "test.ndjson"

        writer = NDJSONWriter(str(output))
        writer.write(create_test_record())
        writer.close()

        with pytest.raises(ValueError, match="closed"):
            writer.write(create_test_record())

    def test_flush(self, tmp_path: Path):
        """Test flushing buffered data."""
        output = tmp_path / "test.ndjson"

        with NDJSONWriter(str(output)) as writer:
            writer.write(create_test_record())
            writer.flush()

            # File should have content after flush
            content = output.read_text()
            assert len(content) > 0

    def test_context_manager(self, tmp_path: Path):
        """Test using writer as context manager."""
        output = tmp_path / "test.ndjson"

        with NDJSONWriter(str(output)) as writer:
            writer.write(create_test_record())
            count = writer.count

        # File should be closed and have content
        assert count == 1
        assert output.exists()
        assert len(output.read_text()) > 0

    def test_count_property(self, tmp_path: Path):
        """Test count property tracks writes."""
        output = tmp_path / "test.ndjson"

        with NDJSONWriter(str(output)) as writer:
            assert writer.count == 0
            writer.write(create_test_record())
            assert writer.count == 1
            writer.write(create_test_record())
            assert writer.count == 2


class TestGzipNDJSONWriter:
    """Tests for GzipNDJSONWriter."""

    def test_write_compressed(self, tmp_path: Path):
        """Test writing gzip-compressed records."""
        output = tmp_path / "test.ndjson.gz"

        with GzipNDJSONWriter(str(output)) as writer:
            writer.write(create_test_record("/users", 200))
            writer.write(create_test_record("/posts", 201))

        # Verify file is gzip compressed
        with gzip.open(output, "rt", encoding="utf-8") as f:
            lines = f.read().strip().split("\n")

        assert len(lines) == 2
        assert writer.count == 2

        parsed = json.loads(lines[0])
        assert parsed["request"]["path"] == "/users"

    def test_compression_level(self, tmp_path: Path):
        """Test different compression levels produce valid output."""
        for level in [1, 5, 9]:
            output = tmp_path / f"test-level-{level}.ndjson.gz"

            with GzipNDJSONWriter(str(output), compression_level=level) as writer:
                writer.write(create_test_record())

            # Verify file can be decompressed
            with gzip.open(output, "rt", encoding="utf-8") as f:
                content = f.read()

            assert len(content) > 0

    def test_write_after_close_raises(self, tmp_path: Path):
        """Test that writing after close raises an error."""
        output = tmp_path / "test.ndjson.gz"

        writer = GzipNDJSONWriter(str(output))
        writer.write(create_test_record())
        writer.close()

        with pytest.raises(ValueError, match="closed"):
            writer.write(create_test_record())

    def test_flush(self, tmp_path: Path):
        """Test flushing buffered data."""
        output = tmp_path / "test.ndjson.gz"

        with GzipNDJSONWriter(str(output)) as writer:
            writer.write(create_test_record())
            writer.flush()
            # Just verify no errors

    def test_compressed_smaller_than_uncompressed(self, tmp_path: Path):
        """Test that compressed output is smaller."""
        uncompressed = tmp_path / "test.ndjson"
        compressed = tmp_path / "test.ndjson.gz"

        # Create a record with repetitive data (compresses well)
        record = IRRecord(
            request=Request(
                method=RequestMethod.GET,
                path="/users",
                headers={f"header-{i}": f"value-{i}" for i in range(20)},
            ),
            response=Response(
                status=200,
                body={"data": ["item"] * 100},
            ),
            id="test-id",
            timestamp="2024-01-01T00:00:00Z",
        )

        with NDJSONWriter(str(uncompressed)) as writer:
            for _ in range(10):
                writer.write(record)

        with GzipNDJSONWriter(str(compressed)) as writer:
            for _ in range(10):
                writer.write(record)

        assert compressed.stat().st_size < uncompressed.stat().st_size
