"""
Test configuration for analytics worker
"""

import pytest


@pytest.fixture
def sample_config():
    """Sample configuration fixture"""
    return {
        "redis_url": "redis://localhost:6379",
        "database_url": "postgresql://localhost:5432/shinkansen",
        "log_level": "INFO",
    }
