# 5. Expose the CLI as an in-process `Run(args, stdin, stdout, stderr) int`

- Status: accepted
- Date: 2026-08-02 _(recorded retrospectively 2026-08-08)_

The whole CLI is one function `Run(args []string, stdin io.Reader, stdout, stderr
io.Writer) int` — no `os.Exit`, no global streams. `main.go` only wires the real
`os` values and calls `os.Exit(Run(...))`; golden tests (ADR 0004) call `Run` and
capture the (stdout, stderr, exit code) triple.

The full CLI contract is exercised in-process; every path threads the four
parameters instead of reaching for globals.
