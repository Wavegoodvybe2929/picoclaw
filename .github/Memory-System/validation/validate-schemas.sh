#!/bin/bash
# Schema Validation Script
# Validates all JSON files in Memory-System against their declared schemas

set -e

echo "🔍 Validating JSON/YAML files against schemas..."
echo ""

# Check if required tools are installed
command -v jq >/dev/null 2>&1 || { echo "❌ jq is required but not installed. Install with: brew install jq"; exit 1; }
command -v yq >/dev/null 2>&1 || { echo "❌ yq is required but not installed. Install with: brew install yq"; exit 1; }

# Colors for output
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
NC='\033[0m' # No Color

total_files=0
valid_files=0
invalid_files=0
skipped_files=0

echo "📋 JSON File Validation"
echo "━━━━━━━━━━━━━━━━━━━━━━"

# Find all JSON files in Memory-System
while IFS= read -r json_file; do
    total_files=$((total_files + 1))
    
    # Check if file is valid JSON
    if ! jq empty "$json_file" 2>/dev/null; then
        echo -e "${RED}✗${NC} $json_file - Invalid JSON syntax"
        invalid_files=$((invalid_files + 1))
        continue
    fi
    
    # Get schema reference
    schema_ref=$(jq -r '."$schema" // empty' "$json_file" 2>/dev/null)
    
    if [ -z "$schema_ref" ]; then
        echo -e "${YELLOW}⊘${NC} $json_file - No schema reference"
        skipped_files=$((skipped_files + 1))
        continue
    fi
    
    # Construct schema file path
    schema_file="Memory-System/schemas/$(basename "$schema_ref")"
    
    if [ ! -f "$schema_file" ]; then
        echo -e "${YELLOW}⊘${NC} $json_file - Schema not found: $schema_file"
        skipped_files=$((skipped_files + 1))
        continue
    fi
    
    # Check if ajv is available for validation
    if command -v ajv >/dev/null 2>&1; then
        if ajv validate -s "$schema_file" -d "$json_file" 2>/dev/null; then
            echo -e "${GREEN}✓${NC} $json_file"
            valid_files=$((valid_files + 1))
        else
            echo -e "${RED}✗${NC} $json_file - Schema validation failed"
            invalid_files=$((invalid_files + 1))
        fi
    else
        # Without ajv, just check JSON syntax
        echo -e "${GREEN}✓${NC} $json_file (syntax only, install ajv-cli for full validation)"
        valid_files=$((valid_files + 1))
    fi
    
done < <(find Memory-System -name "*.json" -type f 2>/dev/null)

echo ""
echo "📋 YAML File Validation"
echo "━━━━━━━━━━━━━━━━━━━━━━"

# Find all YAML files in Memory-System
while IFS= read -r yaml_file; do
    total_files=$((total_files + 1))
    
    # Check if file is valid YAML
    if yq eval '.' "$yaml_file" >/dev/null 2>&1; then
        echo -e "${GREEN}✓${NC} $yaml_file"
        valid_files=$((valid_files + 1))
    else
        echo -e "${RED}✗${NC} $yaml_file - Invalid YAML syntax"
        invalid_files=$((invalid_files + 1))
    fi
    
done < <(find Memory-System -name "*.yaml" -o -name "*.yml" -type f 2>/dev/null)

echo ""
echo "━━━━━━━━━━━━━━━━━━━━━━"
echo "📊 Validation Summary"
echo "━━━━━━━━━━━━━━━━━━━━━━"
echo "Total files:   $total_files"
echo -e "Valid files:   ${GREEN}$valid_files${NC}"
echo -e "Invalid files: ${RED}$invalid_files${NC}"
echo -e "Skipped files: ${YELLOW}$skipped_files${NC}"
echo ""

if [ $invalid_files -eq 0 ]; then
    echo -e "${GREEN}✅ All files validated successfully!${NC}"
    exit 0
else
    echo -e "${RED}❌ Validation failed with $invalid_files errors${NC}"
    echo ""
    echo "💡 Tips:"
    echo "  - Check JSON/YAML syntax"
    echo "  - Ensure schema references are correct"
    echo "  - Install ajv-cli for full JSON schema validation: npm install -g ajv-cli"
    exit 1
fi
