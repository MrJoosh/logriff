from pathlib import Path

import pytest

REPO_ROOT = Path(__file__).resolve().parents[2]
DATA_DIR = REPO_ROOT / "data"


@pytest.fixture
def data_dir() -> Path:
    return DATA_DIR


@pytest.fixture
def sample_log(data_dir: Path) -> Path:
    return data_dir / "sample.log"


@pytest.fixture
def malformed_sample_log(data_dir: Path) -> Path:
    return data_dir / "malformed_sample.log"


@pytest.fixture
def write_log(tmp_path: Path):
    def _write_log(content: str, name: str = "input.log") -> Path:
        path = tmp_path / name
        path.write_text(content, encoding="utf-8")
        return path

    return _write_log


@pytest.fixture
def make_log_line():
    def _make_log_line(
        *,
        addr: str = "192.168.1.10",
        method: str = "GET",
        path: str = "/api/v1/health",
        protocol: str = "HTTP/1.1",
        status: int | str = 200,
        bytes_sent: int | str = 42,
        duration: float | str = 1.2,
        user_agent: str = "test-agent",
    ) -> str:
        return (
            f'{addr} - - [06/Nov/2026:08:12:01 +0000] '
            f'"{method} {path} {protocol}" '
            f'{status} {bytes_sent} {duration} "{user_agent}"'
        )

    return _make_log_line
