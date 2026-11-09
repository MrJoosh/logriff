#!/usr/bin/env python3
"""Synthetic access-log generator for LogRiff test data.

Generated logs include a mix of IPv4 and IPv6 clients, and paths both with and
without query strings, as permitted by the LogRiff input specification.

Usage:
    python3 data/generate_logs.py --lines 1000000 --seed 42 --out data/test_1m.log
    python3 data/generate_logs.py --lines 100 --seed 42 --malformed 5 --out data/malformed.log
"""

import argparse
import contextlib
import random
import sys
from datetime import UTC, datetime, timedelta

METHODS = ["GET", "POST", "PUT", "DELETE", "HEAD", "PATCH", "OPTIONS"]
PATHS = [
    "/api/v1/health",
    "/api/v1/items",
    "/api/v1/auth/login",
    "/api/v1/users",
    "/api/v1/sessions/current",
    "/assets/main.css",
    "/assets/app.js",
]
IPV6_PREFIXES = ["2001:db8", "fe80"]
QUERY_KEYS = ["page", "ref", "sort"]
STATUSES = [200, 200, 200, 200, 200, 201, 204, 301, 304, 400, 401, 403, 404, 500, 503]
USER_AGENTS = [
    "Go-http-client/1.1",
    "curl/8.1.2",
    "PostmanRuntime/7.32.3",
    "Mozilla/5.0 (Windows NT 10.0; Win64; x64)",
    "Mozilla/5.0 (Macintosh; Intel Mac OS X 10_15_7)",
]
MALFORMED_LINES = [
    "this is not an access log line",
    '10.0.0.5 - - [06/Nov/2026:08:12:02 +0000] "POST /api/v1/auth/login HTTP/1.1" 20x 128 85.4 "Mozilla/5.0"',
    '10.0.0.5 - - [06/Nov/2026:08:12:02 +0000] "POST /api/v1/auth/login HTTP/1.1" 401 -12 85.4 "Mozilla/5.0"',
    '10.0.0.5 - - [06/Nov/2026:08:12:02 +0000] "POST /api/v1/auth/login HTTP/1.1" 401 128 slow "Mozilla/5.0"',
    '10.0.0.5 - - [06/Nov/2026:08:12:02 +0000] "POST /api/v1/auth/login HTTP/1.1" 401 128 85.4',
    '10.0.0.5 - - [06/Nov/2026:08:12:02 +0000] "FLY /api/v1/health HTTP/1.1" 200 42 1.2 "agent"',
    "10.0.0.5 - - [06/Nov/2026:08:12:02 +0000]",
    "",
]


def parse_args(argv=None):
    parser = argparse.ArgumentParser(
        description="Generate synthetic LogRiff access logs."
    )
    parser.add_argument(
        "--lines", type=int, default=1000, help="number of lines to generate"
    )
    parser.add_argument(
        "--seed", type=int, default=42, help="random seed for deterministic output"
    )
    parser.add_argument(
        "--malformed",
        type=int,
        default=0,
        help="number of malformed lines to inject (deterministic for a given seed)",
    )
    parser.add_argument("--out", default="-", help="output file, or - for stdout")
    return parser.parse_args(argv)


def generate(lines, seed, malformed=0):
    rng = random.Random(seed)
    malformed_rng = random.Random(seed ^ 0x5EED)
    malformed_indices = set(malformed_rng.sample(range(lines), malformed))
    start = datetime(2026, 11, 6, 8, 0, 0, tzinfo=UTC)

    for i in range(lines):
        if i % 7 == 3:
            remote_addr = (
                f"{rng.choice(IPV6_PREFIXES)}::"
                f"{rng.randrange(1, 0xFFFF):x}:{rng.randrange(1, 0xFFFF):x}"
            )
        else:
            remote_addr = (
                f"10.{rng.randrange(256)}.{rng.randrange(256)}.{rng.randrange(1, 255)}"
            )
        timestamp = (start + timedelta(seconds=i)).strftime("%d/%b/%Y:%H:%M:%S +0000")
        method = rng.choice(METHODS)
        path = rng.choice(PATHS)
        if i % 5 == 2:
            path = f"{path}?{rng.choice(QUERY_KEYS)}={rng.randrange(1, 1000)}"
        protocol = rng.choice(["HTTP/1.1", "HTTP/2.0"])
        status = rng.choice(STATUSES)
        bytes_sent = 0 if status == 304 else rng.randrange(0, 20_000)
        response_time_ms = round(rng.expovariate(1 / 25.0) + 0.1, 1)
        user_agent = rng.choice(USER_AGENTS)

        line = (
            f'{remote_addr} - - [{timestamp}] "{method} {path} {protocol}" '
            f'{status} {bytes_sent} {response_time_ms} "{user_agent}"'
        )
        if i in malformed_indices:
            yield malformed_rng.choice(MALFORMED_LINES)
        else:
            yield line


def main(argv=None):
    args = parse_args(argv)

    if args.lines < 0:
        print("error: --lines must be non-negative", file=sys.stderr)
        return 1
    if not 0 <= args.malformed <= args.lines:
        print("error: --malformed must be between 0 and --lines", file=sys.stderr)
        return 1

    with contextlib.ExitStack() as stack:
        if args.out == "-":
            out = sys.stdout
        else:
            out = stack.enter_context(open(args.out, "w", encoding="utf-8"))

        out.writelines(
            f"{line}\n"
            for line in generate(args.lines, args.seed, args.malformed)
        )

    return 0


if __name__ == "__main__":
    raise SystemExit(main())
