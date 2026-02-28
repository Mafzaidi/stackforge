#!/bin/bash

# Script to verify Clean Architecture dependency rules
# Requirements: 5.1, 5.2, 5.3, 5.4, 5.5

set -e

echo "=== Clean Architecture Dependency Rule Verification ==="
echo ""

# Colors for output
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
NC='\033[0m' # No Color

VIOLATIONS=0

# Function to check imports in a file
check_file_imports() {
    local file=$1
    local layer=$2
    local allowed_patterns=$3
    
    # Get all import statements
    imports=$(grep -E '^import \(|^\s+"' "$file" | grep -v '//' || true)
    
    # Check each import against disallowed patterns
    while IFS= read -r import_line; do
        # Skip empty lines
        [ -z "$import_line" ] && continue
        
        # Extract the import path
        import_path=$(echo "$import_line" | sed -n 's/.*"\([^"]*\)".*/\1/p')
        
        # Skip standard library and external packages
        if [[ ! "$import_path" =~ ^github.com/mafzaidi/stackforge/internal/ ]]; then
            continue
        fi
        
        # Check if import violates layer rules
        case "$layer" in
            "domain")
                # Domain should have NO dependencies on other internal layers
                if [[ "$import_path" =~ (usecase|delivery|infrastructure|pkg) ]]; then
                    echo -e "${RED}VIOLATION${NC}: $file"
                    echo "  Domain layer imports: $import_path"
                    VIOLATIONS=$((VIOLATIONS + 1))
                fi
                ;;
            "usecase")
                # Use case should only depend on domain
                if [[ "$import_path" =~ (delivery|infrastructure) ]]; then
                    echo -e "${RED}VIOLATION${NC}: $file"
                    echo "  Use case layer imports: $import_path"
                    VIOLATIONS=$((VIOLATIONS + 1))
                fi
                ;;
            "delivery")
                # Delivery should only depend on usecase and domain
                if [[ "$import_path" =~ infrastructure ]]; then
                    echo -e "${RED}VIOLATION${NC}: $file"
                    echo "  Delivery layer imports: $import_path"
                    VIOLATIONS=$((VIOLATIONS + 1))
                fi
                ;;
            "infrastructure")
                # Infrastructure should only depend on domain interfaces
                if [[ "$import_path" =~ (usecase|delivery) ]]; then
                    echo -e "${RED}VIOLATION${NC}: $file"
                    echo "  Infrastructure layer imports: $import_path"
                    VIOLATIONS=$((VIOLATIONS + 1))
                fi
                ;;
        esac
    done <<< "$imports"
}

echo "1. Checking Domain Layer (internal/domain/)"
echo "   Rule: Domain layer must have NO dependencies on other layers"
echo ""

for file in $(find internal/domain -name "*.go" -not -name "*_test.go"); do
    check_file_imports "$file" "domain"
done

echo -e "${GREEN}✓${NC} Domain layer check complete"
echo ""

echo "2. Checking Use Case Layer (internal/usecase/)"
echo "   Rule: Use case layer must depend ONLY on domain"
echo ""

for file in $(find internal/usecase -name "*.go" -not -name "*_test.go"); do
    check_file_imports "$file" "usecase"
done

echo -e "${GREEN}✓${NC} Use case layer check complete"
echo ""

echo "3. Checking Delivery Layer (internal/delivery/)"
echo "   Rule: Delivery layer must depend ONLY on usecase and domain"
echo ""

for file in $(find internal/delivery -name "*.go" -not -name "*_test.go" 2>/dev/null || true); do
    check_file_imports "$file" "delivery"
done

echo -e "${GREEN}✓${NC} Delivery layer check complete"
echo ""

echo "4. Checking Infrastructure Layer (internal/infrastructure/)"
echo "   Rule: Infrastructure must depend ONLY on domain interfaces"
echo ""

for file in $(find internal/infrastructure -name "*.go" -not -name "*_test.go"); do
    check_file_imports "$file" "infrastructure"
done

echo -e "${GREEN}✓${NC} Infrastructure layer check complete"
echo ""

echo "5. Checking for Circular Dependencies"
echo ""

# Use go mod graph to check for circular dependencies
if go list -f '{{.ImportPath}}' ./internal/... > /dev/null 2>&1; then
    echo -e "${GREEN}✓${NC} No circular dependencies detected"
else
    echo -e "${RED}✗${NC} Circular dependency detected"
    VIOLATIONS=$((VIOLATIONS + 1))
fi

echo ""
echo "=== Summary ==="
if [ $VIOLATIONS -eq 0 ]; then
    echo -e "${GREEN}✓ All dependency rules satisfied!${NC}"
    echo "  - Domain layer has no external dependencies"
    echo "  - Use case layer depends only on domain"
    echo "  - Delivery layer depends on use case and domain"
    echo "  - Infrastructure implements domain interfaces"
    echo "  - No circular dependencies detected"
    exit 0
else
    echo -e "${RED}✗ Found $VIOLATIONS dependency rule violation(s)${NC}"
    exit 1
fi
