# 1. Extract images by decoding manifests with the typed Kubernetes scheme

- Status: accepted — **under reconsideration (#26)**
- Date: 2025-03-13 _(recorded retrospectively 2026-08-08)_

Decode each document with the typed client-go scheme and read the PodSpec via a
type switch over a fixed set of workload kinds (`Pod`, `Deployment`, `DaemonSet`,
`ReplicaSet`, `StatefulSet`, `Job`, `CronJob`). Anything else yields no images.

Precise and dependency-light, but only the hardcoded kinds are understood:
custom resources that embed a PodSpec (Argo Rollouts, Knative, …) are silently
missed. #26 proposes finding the PodSpec structurally instead (Cue in #27–#30)
and #75 catalogs what's missed — hence "under reconsideration".
