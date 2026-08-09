# 9. Find PodSpecs at configured per-kind paths

- Status: **proposed** — one of four candidate answers to #26, supersedes [0001](0001-typed-kubernetes-decoding.md) if accepted
- Date: 2026-08-09

## Context

[ADR 0001](0001-typed-kubernetes-decoding.md) decodes each document with the
typed client-go scheme and reads the PodSpec through a type switch over seven
hardcoded kinds. #26 asks for the PodSpec to be *found* rather than looked up,
mainly so that custom resources embedding one can work.

The kinds and their paths are not wrong — they are simply *in Go*, which is why
a new one needs a release.

## Decision

Keep the lookup, and move it out of Go into configuration.

`k8s/resources.yaml`, embedded, lists each kind and the paths at which it holds
a PodSpec (`spec`, `spec.template.spec`, `spec.jobTemplate.spec.template.spec`).
A `documents` path names nodes to process as objects in their own right, which
is how a `List` unwraps its items — the special case becomes two lines of
configuration.

`kir --config my.yaml` merges a user's file over the built-in one. Entries are
keyed by kind, so a custom resource can be added and a built-in corrected.

Documents are decoded into plain Go values, since a custom resource has to be
readable without the scheme.

## Consequences

This is the cheapest option by every mechanical measure, because the lookup
never has to *decide* anything: 151 ms → 107 ms over 1000 documents (faster than
today, having dropped typed decoding), a 3.9 MB binary against today's 28.9 MB,
and a dependency list that goes from 99 `go.sum` lines to 22 — both
`k8s.io/client-go` and `k8s.io/api` fall away, leaving `sigs.k8s.io/yaml`.

Precision is exact by construction. A path either matches or it does not, so
there are no false positives to guard against and no schema to keep in step with
the Kubernetes API.

The cost is that **it does not answer #26's second motivation on its own**. A
custom resource stays invisible until somebody describes it: `TestCustomResource`
pins both halves, an Argo Rollout yielding nothing by default and its images
under `--config`. Every user of Argo, Knative, or an in-house CRD has to write
that file, and a kind whose PodSpec moves in a later API version needs it
updated. Options B and C recognise those resources with no configuration at all.

The honest framing is that this is the *complement* of structural discovery
rather than a competitor: precise where it is configured, blind where it is not.
It also composes — structural discovery could use a file like this to override
what it infers, which is roughly how Kyverno's `imageExtractors` work alongside
its built-in knowledge.
