"""Tests for the standalone synthetic log generator.

These live apart from the Python implementation tests because the generator is
shared test tooling for every language implementation.
"""

import importlib.util
import subprocess
import sys
from datetime import UTC, datetime
from pathlib import Path

import pytest

GENERATOR_PATH = Path(__file__).resolve().parents[1] / "generate_logs.py"


@pytest.fixture(scope="module")
def generator():
    spec = importlib.util.spec_from_file_location("generate_logs", GENERATOR_PATH)
    assert spec is not None
    assert spec.loader is not None

    module = importlib.util.module_from_spec(spec)
    spec.loader.exec_module(module)
    return module


def _run_generator(*args: str) -> subprocess.CompletedProcess[str]:
    return subprocess.run(
        [sys.executable, str(GENERATOR_PATH), *args],
        capture_output=True,
        text=True,
        check=False,
    )


def test_generation_is_deterministic(generator) -> None:
    assert list(generator.generate(100, seed=42, malformed=5)) == list(
        generator.generate(100, seed=42, malformed=5)
    )


def test_generate_zero_lines(generator) -> None:
    assert list(generator.generate(0, seed=42, malformed=0)) == []


def test_exact_number_of_malformed_lines(generator) -> None:
    lines = list(generator.generate(50, seed=42, malformed=10))

    assert len(lines) == 50
    assert sum(line in generator.MALFORMED_LINES for line in lines) == 10


def test_no_malformed_lines_when_requested_count_is_zero(generator) -> None:
    lines = list(generator.generate(50, seed=42, malformed=0))

    assert all(line not in generator.MALFORMED_LINES for line in lines)


def _line_index(line: str) -> int:
    timestamp = line.split("[", 1)[1].split("]", 1)[0]
    parsed = datetime.strptime(timestamp, "%d/%b/%Y:%H:%M:%S %z")
    start = datetime(2026, 11, 6, 8, 0, 0, tzinfo=UTC)
    return int((parsed - start).total_seconds())


def test_malformed_lines_do_not_change_valid_line_sequence(generator) -> None:
    clean_lines = list(generator.generate(100, seed=42, malformed=0))
    mixed_lines = list(generator.generate(100, seed=42, malformed=20))
    valid_lines = [
        line for line in mixed_lines if line not in generator.MALFORMED_LINES
    ]

    assert len(valid_lines) == 80
    for line in valid_lines:
        assert clean_lines[_line_index(line)] == line


def test_output_includes_ipv6_and_query_paths(generator) -> None:
    lines = list(generator.generate(20, seed=42, malformed=0))

    assert any(":" in line.split(" ", 1)[0] for line in lines)
    assert any("?" in line for line in lines)


def test_cli_output_is_deterministic(tmp_path: Path) -> None:
    first = tmp_path / "first.log"
    second = tmp_path / "second.log"

    first_result = _run_generator(
        "--lines", "30", "--seed", "7", "--malformed", "3", "--out", str(first)
    )
    second_result = _run_generator(
        "--lines", "30", "--seed", "7", "--malformed", "3", "--out", str(second)
    )

    assert first_result.returncode == 0
    assert second_result.returncode == 0
    assert first.read_text(encoding="utf-8") == second.read_text(encoding="utf-8")


def test_cli_writes_requested_number_of_lines(tmp_path: Path) -> None:
    output = tmp_path / "output.log"

    result = _run_generator(
        "--lines", "30", "--seed", "7", "--malformed", "3", "--out", str(output)
    )

    assert result.returncode == 0
    assert len(output.read_text(encoding="utf-8").splitlines()) == 30


def test_cli_writes_to_stdout_by_default() -> None:
    result = _run_generator("--lines", "5", "--seed", "7")

    assert result.returncode == 0
    assert len(result.stdout.splitlines()) == 5


def test_cli_rejects_more_malformed_lines_than_total_lines() -> None:
    result = _run_generator("--lines", "5", "--malformed", "6")

    assert result.returncode == 1
    assert "--malformed" in result.stderr


def test_cli_rejects_negative_malformed_count() -> None:
    result = _run_generator("--lines", "5", "--malformed", "-1")

    assert result.returncode == 1
    assert "--malformed" in result.stderr
