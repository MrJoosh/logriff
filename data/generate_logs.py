#!/usr/bin/env python3
"""Synthetic access-log generator for LogPulse test data.

Usage:
    python3 data/generate_logs.py --lines 1000000 --seed 42 --out data/test_1m.log
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
STATUSES = [200, 200, 200, 200, 200, 201, 204, 301, 304, 400, 401, 403, 404, 500, 503]
USER_AGENTS = [
    "Go-http-client/1.1",
    "curl/8.1.2",
    "PostmanRuntime/7.32.3",
    "Mozilla/5.0 (Windows NT 10.0; Win64; x64)",
    "Mozilla/5.0 (Macintosh; Intel Mac OS X 10_15_7)",
]


def parse_args(argv=None):
    parser = argparse.ArgumentParser(
        description="Generate synthetic LogPulse access logs."
    )
    parser.add_argument(
        "--lines", type=int, default=1000, help="number of lines to generate"
    )
    parser.add_argument(
        "--seed", type=int, default=42, help="random seed for deterministic output"
    )
    parser.add_argument("--out", default="-", help="output file, or - for stdout")
    return parser.parse_args(argv)


def generate(lines, seed):
    rng = random.Random(seed)
    start = datetime(2026, 11, 6, 8, 0, 0, tzinfo=UTC)

    for i in range(lines):
        remote_addr = (
            f"10.{rng.randrange(256)}.{rng.randrange(256)}.{rng.randrange(1, 255)}"
        )
        timestamp = (start + timedelta(seconds=i)).strftime("%d/%b/%Y:%H:%M:%S +0000")
        method = rng.choice(METHODS)
        path = rng.choice(PATHS)
        protocol = rng.choice(["HTTP/1.1", "HTTP/2.0"])
        status = rng.choice(STATUSES)
        bytes_sent = 0 if status == 304 else rng.randrange(0, 20_000)
        response_time_ms = round(rng.expovariate(1 / 25.0) + 0.1, 1)
        user_agent = rng.choice(USER_AGENTS)

        yield (
            f'{remote_addr} - - [{timestamp}] "{method} {path} {protocol}" '
            f'{status} {bytes_sent} {response_time_ms} "{user_agent}"'
        )


def main(argv=None):
    args = parse_args(argv)

    if args.lines < 0:
        print("error: --lines must be non-negative", file=sys.stderr)
        return 1

    with contextlib.ExitStack() as stack:
        if args.out == "-":
            out = sys.stdout
        else:
            out = stack.enter_context(open(args.out, "w", encoding="utf-8"))

        out.writelines(f"{line}\n" for line in generate(args.lines, args.seed))

    return 0


if __name__ == "__main__":
    raise SystemExit(main())
