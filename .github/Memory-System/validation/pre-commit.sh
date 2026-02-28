#!/bin/bash
# Pre-commit Hook for Picoclaw Enhanced Agent System
# Install: cp Memory-System/validation/pre-commit.sh .git/hooks/pre-commit && chmod +x .git/hooks/pre-commit

set -e

echo "🔍 Running pre-commit checks..."
echo ""

# Colors
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
NC='\033[0m'

has_errors=0

# Check 1: Go formatting
echo "1️⃣  Checking Go formatting..."
if command -v gofmt >/dev/null 2>&1; then
    unformatted=$(gofmt -l . 2>/dev/null | grep -v '^vendor/')
    if [ -n "$unformatted" ]; then
        echo -e "${RED}✗${NC} Code not formatted:"
        echo "$unformatted"
        echo ""
        echo "Run: gofmt -w ."
        has_errors=1
    else
        echo -e "${GREEN}✓${NC} Go formatting OK"
    fi
else
    echo -e "${YELLOW}⊘${NC} gofmt not found - skipping"
fi
echo ""

# Check 2: JSON validation
echo "2️⃣  Validating JSON files..."
json_errors=0
if command -v jq >/dev/null 2>&1; then
    for file in $(git diff --cached --name-only | grep '\.json$'); do
        if [ -f "$file" ]; then
            if ! jq empty "$file" 2>/dev/null; then
                echo -e "${RED}✗${NC} Invalid JSON: $file"
                json_errors=$((json_errors + 1))
                has_errors=1
            fi
        fi
    done
    
    if [ $json_errors -eq 0 ]; then
        echo -e "${GREEN}✓${NC} JSON files valid"
    fi
else
    echo -e "${YELLOW}⊘${NC} jq not found - skipping JSON validation"
fi
echo ""

# Check 3: YAML validation
echo "3️⃣  Validating YAML files..."
yaml_errors=0
if command -v yq >/dev/null 2>&1; then
    for file in $(git diff --cached --name-only | grep -E '\.(yaml|yml)$'); do
        if [ -f "$file" ]; then
            if ! yq eval '.' "$file" >/dev/null 2>&1; then
                echo -e "${RED}✗${NC} Invalid YAML: $file"
                yaml_errors=$((yaml_errors + 1))
                has_errors=1
            fi
        fi
    done
    
    if [ $yaml_errors -eq 0 ]; then
        echo -e "${GREEN}✓${NC} YAML files valid"
    fi
else
    echo -e "${YELLOW}⊘${NC} yq not found - skipping YAML validation"
fi
echo ""

# Check 4: Go vet (if Go files changed)
go_files=$(git diff --cached --name-only | grep '\.go$' | wc -l)
if [ "$go_files" -gt 0 ]; then
    echo "4️⃣  Running go vet..."
    if command -v go >/dev/null 2>&1; then
        if go vet ./... 2>&1; then
            echo -e "${GREEN}✓${NC} go vet passed"
        else
            echo -e "${RED}✗${NC} go vet found issues"
            has_errors=1
        fi
    else
        echo -e "${YELLOW}⊘${NC} go not found - skipping go vet"
    fi
    echo ""
fi

# Check 5: Run tests (if Go files changed)
if [ "$go_files" -gt 0 ]; then
    echo "5️⃣  Running tests..."
    if command -v go >/dev/null 2>&1; then
        if go test ./... 2>&1 | grep -E '(PASS|FAIL|ok|FAIL)'; then
            test_result=${PIPESTATUS[0]}
            if [ $test_result -eq 0 ]; then
                echo -e "${GREEN}✓${NC} Tests passed"
            else
                echo -e "${RED}✗${NC} Tests failed"
                has_errors=1
            fi
        else
            echo -e "${YELLOW}⊘${NC} No tests to run"
        fi
    else
        echo -e "${YELLOW}⊘${NC} go not found - skipping tests"
    fi
    echo ""
fi

# Summary
echo "━━━━━━━━━━━━━━━━━━━━━━"
if [ $has_errors -eq 0 ]; then
    echo -e "${GREEN}✅ All pre-commit checks passed!${NC}"
    exit 0
else
    echo -e "${RED}❌ Pre-commit checks failed${NC}"
    echo ""
    echo "Fix the issues above and try again."
    echo "To bypass (not recommended): git commit --no-verify"
    exit 1
fi
