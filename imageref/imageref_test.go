package imageref

import (
	"strings"
	"testing"
)

// One case per way a reference can be unreportable, rather than one per hostile
// byte — the parser treats them all the same, so byte variants would pin
// nothing extra. Hostile bytes live here and not in an approvals fixture: a
// checked-in .yaml would have its escapes and trailing whitespace normalised
// (see AGENTS.md).
func TestValidateRejects(t *testing.T) {
	for name, image := range map[string]string{
		"empty":                    "",
		"newline forges an entry":  "evil\nsecond-line",
		"escape sequence spoofs":   "nginx:1.0\x1b[2K\rregistry.io/trusted:safe",
		"whitespace splits args":   "a b c",
		"leading dash is a flag":   "--platform=linux/amd64",
		"empty tag":                "nginx:",
		"unrendered helm template": "{{.Values.image}}",
		"unexpanded variable":      "$IMAGE",
	} {
		t.Run(name, func(t *testing.T) {
			if err := Validate(image); err == nil {
				t.Errorf("Validate(%q) = nil, want an error", image)
			}
		})
	}
}

// Over-rejection is the failure that matters: a dropped image is
// indistinguishable from a manifest that had none. Every shape a registry could
// serve must pass.
func TestValidateAccepts(t *testing.T) {
	for _, image := range []string{
		"nginx",
		"nginx:1.28",
		"registry.k8s.io/nginx-slim:0.8",
		"kiwigrid/k8s-sidecar",
		"localhost:5000/kir:0.4.3",
		"quay.io/org/sub/repo:v1.2.3-rc.1",
		"registry.example.com:5000/team/app:2026-08-10_build.7",
		// Digest-pinned references are what break when go-digest cannot resolve
		// sha256; see the crypto/sha256 import in imageref.go.
		"ghcr.io/mpv/kir@sha256:66b7bf84cfc7c1b2d4a7d9848114270ce6049a04d5dab67d767d8ab5c0b3412a",
		"nginx:1.0@sha256:76944c9752702d324e06e2a9fa791c38f2b654ee4100c85327729eb3377a4284",
	} {
		t.Run(image, func(t *testing.T) {
			if err := Validate(image); err != nil {
				t.Errorf("Validate(%q) = %v, want nil", image, err)
			}
		})
	}
}

// The parser echoes the offending value back in its message, so reporting a
// rejection must not pass those bytes through to a terminal.
func TestValidateErrorEscapesTheValue(t *testing.T) {
	err := Validate("nginx:1.0\x1b[2K\rregistry.io/trusted:safe")
	if err == nil {
		t.Fatal("Validate() = nil, want an error")
	}
	if strings.ContainsAny(err.Error(), "\x1b\r\n") {
		t.Errorf("error message carries raw control bytes: %q", err.Error())
	}
}
