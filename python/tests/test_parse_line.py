from http import HTTPMethod

import pytest

from logpulse.types import LineData
from logpulse.utils import parse_line


def test_parse_line_extracts_all_fields(make_log_line) -> None:
    line = make_log_line(
        addr="192.168.1.10",
        method="GET",
        path="/api/v1/health",
        protocol="HTTP/1.1",
        status=200,
        bytes_sent=42,
        duration=1.2,
        user_agent="Go-http-client/1.1",
    )

    parsed = parse_line(line)

    assert parsed == LineData(
        remote_addr="192.168.1.10",
        timestamp="06/Nov/2026:08:12:01 +0000",
        method=HTTPMethod.GET,
        path="/api/v1/health",
        protocol="HTTP/1.1",
        status=200,
        bytes_sent=42,
        response_time_ms=1.2,
        user_agent="Go-http-client/1.1",
    )


def test_parse_line_preserves_user_agent_with_spaces(make_log_line) -> None:
    user_agent = "Mozilla/5.0 (Macintosh; Intel Mac OS X 10_15_7)"
    line = make_log_line(user_agent=user_agent)

    parsed = parse_line(line)

    assert parsed is not None
    assert parsed.user_agent == user_agent


def test_parse_line_accepts_zero_bytes(make_log_line) -> None:
    parsed = parse_line(make_log_line(bytes_sent=0))

    assert parsed is not None
    assert parsed.bytes_sent == 0


def test_parse_line_accepts_large_byte_counts(make_log_line) -> None:
    parsed = parse_line(make_log_line(bytes_sent=10_000_000_000))

    assert parsed is not None
    assert parsed.bytes_sent == 10_000_000_000


def test_parse_line_accepts_http_3(make_log_line) -> None:
    parsed = parse_line(make_log_line(protocol="HTTP/3"))

    assert parsed is not None
    assert parsed.protocol == "HTTP/3"


@pytest.mark.parametrize(
    "line",
    [
        "",
        "not an access log line",
        "10.0.0.1 - - [06/Nov/2026:08:12:01 +0000]",
        '10.0.0.1 - - [06/Nov/2026:08:12:01 +0000] "GET / HTTP/1.1" 200 42 1.2',
    ],
)
def test_obviously_malformed_lines_are_rejected(line: str) -> None:
    assert parse_line(line) is None


@pytest.mark.parametrize(
    "overrides",
    [
        {"status": "20x"},
        {"bytes_sent": "-1"},
        {"duration": "slow"},
        {"duration": "1.2.3"},
        {"method": "FLY"},
    ],
)
def test_invalid_fields_are_rejected(make_log_line, overrides: dict) -> None:
    assert parse_line(make_log_line(**overrides)) is None


def test_ipv6_record_is_accepted(make_log_line) -> None:
    parsed = parse_line(make_log_line(addr="2001:db8::1"))

    assert parsed is not None
    assert parsed.remote_addr == "2001:db8::1"


def test_query_string_path_is_accepted(make_log_line) -> None:
    parsed = parse_line(make_log_line(path="/api/v1/items?page=2"))

    assert parsed is not None


@pytest.mark.parametrize("status", ["099", "600", "999"])
def test_status_outside_100_to_599_is_rejected(make_log_line, status: str) -> None:
    assert parse_line(make_log_line(status=status)) is None


@pytest.mark.parametrize(
    "address",
    [
        "999.999.999.999",
        "256.1.1.1",
    ],
)
def test_invalid_ipv4_octets_are_rejected(make_log_line, address: str) -> None:
    assert parse_line(make_log_line(addr=address)) is None


@pytest.mark.parametrize(
    "path",
    [
        "/files/hello-world",
        "/files/a%20b.txt",
    ],
)
def test_paths_with_common_url_characters_are_accepted(
    make_log_line, path: str
) -> None:
    assert parse_line(make_log_line(path=path)) is not None
