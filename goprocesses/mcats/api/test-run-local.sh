#!/bin/bash

# This script is for local testing, e.g. imitating a request being submitted to the VA Controller.
# First arg must be a JSON file that containing typical VA Controller request keys.
# Example usage: ./test-run-local.sh "test-params-999999999999_1-preprocess.json"

set -euo pipefail
echo "Running with args from: ${1}"
go build main.go && ./main "`jq . ${1}`"