#!/bin/bash

# Script to validate that swagger docs are in sync with main.go annotations
# This ensures we have a single source of truth for swagger metadata

set -e

# Extract contact info from main.go annotations
MAIN_FILE="cmd/meos-graphics/main.go"
DOCS_FILE="docs/docs.go"

# Extract values from main.go
CONTACT_NAME=$(grep -oP '@contact\.name\s+\K.*' "$MAIN_FILE" | xargs)
CONTACT_EMAIL=$(grep -oP '@contact\.email\s+\K.*' "$MAIN_FILE" | xargs)
TITLE=$(grep -oP '@title\s+\K.*' "$MAIN_FILE" | xargs)

# Check if docs.go contains the correct values
echo "Validating swagger documentation..."
echo "Expected values from main.go:"
echo "  Title: $TITLE"
echo "  Contact Name: $CONTACT_NAME"
echo "  Contact Email: $CONTACT_EMAIL"

# Check title in docs.go
if ! grep -q "\"name\": \"$CONTACT_NAME\"" "$DOCS_FILE"; then
    echo "ERROR: Contact name mismatch in $DOCS_FILE"
    echo "Expected: $CONTACT_NAME"
    echo "Run 'swag init -g cmd/meos-graphics/main.go --parseDependency --parseInternal' to regenerate"
    exit 1
fi

if ! grep -q "\"email\": \"$CONTACT_EMAIL\"" "$DOCS_FILE"; then
    echo "ERROR: Contact email mismatch in $DOCS_FILE"
    echo "Expected: $CONTACT_EMAIL"
    echo "Run 'swag init -g cmd/meos-graphics/main.go --parseDependency --parseInternal' to regenerate"
    exit 1
fi

if ! grep -q "Title:.*\"$TITLE\"" "$DOCS_FILE"; then
    echo "ERROR: Title mismatch in $DOCS_FILE"
    echo "Expected: $TITLE"
    echo "Run 'swag init -g cmd/meos-graphics/main.go --parseDependency --parseInternal' to regenerate"
    exit 1
fi

echo "✓ Swagger documentation is in sync with main.go annotations"