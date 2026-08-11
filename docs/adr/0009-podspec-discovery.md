# 9. Infer images structurally, with configured overrides

- Status: **proposed** — a fifth candidate for #26, combining the approaches in #82 and #84; supersedes [0001](0001-typed-kubernetes-decoding.md) if accepted
- Date: 2026-08-09

## Context

[ADR 0001](0001-typed-kubernetes-decoding.md) decodes each document with the
typed client-go scheme and reads the PodSpec through a type switch over seven
hardcoded kinds. #26 asks for images to be *found* rather than looked up, so
custom resources work too.

Four candidates were raised and measured (#81–#84). Two of them are the ones
that matter, and each fails where the other succeeds:

- **Structural inference** (#82) recognises anything shaped like a PodSpec, with
  no configuration — but a resource holding *bare containers* has no PodSpec
  shape to match, so an Argo `Workflow` is invisible to it, with no way for a
  user to say otherwise.
- **Configured expressions** (#84) reach anything a user can write an expression
  for, exactly — but every Argo, Knative or in-house CRD stays invisible until
  somebody writes that file, including resources whose shape already speaks for
  itself.

## Decision

Do both, with configuration taking precedence per kind.

`Config.FindImages` looks up the document's kind. An entry decides on its own:
its expressions are followed and the walk is not consulted. Everything else is
inferred by the structural walk. The two never both contribute to one document,
so an image cannot be reported twice.

The built-in `resources.yaml` keeps entries for the built-in kinds, but **as an
accelerator, not as knowledge**. Deleting every entry changes no answer, only
speed — `TestBuiltInConfigIsRedundant` pins exactly that, comparing each
built-in kind's configured result against its inferred one. Without that test
`resources.yaml` would quietly become the hardcoded kind list #26 set out to
remove.

## Consequences

The union of both reaches, and one capability neither has alone:

- Built-in workloads, and custom resources embedding a PodSpec (Argo `Rollout`),
  need **no configuration** — inferred.
- Resources holding bare containers (Argo `Workflow`) are reachable, which
  inference alone cannot do.
- An entry with **no expressions silences a kind**, so a user can overrule the
  walk when it reads something wrongly. Inference alone cannot be told to
  ignore; configuration alone has nothing to ignore.

It is also **cheaper than inference alone on typical input**: with the built-ins
configured, ordinary manifests take the exact lookup and the walk never runs.
Over 1000 Deployments, 107 ms — against 171 ms for inference alone (#82), and
near configuration alone's 92 ms (#84). Only kinds nobody has described pay for
the walk, which is the reverse of the usual cost of combining two mechanisms.

The costs:

- **Two mechanisms** to document and reason about, where each of #82 and #84 has
  one. The precedence rule is the whole of the extra contract, but it is a
  contract.
- `k8s.io/api` (the inference schema) *and* `go-jmespath` (the expression
  engine): 12.6 MB and 78 `go.sum` lines. Larger than configuration alone
  (4.1 MB, 42) and barely above inference alone (12.4 MB, 72); still less than
  half of today's 27.2 MB.
- Precision is uneven by design. The walk validates shape against the Kubernetes
  types and rejects lookalikes; a configured expression is taken at its word. An
  entry selecting something that is not a container reports nonsense, and
  nothing checks it.

This also makes the "seen but not detected" warning planned in #75 both rare and
actionable for the first time: a document that is neither configured nor yields
anything from the walk is precisely the case worth reporting, and `--config` is
the remedy to point the user at. Under inference alone the warning has no
remedy; under configuration alone nearly every custom resource trips it.
