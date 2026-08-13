build:
	go build -o bin/ ./...

test:
	go test ./... -v

# Regenerate the CUE definitions kir matches against from the Kubernetes Go
# types. Run after bumping k8s.io/api; the output is checked in.
#
# GOTOOLCHAIN is pinned because `go run pkg@version` builds the cue command with
# the Go version in *cue's* go.mod (1.25), while k8s.io/api now requires 1.26 —
# without it the generator fails with "package requires newer Go version".
schema:
	cd k8s && GOTOOLCHAIN=go1.26.5 go run cuelang.org/go/cmd/cue@v0.17.1 get go k8s.io/api/core/v1
