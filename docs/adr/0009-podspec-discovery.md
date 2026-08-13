# 9. Find PodSpecs structurally, validated against the Kubernetes Go types

- Status: **proposed** — one of four candidate answers to #26, supersedes [0001](0001-typed-kubernetes-decoding.md) if accepted
- Date: 2026-08-09

## Context

[ADR 0001](0001-typed-kubernetes-decoding.md) decodes each document with the
typed client-go scheme and reads the PodSpec through a type switch over seven
hardcoded kinds. #26 asks for the PodSpec to be *found* rather than looked up,
so that custom resources embedding one work too.

## Decision

Stop decoding into typed Kubernetes objects. Decode each document into plain Go
values and walk it, testing each node for PodSpec shape.

The shape test is the load-bearing part, and it is not a field-name heuristic: a
candidate's `containers` / `initContainers` / `ephemeralContainers` are decoded
into the real `corev1` types with unknown fields rejected. The Kubernetes Go
types are the schema. A node matches when Kubernetes itself would call it a
PodSpec.

Documents are still required to carry a `kind`, which keeps kir aimed at
manifests rather than at any YAML containing something image-like.

## Consequences

Custom resources work — an Argo `Rollout` yields its images with nothing in kir
naming that kind — and so do the built-ins the old list omitted
(`ReplicationController`, `PodTemplate`). `List` stops being a special case: its
items are just more nodes, so the kind allow-list and the unstructured item
handling both go.

Dropping typed decoding drops `k8s.io/client-go` entirely: the binary goes from
27.2 MB to 12.4 MB. `k8s.io/api` stays, as the schema.

The costs, and they are real:

- **Slower**, since every candidate node is decode-tested: 105 ms → 133 ms over
  1000 documents (~28 µs per document). Acceptable for a tool that shells out to
  an image scanner afterwards.
- **Precision now rests on the strict decode.** A field named `containers`
  holding anything else is rejected, and there is a golden fixture
  (`TestSkipsNonWorkloads.Lookalike`) to keep it that way. But a custom resource
  that inlines a PodSpec *alongside* its own fields would fail the strict decode
  and be missed — the failure mode moves from "kind not listed" to "shape not
  matched".
- **Bound to the vendored `k8s.io/api`**: a container field newer than the
  vendored version fails the strict decode. Bumping the dependency is the fix,
  and a stale bump is now a correctness issue rather than only a hygiene one.
  In practice the binding is loose — the v0.32.3 → v0.36.3 bump moved no
  goldens and needed no code change — but it is the thing to watch.

Alternatives considered: reflecting over the typed scheme (option A) keeps
perfect precision but cannot see custom resources at all; a CUE schema (option
C) buys a user-editable schema for a large dependency; configurable per-kind
paths (option D) keep precision but push the work onto users.
