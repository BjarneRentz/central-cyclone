# Central Cyclone — conventions

Central Cyclone generates SBOMs (via `cdxgen`) for configured repos and uploads them to
DependencyTrack. See `README.md` and `GitOps.md` for behavior; this file is about how code
here is written.

## Testing

- stdlib `testing` only. No testify/gomega/gomock — don't add a test dependency.
- Table-driven for pure functions: anonymous struct slice named `tests`, fields
  `name`/`want`/`wantErr`, `for _, tt := range tests { t.Run(tt.name, ...) }`.
  See `internal/workspace/repomapping_test.go`.
- `TestType_Method_Case` one-test-per-scenario for stateful/orchestration code.
  See `internal/gitops/sync_test.go`.
- Hand-written `Mock*` doubles at the top of the consuming test file: pointer receivers,
  recording call counts/args plus a canned `err`. No `go:generate`.
- `t.TempDir()` (never `os.MkdirTemp`), file modes as `0o644`/`0o755`, inline byte-slice
  fixtures — there is no `testdata/` directory.

## Errors and logging

- `log/slog` package-level funcs, structured KV, `"error", err` as the final pair.
  Emoji prefixes on user-facing progress lines are established style (`🔎`, `💾`, `⬆️`).
- Wrap with `%w`, never `%v`. Newer code (`internal/gitops`) uses a terse
  `verb target: %w` message; keep that style there.
- No sentinel errors in this codebase — don't introduce one unless a caller genuinely
  needs `errors.Is`.

## Structure and DI

- Interfaces are defined producer-side, in the package owning the implementation, kept
  small (1–4 methods): `gittool.Cloner`, `workspace.Workspace`, `analyzer.Analyzer`,
  `upload.Uploader`, `query.ValueExtractor`, `dt.Client`.
- Constructor DI: dependencies passed as interfaces, stored unexported. `New*` returns a
  concrete pointer; `Create*` returns an interface.
- Wiring is manual and top-down in the cobra `RunE`. No DI framework.
- Config reaches commands via `extensions.RequireConfig(cmd)` / `extensions.GetSettings(cmd)`
  (cobra `PreRunE` + typed context key) — reuse these rather than re-reading the file or
  adding globals.

## Workflow

- Conventional Commits, lowercase, imperative, no scope, no trailing period:
  `feat:`, `fix:`, `docs:`, `chore(deps):`.
- Land work via PR from a short-lived topic branch; never commit to `main` directly.
- Verify with `go build ./... && go test ./...`. There is no Makefile and no linter.
