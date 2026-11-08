import json
import re
import sys
from argparse import Namespace
from http import HTTPMethod
from statistics import fmean, quantiles

from logpulse.types import Filters, LineData, OutputDict, OutputTopRoute, Stats

LOG_LINE_REGEX_STR = r'(?P<remote_addr>\d{1,3}\.\d{1,3}\.\d{1,3}\.\d{1,3}) - - \[(?P<timestamp>[\w/: +]+)\] "(?P<method>[A-Z]{3,7}) (?P<path>[\w/.]+) (?P<protocol>[A-Z0-9./]+)" (?P<status>\d{3}) (?P<bytes_sent>\d+) (?P<response_time_ms>[\d.]+) "(?P<user_agent>.+)"'


def printerr(msg: str) -> None:
    print(f"Error: {msg}", file=sys.stderr)


def validate_args(args: Namespace) -> bool:
    if args.status_min and args.status_max and args.status_min > args.status_max:
        printerr("Status min filter can't be greater than status max filter.")
        return False
    if args.top_routes and args.top_routes < 0:
        printerr("Can't specify a negative numer of routes.")
        return False
    return True


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


def generate_output(format: str, num_top_routes: int, stats: Stats) -> str:
    # Build the dict format, and then use that to build the text output if required
    output_dict: OutputDict = {
        "summary": {
            "total_lines": stats.lines,
            "matched_lines": stats.matched,
            "error_lines": stats.error,
        },
        "traffic": {
            "total_bytes": stats.bytes_volume,
        },
        "latency_ms": {
            "min": min(stats.response_times),
            "mean": fmean(stats.response_times),
            "p95": quantiles(sorted(stats.response_times), n=100, method="inclusive")[
                94
            ],
            "max": max(stats.response_times),
        },
        "status_codes": {
            str(code): count
            for code, count in sorted(stats.status_code_distribution.items())
        },
        "top_routes": [
            OutputTopRoute(path=path, count=count)
            for path, count in {
                k: v
                for k, v in sorted(
                    stats.route_distribution.items(),
                    key=lambda x: x[1],
                    reverse=True,
                )
            }.items()
        ][:num_top_routes],
    }
    if format == "text":
        # Build status code output
        status_code_lines: list[str] = [
            f"{status_code} : {count:8,}  ({((count / sum(output_dict['status_codes'].values())) * 100):6.2f}%)"
            for status_code, count in output_dict["status_codes"].items()
        ]
        top_route_lines: list[str] = [
            f"{index}. {route['path']:26} : {route['count']:8,}"
            for index, route in enumerate(output_dict["top_routes"])
        ]

        return f"""=== LogPulse Summary ===
Lines Processed : {output_dict["summary"]["total_lines"]:,}
Matched Records : {output_dict["summary"]["matched_lines"]:,}
Skipped (Errors): {output_dict["summary"]["error_lines"]:,}

--- Traffic & Bandwidth ---
Total Volume    : {(output_dict["traffic"]["total_bytes"] / 1000000000):.2f} GB ({output_dict["traffic"]["total_bytes"]:,} bytes)

--- Latency (ms) ---
Min             : {output_dict["latency_ms"]["min"]:.2f} ms
Mean            : {output_dict["latency_ms"]["mean"]:.2f} ms
p95             : {output_dict["latency_ms"]["p95"]:.2f} ms
Max             : {output_dict["latency_ms"]["max"]:.2f} ms

--- Status Codes ---
{"\n".join(status_code_lines)}

--- Top Routes (Top {num_top_routes}) ---
{"\n".join(top_route_lines)}
========================
        """

    return json.dumps(output_dict)
