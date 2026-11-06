"""Command-line interface for LogPulse.

See the repository root README.md for the full CLI specification and output
formats.
"""

import argparse
import re
import sys
from http import HTTPMethod

from logpulse.types import LineData, Stats

LOG_LINE_REGEX_STR = r'(?P<remote_addr>\d{1,3}\.\d{1,3}\.\d{1,3}\.\d{1,3}) - - \[(?P<timestamp>[\w/: +]+)\] "(?P<method>[A-Z]{3,7}) (?P<path>[\w/.]+) (?P<protocol>[A-Z0-9./]+)" (?P<status>\d{3}) (?P<bytes_sent>\d+) (?P<response_time_ms>[\d.]+) "(?P<user_agent>.+)"'


def build_parser() -> argparse.ArgumentParser:
    parser = argparse.ArgumentParser(
        prog="logpulse",
        description="Filter and aggregate access logs.",
    )
    parser.add_argument(
        "file", nargs="?", default="-", help="input file, or - for stdin"
    )
    parser.add_argument(
        "-s", "--status-min", type=int, help="keep records with status >= MIN"
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
    return parser


def main(argv: list[str] | None = None) -> int:
    args = build_parser().parse_args(argv)

    stats = Stats()

    # TODO: stream input, apply filters, aggregate metrics, and emit output.
    with (
        sys.stdin if args.file == "-" else open(args.file, "r", encoding="utf8") as logs
    ):
        for line in logs:
            stats.lines += 1
            # Attempt to parse the log line
            parsed = parse_line(line.strip())
            if not parsed:
                if args.strict:
                    return 1
                stats.malformed += 1
                continue

            # Filter the log
            filtered = filter_log(parsed)
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
    return 1


def parse_line(line: str) -> LineData | None:
    print(f"{line=}")
    parsed = re.fullmatch(LOG_LINE_REGEX_STR, line)
    if not parsed:
        return None
    print(f"{parsed.groupdict()=}")
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


def filter_log(log: LineData) -> LineData | None:
    pass
