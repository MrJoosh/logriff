import builtins
import io
import sys
from unittest.mock import MagicMock

from logriff import main
from logriff.types import ExitCode


def test_file_not_found(tmp_path, capsys) -> None:
    missing = tmp_path / "missing.log"

    rc = main([str(missing)])

    captured = capsys.readouterr()

    assert rc == ExitCode.IO_ERROR
    assert "not found" in captured.err.lower()
    assert captured.out == ""


def test_permission_denied(monkeypatch, tmp_path, capsys) -> None:
    path = tmp_path / "secret.log"
    mock_open = MagicMock(
        side_effect=PermissionError(13, "Permission denied", str(path))
    )
    monkeypatch.setattr(builtins, "open", mock_open)

    rc = main([str(path)])

    captured = capsys.readouterr()

    assert rc == ExitCode.IO_ERROR
    assert "Permission denied" in captured.err


def test_permission_denied_mid_stream(
    monkeypatch, capsys, make_log_line
) -> None:
    class PermissionDeniedAfterOneLine(io.StringIO):
        def __iter__(self):
            yield make_log_line() + "\n"
            raise PermissionError(13, "Permission denied", "stdin")

    monkeypatch.setattr(sys, "stdin", PermissionDeniedAfterOneLine())

    rc = main([])

    captured = capsys.readouterr()

    assert rc == ExitCode.IO_ERROR
    assert captured.out == ""


def test_pipe_closed_before_read(monkeypatch, capsys) -> None:
    closed_stream = io.StringIO()
    closed_stream.close()
    monkeypatch.setattr(sys, "stdin", closed_stream)

    rc = main([])

    captured = capsys.readouterr()

    assert rc == ExitCode.IO_ERROR
    assert "I/O operation on closed file" in captured.err


def test_pipe_immediate_eof(monkeypatch, capsys) -> None:
    monkeypatch.setattr(sys, "stdin", io.StringIO(""))

    rc = main([])

    captured = capsys.readouterr()

    assert rc == ExitCode.IO_ERROR
    assert "No data to process" in captured.err


def test_empty_file_returns_io_error(tmp_path, capsys) -> None:
    path = tmp_path / "empty.log"
    path.write_text("", encoding="utf-8")

    rc = main([str(path)])

    captured = capsys.readouterr()

    assert rc == ExitCode.IO_ERROR
    assert "No data to process" in captured.err


def test_stdin_success(monkeypatch, capsys, make_log_line) -> None:
    monkeypatch.setattr(sys, "stdin", io.StringIO(make_log_line() + "\n"))

    rc = main([])

    captured = capsys.readouterr()

    assert rc == ExitCode.SUCCESS
    assert "Lines Processed : 1" in captured.out
    assert captured.err == ""


def test_explicit_stdin_marker(monkeypatch, capsys, make_log_line) -> None:
    monkeypatch.setattr(sys, "stdin", io.StringIO(make_log_line() + "\n"))

    rc = main(["-"])

    captured = capsys.readouterr()

    assert rc == ExitCode.SUCCESS
    assert "Lines Processed : 1" in captured.out


def test_input_without_trailing_newline(tmp_path, capsys, make_log_line) -> None:
    path = tmp_path / "no-final-newline.log"
    path.write_text(make_log_line(), encoding="utf-8")

    rc = main([str(path)])

    captured = capsys.readouterr()

    assert rc == ExitCode.SUCCESS
    assert "Lines Processed : 1" in captured.out


def test_directory_input_returns_io_error(tmp_path, capsys) -> None:
    directory = tmp_path / "logs"
    directory.mkdir()

    rc = main([str(directory)])

    captured = capsys.readouterr()

    assert rc == ExitCode.IO_ERROR
    assert captured.out == ""


def test_invalid_utf8_returns_io_error(tmp_path, capsys) -> None:
    path = tmp_path / "invalid-utf8.log"
    path.write_bytes(b"\xff\xfe\x00\n")

    rc = main([str(path)])

    captured = capsys.readouterr()

    assert rc == ExitCode.IO_ERROR
    assert captured.out == ""


def test_unexpected_value_error_returns_io_error(
    monkeypatch, tmp_path, capsys
) -> None:
    path = tmp_path / "input.log"
    mock_open = MagicMock(side_effect=ValueError("Test error."))
    monkeypatch.setattr(builtins, "open", mock_open)

    rc = main([str(path)])

    captured = capsys.readouterr()

    assert rc == ExitCode.IO_ERROR
    assert captured.out == ""


class BrokenStdout:
    def write(self, _text: str) -> int:
        raise BrokenPipeError(32, "Broken pipe")

    def flush(self) -> None:
        pass


def test_broken_stdout_returns_io_error(monkeypatch, sample_log) -> None:
    monkeypatch.setattr(sys, "stdout", BrokenStdout())

    rc = main([str(sample_log)])

    assert rc == ExitCode.IO_ERROR
