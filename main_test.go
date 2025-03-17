package main

import (
	"bytes"
	"log"
	"os"
	"testing"

	"github.com/mpv/kir/cmd"
	appsv1 "k8s.io/api/apps/v1"
	corev1 "k8s.io/api/core/v1"
	v1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/runtime/serializer/json"
)

func TestMainExecute(t *testing.T) {
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
	cmd.Execute(filePaths)

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
