"""Command-line interface for LogPulse.

See the repository root README.md for the full CLI specification and output
formats.
"""

import re
import sys
from argparse import ArgumentParser, Namespace, ArgumentError
from http import HTTPMethod

from logpulse.types import Filters, LineData, Stats

LOG_LINE_REGEX_STR = r'(?P<remote_addr>\d{1,3}\.\d{1,3}\.\d{1,3}\.\d{1,3}) - - \[(?P<timestamp>[\w/: +]+)\] "(?P<method>[A-Z]{3,7}) (?P<path>[\w/.]+) (?P<protocol>[A-Z0-9./]+)" (?P<status>\d{3}) (?P<bytes_sent>\d+) (?P<response_time_ms>[\d.]+) "(?P<user_agent>.+)"'


def build_parser() -> ArgumentParser:
    parser = ArgumentParser(
        prog="logpulse",
        description="Filter and aggregate logs.",
    )
    parser.add_argument(
        "-s",
        "--status-min",
        type=int,
        help="keep records with status >= MIN",
    )
    parser.add_argument(
        "-e", "--status-max", type=int, help="keep records with status <= MAX"
    )
    parser.add_argument("-m", "--method", help="filter by HTTP verb (case-insensitive)")
    parser.add_argument(
        "-p", "--path-prefix", help="keep paths starting with this prefix"
    )
    parser.add_argument(
        "--min-duration",
        type=float,
        help="keep records with response_time_ms >= duration",
    )
    parser.add_argument(
        "-n", "--top-routes", type=int, default=5, help="number of top routes to show"
    )
    parser.add_argument(
        "-f", "--format", choices=("text", "json"), default="text", help="output format"
    )
    parser.add_argument(
        "--strict", action="store_true", help="exit non-zero on malformed lines"
    )
    parser.add_argument(
        "file", nargs="?", default="-", help="input file, or - for stdin"
    )
    return parser


def main(argv: list[str] | None = None) -> int:

    # Initialise the stats
    stats = Stats()
    # Try to parse the filters and raise for wrong formats
    try:
        args = build_parser().parse_args(argv)
        filters = Filters(
            status_min=args.status_min,
            status_max=args.status_max,
            method=HTTPMethod(args.method) if args.method else None,
            path_prefix=args.path_prefix,
            min_duration=args.min_duration,
        )
    except (ValueError, ArgumentError) as e:
        print(e)
        return 1

    with (
        sys.stdin if args.file == "-" else open(args.file, "r", encoding="utf8") as logs
    ):
        for line in logs:
            stats.lines += 1
            # Attempt to parse the log line
            parsed = parse_line(line.strip())
            if not parsed:
                if args.strict:
                    return 3
                stats.malformed += 1
                continue

            # Filter the log
            filtered = filter_log(filters, parsed)
            if not filtered:
                continue
            stats.matched += 1
            stats.bytes_volume += filtered.bytes_sent
            stats.response_times.append(filtered.response_time_ms)
            stats.status_code_distribution[filtered.status] = (
                stats.status_code_distribution.get(filtered.status, 0) + 1
            )
            stats.route_distribution[filtered.path] = (
                stats.route_distribution.get(filtered.path, 0) + 1
            )

    # Present the results
    print(stats)
    return 0


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
