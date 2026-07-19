# Contributing

The core is intentionally strict.

1. Keep public syntax smaller than the equivalent raw Bot API request.
2. Do not add a dependency to solve a standard-library-sized problem.
3. Do not add reflection to update routing or request dispatch.
4. Stream files; never buffer entire uploads.
5. Add unit and race tests for every concurrent behavior.
6. Add benchmarks for changes to routing, decoding, or request construction.
7. Preserve the raw escape hatch when adding typed abstractions.

Before submitting:

```bash
go generate ./types
./scripts/check-generated.sh
gofmt -w .
go vet ./...
go test -race -shuffle=on ./...
```

Maintainers preparing a release must use the actual target toolchain and run
`./scripts/release-check.sh`; see [`docs/releasing.md`](docs/releasing.md).
