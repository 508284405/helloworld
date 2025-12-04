# Repository Guidelines

This repository is a minimal Go module (Go 1.21) that builds a single executable from `main.go`. Use the standards below to keep changes simple, consistent, and easy to review.

## Project Structure & Module Organization

- Root: `go.mod`, `main.go` (program entry).
- Tests live next to code as `*_test.go` files.
- When the project grows, prefer:
  - `cmd/helloworld/main.go` for the CLI entry.
  - `pkg/<name>` for reusable packages; `internal/<name>` for private packages.
  - `testdata/` for fixtures used by tests.

## Build, Test, and Development Commands

- Run locally: `go run .`
- Build binary: `go build -o bin/helloworld .` then `./bin/helloworld`
- Format code: `gofmt -s -w .`
- Static checks: `go vet ./...`
- Tests (with coverage): `go test ./... -race -cover`
- Tidy modules: `go mod tidy`

## Coding Style & Naming Conventions

- Always run `gofmt` before committing; use tabs (Go default).
- Packages: short, lowercase, no underscores (e.g., `greeting`).
- Files: lowercase, underscores allowed when helpful (e.g., `http_client.go`).
- Exported identifiers use CamelCase starting with an uppercase letter; unexported start lowercase.
- Errors: return errors (avoid `panic` in libraries); wrap with `fmt.Errorf("...: %w", err)` for context.

## Testing Guidelines

- Use the standard `testing` package; prefer table-driven tests.
- Test files: `*_test.go`; test funcs: `TestXxx(t *testing.T)`.
- Run full suite: `go test ./... -race -cover`.
- Keep tests deterministic; place sample inputs under `testdata/`.

## Commit & Pull Request Guidelines

- Commit messages: follow Conventional Commits (e.g., `feat: add personalized greeting`, `fix: handle edge case in loop`).
- Scope small, single-purpose commits; include why, not just what.
- PRs must include: clear description, steps to run/test, linked issues (if any), and relevant output or screenshots.
- Add/update tests for behavior changes; note breaking changes in the PR description.

## Security & Configuration Tips

- Do not commit secrets; prefer environment variables. If used, ignore `.env` locally.
- Keep dependencies tidy and reproducible with `go mod tidy` and `go build` on a clean workspace.

