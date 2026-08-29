# Tests

Go unit and integration tests live beside the package they exercise. Run the
normal backend suite and its coverage regression check with:

```sh
./tools/go-test.sh
```

Database integration tests require `TEST_DATABASE_URL`. They currently share
database state, so run packages serially (`go test -p 1`) against a disposable
database. Cross-process smoke tests live in `tests/smoke`.

## Smoke tests

- `api-workflow.sh` starts PostgreSQL and the API locally, then exercises the
  primary API workflow.
- `kubernetes-api-workflow.sh` exercises the same workflow against the local
  Kubernetes deployment.
- `legacy/` contains milestone-specific snapshots retained for historical
  debugging. They are not part of the current CI contract.

Run the local suite from the repository root:

```sh
./tests/smoke/api-workflow.sh
```
