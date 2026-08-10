# 8. Best-effort processing; failures surface via the exit code

- Status: accepted
- Date: 2026-08-09

`kir` processes a whole batch of inputs (files, globs, a stdin stream). One bad
input must neither abort the batch nor pass silently: `kir` processes every
input, prints every image it finds, logs each failure to stderr, and returns a
non-zero exit code if any input failed. Successes are still printed.

An argument that matches no files (a typo, or a glob with no hits) is itself
such a failure — reported on stderr with a non-zero exit, not a silent no-op, so
a mistyped path can't masquerade as "no images found".

The same rule applies **within** an input, not only between them. A stream is a
batch of documents, so one document `kir` cannot process is reported, counted
against the exit code, and skipped — the documents around it are still read and
their images still printed. Every failure in the stream is reported, not just the
first. This matters most where a whole cluster arrives as one input
(`kubectl get pod -A -o yaml | kir -`): treating a single unparseable object as
fatal to the stream would discard every image beside it and leave "no images
found" as the only visible result.

This lets a pipeline (`kir manifests/* | xargs grype`) trust the exit code —
zero means every input was understood, non-zero means at least one wasn't —
without losing the images `kir` did find. Which inputs count as a failure vs. a
silent skip is ADR 0007.
