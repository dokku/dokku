# NOTICE

Files in this directory are derived from
[github.com/databus23/helm-diff](https://github.com/databus23/helm-diff),
which is distributed under the Apache License 2.0. See `LICENSE` for the
full text.

Source files merged into this package:

- `diff/diff.go`
- `diff/report.go`
- `manifest/parse.go`

Modifications:

- Merged the upstream `diff` and `manifest` packages into a single `helmdiff`
  package; references to `manifest.MappingResult` are now bare `MappingResult`.
- Removed three-way-merge support (the upstream `manifest/generate.go` and
  `manifest/util.go` files), which imported Helm v4 and other dependencies not
  used by Dokku.
- Removed `manifest.ParseObject`, which depended on `k8s.io/apimachinery/pkg/runtime`.
- Removed release-ownership tracking (`ManifestsOwnership`, `OwnershipDiff`,
  `Releases`, `reIndexForRelease`).
- Removed the simple, JSON, structured, dyff, and template output formatters and
  the upstream `diff/structured.go` and `diff/constant.go` files. Only the
  default unified-diff (`"diff"`) output format remains.
- `setupReportFormat` is hard-wired to `setupDiffReport`.
- `ReportEntry.Structured` and the `*Options.StructuredOutput()` method removed.
