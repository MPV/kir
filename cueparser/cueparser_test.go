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
			name: "Valid Pod with containers",
			data: []byte(`
apiVersion: v1
kind: Pod
metadata:
  name: test-pod
spec:
  containers:
  - name: test-container
    image: test-image
    command: ["echo", "hello"]
    ports:
    - containerPort: 8080
`),
			want:    []string{"test-image"},
			wantErr: false,
		},
		{
			name: "Valid Pod with initContainers",
			data: []byte(`
apiVersion: v1
kind: Pod
metadata:
  name: test-pod
spec:
  initContainers:
  - name: init-container
    image: init-image
    command: ["echo", "init"]
`),
			want:    []string{"init-image"},
			wantErr: false,
		},
		{
			name: "Valid Pod with both containers and initContainers",
			data: []byte(`
apiVersion: v1
kind: Pod
metadata:
  name: test-pod
spec:
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
			name: "Invalid Pod (missing required fields)",
			data: []byte(`
apiVersion: v1
kind: Pod
spec:
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
			name: "Multiple Pods",
			data: []byte(`
apiVersion: v1
kind: Pod
metadata:
  name: pod1
spec:
  containers:
  - name: container1
    image: image1
---
apiVersion: v1
kind: Pod
metadata:
  name: pod2
spec:
  containers:
  - name: container2
    image: image2
`),
			want:    []string{"image1", "image2"},
			wantErr: false,
		},
		{
			name: "Mixed valid and invalid Pods",
			data: []byte(`
apiVersion: v1
kind: Pod
metadata:
  name: pod1
spec:
  containers:
  - name: container1
    image: image1
---
apiVersion: v1
kind: Pod
spec:
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
