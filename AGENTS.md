# Agent guide

Conventions for anyone — human or AI — changing `kir`. Read
[`CONTRIBUTING.md`](CONTRIBUTING.md) and [`docs/adr/`](docs/adr/) first; this
file is the short version of what trips agents up.

## Architecture decisions

- The significant design decisions live in [`docs/adr/`](docs/adr/). **Read them
  before changing how `kir` is structured or behaves** — they say *why* it is the
  way it is.
- Revisiting a decision? Update or supersede its ADR **in the same PR**. New
  decision? Add a new ADR (next number, dated today). Don't silently diverge.

## Testing

- Behavioural changes are pinned by **golden/approval tests through `cmd.Run`**,
  the one CLI seam — see ADR 0004/0005. Goldens live in `approvals/` as the
  `(stdout, stderr, exitcode)` triple.
- A new test must **fail without your change** — if it passes on the old code,
  it isn't guarding anything. Verify that.
- Put the test where the failure lives: an algorithm bug → a unit test at that
  package; a wiring/CLI-contract change → an approvals golden. Inputs hostile to
  fixture files (trailing whitespace, CRLF) belong in a Go unit test, not a
  checked-in `.yaml`, which tooling silently normalises.
- Run before opening a PR: `make test`, `go vet ./...`, `gofmt -l .` (empty),
  `go test -race ./...`.

## Commits & PRs

- [Conventional Commits](https://www.conventionalcommits.org) on **commits and
  the PR title** — squash-merges use the title, and release-please derives the
  version/changelog from the type (see ADR 0006, `CONTRIBUTING.md`).
- Keep PRs small and self-contained: each commit should carry itself.

## Accuracy

- Any before/after or example in a commit message or PR body must be **run
  against the actual binary**, not written from memory. A plausible-looking but
  wrong example is worse than none.
- Prefer a checked-in approval input as the example's input
  (`approvals/*.input.yaml`): it's reproducible and its output is already pinned
  by the golden, so the example can't drift from reality. If none fits the
  behaviour you're showing, **add a fixture** (and its golden) rather than using
  a throwaway input — then the example and a test guard the same thing. The
  README's Usage examples work this way. (Exception: bytes hostile to fixture
  files — CRLF, trailing whitespace — stay in Go unit tests, per Testing.)
