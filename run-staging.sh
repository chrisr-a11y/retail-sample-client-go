#!/bin/bash
# Wrapper script to run the client with staging environment
# Loads .env.staging and executes the Go client
#
# Usage: ./run-staging.sh
#
# The client checks POLYMARKET_* vars first, then falls back to TEST_*/RETAIL_* vars.
# This script exports the staging env vars, allowing them to be overridden by
# setting POLYMARKET_* vars before running.

set -e

SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
ENV_FILE="${SCRIPT_DIR}/.env.staging"

if [[ ! -f "$ENV_FILE" ]]; then
    echo "Error: $ENV_FILE not found"
    echo "Copy .env.staging from the integration harness first"
    exit 1
fi

# Export variables from .env.staging, but only if not already set
# This allows POLYMARKET_* vars to override TEST_*/RETAIL_* vars
while IFS= read -r line || [[ -n "$line" ]]; do
    # Skip comments and empty lines
    [[ "$line" =~ ^[[:space:]]*# ]] && continue
    [[ -z "$line" ]] && continue

    # Extract variable name and value
    if [[ "$line" =~ ^([A-Za-z_][A-Za-z0-9_]*)=(.*)$ ]]; then
        var_name="${BASH_REMATCH[1]}"
        var_value="${BASH_REMATCH[2]}"

        # Only export if not already set
        if [[ -z "${!var_name}" ]]; then
            export "$var_name=$var_value"
        fi
    fi
done < "$ENV_FILE"

echo "=== Running with staging environment ==="
echo "API URL: ${RETAIL_API_URL:-not set}"
echo "WS URL: ${RETAIL_WS_URL:-not set}"
echo "Market: ${TEST_MARKET_SLUG:-not set}"
echo ""

# Run the Go client
exec go run "${SCRIPT_DIR}/main.go" "$@"
