# 9. Find PodSpecs with a CUE schema generated from the Kubernetes types

- Status: **proposed** — one of four candidate answers to #26, supersedes [0001](0001-typed-kubernetes-decoding.md) if accepted
- Date: 2026-08-09

## Context

[ADR 0001](0001-typed-kubernetes-decoding.md) decodes each document with the
typed client-go scheme and reads the PodSpec through a type switch over seven
hardcoded kinds. #26 asks for the PodSpec to be *found* rather than looked up,
and suggests CUE as the way to describe what is being looked for.

## Decision

Decode each document into plain Go values, walk it, and ask a CUE schema whether
each node is PodSpec-shaped.

The schema is **generated from `k8s.io/api/core/v1` by `cue get go`** (`make
schema`, output checked in under `k8s/cue.mod/gen/`) rather than written by
hand. What kir maintains is nineteen lines that import the generated package:

```cue
import corev1 "k8s.io/api/core/v1"

#Containers: [...corev1.#Container]
#EphemeralContainers: [...corev1.#EphemeralContainer]
```

Generated definitions are CUE definitions and so are closed: a container
carrying a field the Kubernetes API does not define is a mismatch. That is what
stops a resource with a `containers` field holding something else from being
read as a workload.

Because the schema is data, it can be replaced at runtime — `kir --schema
my.cue` — and a user schema is loaded *inside the same module*, so it can import
the official definitions and extend them rather than restate them.

## Consequences

Custom resources work, as do the built-ins the old kind list omitted, and `List`
stops being a special case. Those gains are shared with options B and D. What is
unique here is the **replaceable schema**, covered by `TestCustomSchema`.

Reusing the generated types removes the fidelity objection to a hand-written
CUE schema — there is nothing to drift. It replaces it with three sharper ones:

- **Cost.** 105 ms → **2163 ms** over 1000 documents, ~21× today and ~16× option
  B, which validates the same shapes using the same Kubernetes types as Go code.
  Unifying against the full closed `#Container` and its dependency graph is
  simply expensive.
- **A trap.** Pre-evaluating the definitions (`cue.Value.Eval()`) makes it ~7×
  faster and **silently discards closedness**, so every lookalike starts
  matching. Only the `TestSkipsNonWorkloads.Lookalike` fixture catches it. An
  optimisation that turns a correctness guarantee off without erroring is a
  sharp edge to hand to a future maintainer.
- **An impedance mismatch.** Manifests decode YAML-to-JSON-to-Go, making every
  number a `float64`; offering that to a schema saying `int32` rejects every
  container declaring a port. Candidates therefore reach CUE as JSON rather than
  via `ctx.Encode`. `TestFindImagesIntegerFields` pins it.
- **Code generation that needs maintaining.** 47 generated `.cue` files checked
  in and a `make schema` step to re-run whenever `k8s.io/api` is bumped. The
  v0.32.3 → v0.36.3 bump showed what that costs in practice, twice over:
  `go mod tidy` drops `k8s.io/api` from `go.mod` — kir's Go code does not import
  it, and CUE imports are invisible to the Go toolchain — so the schema's own
  source has to be pinned by a build-tagged blank import (`k8s/schema_source.go`)
  or it silently disappears; and `go run cuelang.org/go/cmd/cue@v0.17.1` builds
  the generator with the Go version in *CUE's* `go.mod` (1.25) while
  `k8s.io/api` v0.36.3 requires 1.26, so generation fails with "package requires
  newer Go version" until `GOTOOLCHAIN` is pinned in the target. Neither is
  hard once known; both are invisible until a bump breaks them.
- The Go floor is no longer a cost: CUE needs 1.25 and `master` is on 1.26.5, so
  this no longer moves `.tool-versions`.

The comparison this sets up: option B gets the same reach and the same lookalike
rejection from `k8s.io/api` — a dependency kir already has — at 1/12th the
runtime, a third of the binary, and no code generation. CUE earns its keep here
only if a user-replaceable schema is worth that.
