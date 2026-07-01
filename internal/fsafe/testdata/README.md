# fsafe security test corpus

`security-cases.json` is the centralized, documented attack corpus for the
filesystem-safety gate (ADR-005). It backs Phase 1 **criterion 17**
("path-traversal + symlink-escape testdata rejected pre-write"). The gate's own
tests (`fsafe_test.go`, `fsafe_unix_test.go`) construct attacks inline; this
corpus is a single auditable list of what must be rejected, consumed by two
data-driven tests:

- `testdata_security_test.go` — `TestSecurityCorpusTraversalRejected`
  (cross-platform, no build constraint).
- `testdata_security_unix_test.go` — `TestSecurityCorpusSymlinkRejected`
  (`//go:build unix`, matching `fsafe_unix_test.go`).

For every case the harness builds a fresh `fsafe.New(t.TempDir())` gate, then
asserts that **both** `Resolve` and `WriteFile` return the expected sentinel via
`errors.Is`, and that **zero** payload bytes land outside the boundary.

## Delivery mechanism

Symlinks are **materialized at runtime** with `os.Symlink`, not committed as live
symlinks — the corpus stays portable (no dependency on Windows `core.symlinks`)
and nothing escaping is ever checked into the repo. The `{{OUTSIDE}}` placeholder
is substituted per case with a temp directory outside the boundary root, so
absolute-outside and symlink-target paths need no machine-specific values.

## Schema

```jsonc
{
  "traversal": [
    { "name": "...", "input": "<path or {{OUTSIDE}}/...>", "expect": "ErrOutsideBoundary", "criterion": 17 }
  ],
  "symlink": [
    { "name": "...", "link": "<symlink path under root>", "target": "{{OUTSIDE}}",
      "input": "<path traversing link>", "expect": "ErrSymlinkEscape", "criterion": 17 }
  ]
}
```

## Cases

| Group | Case | Input | Expected sentinel |
|---|---|---|---|
| traversal | `parent-relative` | `../outside.txt` | `ErrOutsideBoundary` |
| traversal | `nested-then-escape` | `a/../../x.txt` | `ErrOutsideBoundary` |
| traversal | `deep-etc-passwd` | `../../etc/passwd` | `ErrOutsideBoundary` |
| traversal | `absolute-outside-root` | `{{OUTSIDE}}/loot.txt` | `ErrOutsideBoundary` |
| symlink | `dir-symlink-escape` | `link/loot.txt` (via `link` → `{{OUTSIDE}}`) | `ErrSymlinkEscape` |
| symlink | `nested-symlink-escape` | `sub/escape/file.txt` (via `sub/escape` → `{{OUTSIDE}}`) | `ErrSymlinkEscape` |
