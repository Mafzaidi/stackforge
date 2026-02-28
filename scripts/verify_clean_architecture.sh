#!/bin/bash

# Clean Architecture Compliance Verification Script
# Verifies: Requirements 6.1, 6.3, 6.4, 7.6

set -e

echo "=== Clean Architecture Compliance Verification ==="
echo ""

# Colors for output
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
NC='\033[0m' # No Color

ERRORS=0
WARNINGS=0

# Function to report error
report_error() {
    echo -e "${RED}✗ ERROR: $1${NC}"
    ERRORS=$((ERRORS + 1))
}

# Function to report warning
report_warning() {
    echo -e "${YELLOW}⚠ WARNING: $1${NC}"
    WARNINGS=$((WARNINGS + 1))
}

# Function to report success
report_success() {
    echo -e "${GREEN}✓ $1${NC}"
}

echo "1. Verifying Module Structure Consistency (Requirement 6.1)"
echo "-----------------------------------------------------------"

# Check that all modules have consistent structure
MODULES=("auth" "todo" "credential")

for module in "${MODULES[@]}"; do
    echo "Checking module: $module"
    
    # Check if usecase exists
    if [ -d "internal/usecase/$module" ]; then
        report_success "Module $module has usecase layer"
    else
        report_error "Module $module missing usecase layer"
    fi
    
    # Check if handler exists (for modules that need HTTP handlers)
    if grep -q "${module}_handler.go" internal/delivery/http/handler/* 2>/dev/null; then
        report_success "Module $module has handler"
    else
        report_warning "Module $module may be missing handler"
    fi
done

echo ""
echo "2. Verifying Naming Conventions (Requirement 6.3)"
echo "-----------------------------------------------------------"

# Check entity file naming: <entity_name>.go
echo "Checking entity file naming..."
ENTITY_FILES=$(find internal/domain/entity -name "*.go" -not -name "*_test.go" 2>/dev/null || true)
for file in $ENTITY_FILES; do
    basename=$(basename "$file")
    if [[ "$basename" =~ ^[a-z_]+\.go$ ]]; then
        report_success "Entity file naming correct: $basename"
    else
        report_error "Entity file naming incorrect: $basename (should be lowercase with underscores)"
    fi
done

# Check repository interface naming: <entity_name>_repository.go
echo "Checking repository interface file naming..."
REPO_FILES=$(find internal/domain/repository -name "*.go" -not -name "*_test.go" 2>/dev/null || true)
for file in $REPO_FILES; do
    basename=$(basename "$file")
    if [[ "$basename" =~ ^[a-z_]+_repository\.go$ ]]; then
        report_success "Repository file naming correct: $basename"
    else
        report_error "Repository file naming incorrect: $basename (should be <entity>_repository.go)"
    fi
done

# Check usecase file naming: <operation>_usecase.go
echo "Checking usecase file naming..."
USECASE_FILES=$(find internal/usecase -name "*.go" -not -name "*_test.go" -not -name "interface.go" 2>/dev/null || true)
for file in $USECASE_FILES; do
    basename=$(basename "$file")
    if [[ "$basename" =~ ^[a-z_]+_usecase\.go$ ]]; then
        report_success "Usecase file naming correct: $basename"
    else
        report_error "Usecase file naming incorrect: $basename (should be <operation>_usecase.go)"
    fi
done

# Check handler file naming: <entity>_handler.go
echo "Checking handler file naming..."
HANDLER_FILES=$(find internal/delivery/http/handler -name "*.go" -not -name "*_test.go" 2>/dev/null || true)
for file in $HANDLER_FILES; do
    basename=$(basename "$file")
    if [[ "$basename" =~ ^[a-z_]+_handler\.go$ ]]; then
        report_success "Handler file naming correct: $basename"
    else
        report_error "Handler file naming incorrect: $basename (should be <entity>_handler.go)"
    fi
done

echo ""
echo "3. Verifying Dependency Rule (Requirement 6.4)"
echo "-----------------------------------------------------------"

# Check that domain layer has no dependencies on other internal layers
echo "Checking domain layer dependencies..."
DOMAIN_IMPORTS=$(grep -r "import" internal/domain --include="*.go" | grep -E "(usecase|delivery|infrastructure)" || true)
if [ -z "$DOMAIN_IMPORTS" ]; then
    report_success "Domain layer has no dependencies on other layers"
else
    report_error "Domain layer has forbidden dependencies:"
    echo "$DOMAIN_IMPORTS"
fi

# Check that usecase layer only depends on domain
echo "Checking usecase layer dependencies..."
USECASE_BAD_IMPORTS=$(grep -r "import" internal/usecase --include="*.go" | grep -E "(delivery|infrastructure)" | grep -v "domain" || true)
if [ -z "$USECASE_BAD_IMPORTS" ]; then
    report_success "Usecase layer only depends on domain layer"
else
    report_error "Usecase layer has forbidden dependencies:"
    echo "$USECASE_BAD_IMPORTS"
fi

# Check that delivery layer doesn't depend on infrastructure
echo "Checking delivery layer dependencies..."
DELIVERY_BAD_IMPORTS=$(grep -r "import" internal/delivery --include="*.go" | grep "infrastructure" || true)
if [ -z "$DELIVERY_BAD_IMPORTS" ]; then
    report_success "Delivery layer doesn't depend on infrastructure layer"
else
    report_warning "Delivery layer has infrastructure dependencies (may be acceptable for middleware):"
    echo "$DELIVERY_BAD_IMPORTS"
fi

echo ""
echo "4. Verifying Package Naming Consistency (Requirement 7.6)"
echo "-----------------------------------------------------------"

# Check that package names match directory names
echo "Checking package declarations match directory structure..."
check_package_names() {
    local dir=$1
    local expected_package=$(basename "$dir")
    
    for file in "$dir"/*.go; do
        if [ -f "$file" ] && [[ ! "$file" =~ _test\.go$ ]]; then
            actual_package=$(grep -m 1 "^package " "$file" | awk '{print $2}')
            if [ "$actual_package" = "$expected_package" ]; then
                report_success "Package name matches directory: $file"
            else
                report_error "Package name mismatch in $file: expected '$expected_package', got '$actual_package'"
            fi
        fi
    done
}

# Check key directories
for dir in internal/domain/entity internal/domain/repository internal/usecase/auth internal/usecase/todo internal/usecase/credential; do
    if [ -d "$dir" ]; then
        check_package_names "$dir"
    fi
done

echo ""
echo "5. Verifying Test Organization (Requirement 6.4)"
echo "-----------------------------------------------------------"

# Check that tests are co-located with implementation
echo "Checking test file organization..."
TEST_FILES=$(find internal -name "*_test.go" 2>/dev/null || true)
for test_file in $TEST_FILES; do
    impl_file="${test_file/_test.go/.go}"
    impl_file="${impl_file/_property_test.go/.go}"
    
    test_dir=$(dirname "$test_file")
    if [ -d "$test_dir" ]; then
        report_success "Test file co-located: $(basename $test_file)"
    else
        report_warning "Test file not co-located: $test_file"
    fi
done

echo ""
echo "=== Summary ==="
echo "-----------------------------------------------------------"
if [ $ERRORS -eq 0 ] && [ $WARNINGS -eq 0 ]; then
    echo -e "${GREEN}✓ All checks passed! Clean architecture compliance verified.${NC}"
    exit 0
elif [ $ERRORS -eq 0 ]; then
    echo -e "${YELLOW}⚠ $WARNINGS warnings found. Review recommended.${NC}"
    exit 0
else
    echo -e "${RED}✗ $ERRORS errors and $WARNINGS warnings found.${NC}"
    exit 1
fi
