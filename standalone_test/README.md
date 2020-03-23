Standalone tests for go-subvert
===============================

These are just copies of the normal tests, packed into a regular binary.

This serves two purposes:

1. It tests in a normal executable environment
2. It allows all tests to be run under Windows (the testing environment in Windows has some problems)

### To run tests:

    go build
    ./standalone_test
