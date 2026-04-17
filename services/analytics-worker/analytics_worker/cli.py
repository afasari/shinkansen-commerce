"""
CLI interface for Shinkansen Analytics Worker
"""

import os

import click
import structlog


def init_telemetry():
    service_name = os.environ.get("OTEL_SERVICE_NAME", "analytics-worker")
    endpoint = os.environ.get("OTEL_EXPORTER_OTLP_ENDPOINT", "http://localhost:4317")

    try:
        from opentelemetry import trace
        from opentelemetry.exporter.otlp.proto.grpc.trace_exporter import OTLPSpanExporter
        from opentelemetry.sdk.resources import Resource
        from opentelemetry.sdk.trace import TracerProvider
        from opentelemetry.sdk.trace.export import BatchSpanProcessor

        resource = Resource.create({"service.name": service_name})
        provider = TracerProvider(resource=resource)
        exporter = OTLPSpanExporter(endpoint=endpoint, insecure=True)
        provider.add_span_processor(BatchSpanProcessor(exporter))
        trace.set_tracer_provider(provider)
    except Exception:
        pass


def init_logging():
    structlog.configure(
        processors=[
            structlog.contextvars.merge_contextvars,
            structlog.processors.add_log_level,
            structlog.processors.StackInfoRenderer(),
            structlog.dev.set_exc_info,
            structlog.processors.TimeStamper(fmt="iso"),
            structlog.processors.JSONRenderer(),
        ],
        wrapper_class=structlog.make_filtering_bound_logger(20),
        context_class=dict,
        logger_factory=structlog.PrintLoggerFactory(),
        cache_logger_on_first_use=True,
    )


@click.group()
@click.version_option(version="0.1.0", prog_name="shinkansen-analytics")
def main():
    """Shinkansen Analytics Worker CLI"""
    init_logging()
    init_telemetry()


@main.command()
@click.option("--config", type=click.Path(), help="Path to configuration file")
def start(config):
    """Start the analytics worker"""
    logger = structlog.get_logger()
    logger.info("starting_analytics_worker", config=config or "default")


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
