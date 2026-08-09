# Security policy

## Reporting a vulnerability

Report privately through GitHub:
**[Security → Advisories → Report a vulnerability](https://github.com/MPV/kir/security/advisories/new)**.

Please don't open a public issue for something exploitable — the advisory draft
is private until we publish it, and it gives us a place to work on a fix with
you.

Include what you'd need yourself: the manifest (or a reduced version of it) that
triggers the behaviour, the `kir` version from `kir --version`, and what you
expected instead. A reproducer we can run beats a description we have to
reconstruct.

`kir` is maintained by volunteers, so we can't promise a response time. We'll
acknowledge a report as soon as we see it and keep you updated as we work on it.

## What counts as a vulnerability

`kir` decides which images a scanner is asked about. An image it fails to
report isn't flagged as unscanned — it's never considered, and the scan comes
back clean. So the reports we most want are the ones where **`kir` under-reports
without saying so**:

- A manifest whose images `kir` omits from stdout while exiting 0.
- A document `kir` treats as image-less when it does describe containers.
- Anything that makes `kir` exit 0 on input it did not fully understand.

Also in scope:

- Output that misrepresents what will be scanned — an image reference that
  renders as something other than the value passed downstream, or that breaks
  the newline-separated contract stdout is meant to keep.
- Input that makes `kir` consume unbounded resources, or read files outside the
  paths it was given.
- Anything undermining the integrity of a published release: the signing
  pipeline, the release workflow, or the container image.

Out of scope: a malformed or unreadable input causing an `error:` on stderr and
a non-zero exit. That is the designed behaviour — see
[ADR 0008](docs/adr/0008-best-effort-processing-and-exit-codes.md).

## Supported versions

`kir` is pre-1.0 and releases roll forward. Fixes go into the next release from
`master`; there are no maintenance branches for older versions. Report against
the [latest release](https://github.com/MPV/kir/releases/latest) where you can.

## Verifying a release

Binaries and container images are signed keylessly with
[cosign](https://github.com/sigstore/cosign). `CONTRIBUTING.md` has the
[verification commands](CONTRIBUTING.md#verifying-a-release) — worth running
before you trust a downloaded artifact, and worth telling us about if they
fail.
