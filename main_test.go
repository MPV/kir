package main

import (
	"bytes"
	"log"
	"os"
	"testing"

	appsv1 "k8s.io/api/apps/v1"
	corev1 "k8s.io/api/core/v1"
	v1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/runtime/serializer/json"
)

func TestReadPodYAML(t *testing.T) {
	dir := setupTestDir(t)
	defer os.RemoveAll(dir)

	createAsYamlFile(dir, "pod.yaml", &corev1.Pod{
		TypeMeta: v1.TypeMeta{
			APIVersion: "v1",
			Kind:       "Pod",
		},
		ObjectMeta: v1.ObjectMeta{
			Name: "test-pod",
		},
		Spec: corev1.PodSpec{
			Containers: []corev1.Container{
				{
					Name:  "test-container",
					Image: "test-image",
				},
			},
		},
	})

	testReadYAML(t, []string{"pod.yaml"}, "test-image\n")
}

func TestReadDeploymentYAML(t *testing.T) {
	dir := setupTestDir(t)
	defer os.RemoveAll(dir)

	createAsYamlFile(dir, "deployment.yaml", &appsv1.Deployment{
		TypeMeta: v1.TypeMeta{
			APIVersion: "apps/v1",
			Kind:       "Deployment",
		},
		ObjectMeta: v1.ObjectMeta{
			Name: "test-deployment",
		},
		Spec: appsv1.DeploymentSpec{
			Template: corev1.PodTemplateSpec{
				Spec: corev1.PodSpec{
					Containers: []corev1.Container{
						{
							Name:  "test-container",
							Image: "image1",
						},
						{
							Name:  "test-sidecar",
							Image: "sidecar-image2",
						},
					},
				},
			},
		},
	})

	testReadYAML(t, []string{"deployment.yaml"}, "image1\nsidecar-image2\n")
}

func TestReadYAMLFromFolder(t *testing.T) {
	dir := setupTestDir(t)
	defer os.RemoveAll(dir)

	createAsYamlFile(dir, "pod.yaml", &corev1.Pod{
		TypeMeta: v1.TypeMeta{
			APIVersion: "v1",
			Kind:       "Pod",
		},
		ObjectMeta: v1.ObjectMeta{
			Name: "test-pod",
		},
		Spec: corev1.PodSpec{
			Containers: []corev1.Container{
				{
					Name:  "test-container",
					Image: "test-image",
				},
			},
		},
	})

	createAsYamlFile(dir, "deployment.yaml", &appsv1.Deployment{
		TypeMeta: v1.TypeMeta{
			APIVersion: "apps/v1",
			Kind:       "Deployment",
		},
		ObjectMeta: v1.ObjectMeta{
			Name: "test-deployment",
		},
		Spec: appsv1.DeploymentSpec{
			Template: corev1.PodTemplateSpec{
				Spec: corev1.PodSpec{
					Containers: []corev1.Container{
						{
							Name:  "test-container",
							Image: "image1",
						},
						{
							Name:  "test-sidecar",
							Image: "sidecar-image2",
						},
					},
				},
			},
		},
	})

	testReadYAML(t, []string{dir}, "image1\nsidecar-image2\ntest-image\n")
}

func TestReadYAMLFromGlob(t *testing.T) {
	dir := setupTestDir(t)
	defer os.RemoveAll(dir)

	createAsYamlFile(dir, "pod1.yaml", &corev1.Pod{
		TypeMeta: v1.TypeMeta{
			APIVersion: "v1",
			Kind:       "Pod",
		},
		ObjectMeta: v1.ObjectMeta{
			Name: "test-pod1",
		},
		Spec: corev1.PodSpec{
			Containers: []corev1.Container{
				{
					Name:  "test-container1",
					Image: "test-image1",
				},
			},
		},
	})

	createAsYamlFile(dir, "pod2.yaml", &corev1.Pod{
		TypeMeta: v1.TypeMeta{
			APIVersion: "v1",
			Kind:       "Pod",
		},
		ObjectMeta: v1.ObjectMeta{
			Name: "test-pod2",
		},
		Spec: corev1.PodSpec{
			Containers: []corev1.Container{
				{
					Name:  "test-container2",
					Image: "test-image2",
				},
			},
		},
	})

	testReadYAML(t, []string{"pod1.yaml", "pod2.yaml"}, "test-image1\ntest-image2\n")
}

func TestReadYAMLFromStdin(t *testing.T) {
	dir := setupTestDir(t)
	defer os.RemoveAll(dir)

	yamlData := `
apiVersion: v1
kind: Pod
metadata:
  name: test-pod
spec:
  containers:
  - name: test-container
    image: test-image
`
	oldStdin := os.Stdin
	r, w, _ := os.Pipe()
	os.Stdin = r

	go func() {
		w.Write([]byte(yamlData))
		w.Close()
	}()

	testReadYAML(t, []string{"-"}, "test-image\n")

	os.Stdin = oldStdin
}

func setupTestDir(t *testing.T) string {
	dir, err := os.MkdirTemp("", "test")
	if err != nil {
		t.Fatalf("error: %v", err)
	}

	err = os.Chdir(dir)
	if err != nil {
		t.Fatalf("error: %v", err)
	}

	return dir
}

func testReadYAML(t *testing.T, filePaths []string, expectedOutput string) {
	// Redirect stdout
	oldStdout := os.Stdout
	r, w, _ := os.Pipe()
	os.Stdout = w

	// Pass the file paths as arguments
	os.Args = append([]string{"cmd"}, filePaths...)
	main()

	w.Close()
	os.Stdout = oldStdout

	var buf bytes.Buffer
	_, err := buf.ReadFrom(r)
	if err != nil {
		t.Fatalf("Error reading from buffer: %v", err)
	}
	output := buf.String()

	if output != expectedOutput {
		t.Errorf("expected %q, got %q", expectedOutput, output)
	}
}

func createAsYamlFile(dir, name string, obj runtime.Object) {
	scheme := runtime.NewScheme()
	appsv1.AddToScheme(scheme)
	serializer := json.NewSerializerWithOptions(json.DefaultMetaFactory, scheme, scheme, json.SerializerOptions{
		Yaml:   true,
		Pretty: true,
		Strict: true,
	})

	var buf bytes.Buffer
	err := serializer.Encode(obj, &buf)
	if err != nil {
		log.Fatalf("error: %v", err)
	}

	err = os.WriteFile(dir+"/"+name, buf.Bytes(), 0644)
	if err != nil {
		log.Fatalf("error: %v", err)
	}
}
