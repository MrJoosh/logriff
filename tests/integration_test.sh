#!/usr/bin/env bash
# Cross-language output verification for LogPulse.
set -euo pipefail

ROOT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")/.." && pwd)"
SAMPLE_LOG="${ROOT_DIR}/data/sample.log"

# TODO: run each language implementation against SAMPLE_LOG with --format json,
# normalize with jq, and assert that outputs match.
echo "TODO: integration tests are not yet implemented"
echo "sample log: ${SAMPLE_LOG}"
