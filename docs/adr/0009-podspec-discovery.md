# 9. Find PodSpecs with a CUE schema

- Status: **proposed** — one of four candidate answers to #26, supersedes [0001](0001-typed-kubernetes-decoding.md) if accepted
- Date: 2026-08-09

## Context

[ADR 0001](0001-typed-kubernetes-decoding.md) decodes each document with the
typed client-go scheme and reads the PodSpec through a type switch over seven
hardcoded kinds. #26 asks for the PodSpec to be *found* rather than looked up,
and suggests CUE as the way to describe what is being looked for.

## Decision

Stop decoding into typed Kubernetes objects. Decode each document into plain Go
values and walk it, asking a CUE schema whether each node is PodSpec-shaped.

The schema (`k8s/podspec.cue`, embedded) defines `#Container` as a CUE
definition, so it is closed: a container carrying a field the schema does not
list is a mismatch. That is what keeps a resource with a `containers` field
holding something else from being read as a workload.

Because the schema is data rather than Go code, it can be replaced at runtime:
`kir --schema my.cue manifests/` swaps in a different notion of what a container
is, with no rebuild.

## Consequences

Custom resources work — an Argo `Rollout` yields its images with nothing in kir
naming that kind — as do the built-ins the old list omitted
(`ReplicationController`, `PodTemplate`). `List` stops being a special case.
Those gains are shared with options B and D; what is unique here is the
**replaceable schema**, covered by `TestCustomSchema`: the same manifest yields
nothing under the embedded schema and its image under an extended one.

The costs are the reason to think twice:

- **Slowest of the four.** Unifying candidate nodes costs 164 ms → 406 ms over
  1000 documents, roughly 2× the same walk validated by Go types (option B).
- **A large dependency.** `cuelang.org/go` adds 35 module lines to `go.sum` and
  takes the binary from 12.8 MB (option B's walk) to 23.7 MB.
- **A toolchain bump.** CUE requires Go 1.25, so `.tool-versions` moves from
  1.24.3 — and CI reads its Go version from that file.
- **A second schema to maintain.** `podspec.cue` restates part of the container
  API by hand, and it can drift from `k8s.io/api`. Generating it with
  `cue get go k8s.io/api/core/v1` fixes the fidelity but adds a code-generation
  step to the build.

The comparison this sets up: option B gets the same reach and the same
lookalike rejection using `k8s.io/api` as the schema, which kir already depends
on, at a third of the runtime cost and half the binary size. CUE earns its
keep here only if a user-editable schema is worth those numbers.
