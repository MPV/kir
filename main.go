package main

import (
	"bufio"
	"fmt"
	"io"
	"log"
	"os"
	"path/filepath"

	"slices"

	appsv1 "k8s.io/api/apps/v1"
	batchv1 "k8s.io/api/batch/v1"
	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
	"k8s.io/apimachinery/pkg/runtime/serializer"
	"k8s.io/client-go/kubernetes/scheme"
)

var supportedKinds = []string{"Pod", "Deployment", "DaemonSet", "ReplicaSet", "StatefulSet", "Job", "CronJob"}

func main() {
	if len(os.Args) < 2 {
		log.Fatal("Usage: kir <file_path> [<file_path_2> ...] or kir -")
		return
	}

	for i := 1; i < len(os.Args); i++ {
		filePath := os.Args[i]

		if filePath == "-" {
			processStdin()
		} else {
			fileInfo, err := os.Stat(filePath)
			if err != nil {
				log.Fatalf("error: %v", err)
			}

			if fileInfo.IsDir() {
				err := filepath.Walk(filePath, func(path string, info os.FileInfo, err error) error {
					if err != nil {
						return err
					}
					if !info.IsDir() {
						processFile(path)
					}
					return nil
				})
				if err != nil {
					log.Fatalf("error: %v", err)
				}
			} else {
				// Handle glob patterns
				files, err := filepath.Glob(filePath)
				if err != nil {
					log.Fatalf("error: %v", err)
				}

				for _, file := range files {
					processFile(file)
				}
			}
		}
	}
}

func processStdin() {
	reader := bufio.NewReader(os.Stdin)
	var data []byte
	for {
		line, err := reader.ReadBytes('\n')
		if err != nil && err != io.EOF {
			log.Fatalf("error reading stdin: %v", err)
		}
		data = append(data, line...)
		if err == io.EOF {
			break
		}
	}
	processData(data)
}

func processFile(filePath string) {
	// Read the YAML file
	data, err := os.ReadFile(filePath)
	if err != nil {
		log.Printf("error reading file %s: %v", filePath, err)
		return
	}
	processData(data)
}

func processData(data []byte) {
	// Decode the YAML file into a Kubernetes object
	decode := serializer.NewCodecFactory(scheme.Scheme).UniversalDeserializer().Decode
	obj, gvk, err := decode(data, nil, nil)
	if err != nil {
		log.Printf("error decoding data: %v", err)
		return
	}

	// Check if the object has a PodSpec
	if podSpec, err := getPodSpec(obj); err == nil {
		printContainerImages(podSpec.Containers)
		printContainerImages(podSpec.InitContainers)
		return
	}

	// Handle List type separately
	if gvk.Kind == "List" {
		list, ok := obj.(*corev1.List)
		if !ok {
			log.Printf("error: not a List")
			return
		}
		for _, item := range list.Items {
			var unstructuredObj unstructured.Unstructured
			if err := unstructuredObj.UnmarshalJSON(item.Raw); err != nil {
				log.Printf("error unmarshaling item: %v", err)
				continue
			}
			processUnstructured(unstructuredObj)
		}
		return
	}

	log.Printf("error: unsupported kind %s", gvk.Kind)
}

func getPodSpec(obj interface{}) (*corev1.PodSpec, error) {
	switch resource := obj.(type) {
	case *corev1.Pod:
		return &resource.Spec, nil
	case *appsv1.Deployment:
		return &resource.Spec.Template.Spec, nil
	case *appsv1.DaemonSet:
		return &resource.Spec.Template.Spec, nil
	case *appsv1.ReplicaSet:
		return &resource.Spec.Template.Spec, nil
	case *appsv1.StatefulSet:
		return &resource.Spec.Template.Spec, nil
	case *batchv1.Job:
		return &resource.Spec.Template.Spec, nil
	case *batchv1.CronJob:
		return &resource.Spec.JobTemplate.Spec.Template.Spec, nil
	default:
		return nil, fmt.Errorf("object does not have a PodSpec")
	}
}

// printContainerImages prints the container images
func printContainerImages(containers []corev1.Container) {
	for _, container := range containers {
		fmt.Println(container.Image)
	}
}

func processUnstructured(item unstructured.Unstructured) {
	itemData, err := item.MarshalJSON()
	if err != nil {
		log.Printf("error marshaling item: %v", err)
		return
	}
	gvk := item.GroupVersionKind()
	if slices.Contains(supportedKinds, gvk.Kind) {
		processData(itemData)
		return
	}
	log.Printf("error: unsupported kind %s in List", gvk.Kind)
}
