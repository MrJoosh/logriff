import pytest

from logriff import main
from logriff.cli import build_parser
from logriff.types import ExitCode


@pytest.mark.parametrize(
    "args",
    [
        ["-s", "bar"],
        ["--status-min", "bar"],
        ["-e", "bar"],
        ["--status-max", "bar"],
        ["-m", "bar"],
        ["--method", "bar"],
        ["--min-duration", "bar"],
        ["-n", "bar"],
        ["--top-routes", "bar"],
        ["-f", "xml"],
        ["--format", "xml"],
        ["--not-a-real-flag"],
        ["-s"],
        ["first.log", "second.log"],
    ],
)
def test_invalid_arguments_return_validation_error(args: list[str]) -> None:
    assert main(args) == ExitCode.VALIDATION_ERROR


def test_parser_defaults() -> None:
    args = build_parser().parse_args([])

    assert args.file == "-"
    assert args.status_min is None
    assert args.status_max is None
    assert args.method is None
    assert args.path_prefix is None
    assert args.min_duration is None
    assert args.top_routes == 5
    assert args.format == "text"
    assert args.strict is False


def test_help_returns_success(capsys) -> None:
    rc = main(["--help"])

    captured = capsys.readouterr()

    assert rc == ExitCode.SUCCESS
    assert "usage: logriff" in captured.out
    assert "--status-min" in captured.out


def test_short_help_returns_success(capsys) -> None:
    rc = main(["-h"])

    captured = capsys.readouterr()

    assert rc == ExitCode.SUCCESS
    assert "usage: logriff" in captured.out


def test_status_min_greater_than_max_is_validation_error(sample_log) -> None:
    rc = main(["-s", "500", "-e", "200", str(sample_log)])

    assert rc == ExitCode.VALIDATION_ERROR


def test_negative_top_routes_is_validation_error(sample_log) -> None:
    rc = main(["-n", "-1", str(sample_log)])

    assert rc == ExitCode.VALIDATION_ERROR


def test_long_form_filters_are_accepted(sample_log, capsys) -> None:
    rc = main(
        [
            "--status-min",
            "300",
            "--format",
            "json",
            str(sample_log),
        ]
    )

    captured = capsys.readouterr()

    assert rc == ExitCode.SUCCESS
    assert '"matched_lines": 3' in captured.out
