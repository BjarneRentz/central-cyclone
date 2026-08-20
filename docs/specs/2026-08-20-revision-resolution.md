# Spec: resilient revision resolution for GitOps checkout

Status: approved for implementation
Date: 2026-08-20

## Problem

In GitOps mode, Central Cyclone extracts the deployed image tag for an app/environment from a
GitOps repo's `values.yaml` (via a yq path) and passes that raw string as the `revision` to
check out in the app's source repo (`internal/gitops/handler.go` → `ClonedRepo.CheckoutRevision`
in `internal/gittool/cloned_repo.go`).

Some deployments use SemVer build metadata to identify a revision that is a successor of a
released tag but not itself tagged, e.g. `1.2.0+34ab12cd`, where `34ab12cd` is the commit hash
and `1.2.0` is the nearest ancestor tag. `CheckoutRevision` today has no notion of this format:
the whole string is handed to `ResolveRevision`, which fails to resolve `1.2.0+34ab12cd` as a
ref, and the checkout — and therefore the SBOM generation for that app/environment — fails
silently for every deployment using this convention.

## Scope

- Change is confined to `internal/gittool/cloned_repo.go`.
- `internal/gitops/handler.go` is the only caller of `CheckoutRevision` and is unchanged; the
  function's signature (`CheckoutRevision(revision string) error`) does not change.
- No config, interface, or CLI surface changes.

## Terminology

- **revision**: the raw string extracted from the GitOps values file and passed to
  `CheckoutRevision` — may be a full SHA, a ref name, or a SemVer-with-build-metadata string.
- **commit-ish**: the substring after the first `+` in a revision, when present.

## Resolution order (normative)

Given a `revision` string, `CheckoutRevision` resolves a target commit hash as follows:

1. If `revision` is a full 40-character SHA-1 (`plumbing.IsHash`), use it directly. (Existing
   behavior, unchanged.)
2. Otherwise, if `revision` contains a `+`, and the substring after the **first** `+` is
   4–40 characters of `[0-9a-fA-F]` (a plausible short-to-full commit hash), attempt to resolve
   that substring as a revision (named ref lookup, or direct hash if it's a full 40-char hash;
   peel annotated tags to their commit — see step 3's peeling rule). If this resolves
   successfully, its commit is the target and resolution stops here.
3. If step 2 did not apply (no `+`, or the suffix isn't a valid hex commit-ish) or step 2's
   resolution failed, resolve the **entire original `revision` string** as a named reference
   (tag, branch, or short hash) via `ResolveRevision`. If the resolved object is an annotated
   tag, peel it to the commit it points at. (Existing behavior, unchanged.)
4. If step 3 fails, return an error that names the **full original `revision`** string (not
   just the commit-ish suffix), so operators can diagnose from the log alone.

In short: a valid-looking commit-ish suffix is tried first; any failure — malformed suffix or
failed resolution — falls back to resolving the untouched original string, exactly as today.

## Non-goals

- The tag portion before `+` (e.g. `1.2.0`) is not parsed or validated as SemVer, and is not
  cross-checked against the resolved commit (no ancestry/"is this actually a successor of
  1.2.0" check). It is purely discarded once the commit-ish is resolved.
- No support for other metadata conventions (e.g. `-`, `_`, `~` separators) — only `+`.
- No change to checkout behavior itself: checkout remains forced
  (`git.CheckoutOptions{Force: true}`) once a target hash is determined.

## Behavior table

| # | Input revision | Expected result |
|---|---|---|
| 1 | 40-char SHA-1 | Checks out that commit directly (fast path, unchanged) |
| 2 | Plain lightweight tag name | Checks out the tagged commit (unchanged) |
| 3 | Annotated tag name | Checks out the commit the tag points at (unchanged) |
| 4 | `<tag>+<valid short hex, 4-39 chars>` where the hex resolves to a real commit in the repo | Checks out that commit |
| 5 | `<tag>+<valid 40-char hex>` that is a real commit in the repo | Checks out that commit |
| 6 | `<tag>+<suffix containing non-hex characters>` (e.g. `1.2.0+34asdadasd` — contains `s`) | Suffix is rejected as a commit-ish; falls back to resolving the full string `1.2.0+34asdadasd`, which fails; returns an error naming that full string |
| 7 | `<tag>+<suffix shorter than 4 hex chars>` | Suffix rejected; falls back to resolving the full string; error if that fails |
| 8 | `<tag>+` (empty suffix) | Suffix rejected (empty); falls back to resolving the full string; error if that fails |
| 9 | `<tag>+<well-formed hex suffix not present in the repo>` | Suffix resolution fails; falls back to resolving the full original string; error naming the full string if that also fails |

## Logging

Existing log lines are preserved:
- `slog.Info("🔄 Preparing checkout", "repo", ..., "revision", revision)` — logs the original,
  unmodified revision string exactly as received.
- `slog.Info("🏷️  Checking out", "repo", ..., "hash", targetHash.String())` — logs the final
  resolved hash, regardless of which resolution step produced it.

No new log line is required, but the resolved hash log must reflect whichever path (commit-ish
or full string) actually produced the target, so an operator can infer which path was taken by
comparing the logged `revision` and `hash`.
