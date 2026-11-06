"""Command-line interface for LogPulse.

See the repository root README.md for the full CLI specification and output
formats.
"""

import argparse
import sys


def build_parser() -> argparse.ArgumentParser:
    parser = argparse.ArgumentParser(
        prog="logpulse",
        description="Filter and aggregate access logs.",
    )
    parser.add_argument("file", nargs="?", default="-", help="input file, or - for stdin")
    parser.add_argument("-s", "--status-min", type=int, help="keep records with status >= MIN")
    parser.add_argument("-e", "--status-max", type=int, help="keep records with status <= MAX")
    parser.add_argument("-m", "--method", help="filter by HTTP verb (case-insensitive)")
    parser.add_argument("-p", "--path-prefix", help="keep paths starting with this prefix")
    parser.add_argument("--min-duration", type=float, help="keep records with response_time_ms >= duration")
    parser.add_argument("-n", "--top-routes", type=int, default=5, help="number of top routes to show")
    parser.add_argument("-f", "--format", choices=("text", "json"), default="text", help="output format")
    parser.add_argument("--strict", action="store_true", help="exit non-zero on malformed lines")
    return parser


def main(argv: list[str] | None = None) -> int:
    args = build_parser().parse_args(argv)

    # TODO: stream input, apply filters, aggregate metrics, and emit output.
    print("logpulse: Python implementation is not yet written", file=sys.stderr)
    return 1
