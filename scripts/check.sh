#!/usr/bin/env bash
# Runs locally what GitHub Actions CI ran on pull requests (CI is off to save
# paid minutes): the Makefile's check target, the steps of
# .github/workflows/test.yml. The pre-push hook runs it. There is no slow
# suite, so there is no --full.
set -euo pipefail
cd "$(dirname "$0")/.."

case "${1:-}" in
  "") ;;
  *) echo "usage: scripts/check.sh" >&2; exit 2 ;;
esac

command -v make >/dev/null 2>&1 || { echo "check: make is missing; sudo apt-get install make" >&2; exit 1; }
make --no-print-directory check
printf '\ncheck: passed\n'
