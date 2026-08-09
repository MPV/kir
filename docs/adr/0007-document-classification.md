# 7. How kir classifies each document

- Status: accepted
- Date: 2026-08-08

A manifest stream mixes workloads, image-less objects, custom resources, and the
occasional malformed document. Each falls into one of three tiers:

| Tier | Examples | stdout | stderr | exit |
|---|---|---|---|---|
| Workload (has a PodSpec) | Pod, Deployment, …, CronJob | images | — | 0 |
| Known, image-less | Service, ConfigMap, Secret, … | — | — | 0 |
| Unprocessable | malformed YAML, unreadable file | — | `error: …` | non-zero (ADR 0008) |

The load-bearing choice: a valid image-less document is **not** an error and
**not** a warning — it's expected input with nothing to report, so it's skipped
silently. Only unprocessable input hits stderr and the exit code. That keeps
stdout to images, keeps stderr quiet for normal input, and keeps the exit code
trustworthy (an earlier version erred on every Service, which — once failures
became non-zero, ADR 0008 — would make `kir manifests/*` exit non-zero).

Unregistered kinds (CRDs) are a fourth case, today handled like image-less —
skipped silently. But a CRD may embed a PodSpec (Argo Rollouts, Knative, …), so
skipping it silently can drop images (the #49 failure mode). Planned (#75): a
`warning:` on stderr, exit 0 — "seen but not detected" — distinct from the silent
known-image-less tier. Updated when that lands.
