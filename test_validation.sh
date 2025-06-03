#!/bin/bash
# Test script for k8s-diff validation functionality
# This script tests various validation scenarios to ensure proper error handling

# NOTE: This script expects the main Go file to be named 'diff.go'.
# If you rename the main file, update the script accordingly.

echo "Testing k8s-diff validation functionality..."
echo

echo "1. Testing missing apiVersion..."
if go run diff.go test_data/invalid/manifest-missing-apiversion.yaml test_data/scenario1/manifest1.yaml 2>&1 | grep -q "missing required field 'apiVersion'"; then
    echo "✓ PASS: Missing apiVersion validation works"
else
    echo "✗ FAIL: Missing apiVersion validation failed"
fi

echo
echo "2. Testing missing kind..."
if go run diff.go test_data/invalid/manifest-missing-kind.yaml test_data/scenario1/manifest1.yaml 2>&1 | grep -q "missing required field 'kind'"; then
    echo "✓ PASS: Missing kind validation works"
else
    echo "✗ FAIL: Missing kind validation failed"
fi

echo
echo "3. Testing missing metadata..."
if go run diff.go test_data/invalid/manifest-missing-metadata.yaml test_data/scenario1/manifest1.yaml 2>&1 | grep -q "missing required field 'metadata'"; then
    echo "✓ PASS: Missing metadata validation works"
else
    echo "✗ FAIL: Missing metadata validation failed"
fi

echo
echo "4. Testing missing metadata.name..."
if go run diff.go test_data/invalid/manifest-missing-name.yaml test_data/scenario1/manifest1.yaml 2>&1 | grep -q "missing required field 'metadata.name'"; then
    echo "✓ PASS: Missing metadata.name validation works"
else
    echo "✗ FAIL: Missing metadata.name validation failed"
fi

echo
echo "5. Testing empty name..."
if go run diff.go test_data/invalid/manifest-empty-name.yaml test_data/scenario1/manifest1.yaml 2>&1 | grep -q "'metadata.name' must be a non-empty string"; then
    echo "✓ PASS: Empty name validation works"
else
    echo "✗ FAIL: Empty name validation failed"
fi

echo
echo "6. Testing invalid namespace type..."
if go run diff.go test_data/invalid/manifest-invalid-namespace.yaml test_data/scenario1/manifest1.yaml 2>&1 | grep -q "'metadata.namespace' must be a string"; then
    echo "✓ PASS: Invalid namespace type validation works"
else
    echo "✗ FAIL: Invalid namespace type validation failed"
fi

echo
echo "All validation tests completed! ✓"
