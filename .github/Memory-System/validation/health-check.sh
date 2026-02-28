#!/bin/bash
# Memory System Health Check
# Verifies memory system integrity and consistency

set -e

echo "🏥 Memory System Health Check"
echo "━━━━━━━━━━━━━━━━━━━━━━━━━━━━"
echo ""

# Colors
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
BLUE='\033[0;34m'
NC='\033[0m'

checks_passed=0
checks_failed=0
checks_warned=0

check_pass() {
    echo -e "${GREEN}✓${NC} $1"
    checks_passed=$((checks_passed + 1))
}

check_fail() {
    echo -e "${RED}✗${NC} $1"
    checks_failed=$((checks_failed + 1))
}

check_warn() {
    echo -e "${YELLOW}⚠${NC} $1"
    checks_warned=$((checks_warned + 1))
}

echo "📁 Directory Structure Check"
echo "━━━━━━━━━━━━━━━━━━━━━━━━━━"

required_dirs=(
    "Memory-System/short-term"
    "Memory-System/long-term/knowledge-base"
    "Memory-System/long-term/entity-memory"
    "Memory-System/long-term/historical"
    "Memory-System/schemas"
    "Memory-System/archive"
    "Memory-System/validation"
)

for dir in "${required_dirs[@]}"; do
    if [ -d "$dir" ]; then
        check_pass "$dir exists"
    else
        check_fail "$dir missing"
    fi
done

echo ""
echo "📄 Required Files Check"
echo "━━━━━━━━━━━━━━━━━━━━━━"

required_files=(
    "Memory-System/short-term/current-context.json"
    "Memory-System/short-term/active-tasks.yaml"
    "Memory-System/short-term/recent-decisions.json"
    "Memory-System/long-term/knowledge-base/patterns.yaml"
    "Memory-System/long-term/knowledge-base/decisions.json"
    "Memory-System/long-term/entity-memory/components.json"
    "Memory-System/schemas/current-context.schema.json"
    "Memory-System/schemas/patterns.schema.json"
)

for file in "${required_files[@]}"; do
    if [ -f "$file" ]; then
        check_pass "$file exists"
    else
        check_fail "$file missing"
    fi
done

echo ""
echo "🔍 File Integrity Check"
echo "━━━━━━━━━━━━━━━━━━━━━━"

# Check if jq is available
if command -v jq >/dev/null 2>&1; then
    # Validate JSON files syntax
    for json_file in $(find Memory-System -name "*.json" -type f 2>/dev/null); do
        if jq empty "$json_file" 2>/dev/null; then
            check_pass "$(basename "$json_file") - valid JSON"
        else
            check_fail "$(basename "$json_file") - invalid JSON"
        fi
    done
else
    check_warn "jq not installed - skipping JSON validation"
fi

# Check if yq is available
if command -v yq >/dev/null 2>&1; then
    # Validate YAML files syntax
    for yaml_file in $(find Memory-System -name "*.yaml" -o -name "*.yml" -type f 2>/dev/null); do
        if yq eval '.' "$yaml_file" >/dev/null 2>&1; then
            check_pass "$(basename "$yaml_file") - valid YAML"
        else
            check_fail "$(basename "$yaml_file") - invalid YAML"
        fi
    done
else
    check_warn "yq not installed - skipping YAML validation"
fi

echo ""
echo "📊 Size Check"
echo "━━━━━━━━━━━━━━"

# Check short-term memory size
if [ -d "Memory-System/short-term" ]; then
    short_term_size=$(du -sh Memory-System/short-term 2>/dev/null | cut -f1)
    echo -e "${BLUE}ℹ${NC} Short-term memory size: $short_term_size"
    
    # Check if size is concerning (>10MB)
    size_mb=$(du -sm Memory-System/short-term 2>/dev/null | cut -f1)
    if [ "$size_mb" -gt 10 ]; then
        check_warn "Short-term memory size ($short_term_size) exceeds 10MB - consider archival"
    else
        check_pass "Short-term memory size within limits"
    fi
fi

# Check long-term memory size
if [ -d "Memory-System/long-term" ]; then
    long_term_size=$(du -sh Memory-System/long-term 2>/dev/null | cut -f1)
    echo -e "${BLUE}ℹ${NC} Long-term memory size: $long_term_size"
fi

echo ""
echo "🔗 Cross-Reference Check"
echo "━━━━━━━━━━━━━━━━━━━━━"

# Check for broken schema references in JSON files
if command -v jq >/dev/null 2>&1; then
    for json_file in $(find Memory-System -name "*.json" -type f 2>/dev/null); do
        schema_ref=$(jq -r '."$schema" // empty' "$json_file" 2>/dev/null)
        if [ -n "$schema_ref" ]; then
            schema_file="Memory-System/schemas/$(basename "$schema_ref")"
            if [ -f "$schema_file" ]; then
                check_pass "$(basename "$json_file") schema reference valid"
            else
                check_fail "$(basename "$json_file") schema reference broken: $schema_file"
            fi
        fi
    done
fi

echo ""
echo "━━━━━━━━━━━━━━━━━━━━━━"
echo "📊 Health Summary"
echo "━━━━━━━━━━━━━━━━━━━━━━"
echo -e "Checks passed: ${GREEN}$checks_passed${NC}"
echo -e "Checks failed: ${RED}$checks_failed${NC}"
echo -e "Warnings:      ${YELLOW}$checks_warned${NC}"
echo ""

if [ $checks_failed -eq 0 ]; then
    if [ $checks_warned -eq 0 ]; then
        echo -e "${GREEN}✅ Memory system is healthy!${NC}"
    else
        echo -e "${YELLOW}⚠️  Memory system is functional but has warnings${NC}"
    fi
    exit 0
else
    echo -e "${RED}❌ Memory system has issues that need attention${NC}"
    exit 1
fi
