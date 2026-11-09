from logpulse import main
from logpulse.types import ExitCode


def test_success_exit_code(sample_log, capsys) -> None:
    rc = main([str(sample_log)])

    captured = capsys.readouterr()

    assert rc == ExitCode.SUCCESS
    assert "=== LogPulse Summary ===" in captured.out
    assert captured.err == ""
