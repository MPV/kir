//go:build schema

package k8s

// kir's Go code does not import k8s.io/api — the CUE definitions under
// cue.mod/gen are generated from it by `make schema`, and CUE imports are
// invisible to the Go toolchain. This blank import keeps the module in go.mod
// anyway, so `go mod tidy` cannot drop the source the schema is generated from
// and a dependency bump moves the schema's source version with it.
import _ "k8s.io/api/core/v1"
