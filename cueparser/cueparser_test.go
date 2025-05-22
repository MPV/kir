package cueparser

import (
	"testing"
)

func TestProcessData(t *testing.T) {
	tests := []struct {
		name    string
		data    []byte
		want    []string
		wantErr bool
	}{
		{
			name: "Valid PodSpec with containers",
			data: []byte(`
containers:
  - name: test-container
    image: test-image
`),
			want:    []string{"test-image"},
			wantErr: false,
		},
		{
			name: "Valid PodSpec with initContainers",
			data: []byte(`
initContainers:
  - name: init-container
    image: init-image
`),
			want:    []string{"init-image"},
			wantErr: false,
		},
		{
			name: "Valid PodSpec with both containers and initContainers",
			data: []byte(`
containers:
  - name: test-container
    image: test-image
initContainers:
  - name: init-container
    image: init-image
`),
			want:    []string{"test-image", "init-image"},
			wantErr: false,
		},
		{
			name: "Invalid PodSpec (missing required fields)",
			data: []byte(`
containers:
  - image: test-image
`),
			want:    nil,
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := ProcessData(tt.data)
			if (err != nil) != tt.wantErr {
				t.Errorf("ProcessData() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !tt.wantErr {
				if len(got) != len(tt.want) {
					t.Errorf("ProcessData() got %v images, want %v", len(got), len(tt.want))
					return
				}
				for i, img := range got {
					if img != tt.want[i] {
						t.Errorf("ProcessData() got image %v, want %v", img, tt.want[i])
					}
				}
			}
		})
	}
}

func TestProcessKubernetesListYAML(t *testing.T) {
	tests := []struct {
		name    string
		data    []byte
		want    []string
		wantErr bool
	}{
		{
			name: "Multiple PodSpecs",
			data: []byte(`
---
containers:
  - name: container1
    image: image1
---
containers:
  - name: container2
    image: image2
`),
			want:    []string{"image1", "image2"},
			wantErr: false,
		},
		{
			name: "Mixed valid and invalid PodSpecs",
			data: []byte(`
---
containers:
  - name: container1
    image: image1
---
containers:
  - image: image2
`),
			want:    []string{"image1"},
			wantErr: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := ProcessKubernetesListYAML(tt.data)
			if (err != nil) != tt.wantErr {
				t.Errorf("ProcessKubernetesListYAML() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !tt.wantErr {
				if len(got) != len(tt.want) {
					t.Errorf("ProcessKubernetesListYAML() got %v images, want %v", len(got), len(tt.want))
					return
				}
				for i, img := range got {
					if img != tt.want[i] {
						t.Errorf("ProcessKubernetesListYAML() got image %v, want %v", img, tt.want[i])
					}
				}
			}
		})
	}
}
