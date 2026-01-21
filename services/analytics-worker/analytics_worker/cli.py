"""
CLI interface for Shinkansen Analytics Worker
"""

import click


@click.group()
@click.version_option(version="0.1.0", prog_name="shinkansen-analytics")
def main():
    """Shinkansen Analytics Worker CLI"""
    pass


@main.command()
@click.option("--config", type=click.Path(), help="Path to configuration file")
def start(config):
    """Start the analytics worker"""
    click.echo("Starting analytics worker...")
    if config:
        click.echo(f"Using config: {config}")
    else:
        click.echo("Using default configuration")


@main.command()
def status():
    """Check the status of the analytics worker"""
    click.echo("Analytics worker status: Ready")


@main.command()
@click.option("--output", type=click.Path(), default="metrics.json", help="Output file path")
def metrics(output):
    """Export analytics metrics"""
    click.echo(f"Exporting metrics to {output}")


if __name__ == "__main__":
    main()
