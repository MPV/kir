package fileutil

import (
	"os"
	"path/filepath"
	"testing"
)

func TestFindFiles(t *testing.T) {
	// Setup temporary directory and files for testing
	tempDir := t.TempDir()
	file1 := filepath.Join(tempDir, "file1.yaml")
	file2 := filepath.Join(tempDir, "file2.yaml")
	file3 := filepath.Join(tempDir, "file3.yaml")
	os.WriteFile(file1, []byte("content1"), 0644)
	os.WriteFile(file2, []byte("content2"), 0644)
	os.WriteFile(file3, []byte("content3"), 0644)

	tests := []struct {
		name string
		args []string
		want []string
	}{
		{
			name: "Single file",
			args: []string{file1},
			want: []string{file1},
		},
		{
			name: "Directory",
			args: []string{tempDir},
			want: []string{file1, file2, file3},
		},
		{
			name: "Multiple files",
			args: []string{file1, file2},
			want: []string{file1, file2},
		},
		{
			name: "Glob pattern",
			args: []string{filepath.Join(tempDir, "*.yaml")},
			want: []string{file1, file2, file3},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := FindFiles(tt.args)
			if err != nil {
				t.Fatalf("FindFiles() error = %v", err)
			}
			if len(got) != len(tt.want) {
				t.Errorf("FindFiles() = %v, want %v", got, tt.want)
			}
			for _, file := range tt.want {
				found := false
				for _, g := range got {
					if g == file {
						found = true
						break
					}
				}
				if !found {
					t.Errorf("FindFiles() = %v, want %v", got, tt.want)
				}
			}
		})
	}
}
