# 7. How kir classifies each document

- Status: accepted
- Date: 2026-08-08

A manifest stream mixes workloads, image-less objects, custom resources, and the
occasional malformed document. Each falls into one of these tiers:

| Tier | Examples | stdout | stderr | exit |
|---|---|---|---|---|
| Workload (has a PodSpec) | Pod, Deployment, …, CronJob | images | — | 0 |
| Known, image-less | Service, ConfigMap, Secret, … | — | — | 0 |
| Unprocessable | malformed YAML, unreadable file | — | `error: …` | non-zero (ADR 0008) |
| Workload with an unreportable image | an image value that is not a valid reference | its other images | `error: …` | non-zero |

The load-bearing choice: a valid image-less document is **not** an error and
**not** a warning — it's expected input with nothing to report, so it's skipped
silently. Only unprocessable input hits stderr and the exit code. That keeps
stdout to images, keeps stderr quiet for normal input, and keeps the exit code
trustworthy (an earlier version erred on every Service, which — once failures
became non-zero, ADR 0008 — would make `kir manifests/*` exit non-zero).

The last tier classifies an image *value*, because stdout is a contract as much
as a report: one image per line, normally fed straight into another program's
arguments. A line break in a value forges an extra entry in that list, escape
sequences can make a terminal show a registry the scanner was never given, and a
leading dash arrives at the scanner as an option. Each is refused on its own,
leaving the document's other images reported.

Validity is the canonical parser's verdict
([`distribution/reference`](https://github.com/distribution/reference)), not a
hand-written rule set — the load-bearing choice here. kir accepts exactly what a
registry client would, so anything it refuses was never a pullable image, which
is what makes refusing safe; and it refuses nothing a registry would serve, so it
cannot drop a real image. Hand-written rules failed both ways: an earlier draft
of this let `nginx:`, `{{.Values.image}}` and `$IMAGE` through as images.
Normalisation is not borrowed, only validation — kir reports what the manifest
said, so `nginx` stays `nginx`, not `docker.io/library/nginx`.

Unregistered kinds (CRDs) are a further case, today handled like image-less —
skipped silently. But a CRD may embed a PodSpec (Argo Rollouts, Knative, …), so
skipping it silently can drop images (the #49 failure mode). Planned (#75): a
`warning:` on stderr, exit 0 — "seen but not detected" — distinct from the silent
known-image-less tier. Updated when that lands.
