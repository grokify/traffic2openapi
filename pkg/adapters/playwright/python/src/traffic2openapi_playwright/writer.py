"""
NDJSON writers for IR records.
"""

import gzip
import threading
from pathlib import Path
from typing import IO, Optional, Union

from .types import IRRecord


class NDJSONWriter:
    """Writes IR records in NDJSON format (newline-delimited JSON)."""

    def __init__(
        self,
        output: Union[str, Path, IO[str]],
        *,
        flush_interval: int = 1,
    ):
        """
        Initialize NDJSON writer.

        Args:
            output: File path or file-like object to write to.
            flush_interval: Flush after this many writes (0 = no auto-flush).
        """
        self._lock = threading.Lock()
        self._count = 0
        self._flush_interval = flush_interval
        self._closed = False

        if isinstance(output, (str, Path)):
            self._file: IO[str] = open(output, "w", encoding="utf-8")
            self._owns_file = True
        else:
            self._file = output
            self._owns_file = False

    def write(self, record: IRRecord) -> None:
        """
        Write a single IR record.

        Args:
            record: The IR record to write.

        Raises:
            ValueError: If the writer has been closed.
        """
        with self._lock:
            if self._closed:
                raise ValueError("Writer has been closed")

            self._file.write(record.to_json())
            self._file.write("\n")
            self._count += 1

            if self._flush_interval > 0 and self._count % self._flush_interval == 0:
                self._file.flush()

    def flush(self) -> None:
        """Flush any buffered data."""
        with self._lock:
            if not self._closed:
                self._file.flush()

    def close(self) -> None:
        """Close the writer and underlying file."""
        with self._lock:
            if self._closed:
                return
            self._closed = True
            self._file.flush()
            if self._owns_file:
                self._file.close()

    @property
    def count(self) -> int:
        """Number of records written."""
        return self._count

    def __enter__(self) -> "NDJSONWriter":
        return self

    def __exit__(self, exc_type, exc_val, exc_tb) -> None:
        self.close()


class GzipNDJSONWriter:
    """Writes IR records in gzip-compressed NDJSON format."""

    def __init__(
        self,
        output: Union[str, Path],
        *,
        compression_level: int = 9,
        flush_interval: int = 10,
    ):
        """
        Initialize gzip NDJSON writer.

        Args:
            output: File path to write to.
            compression_level: Gzip compression level (1-9).
            flush_interval: Flush after this many writes (0 = no auto-flush).
        """
        self._lock = threading.Lock()
        self._count = 0
        self._flush_interval = flush_interval
        self._closed = False

        self._file = gzip.open(
            output,
            "wt",
            encoding="utf-8",
            compresslevel=compression_level,
        )

    def write(self, record: IRRecord) -> None:
        """
        Write a single IR record.

        Args:
            record: The IR record to write.

        Raises:
            ValueError: If the writer has been closed.
        """
        with self._lock:
            if self._closed:
                raise ValueError("Writer has been closed")

            self._file.write(record.to_json())
            self._file.write("\n")
            self._count += 1

            if self._flush_interval > 0 and self._count % self._flush_interval == 0:
                self._file.flush()

    def flush(self) -> None:
        """Flush any buffered data."""
        with self._lock:
            if not self._closed:
                self._file.flush()

    def close(self) -> None:
        """Close the writer and underlying file."""
        with self._lock:
            if self._closed:
                return
            self._closed = True
            self._file.close()

    @property
    def count(self) -> int:
        """Number of records written."""
        return self._count

    def __enter__(self) -> "GzipNDJSONWriter":
        return self

    def __exit__(self, exc_type, exc_val, exc_tb) -> None:
        self.close()
