# Contributing to hexago

Thanks for your interest in contributing.

## Reporting bugs

Open an issue using the bug report template. Include:
- hexago version (`hexago version`)
- Go version (`go version`)
- OS and architecture
- Steps to reproduce
- Expected vs actual behavior

## Proposing features

Open an issue using the feature request template.
Describe the use case before proposing a solution.

## Development setup

```shell
git clone https://github.com/padiazg/hexago.git
cd hexago
make build
make test
make lint
```

Requirements: Go 1.25+, golangci-lint.

## Pull request process

1. Fork the repo and create a feature branch from `main`.
2. Make your changes. Keep commits atomic.
3. Run `make test` and `make lint` — both must pass.
4. Open a PR against `main`. Fill in the PR template.
5. One approval required before merge.

## Code style

- Tests: table-driven with `testify/assert`. Use `go-testgen` for scaffolding.
- Commits: conventional commits (`feat:`, `fix:`, `docs:`, `refactor:`, `test:`).
- No `os.Exit` outside `main` or Cobra's `Execute`.
- Wrap errors with `%w`. No swallowed errors without a comment explaining why.

## License

By contributing you agree your contributions are licensed under the MIT License.
