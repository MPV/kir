package yamlparser

import (
	"strings"
	"testing"
)

// ProcessReader must collect images from every document in a stream and split
// documents robustly. These inputs are ones a naive bytes.Split on "\n---\n"
// mishandled: a leading separator, a separator with trailing whitespace, CRLF
// line endings, and a final document with no trailing newline.
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
			got, err := ProcessReader(strings.NewReader(tt.data))
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
