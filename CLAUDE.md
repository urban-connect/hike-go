# CLAUDE.md

This file provides guidance to Claude Code (claude.ai/code) when working with code in this repository.

## What this is

`github.com/urban-connect/hike-go` is a shared Go library ("Hike") holding bits and pieces of code
reused across Go projects: configuration loading, encryption and secret token handling, HTTP request
parsing. It has no `main` package and no binary — it is consumed as a module by other repos.

## Keep it general

Hike is a library, not an application. Every exported API is written to serve a caller whose types it
does not know about: readers take `any` and fill a config struct the caller owns, `NewLogger` takes an
embeddable `BaseConfig` rather than a concrete config type, and `config.FromEncryptedFile` accepts a
`config.Crypto` interface rather than importing the `crypto` package.

**Preserve that generality when changing anything.** A change that only makes sense for one consuming
application belongs in that application, not here.

## Commands

```bash
go build ./...
go test ./...
go test -run TestName ./crypto   # single test
go vet ./...
gofmt -l .         # must print nothing; `gofmt -w .` to fix

# CI gates on this analyzer:
go run golang.org/x/tools/go/analysis/passes/modernize/cmd/modernize@latest ./...
go run golang.org/x/tools/go/analysis/passes/modernize/cmd/modernize@latest -fix ./...
```

`.github/workflows/ci.yml` runs `go test`, `go vet`, `gofmt`, and `modernize` on every PR and push to
main. Keep them green. Tests use `testify` (`assert`/`require`) and need no external services.

## Architecture

Three independent packages; none imports another. The only direct third-party dependencies are
`go-envconfig` and `golang.org/x/crypto` — logging is stdlib. Keep that list short.

**`config`** — layered configuration. `Reader`/`ReaderFunc` is the core abstraction: each reader
mutates a caller-supplied config struct, so callers compose them by calling several in sequence (later
readers overwrite earlier ones). Three are provided: `FromEnv(prefix)`, `FromFile(name, optional)`,
`FromEncryptedFile(name, optional, decryptor)`. The `optional` flag makes a missing file a no-op rather
than an error.

`BaseConfig` is meant to be embedded in an app's own config struct. Fields carry both `json` and `env`
tags because the same struct is filled from either source. Env lookup goes through `UpcaseLookuper`, so
an `env:"crypto_key"` tag with prefix `APP_` reads `APP_CRYPTO_KEY`.

`FromEncryptedFile` takes a `config.Crypto` interface (`Encrypt`/`Decrypt`) rather than importing
`crypto` — the `crypto.AES` type satisfies it, and the caller wires them together. Keep that decoupling
when extending.

`NewLogger` returns a stdlib `*slog.Logger`: a JSON handler at info level when `Env` is `production`, a
text handler at debug level otherwise. It takes `BaseConfig` (not the app's concrete config) and cannot
fail, so it returns no error. This library has no logging dependency — do not introduce one.

**`crypto`** — `AES` does AES-GCM with a *fixed* key and nonce, both base64-decoded from config strings.
Reusing a nonce is intentional here (it makes encrypted config files reproducible), but it means the
type is unsuitable for encrypting multiple distinct messages under one key. `Token` wraps a random
string, hashes it with bcrypt for storage (`Digest`), and compares against a stored digest (`Validate`).

**`api`** — `ParseParams(r, &v)` reflects over a struct and populates fields from an `*http.Request`.
Struct tag format is `params:"key,source"` where source is `path`, `query`, or `header`; with no tag the
field name is lowercased and read from the query string. Path values use `r.PathValue`, i.e. Go 1.22
`net/http` routing patterns. Supported field kinds are string, int, int64, float32, float64, bool —
anything else is an error. Empty values leave the field at its zero value rather than failing.

## Conventions

- Wrap every returned error with `fmt.Errorf("...: %w", err)` and a lowercase description of what failed.
- Assign, blank line, then check `err` — the existing code separates the call from its error check.
- Constructors are `NewX` returning `(*X, error)`; keep struct fields unexported and expose behavior
  through methods.
- Keep examples and test fixtures generic. This is a public repository — do not use real hostnames,
  internal project names, or anything else specific to a private codebase.

## Git & PR Conventions

- Do not add Claude as a co-author to commit messages and PR descriptions
- Provide short description for the PR
- **Do not add test cases to PR descriptions**
- Always assign the PR to its author when creating it
- Do not commit unless explicitly asked to
- Do not push branches unless explicitly asked to
- Never force push (`git push --force` / `git push -f`) branches to GitHub
- To resolve conflicts with main, use `git pull origin main` instead of `git rebase` (rebase requires
  force push)
- Always ask before using `--admin` flag when merging PRs — it bypasses branch protection checks
- Try to keep branch names short but readable
- Never prefix branch names (no `feat/`, `fix/`, `chore/`, etc.) — use a short, readable descriptive
  name only
