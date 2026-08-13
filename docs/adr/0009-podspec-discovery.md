# 9. Find PodSpecs by reflecting over the decoded object

- Status: **proposed** — one of four candidate answers to #26, supersedes [0001](0001-typed-kubernetes-decoding.md) if accepted
- Date: 2026-08-09

## Context

[ADR 0001](0001-typed-kubernetes-decoding.md) decodes each document with the
typed client-go scheme and reads the PodSpec through a type switch over seven
hardcoded kinds. #26 asks for the PodSpec to be *found* rather than looked up.

## Decision

Keep typed decoding. Replace the type switch with a reflective walk of the
decoded Go value that collects every field of type `corev1.PodSpec`, wherever it
sits in the struct.

Kinds and paths both disappear: `k8s.FindPodSpecs` matches on type, so any type
the scheme can decode is understood, and the kind allow-list in `yamlparser`
(previously consulted for `List` items) goes with it.

## Consequences

Two built-in kinds that the type switch omitted now work with no code dedicated
to them — `ReplicationController` and `PodTemplate` — and a future workload kind
added to `k8s.io/api` needs no change here.

Precision is unchanged: only real `corev1.PodSpec` values match, so there are no
false positives, and no measurable cost (decoding dominates; see the pull
request for numbers).

The limit is inherited from 0001 and is the reason this may not be the answer to
#26: reflection can only see what the scheme decodes, so a custom resource
embedding a PodSpec — Argo Rollouts, Knative — is still invisible. This ADR
addresses "less hardcoded kinds/structures" but not "easier to use with other
(custom) resources". Options B (structural matching), C (CUE schema), and D
(configurable paths) trade precision or dependencies for that reach.
