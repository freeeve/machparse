#!/bin/bash
# Run all fuzz tests sequentially with 8 workers each
# Usage: ./scripts/fuzz.sh [duration]
# Default duration is 10m (10 minutes)

DURATION=${1:-10m}
PARALLEL=${2:-8}

echo "Running fuzz tests with $PARALLEL workers, $DURATION each"
echo "================================================"

TESTS=(
    "FuzzParse"
    "FuzzLexer"
    "FuzzParseAll"
    "FuzzWalk"
    "FuzzRewrite"
    "FuzzFormat"
    "FuzzPooling"
    "FuzzDialects"
)

FAILED=()
PASSED=()

for test in "${TESTS[@]}"; do
    echo ""
    echo ">>> Running $test for $DURATION..."
    echo "---"

    if go test ./fuzz/... -fuzz="^${test}$" -fuzztime="$DURATION" -parallel="$PARALLEL" 2>&1; then
        PASSED+=("$test")
        echo "<<< $test PASSED"
    else
        FAILED+=("$test")
        echo "<<< $test FAILED"
    fi
done

echo ""
echo "================================================"
echo "Summary:"
echo "  Passed: ${#PASSED[@]} (${PASSED[*]:-none})"
echo "  Failed: ${#FAILED[@]} (${FAILED[*]:-none})"

if [ ${#FAILED[@]} -gt 0 ]; then
    exit 1
fi
