# Shinkansen Analytics Worker

Python-based analytics and data processing service for Shinkansen e-commerce platform.

## Development Setup

### Prerequisites

- Python 3.11 or higher
- [uv](https://github.com/astral-sh/uv) package manager

### Installing uv

```bash
curl -LsSf https://astral.sh/uv/install.sh | sh
```

Or using pip:
```bash
pip install uv
```

### Setting up the environment

```bash
# Install dependencies
uv sync

# Install dev dependencies
uv sync --extra dev
```

### Running the service

```bash
# Using uv run
uv run shinkansen-analytics start

# Or using make
make uv-run CMD="shinkansen-analytics start"
```

### Development commands

```bash
# Format code
make format-python

# Lint code
make lint-python

# Run tests
make test-python

# Add a dependency
make uv-add PACKAGE=package-name

# Add a dev dependency
make uv-add-dev PACKAGE=package-name
```

## Project Structure

```
analytics-worker/
├── analytics_worker/       # Main package
│   ├── __init__.py
│   └── cli.py            # CLI interface
├── tests/                 # Test files
├── pyproject.toml         # Project configuration
├── Dockerfile             # Docker build
└── README.md             # This file
```

## Dependencies

### Production
- `grpcio`: gRPC client for service communication
- `protobuf`: Protocol buffers support
- `redis`: Redis client for caching
- `psycopg2-binary`: PostgreSQL adapter
- `pydantic`: Data validation
- `click`: CLI framework

### Development
- `pytest`: Testing framework
- `pytest-cov`: Coverage reporting
- `ruff`: Fast Python linter and formatter
- `mypy`: Static type checker
- `black`: Code formatter

## Building Docker image

```bash
docker build -t shinkansen/analytics-worker:latest -f services/analytics-worker/Dockerfile .
```

## Running in production

```bash
docker run -p 8000:8000 shinkansen/analytics-worker:latest
```

## Configuration

Environment variables can be set via `.env` file or passed directly:

- `REDIS_URL`: Redis connection string
- `DATABASE_URL`: PostgreSQL connection string
- `LOG_LEVEL`: Logging level (DEBUG, INFO, WARNING, ERROR)

## Contributing

1. Create a new branch
2. Make your changes
3. Run `make lint-python` and `make test-python`
4. Submit a pull request

## License

MIT
