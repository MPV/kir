# Architecture Decision Records

Short records of the significant architectural decisions in `kir`, in the order
they were made. See Michael Nygard's
[Documenting architecture decisions](https://cognitect.com/blog/2011/11/15/documenting-architecture-decisions)
for the format.

ADRs 0001–0006 were recorded retrospectively (on 2026-08-08); each carries the
date the decision was actually made.

| ADR | Decision | Date |
| --- | --- | --- |
| [0001](0001-typed-kubernetes-decoding.md) | Extract images by decoding manifests with the typed Kubernetes scheme (under reconsideration — #26) | 2025-03-13 |
| [0002](0002-cli-input-model.md) | Take manifest sources as positional arguments, with `-` for stdin | 2025-03-13 |
| [0003](0003-package-layout.md) | Layer the code as cmd → processor → yamlparser → k8s | 2025-03-17 |
| [0004](0004-approval-testing.md) | Pin behaviour with golden (approval) tests | 2025-03-18 |
| [0005](0005-run-entry-point-seam.md) | Expose the CLI as an in-process `Run(args, stdin, stdout, stderr) int` | 2026-08-02 |
| [0006](0006-conventional-commits-and-releases.md) | Automate releases from Conventional Commits | 2026-08-06 |
