from dataclasses import dataclass, field
from http import HTTPMethod
from typing import TypedDict


@dataclass
class Stats:
    lines: int = 0
    matched: int = 0
    error: int = 0
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


class OutputSummary(TypedDict):
    total_lines: int
    matched_lines: int
    error_lines: int


class OutputTraffic(TypedDict):
    total_bytes: int


class OutputLatency(TypedDict):
    min: float
    mean: float
    p95: float
    max: float


class OutputTopRoute(TypedDict):
    path: str
    count: int


class OutputDict(TypedDict):
    summary: OutputSummary
    traffic: OutputTraffic
    latency_ms: OutputLatency
    status_codes: dict[str, int]
    top_routes: list[OutputTopRoute]
