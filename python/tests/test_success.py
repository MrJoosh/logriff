from logriff import main
from logriff.types import ExitCode


def test_success_exit_code(sample_log, capsys) -> None:
    rc = main([str(sample_log)])

    captured = capsys.readouterr()

    assert rc == ExitCode.SUCCESS
    assert "=== LogRiff Summary ===" in captured.out
    assert captured.err == ""
