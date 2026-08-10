// Package imageref decides whether a container image reference is safe for kir
// to report. See docs/adr/0007-document-classification.md for why kir refuses
// some values outright.
package imageref

import (
	// go-digest resolves sha256 only when that hash is linked in, and rejects
	// every digest-pinned reference without it. kir links it anyway today via
	// client-go, so this changes nothing now; it keeps imageref standing on its
	// own for when #26's candidates drop client-go. No test here can pin it —
	// the test binary links the hash regardless.
	_ "crypto/sha256"
	"fmt"
	"strconv"

	"github.com/distribution/reference"
)

// Validate returns an error describing why image cannot be reported, or nil if
// it can. The canonical parser decides, so kir accepts exactly what a registry
// client would. The parsed value is discarded: kir reports what the manifest
// said, so "nginx" stays "nginx" rather than becoming "docker.io/library/nginx".
func Validate(image string) error {
	if _, err := reference.ParseNormalizedNamed(image); err != nil {
		// Quoted, not interpolated: the parser echoes the offending value back,
		// and an escape sequence in it must not reach the terminal this is
		// reported on.
		return fmt.Errorf("invalid image reference %q: %s", image, strconv.Quote(err.Error()))
	}
	return nil
}
