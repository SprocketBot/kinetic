# Development tools

This directory contains developer and release commands, not tests.

- `start-dev.sh` and `seed-*.sh` manage a development environment.
- `go-test.sh` runs backend tests and enforces the current coverage floor.
- `quality-gate.sh`, `web-quality-gate.sh`, and `ci-verify.sh` compose quality
  checks used locally and in CI.
- `release-evidence.sh` produces release evidence bundles.

Executable test suites are organized under `tests/`.
