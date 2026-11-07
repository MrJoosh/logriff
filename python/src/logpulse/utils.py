import json
import re
from http import HTTPMethod
from statistics import fmean, quantiles

from logpulse.types import Filters, LineData, OutputDict, OutputTopRoute, Stats

LOG_LINE_REGEX_STR = r'(?P<remote_addr>\d{1,3}\.\d{1,3}\.\d{1,3}\.\d{1,3}) - - \[(?P<timestamp>[\w/: +]+)\] "(?P<method>[A-Z]{3,7}) (?P<path>[\w/.]+) (?P<protocol>[A-Z0-9./]+)" (?P<status>\d{3}) (?P<bytes_sent>\d+) (?P<response_time_ms>[\d.]+) "(?P<user_agent>.+)"'


def parse_line(line: str) -> LineData | None:
    """Helper to parse the log line into the relevant component pieces."""
    parsed = re.fullmatch(LOG_LINE_REGEX_STR, line)
    if not parsed:
        return None
    try:
        fields = parsed.groupdict()
        return LineData(
            remote_addr=fields["remote_addr"],
            timestamp=fields["timestamp"],
            method=HTTPMethod(fields["method"]),
            path=fields["path"],
            protocol=fields["protocol"],
            status=int(fields["status"]),
            bytes_sent=int(fields["bytes_sent"]),
            response_time_ms=float(fields["response_time_ms"]),
            user_agent=fields["user_agent"],
        )
    except ValueError:
        return None


def filter_log(filters: Filters, log: LineData) -> LineData | None:
    """Helper to apply the given filters to the log line."""
    if filters.status_min and log.status < filters.status_min:
        return None
    if filters.status_max and log.status > filters.status_max:
        return None
    if filters.method and log.method != filters.method:
        return None
    if filters.path_prefix and not log.path.startswith(filters.path_prefix):
        return None
    if filters.min_duration and log.response_time_ms < filters.min_duration:
        return None
    return log

