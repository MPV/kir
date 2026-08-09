# 4. Pin behaviour with golden (approval) tests

- Status: accepted
- Date: 2025-03-18 _(recorded retrospectively 2026-08-08)_

Use [`go-approval-tests`](https://github.com/approvals/go-approval-tests): run
real manifests through `kir` and compare output against committed
`*.approved.txt` goldens (replacing an informal `examples/` dir). Later refined
to capture stdout, stderr, and the exit code separately through one seam
(ADR 0005).

Adding a case is cheap (drop a manifest, approve output) and the goldens double
as documentation; intentional output changes require an explicit approve step.
