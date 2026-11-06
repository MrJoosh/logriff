from dataclasses import dataclass, field
from http import HTTPMethod


@dataclass
class Stats:
    lines: int = 0
    matched: int = 0
    malformed: int = 0
    bytes_volume: int = 0
    response_times: list[float] = field(default_factory=list)
    status_code_distribution: dict[int, int] = field(default_factory=dict)
    route_distribution: dict[str, int] = field(default_factory=dict)


@dataclass
class LineData:
    remote_addr: str
    timestamp: str
    method: HTTPMethod
    path: str
    protocol: str
    status: int
    bytes_sent: int
    response_time_ms: float
    user_agent: str


@dataclass
class Filters:
    status_min: int | None
    status_max: int | None
    method: HTTPMethod | None
    path_prefix: str | None
    min_duration: float | None


text_output_template: str = """=== LogPulse Summary ===
Lines Processed : {lines:,}
Matched Records : {matched:,}
Skipped (Errors): {malformed}

--- Traffic & Bandwidth ---
Total Volume    : {bytes_gb:.2f} GB ({bytes_volume:,} bytes)

--- Latency (ms) ---
Min             : {lat_min:.2f} ms
Mean            : {lat_avg:.2f} ms
p95             : {lat_95th:.2f} ms
Max             : {lat_max:.2f} ms

--- Status Codes ---
{status_codes}

--- Top Routes (Top {num_routes}) ---
{top_routes}
========================
"""

status_code_template: str = "{status_code} : {number:8,}  ({percentage:6.2f}%)"
route_template: str = "{index}. {path:26} : {number:8,}"
