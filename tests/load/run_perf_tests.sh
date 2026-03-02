#!/bin/bash
# Performance test script for Shinkansen Commerce

set -e

# Colors for output
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[0;33m'
NC='\033[0m' # No Color

# Configuration
API_URL="${API_URL:-http://localhost:8080}"
K6_VUS="${K6_VUS:-100}"
K6_DURATION="${K6_DURATION:-10m}"
OUTPUT_DIR="./test-results"

echo -e "${GREEN}=== Shinkansen Commerce Performance Testing ===${NC}"
echo "API URL: $API_URL"
echo "Virtual Users: $K6_VUS"
echo "Duration: $K6_DURATION"
echo ""

# Create output directory
mkdir -p "$OUTPUT_DIR"

# Check dependencies
check_dependencies() {
    echo -e "${YELLOW}Checking dependencies...${NC}"

    if ! command -v k6 &> /dev/null; then
        echo -e "${RED}k6 not found. Install from: https://k6.io${NC}"
        exit 1
    fi

    if ! command -v locust &> /dev/null; then
        echo -e "${YELLOW}Locust not found. Install with: pip install locust${NC}"
    fi

    echo -e "${GREEN}Dependencies OK${NC}"
}

# Wait for service to be ready
wait_for_service() {
    echo -e "${YELLOW}Waiting for service to be ready...${NC}"

    local max_attempts=30
    local attempt=1

    while [ $attempt -le $max_attempts ]; do
        if curl -s "$API_URL/health" > /dev/null; then
            echo -e "${GREEN}Service is ready!${NC}"
            return 0
        fi

        echo "Waiting... ($attempt/$max_attempts)"
        sleep 2
        ((attempt++))
    done

    echo -e "${RED}Service not ready after $max_attempts attempts${NC}"
    return 1
}

# Run K6 load test
run_k6_test() {
    echo -e "${YELLOW}Running K6 load test...${NC}"

    k6 run \
        --out json="$OUTPUT_DIR/k6-results.json" \
        --out influxdb=http://localhost:8086/k6 \
        -e API_URL="$API_URL" \
        -e K6_VUS="$K6_VUS" \
        -e K6_DURATION="$K6_DURATION" \
        tests/load/api_load_test.js

    echo -e "${GREEN}K6 test completed${NC}"
}

# Run Locust load test
run_locust_test() {
    echo -e "${YELLOW}Running Locust load test (headless)...${NC}"

    locust \
        -f tests/load/locust_load_test.py \
        --headless \
        --host "$API_URL" \
        --users 100 \
        --spawn-rate 10 \
        --run-time "$K6_DURATION" \
        --html "$OUTPUT_DIR/locust-report.html" \
        --csv "$OUTPUT_DIR/locust-stats"

    echo -e "${GREEN}Locust test completed${NC}"
}

# Run database query performance test
run_db_query_test() {
    echo -e "${YELLOW}Running database query performance test...${NC}"

    # This would run database-specific performance tests
    # For now, just a placeholder
    echo "Database query testing skipped (requires DB connection)"
}

# Run cache performance test
run_cache_test() {
    echo -e "${YELLOW}Running cache performance test...${NC}"

    # Test Redis cache performance
    redis-cli -h localhost -p 6379 ping > /dev/null 2>&1
    if [ $? -eq 0 ]; then
        echo -e "${GREEN}Redis is responding${NC}"

        # Benchmark cache operations
        echo "Benchmarking cache SET/GET operations..."
        redis-benchmark -h localhost -p 6379 -t set,get -n 100000 -c 20
    else
        echo -e "${YELLOW}Redis not available for cache testing${NC}"
    fi
}

# Generate performance report
generate_report() {
    echo -e "${YELLOW}Generating performance report...${NC}"

    cat > "$OUTPUT_DIR/summary.md" << EOF
# Performance Test Summary

**Date:** $(date)
**API URL:** $API_URL
**Virtual Users:** $K6_VUS
**Duration:** $K6_DURATION

## Test Results

### K6 Results
See \`k6-results.json\` for detailed metrics.

### Locust Results
See \`locust-report.html\` for detailed metrics.

## Recommendations

- Review response times (p95, p99)
- Check error rates
- Monitor resource utilization
- Identify slow endpoints

## Next Steps

1. Analyze slow endpoints
2. Optimize database queries
3. Add caching where needed
4. Consider horizontal scaling
EOF

    echo -e "${GREEN}Report generated: $OUTPUT_DIR/summary.md${NC}"
}

# Cleanup function
cleanup() {
    echo -e "${YELLOW}Cleaning up...${NC}"
    # Cleanup any test data if needed
}

# Main execution
main() {
    check_dependencies
    wait_for_service

    trap cleanup EXIT

    # Run all tests
    run_k6_test
    run_locust_test
    run_db_query_test
    run_cache_test

    generate_report

    echo -e "${GREEN}=== All tests completed ===${NC}"
}

# Parse command line arguments
case "${1:-}" in
    k6)
        check_dependencies
        wait_for_service
        run_k6_test
        ;;
    locust)
        check_dependencies
        wait_for_service
        run_locust_test
        ;;
    cache)
        run_cache_test
        ;;
    db)
        run_db_query_test
        ;;
    report)
        generate_report
        ;;
    all|*)
        main
        ;;
esac
