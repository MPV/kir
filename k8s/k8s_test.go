package k8s

import (
	"testing"

	appsv1 "k8s.io/api/apps/v1"
	batchv1 "k8s.io/api/batch/v1"
	corev1 "k8s.io/api/core/v1"
)

// Test that GetPodSpec works for the kinds that have a PodSpec:
func TestGetPodSpec(t *testing.T) {
	tests := []struct {
		name    string
		obj     interface{}
		wantErr bool
	}{
		{"Pod", &corev1.Pod{}, false},
		{"Deployment", &appsv1.Deployment{}, false},
		{"DaemonSet", &appsv1.DaemonSet{}, false},
		{"ReplicaSet", &appsv1.ReplicaSet{}, false},
		{"StatefulSet", &appsv1.StatefulSet{}, false},
		{"Job", &batchv1.Job{}, false},
		{"CronJob", &batchv1.CronJob{}, false},
		{"Invalid", "invalid", true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			_, err := GetPodSpec(tt.obj)
			if (err != nil) != tt.wantErr {
				t.Errorf("GetPodSpec() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

// Test that GetPodSpec returns the correct PodSpec:
func TestGetPodSpecPod(t *testing.T) {
	pod := &corev1.Pod{
		Spec: corev1.PodSpec{
			Containers: []corev1.Container{
				{
					Name:  "test-container",
					Image: "test-image",
				},
			},
		},
	}

	spec, err := GetPodSpec(pod)
	if err != nil {
		t.Fatalf("GetPodSpec() error = %v", err)
	}

	if len(spec.Containers) != 1 {
		t.Fatalf("expected 1 container, got %d", len(spec.Containers))
	}

	if spec.Containers[0].Image != "test-image" {
		t.Errorf("expected image %q, got %q", "test-image", spec.Containers[0].Image)
	}
}

// Test that GetPodSpec returns the correct PodSpec for all supported kinds:
func TestGetPodSpecSupported(t *testing.T) {
	commonPodSpec := corev1.PodSpec{
		Containers: []corev1.Container{
			{
				Name:  "test-container",
				Image: "test-image",
			},
		},
	}

	tests := []struct {
		name string
		obj  interface{}
	}{
		{"Deployment", &appsv1.Deployment{
			Spec: appsv1.DeploymentSpec{
				Template: corev1.PodTemplateSpec{
					Spec: commonPodSpec,
				},
			},
		}},
		{"DaemonSet", &appsv1.DaemonSet{
			Spec: appsv1.DaemonSetSpec{
				Template: corev1.PodTemplateSpec{
					Spec: commonPodSpec,
				},
			},
		}},
		{"ReplicaSet", &appsv1.ReplicaSet{
			Spec: appsv1.ReplicaSetSpec{
				Template: corev1.PodTemplateSpec{
					Spec: commonPodSpec,
				},
			},
		}},
		{"StatefulSet", &appsv1.StatefulSet{
			Spec: appsv1.StatefulSetSpec{
				Template: corev1.PodTemplateSpec{
					Spec: commonPodSpec,
				},
			},
		}},
		{"Job", &batchv1.Job{
			Spec: batchv1.JobSpec{
				Template: corev1.PodTemplateSpec{
					Spec: commonPodSpec,
				},
			},
		}},
		{"CronJob", &batchv1.CronJob{
			Spec: batchv1.CronJobSpec{
				JobTemplate: batchv1.JobTemplateSpec{
					Spec: batchv1.JobSpec{
						Template: corev1.PodTemplateSpec{
							Spec: commonPodSpec,
						},
					},
				},
			},
		}},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			spec, err := GetPodSpec(tt.obj)
			if err != nil {
				t.Fatalf("GetPodSpec() error = %v", err)
			}

			if len(spec.Containers) != 1 {
				t.Fatalf("expected 1 container, got %d", len(spec.Containers))
			}

			if spec.Containers[0].Image != "test-image" {
				t.Errorf("expected image %q, got %q", "test-image", spec.Containers[0].Image)
			}
		})
	}
}

// Test that GetPodSpec fails for an object that does not have a PodSpec:
func TestGetPodSpecInvalid(t *testing.T) {
	_, err := GetPodSpec("invalid")

	// Assert correct error message ()"object does not have a PodSpec"):
	if err == nil || err.Error() != "object does not have a PodSpec" {
		t.Fatalf("GetPodSpec() error = %v, want %q", err, "object does not have a PodSpec")
	}
}

func TestGetContainerImages(t *testing.T) {
	containers := []corev1.Container{
		{Name: "container1", Image: "image1"},
		{Name: "container2", Image: "image2"},
	}

	expected := []string{"image1", "image2"}
	images := GetContainerImages(containers)

	if len(images) != len(expected) {
		t.Fatalf("expected %d images, got %d", len(expected), len(images))
	}

	for i, img := range images {
		if img != expected[i] {
			t.Errorf("expected image %q, got %q", expected[i], img)
		}
	}
}

func TestGetContainersFromObject(t *testing.T) {
	tests := []struct {
		name    string
		obj     interface{}
		want    []corev1.Container
		wantErr bool
	}{
		{
			name: "Pod",
			obj: &corev1.Pod{
				Spec: corev1.PodSpec{
					Containers: []corev1.Container{
						{Name: "container1", Image: "image1"},
					},
					InitContainers: []corev1.Container{
						{Name: "init-container1", Image: "init-image1"},
					},
				},
			},
			want: []corev1.Container{
				{Name: "container1", Image: "image1"},
				{Name: "init-container1", Image: "init-image1"},
			},
			wantErr: false,
		},
		{
			name: "Deployment",
			obj: &appsv1.Deployment{
				Spec: appsv1.DeploymentSpec{
					Template: corev1.PodTemplateSpec{
						Spec: corev1.PodSpec{
							Containers: []corev1.Container{
								{Name: "container1", Image: "image1"},
							},
							InitContainers: []corev1.Container{
								{Name: "init-container1", Image: "init-image1"},
							},
						},
					},
				},
			},
			want: []corev1.Container{
				{Name: "container1", Image: "image1"},
				{Name: "init-container1", Image: "init-image1"},
			},
			wantErr: false,
		},
		{
			name:    "Invalid",
			obj:     "invalid",
			want:    nil,
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := GetContainersFromObject(tt.obj)
			if (err != nil) != tt.wantErr {
				t.Errorf("GetContainersFromObject() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if len(got) != len(tt.want) {
				t.Errorf("expected %d containers, got %d", len(tt.want), len(got))
			}
			for i, container := range got {
				if container.Name != tt.want[i].Name || container.Image != tt.want[i].Image {
					t.Errorf("expected container %v, got %v", tt.want[i], container)
				}
			}
		})
	}
}
