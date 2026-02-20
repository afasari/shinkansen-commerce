#!/usr/bin/env bash
# Generate VitePress API documentation from proto files
# This script parses proto definitions and creates markdown files in docs/api/

set -e

PROTO_DIR="proto"
DOCS_DIR="docs/api"

mkdir -p "$DOCS_DIR"

echo "📝 Generating API documentation from proto files..."

# Function to extract service info from proto
extract_service_info() {
    local proto_file="$1"
    local service_name="$2"
    local language="${3:-Go}"
    local proto_file_path="$PROTO_DIR/${proto_file}"

    if [ ! -f "$proto_file_path" ]; then
        echo "⚠️  Proto file not found: $proto_file_path"
        return
    fi

    local doc_file="$DOCS_DIR/${service_name}.md"

    echo "# ${service_name^} API" > "$doc_file"
    echo "" >> "$doc_file"
    echo "Service: shinkansen.${service_name}" >> "$doc_file"
    echo "" >> "$doc_file"
    echo "## Overview" >> "$doc_file"
    echo "" >> "$doc_file"
    echo "The ${service_name} service provides APIs for managing ${service_name}-related operations." >> "$doc_file"
    echo "" >> "$doc_file"

    # Extract RPC methods using simplified approach
    echo "## RPC Methods" >> "$doc_file"
    echo "" >> "$doc_file"
    
    # Read proto file and extract RPCs with simpler parsing
    grep -E '^\s*rpc\s+[A-Z]' "$proto_file_path" | while read -r line; do
        # Extract method (first word after "rpc")
        method=$(echo "$line" | sed -E 's/^\s*rpc\s+([A-Z][a-zA-Z]+).*/\1/')
        
        # Extract request type (text in first parentheses)
        request=$(echo "$line" | sed -E 's/.*\(([A-Z][a-zA-Z]+Request).*/\1/')
        
        # Extract response type - handle both formats:
        # - returns (GetStockResponse)
        # - returns (shinkansen.common.Empty)
        
        # Extract everything between parentheses after "returns"
        response=$(echo "$line" | sed -E 's/.*returns\s+\(([^)]+)\).*/\1/')
        
        # If response contains a dot (nested type), use it as-is
        # If response is a simple word and not "Empty", it's likely Empty type
        if [ -n "$response" ]; then
            if [[ ! "$response" =~ \. ]]; then
                if [ "$response" = "Empty" ]; then
                    response="shinkansen.common.Empty"
                fi
            fi
        fi
        
        # Apply Empty fallback if response doesn't end with Response
        if [ -n "$response" ] && [[ ! "$response" =~ Response$ ]]; then
            if [[ "$response" =~ ^[A-Z] ]]; then
                # If it's just a word without dots, assume it's shinkansen.common.Empty
                case "$response" in
                    Empty) response="shinkansen.common.Empty" ;;
                esac
            fi
        fi
        
        if [ -n "$method" ] && [ -n "$request" ]; then
            echo "### $method" >> "$doc_file"
            echo "" >> "$doc_file"
            echo "**Request:** \`$request\`" >> "$doc_file"
            echo "" >> "$doc_file"
            echo "**Response:** \`$response\`" >> "$doc_file"
            echo "" >> "$doc_file"
        fi
    done
    echo "" >> "$doc_file"

    # Extract HTTP routes
    echo "## HTTP Endpoints" >> "$doc_file"
    echo "" >> "$doc_file"
    echo "| Method | Path |" >> "$doc_file"
    echo "|--------|------|" >> "$doc_file"

    # Extract HTTP method and path using grep
    grep -E "option.*http.*=" "$proto_file_path" | while IFS= read -r line; do
        # Extract HTTP method
        http_method=$(echo "$line" | grep -oE '\b(get|put|post|delete):' | tr '[:lower:]' '[:upper:]' | tr -d ':')
        
        # Extract path from quotes
        http_path=$(echo "$line" | grep -oE '"[^"]+"' | tr -d '"')
        
        if [ -n "$http_method" ] && [ -n "$http_path" ]; then
            echo "| $http_method | \`$http_path\` |" >> "$doc_file"
        fi
    done
    echo "" >> "$doc_file"

    # List message types from messages file
    echo "## Message Types" >> "$doc_file"
    echo "" >> "$doc_file"
    
    local messages_file=$(echo "$proto_file" | sed 's/_service.proto/_messages.proto/')
    if [ -f "$PROTO_DIR/$messages_file" ]; then
        echo "Message types are defined in \`$messages_file\`" >> "$doc_file"
        echo "" >> "$doc_file"
        
        grep "^message " "$PROTO_DIR/$messages_file" | awk '{print $2}' | while read -r msg; do
            if [[ ! "$msg" =~ (Request|Response)$ ]]; then
                echo "### $msg" >> "$doc_file"
                echo "" >> "$doc_file"
                echo "Data structure for ${service_name} operations." >> "$doc_file"
                echo "" >> "$doc_file"
            fi
        done
    fi

    echo "## Implementation" >> "$doc_file"
    echo "" >> "$doc_file"
    echo "**Language:** $language" >> "$doc_file"
    echo "**Location:** \`services/${service_name}-service/\`" >> "$doc_file"
    echo "" >> "$doc_file"

    echo "## Testing" >> "$doc_file"
    echo "" >> "$doc_file"
    echo '```bash' >> "$doc_file"
    echo "# Example gRPC call using grpcurl" >> "$doc_file"
    
    # Get first RPC method for example
    first_method=$(grep -E '^\s*rpc\s+[A-Z]' "$proto_file_path" | head -1 | sed -E 's/^\s*rpc\s+([A-Z][a-zA-Z]+).*/\1/')
    if [ -n "$first_method" ]; then
        echo "grpcurl -plaintext localhost:<port> shinkansen.${service_name}.${service_name^}Service/${first_method}" >> "$doc_file"
    else
        echo "grpcurl -plaintext localhost:<port> shinkansen.${service_name}.<Service>/<RPCMethod>" >> "$doc_file"
    fi
    echo '```' >> "$doc_file"
    echo "" >> "$doc_file"

    echo "✅ Generated $doc_file"
}

# Generate documentation for each service
extract_service_info "product/product_service.proto" "product" "Go"
extract_service_info "order/order_service.proto" "order" "Go"
extract_service_info "user/user_service.proto" "user" "Go"
extract_service_info "payment/payment_service.proto" "payment" "Go"
extract_service_info "inventory/inventory_service.proto" "inventory" "Rust"
extract_service_info "delivery/delivery_service.proto" "delivery" "Go"

# Generate index with all services
echo "# API Reference" > "$DOCS_DIR/index.md"
echo "" >> "$DOCS_DIR/index.md"
echo "Complete API reference for all Shinkansen Commerce services." >> "$DOCS_DIR/index.md"
echo "" >> "$DOCS_DIR/index.md"
echo "## Services" >> "$DOCS_DIR/index.md"
echo "" >> "$DOCS_DIR/index.md"
echo "| Service | RPC Methods | Language |" >> "$DOCS_DIR/index.md"
echo "|---------|-------------|----------|" >> "$DOCS_DIR/index.md"

for service in product order user payment inventory delivery; do
    service_file="proto/${service}/${service}_service.proto"
    if [ -f "$service_file" ]; then
        rpc_count=$(grep -cE '^\s*rpc\s+[A-Z]' "$service_file" 2>/dev/null || echo "0")
        lang="Go"
        if [ "$service" = "inventory" ]; then
            lang="Rust"
        fi
        echo "| [$service](${service}.md) | $rpc_count | $lang |" >> "$DOCS_DIR/index.md"
    fi
done

# Add getting started section
echo "" >> "$DOCS_DIR/index.md"
echo "## Getting Started" >> "$DOCS_DIR/index.md"
echo "" >> "$DOCS_DIR/index.md"
echo "1. [Quick Start](../quickstart) - Get started in 5 minutes" >> "$DOCS_DIR/index.md"
echo "2. [API Overview](./overview) - Learn about authentication, errors, and pagination" >> "$DOCS_DIR/index.md"
echo "3. [Products API](./products) - Product catalog management" >> "$DOCS_DIR/index.md"
echo "4. [Orders API](./orders) - Order processing" >> "$DOCS_DIR/index.md"
echo "5. [Users API](./users) - User management" >> "$DOCS_DIR/index.md"
echo "6. [Payments API](./payments) - Payment processing" >> "$DOCS_DIR/index.md"
echo "7. [Inventory API](./inventory) - Stock management" >> "$DOCS_DIR/index.md"
echo "8. [Delivery API](./delivery) - Delivery logistics" >> "$DOCS_DIR/index.md"
echo "9. [Authentication API](./authentication) - Auth and tokens" >> "$DOCS_DIR/index.md"

echo ""
echo "✅ API documentation generation completed!"
echo "📂 Generated files in $DOCS_DIR/"
echo "💡 Tip: Run 'make proto-watch' to automatically regenerate docs when proto files change"
