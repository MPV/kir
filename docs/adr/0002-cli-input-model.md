# 2. Take manifest sources as positional arguments, with `-` for stdin

- Status: accepted
- Date: 2025-03-13 _(recorded retrospectively 2026-08-08)_

Manifest sources are positional arguments — each a file, a directory (expanded),
or a shell glob; a single `-` reads a stream from stdin. No input flags.

Composes with the shell (`kir manifests/*.yaml`, `kubectl get … -o yaml | kir -`)
and follows the Unix filter convention. `-` is an overloaded sentinel, so
argument resolution (`fileutil`) special-cases it against real paths.
