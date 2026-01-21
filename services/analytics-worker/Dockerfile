FROM python:3.11-slim

# Install uv
COPY --from=ghcr.io/astral-sh/uv:latest /uv /usr/local/bin/uv

WORKDIR /app

# Set environment variables
ENV PYTHONUNBUFFERED=1
ENV PYTHONDONTWRITEBYTECODE=1

# Copy project files
COPY services/analytics-worker/pyproject.toml .
COPY services/analytics-worker/analytics_worker ./analytics_worker

# Create virtual environment and install dependencies
RUN uv sync

# Expose port (if needed)
EXPOSE 8000

# Set the entrypoint
ENTRYPOINT ["uv", "run", "shinkansen-analytics"]
CMD ["start"]
