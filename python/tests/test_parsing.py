import json

import pytest

from logriff import main
from logriff.types import ExitCode


def _json_output(capsys) -> dict:
    return json.loads(capsys.readouterr().out)


def test_strict_parse_return_code(malformed_sample_log) -> None:
    rc = main(["--strict", str(malformed_sample_log)])

    assert rc == ExitCode.STRICT_ERROR


def test_strict_parse_prints_no_summary(malformed_sample_log, capsys) -> None:
    rc = main(["--strict", str(malformed_sample_log)])

    captured = capsys.readouterr()

    assert rc == ExitCode.STRICT_ERROR
    assert captured.out == ""


def test_strict_mode_with_valid_input(sample_log, capsys) -> None:
    rc = main(["--strict", str(sample_log)])

    captured = capsys.readouterr()

    assert rc == ExitCode.SUCCESS
    assert "=== LogRiff Summary ===" in captured.out


def test_strict_mode_with_missing_file(tmp_path) -> None:
    rc = main(["--strict", str(tmp_path / "missing.log")])

    assert rc == ExitCode.IO_ERROR


def test_malformed_count(malformed_sample_log, capsys) -> None:
    rc = main(["--format", "json", str(malformed_sample_log)])

    output_json = _json_output(capsys)

    assert rc == ExitCode.SUCCESS
    assert output_json["summary"]["total_lines"] == 50
    assert output_json["summary"]["matched_lines"] == 40
    assert output_json["summary"]["error_lines"] == 10


def test_all_malformed_non_strict_returns_success(write_log, capsys) -> None:
    path = write_log("garbage\nmore garbage\n")

    rc = main(["--format", "json", str(path)])

    output_json = _json_output(capsys)

    assert rc == ExitCode.SUCCESS
    assert output_json["summary"] == {
        "total_lines": 2,
        "matched_lines": 0,
        "error_lines": 2,
    }
    assert output_json["traffic"] == {"total_bytes": 0}
    assert output_json["status_codes"] == {}
    assert output_json["top_routes"] == []


def test_zero_matched_records_is_a_successful_run(sample_log, capsys) -> None:
    rc = main(["-s", "600", "--format", "json", str(sample_log)])

    output_json = _json_output(capsys)

    assert rc == ExitCode.SUCCESS
    assert output_json["summary"] == {
        "total_lines": 5,
        "matched_lines": 0,
        "error_lines": 0,
    }
    assert output_json["traffic"] == {"total_bytes": 0}
    assert output_json["status_codes"] == {}
    assert output_json["top_routes"] == []


def test_text_format(sample_log, capsys) -> None:
    rc = main([str(sample_log)])

    captured = capsys.readouterr()

    assert rc == ExitCode.SUCCESS
    assert "=== LogRiff Summary ===" in captured.out
    assert "Lines Processed : 5" in captured.out
    assert "Matched Records : 5" in captured.out
    assert "Skipped (Errors): 0" in captured.out
    assert "Total Volume    : 0.00 GB (9,330 bytes)" in captured.out
    assert "Min             : 0.80 ms" in captured.out
    assert "Mean            : 62.56 ms" in captured.out
    assert "p95             : 185.56 ms" in captured.out
    assert "Max             : 210.60 ms" in captured.out
    assert "200 :        2  ( 40.00%)" in captured.out
    assert "304 :        1  ( 20.00%)" in captured.out
    assert "1. /assets/main.css" in captured.out
    assert "0. /assets/main.css" not in captured.out


def test_json_output_contains_all_metrics(sample_log, capsys) -> None:
    rc = main(["--format", "json", str(sample_log)])

    output_json = _json_output(capsys)

    assert rc == ExitCode.SUCCESS
    assert output_json["summary"] == {
        "total_lines": 5,
        "matched_lines": 5,
        "error_lines": 0,
    }
    assert output_json["traffic"] == {"total_bytes": 9_330}
    assert output_json["latency_ms"]["min"] == 0.8
    assert output_json["latency_ms"]["mean"] == pytest.approx(62.56)
    assert output_json["latency_ms"]["p95"] == pytest.approx(185.56)
    assert output_json["latency_ms"]["max"] == 210.6
    assert output_json["status_codes"] == {
        "200": 2,
        "304": 1,
        "401": 1,
        "500": 1,
    }
    assert len(output_json["top_routes"]) == 5
    assert {route["path"] for route in output_json["top_routes"]} == {
        "/api/v1/health",
        "/api/v1/auth/login",
        "/api/v1/items/4912",
        "/assets/main.css",
        "/api/v1/sessions/current",
    }
    assert all(route["count"] == 1 for route in output_json["top_routes"])


def test_filtered_metrics_only_include_matched_records(sample_log, capsys) -> None:
    rc = main(["-s", "300", "--format", "json", str(sample_log)])

    output_json = _json_output(capsys)

    assert rc == ExitCode.SUCCESS
    assert output_json["summary"] == {
        "total_lines": 5,
        "matched_lines": 3,
        "error_lines": 0,
    }
    assert output_json["traffic"] == {"total_bytes": 368}
    assert output_json["status_codes"] == {
        "304": 1,
        "401": 1,
        "500": 1,
    }
    assert output_json["latency_ms"]["min"] == 0.8
    assert output_json["latency_ms"]["mean"] == pytest.approx(98.93333333333334)
    assert output_json["latency_ms"]["p95"] == pytest.approx(198.08)
    assert output_json["latency_ms"]["max"] == 210.6
    assert {route["path"] for route in output_json["top_routes"]} == {
        "/api/v1/auth/login",
        "/assets/main.css",
        "/api/v1/sessions/current",
    }


def test_combined_filters(sample_log, capsys) -> None:
    rc = main(
        [
            "-s",
            "200",
            "-e",
            "400",
            "-m",
            "GET",
            "-p",
            "/api",
            "--min-duration",
            "10",
            "--format",
            "json",
            str(sample_log),
        ]
    )

    output_json = _json_output(capsys)

    assert rc == ExitCode.SUCCESS
    assert output_json["summary"]["matched_lines"] == 1
    assert output_json["traffic"] == {"total_bytes": 8_920}
    assert output_json["top_routes"] == [{"path": "/api/v1/items/4912", "count": 1}]


def test_status_filters_are_inclusive(write_log, capsys, make_log_line) -> None:
    content = "\n".join(make_log_line(status=status) for status in (299, 300, 301))
    path = write_log(content + "\n")

    rc = main(["-s", "300", "-e", "300", "--format", "json", str(path)])

    output_json = _json_output(capsys)

    assert rc == ExitCode.SUCCESS
    assert output_json["summary"]["matched_lines"] == 1
    assert output_json["status_codes"] == {"300": 1}


def test_method_filter_is_case_insensitive(sample_log, capsys) -> None:
    rc = main(["-m", "get", "--format", "json", str(sample_log)])

    output_json = _json_output(capsys)

    assert rc == ExitCode.SUCCESS
    assert output_json["summary"]["matched_lines"] == 3


def test_top_routes_limit(write_log, capsys, make_log_line) -> None:
    content = "\n".join(
        [
            make_log_line(path="/a"),
            make_log_line(path="/a"),
            make_log_line(path="/a"),
            make_log_line(path="/b"),
            make_log_line(path="/b"),
            make_log_line(path="/c"),
        ]
    )
    path = write_log(content + "\n")

    rc = main(["-n", "2", "--format", "json", str(path)])

    output_json = _json_output(capsys)

    assert rc == ExitCode.SUCCESS
    assert output_json["top_routes"] == [
        {"path": "/a", "count": 3},
        {"path": "/b", "count": 2},
    ]


def test_top_routes_larger_than_route_count(write_log, capsys, make_log_line) -> None:
    path = write_log(make_log_line(path="/only") + "\n")

    rc = main(["-n", "10", "--format", "json", str(path)])

    output_json = _json_output(capsys)

    assert rc == ExitCode.SUCCESS
    assert output_json["top_routes"] == [{"path": "/only", "count": 1}]


def test_ipv6_record_is_parsed(write_log, capsys, make_log_line) -> None:
    path = write_log(make_log_line(addr="2001:db8::1") + "\n")

    rc = main(["--format", "json", str(path)])

    output_json = _json_output(capsys)

    assert rc == ExitCode.SUCCESS
    assert output_json["summary"] == {
        "total_lines": 1,
        "matched_lines": 1,
        "error_lines": 0,
    }


def test_query_strings_are_excluded_from_route_aggregation(
    write_log, capsys, make_log_line
) -> None:
    content = "\n".join(
        [
            make_log_line(path="/api/v1/items?page=1"),
            make_log_line(path="/api/v1/items?page=2"),
        ]
    )
    path = write_log(content + "\n")

    rc = main(["--format", "json", str(path)])

    output_json = _json_output(capsys)

    assert rc == ExitCode.SUCCESS
    assert output_json["summary"]["matched_lines"] == 2
    assert output_json["top_routes"] == [{"path": "/api/v1/items", "count": 2}]


def test_blank_line_is_counted_as_malformed(write_log, capsys, make_log_line) -> None:
    path = write_log(make_log_line() + "\n\n")

    rc = main(["--format", "json", str(path)])

    output_json = _json_output(capsys)

    assert rc == ExitCode.SUCCESS
    assert output_json["summary"] == {
        "total_lines": 2,
        "matched_lines": 1,
        "error_lines": 1,
    }
