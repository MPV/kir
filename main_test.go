package main

import (
	"bytes"
	"flag"
	"io/ioutil"
	"log"
	"os"
	"testing"

	appsv1 "k8s.io/api/apps/v1"
	corev1 "k8s.io/api/core/v1"
	v1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/runtime/serializer/json"
)

func TestMain(m *testing.M) {
	// Set up a temporary directory for test files
	dir, err := ioutil.TempDir("", "test")
	if err != nil {
		log.Fatalf("error: %v", err)
	}
	defer os.RemoveAll(dir)

	// Create test files

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
							Image: "test-image",
						},
					},
				},
			},
		},
	})

	// Change working directory to the temporary directory
	err = os.Chdir(dir)
	if err != nil {
		log.Fatalf("error: %v", err)
	}

	// Run tests
	os.Exit(m.Run())
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

func TestReadPodYAML(t *testing.T) {
	// Set the flag
	err := flag.Set("file", "pod.yaml")
	if err != nil {
		t.Fatalf("Error setting flag: %v", err)
	}
	flag.Parse()

	// Redirect stdout
	oldStdout := os.Stdout
	r, w, _ := os.Pipe()
	os.Stdout = w

	main()

	w.Close()
	os.Stdout = oldStdout

	var buf bytes.Buffer
	_, err = buf.ReadFrom(r)
	if err != nil {
		t.Fatalf("Error reading from buffer: %v", err)
	}
	output := buf.String()

	expected := "test-image\n"
	if output != expected {
		t.Errorf("expected %q, got %q", expected, output)
	}
}

func TestReadDeploymentYAML(t *testing.T) {
	// Set the flag
	err := flag.Set("file", "deployment.yaml")
	if err != nil {
		t.Fatalf("Error setting flag: %v", err)
	}
	flag.Parse()

	// Redirect stdout
	oldStdout := os.Stdout
	r, w, _ := os.Pipe()
	os.Stdout = w

	main()

	w.Close()
	os.Stdout = oldStdout

	var buf bytes.Buffer
	_, err = buf.ReadFrom(r)
	if err != nil {
		t.Fatalf("Error reading from buffer: %v", err)
	}
	output := buf.String()

	expected := "test-image\n"
	if output != expected {
		t.Errorf("expected %q, got %q", expected, output)
	}
}
