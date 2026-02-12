"""
Tests for analytics worker CLI
"""

from click.testing import CliRunner

from analytics_worker.cli import main


def test_cli_status():
    """Test CLI status command"""
    runner = CliRunner()
    result = runner.invoke(main, ["status"])
    assert result.exit_code == 0
    assert "Ready" in result.output


def test_cli_start():
    """Test CLI start command"""
    runner = CliRunner()
    result = runner.invoke(main, ["start"])
    assert result.exit_code == 0
    assert "Starting analytics worker" in result.output


def test_cli_metrics():
    """Test CLI metrics command"""
    runner = CliRunner()
    result = runner.invoke(main, ["metrics"])
    assert result.exit_code == 0
    assert "Exporting metrics" in result.output


def test_cli_version():
    """Test CLI version command"""
    runner = CliRunner()
    result = runner.invoke(main, ["--version"])
    assert result.exit_code == 0
    assert "0.1.0" in result.output
