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
		log.Fatal("Usage: oci-images-from-k8s-yaml <file_path> [<file_path_2> ...] or oci-images-from-k8s-yaml -")
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

	// Handle different types of Kubernetes objects
	switch gvk.Kind {
	case "Pod":
		pod, ok := obj.(*corev1.Pod)
		if !ok {
			log.Printf("error: not a Pod")
			return
		}
		printContainerImages(pod.Spec.Containers)
		printContainerImages(pod.Spec.InitContainers)
	case "Deployment":
		deployment, ok := obj.(*appsv1.Deployment)
		if !ok {
			log.Printf("error: not a Deployment")
			return
		}
		printContainerImages(deployment.Spec.Template.Spec.Containers)
		printContainerImages(deployment.Spec.Template.Spec.InitContainers)
	case "DaemonSet":
		daemonSet, ok := obj.(*appsv1.DaemonSet)
		if !ok {
			log.Printf("error: not a DaemonSet")
			return
		}
		printContainerImages(daemonSet.Spec.Template.Spec.Containers)
		printContainerImages(daemonSet.Spec.Template.Spec.InitContainers)
	case "ReplicaSet":
		replicaSet, ok := obj.(*appsv1.ReplicaSet)
		if !ok {
			log.Printf("error: not a ReplicaSet")
			return
		}
		printContainerImages(replicaSet.Spec.Template.Spec.Containers)
		printContainerImages(replicaSet.Spec.Template.Spec.InitContainers)
	case "StatefulSet":
		statefulSet, ok := obj.(*appsv1.StatefulSet)
		if !ok {
			log.Printf("error: not a StatefulSet")
			return
		}
		printContainerImages(statefulSet.Spec.Template.Spec.Containers)
		printContainerImages(statefulSet.Spec.Template.Spec.InitContainers)
	case "Job":
		job, ok := obj.(*batchv1.Job)
		if !ok {
			log.Printf("error: not a Job")
			return
		}
		printContainerImages(job.Spec.Template.Spec.Containers)
		printContainerImages(job.Spec.Template.Spec.InitContainers)
	case "CronJob":
		cronJob, ok := obj.(*batchv1.CronJob)
		if !ok {
			log.Printf("error: not a CronJob")
			return
		}
		printContainerImages(cronJob.Spec.JobTemplate.Spec.Template.Spec.Containers)
		printContainerImages(cronJob.Spec.JobTemplate.Spec.Template.Spec.InitContainers)
	case "List":
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
	default:
		log.Printf("error: unsupported kind %s", gvk.Kind)
		return
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
