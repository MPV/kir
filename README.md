# 🥂 `kir` (Kubernetes Image Retriever)

## What

- Send it a k8s manifest file, and you get a (newline separated) list of the OCI images those pods would run.

## Why

- Sometimes you want to do things for the list of images in a given set of kubernetes manifests
- ...like scanning them for vulnerabilities.

## Alternatives considered

1. If one can pick Syft/Grype, this looks like it'll solve the same problem:
   - https://github.com/anchore/syft/issues/2729
   - https://github.com/anchore/grype/issues/1259
   - https://github.com/anchore/syft/issues/562
1. But if one must use another image scanning tool (🙉), building this myself is the best I've found (yet?).

## Usage

### Get images for a manifest:

```shell
$ go run main.go approvals/kir_test.TestKind.StatefulSet.input.yaml
registry.k8s.io/nginx-slim:0.8
gcr.io/google-containers/sidecar
kiwigrid/k8s-sidecar
```

### Get images for all manifests matching a glob:

```shell
$ go run main.go approvals/kir_test.TestKind.*.input.yaml | sort -u
busybox:1.28
gcr.io/google-containers/busybox
gcr.io/google-containers/sidecar
kiwigrid/k8s-sidecar
nginx
perl
registry.k8s.io/nginx-slim:0.8
```

### Get images from a running cluster:

```shell
$ kubectl -n kube-system get pod kube-proxy-mzp9j -o yaml | go run main.go -
registry.k8s.io/kube-proxy:v1.31.7

$ kubectl get pod -A -o yaml | go run main.go - | sort -u
# [...]
```

### Scan images from a manifest:

```shell
# Syft:
$ go run main.go approvals/kir_test.TestKind.Job.input.yaml | xargs syft

# Snyk:
$ go run main.go approvals/kir_test.TestKind.Job.input.yaml | xargs snyk container test

# Docker Scout
$ go run main.go approvals/kir_test.TestKind.Job.input.yaml | xargs docker scout cves
```

## How `kir` treats each document

A manifest stream usually mixes workloads with other objects. `kir` handles each by kind:

| Document | Result |
| --- | --- |
| A configured kind — `Pod`, `Deployment`, `DaemonSet`, `ReplicaSet`, `ReplicationController`, `PodTemplate`, `StatefulSet`, `Job`, `CronJob`, `List` | its images are printed to stdout |
| A valid object with no images — `Service`, `ConfigMap`, `Secret`, … | skipped silently (exit 0) |
| Malformed or unreadable input | reported on stderr, non-zero exit |
| A workload whose image value isn't a valid image reference | that image is reported on stderr with a non-zero exit; the document's other images are still printed |
| A custom resource kir has not been told about | skipped silently (exit 0) — describe it with `--config` |

So stdout carries only images and stderr stays quiet for normal input. See [ADR 0007](docs/adr/0007-document-classification.md) for the rationale.

### Teaching `kir` a custom resource

Which kinds hold images, and where, is configuration rather than Go code — see [`k8s/resources.yaml`](k8s/resources.yaml). Locations are [JMESPath](https://jmespath.org) expressions. Point `--config` at your own file to describe a custom resource; entries are merged over the built-in ones, keyed by kind:

```yaml
# rollouts.yaml
resources:
  - kind: Rollout
    podSpecs: [spec.template.spec]
```

```shell
$ kir --config rollouts.yaml manifests/
```

Use `containers` for a resource that holds bare containers instead of a `PodSpec`. An Argo `Workflow` keeps a list of templates, each holding a container, a script, or neither — one expression covers all of it:

```yaml
resources:
  - kind: Workflow
    containers: ["spec.templates[*].[container, script][]"]
```

Expressions are compiled when the file loads, so a typo is reported then rather than silently matching nothing.
