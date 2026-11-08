"""Command-line interface for LogPulse.

See the repository root README.md for the full CLI specification and output
formats.
"""

import sys
from argparse import ArgumentParser
from http import HTTPMethod

from logpulse.types import ExitCode, Filters, Stats
from logpulse.utils import (
    filter_log,
    generate_output,
    parse_line,
    printerr,
    validate_args,
)


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
        "-e",
        "--status-max",
        type=int,
        help="keep records with status <= MAX",
    )
    parser.add_argument(
        "-m",
        "--method",
        help="filter by HTTP verb (case-insensitive)",
        choices=[m.value for m in HTTPMethod],
        type=str.upper,
    )
    parser.add_argument(
        "-p",
        "--path-prefix",
        help="keep paths starting with this prefix",
    )
    parser.add_argument(
        "--min-duration",
        type=float,
        help="keep records with response_time_ms >= duration",
    )
    parser.add_argument(
        "-n",
        "--top-routes",
        type=int,
        default=5,
        help="number of top routes to show",
    )
    parser.add_argument(
        "-f",
        "--format",
        choices=("text", "json"),
        default="text",
        help="output format",
    )
    parser.add_argument(
        "--strict",
        action="store_true",
        help="exit non-zero on malformed lines",
    )
    parser.add_argument(
        "file",
        nargs="?",
        default="-",
        help="input file, or - for stdin",
    )
    return parser


def main(argv: list[str] | None = None) -> int:

    # Initialise the stats
    stats = Stats()
    # Try to parse the filters and raise for wrong formats
    try:
        args = build_parser().parse_args(argv)
        if not validate_args(args):
            return ExitCode.VALIDATION_ERROR
        filters = Filters(
            status_min=args.status_min,
            status_max=args.status_max,
            method=HTTPMethod(args.method) if args.method else None,
            path_prefix=args.path_prefix,
            min_duration=args.min_duration,
        )
    except SystemExit as e:
        if e.code == 2:
            return ExitCode.VALIDATION_ERROR
        else:
            raise
    try:
        with (
            sys.stdin
            if args.file == "-"
            else open(args.file, "r", encoding="utf8") as logs
        ):
            for line in logs:
                stats.lines += 1
                # Attempt to parse the log line
                parsed = parse_line(line.strip())
                if not parsed:
                    if args.strict:
                        return ExitCode.STRICT_ERROR
                    stats.error += 1
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
    except FileNotFoundError:
        printerr(f"File '{args.file}' not found.")
        return ExitCode.IO_ERROR
    except ValueError as e:
        msg = e.args[0]
        if msg == "I/O operation on closed file.":
            print(f"Error: {msg}", file=sys.stderr)
            return ExitCode.IO_ERROR
        raise
    except PermissionError as e:
        printerr(str(e))

    # Present the results
    if stats.lines == 0:
        printerr("No data to process.")
        return ExitCode.IO_ERROR
    print(generate_output(args.format, args.top_routes, stats))
    return ExitCode.SUCCESS
