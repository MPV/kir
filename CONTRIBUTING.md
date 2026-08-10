# Contributing

## Development

```shell
make test    # go test ./... -v
make build   # go build -o bin/ ./...
go vet ./...
gofmt -l .   # should print nothing
```

CI runs `gofmt`, `go vet`, `go mod tidy` verification, and `go test -race`, so
run them locally before opening a pull request.

## Approving changed golden output

Behaviour is pinned by approval tests (ADR 0004/0005), so a change to what `kir`
prints shows up as failing goldens. A failing test leaves what it produced beside
its golden as `*.received.txt` and `make test` prints both files' contents; a
passing test deletes its own received file. When the new output is what you
meant:

```shell
make test      # prints received vs approved for each failure
make approve   # renames every *.received.txt over its *.approved.txt
```

`make approve` accepts every received file it finds, so run it after a full
`make test` — a filtered `go test -run ...` leaves the untouched tests' files
behind.

Read the output before approving. A golden diff *is* the behaviour change, so it
belongs in the pull request description; approving one to clear a red test hides
the very thing review needs to see.

## Architecture decisions

Design decisions are recorded as ADRs in [`docs/adr/`](docs/adr/). Read them
before changing how `kir` is structured or behaves — they explain why the
current design is the way it is. If a change revisits a decision, update or
supersede the relevant ADR in the same pull request; add a new ADR for a new
decision. This applies to human and AI contributors alike.

## Commit messages

This project uses [Conventional Commits](https://www.conventionalcommits.org).
The version and changelog are derived automatically from commit history, so the
prefix determines the release:

| Commit type                                          | Bump              |
| ---------------------------------------------------- | ----------------- |
| `fix:` / `perf:`                                     | patch             |
| `feat:`                                              | minor             |
| `feat!:` / any `BREAKING CHANGE:` footer             | minor (pre-1.0)   |
| `docs:` `ci:` `build:` `refactor:` `test:` `chore:`  | none              |

Whatever lands on `master` feeds the version and changelog, so use Conventional
Commit prefixes on your commits. Note that if a pull request is squash-merged the
**PR title** becomes the commit, so give it a conventional prefix too.

## Releasing

Versioning is driven by
[release-please](https://github.com/googleapis/release-please), with
[GoReleaser](https://goreleaser.com/) building the artifacts.

1. **release-please** watches `master`. As Conventional Commits land, it keeps a
   "release" pull request open that accumulates the next version and the
   `CHANGELOG.md`.

2. **Merging that release PR** creates the `vX.Y.Z` tag and a GitHub Release with
   the generated notes.

3. **GoReleaser** then builds and attaches the release artifacts: multi-platform
   binaries (`linux`/`darwin`/`windows`, `amd64`/`arm64`) with checksums and
   SBOMs, and a multi-arch container image at `ghcr.io/mpv/kir:X.Y.Z` (and
   `:latest`). Binaries and image are signed keylessly with
   [cosign](https://github.com/sigstore/cosign).

`kir --version` prints the embedded version, commit, and build date.

### Verifying a release

Binaries are signed keylessly; the signature is a Sigstore bundle
(`checksums.txt.sigstore.json`) covering `checksums.txt`:

```shell
cosign verify-blob \
  --bundle checksums.txt.sigstore.json \
  --certificate-identity-regexp 'https://github.com/mpv/kir/.*' \
  --certificate-oidc-issuer https://token.actions.githubusercontent.com \
  checksums.txt
```

Then check a downloaded binary's archive against `checksums.txt`. The container
image is signed too: `cosign verify ghcr.io/mpv/kir:X.Y.Z` (with the same
identity flags).

Dependency updates are managed by [Renovate](https://docs.renovatebot.com/),
which follows the same commit conventions, so routine bumps flow through the same
process.
