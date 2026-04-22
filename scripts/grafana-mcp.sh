#!/usr/bin/env bash

GRAFANA_URL="http://grafana.notopia.localhost"

# 1. Create the Service Account and get its ID
SA_ID=$(curl -u admin:admin -X POST "$GRAFANA_URL/api/serviceaccounts" \
  -H "Content-Type: application/json" \
  -d '{"name":"mcp-token", "role":"Admin"}' | jq -r '.id')

# 2. Generate the Token for that ID
TOKEN=$(curl -u admin:admin -X POST "$GRAFANA_URL/api/serviceaccounts/$SA_ID/tokens" \
  -H "Content-Type: application/json" \
  -d '{"name":"mcp-key"}' | jq -r '.key')

echo "Your Token: $TOKEN"
