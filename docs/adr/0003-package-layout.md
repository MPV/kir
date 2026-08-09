# 3. Layer the code as cmd → processor → yamlparser → k8s

- Status: accepted
- Date: 2025-03-17 _(recorded retrospectively 2026-08-08)_

One-directional flow `main → cmd → processor → yamlparser → k8s`, with `fileutil`
for argument/glob resolution: `cmd` wires the CLI, `processor` orchestrates per
source, `yamlparser` decodes documents, `k8s` pulls the PodSpec and images.

Each layer is unit-testable in isolation and new behavior has an obvious home, at
the cost of more packages than a tool this size strictly needs.
