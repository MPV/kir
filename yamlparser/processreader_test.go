package yamlparser

import (
	"strings"
	"testing"

	"github.com/mpv/kir/k8s"
)

// ProcessReader must collect images from every document in a stream and split
// documents robustly. A naive bytes.Split on "\n---\n" drops every document
// after the first when the separator isn't exactly that literal — a separator
// with trailing whitespace, or CRLF line endings. The leading-separator and
// no-trailing-newline cases are general correctness checks: a plain bytes.Split
// happens to handle those, but the YAML reader must too.
func TestProcessReader(t *testing.T) {
	pod := func(name, image string) string {
		return "apiVersion: v1\nkind: Pod\nmetadata:\n  name: " + name +
			"\nspec:\n  containers:\n  - name: c\n    image: " + image + "\n"
	}

	tests := []struct {
		name string
		data string
		want []string
	}{
		{
			name: "multiple documents",
			data: pod("one", "image-one") + "---\n" + pod("two", "image-two"),
			want: []string{"image-one", "image-two"},
		},
		{
			name: "leading separator",
			data: "---\n" + pod("one", "image-one"),
			want: []string{"image-one"},
		},
		{
			name: "separator with trailing whitespace",
			data: pod("one", "image-one") + "---   \n" + pod("two", "image-two"),
			want: []string{"image-one", "image-two"},
		},
		{
			name: "crlf line endings",
			data: strings.ReplaceAll(pod("one", "image-one")+"---\n"+pod("two", "image-two"), "\n", "\r\n"),
			want: []string{"image-one", "image-two"},
		},
		{
			name: "no trailing newline",
			data: strings.TrimRight(pod("one", "image-one"), "\n"),
			want: []string{"image-one"},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := ProcessReader(k8s.DefaultMatcher(), strings.NewReader(tt.data))
			if err != nil {
				t.Fatalf("ProcessReader() error = %v", err)
			}
			if len(got) != len(tt.want) {
				t.Fatalf("expected %d images, got %d: %v", len(tt.want), len(got), got)
			}
			for i, img := range got {
				if img != tt.want[i] {
					t.Errorf("image %d: expected %q, got %q", i, tt.want[i], img)
				}
			}
		})
	}
}
