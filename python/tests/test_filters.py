from http import HTTPMethod

import pytest

from logpulse.types import Filters, LineData
from logpulse.utils import filter_log


def make_line(**overrides) -> LineData:
    values = {
        "remote_addr": "192.168.1.10",
        "timestamp": "06/Nov/2026:08:12:01 +0000",
        "method": HTTPMethod.GET,
        "path": "/api/v1/items",
        "protocol": "HTTP/1.1",
        "status": 200,
        "bytes_sent": 42,
        "response_time_ms": 1.5,
        "user_agent": "test-agent",
    }
    values.update(overrides)
    return LineData(**values)


def make_filters(**overrides) -> Filters:
    values = {
        "status_min": None,
        "status_max": None,
        "method": None,
        "path_prefix": None,
        "min_duration": None,
    }
    values.update(overrides)
    return Filters(**values)


@pytest.mark.parametrize(
    ("filter_overrides", "expected_match"),
    [
        ({}, True),
        ({"status_min": 200}, True),
        ({"status_min": 201}, False),
        ({"status_max": 200}, True),
        ({"status_max": 199}, False),
        ({"method": HTTPMethod.GET}, True),
        ({"method": HTTPMethod.POST}, False),
        ({"path_prefix": "/api"}, True),
        ({"path_prefix": "/api/v1/items"}, True),
        ({"path_prefix": "/assets"}, False),
        ({"min_duration": 1.5}, True),
        ({"min_duration": 1.6}, False),
        (
            {
                "status_min": 200,
                "status_max": 200,
                "method": HTTPMethod.GET,
                "path_prefix": "/api",
                "min_duration": 1.5,
            },
            True,
        ),
        (
            {
                "status_min": 201,
                "status_max": 200,
                "method": HTTPMethod.POST,
                "path_prefix": "/assets",
                "min_duration": 2.0,
            },
            False,
        ),
    ],
)
def test_filter_log(filter_overrides: dict, expected_match: bool) -> None:
    line = make_line()
    filters = make_filters(**filter_overrides)

    result = filter_log(filters, line)

    if expected_match:
        assert result is line
    else:
        assert result is None
