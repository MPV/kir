package main

import (
	"fmt"
	"log"
	"os"

	appsv1 "k8s.io/api/apps/v1"
	batchv1 "k8s.io/api/batch/v1"
	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/runtime/serializer"
	"k8s.io/client-go/kubernetes/scheme"
)

func main() {
	var filePath string

	if len(os.Args) < 2 {
		log.Fatal("Usage: oci-images-from-k8s-yaml <file_path>")
		return
	}

	filePath = os.Args[1]

	// Read the YAML file
	data, err := os.ReadFile(filePath)
	if err != nil {
		log.Fatalf("error: %v", err)
	}

	// Decode the YAML file into a Kubernetes object
	decode := serializer.NewCodecFactory(scheme.Scheme).UniversalDeserializer().Decode
	obj, gvk, err := decode(data, nil, nil)
	if err != nil {
		log.Fatalf("error: %v", err)
	}

	// Handle different types of Kubernetes objects
	switch gvk.Kind {
	case "Pod":
		pod, ok := obj.(*corev1.Pod)
		if !ok {
			log.Fatalf("error: not a Pod")
		}
		printContainerImages(pod.Spec.Containers)
		printContainerImages(pod.Spec.InitContainers)
	case "Deployment":
		deployment, ok := obj.(*appsv1.Deployment)
		if !ok {
			log.Fatalf("error: not a Deployment")
		}
		printContainerImages(deployment.Spec.Template.Spec.Containers)
		printContainerImages(deployment.Spec.Template.Spec.InitContainers)
	case "DaemonSet":
		daemonSet, ok := obj.(*appsv1.DaemonSet)
		if !ok {
			log.Fatalf("error: not a DaemonSet")
		}
		printContainerImages(daemonSet.Spec.Template.Spec.Containers)
		printContainerImages(daemonSet.Spec.Template.Spec.InitContainers)
	case "ReplicaSet":
		replicaSet, ok := obj.(*appsv1.ReplicaSet)
		if !ok {
			log.Fatalf("error: not a ReplicaSet")
		}
		printContainerImages(replicaSet.Spec.Template.Spec.Containers)
		printContainerImages(replicaSet.Spec.Template.Spec.InitContainers)
	case "StatefulSet":
		statefulSet, ok := obj.(*appsv1.StatefulSet)
		if !ok {
			log.Fatalf("error: not a StatefulSet")
		}
		printContainerImages(statefulSet.Spec.Template.Spec.Containers)
		printContainerImages(statefulSet.Spec.Template.Spec.InitContainers)
	case "Job":
		job, ok := obj.(*batchv1.Job)
		if !ok {
			log.Fatalf("error: not a Job")
		}
		printContainerImages(job.Spec.Template.Spec.Containers)
		printContainerImages(job.Spec.Template.Spec.InitContainers)
	case "CronJob":
		cronJob, ok := obj.(*batchv1.CronJob)
		if !ok {
			log.Fatalf("error: not a CronJob")
		}
		printContainerImages(cronJob.Spec.JobTemplate.Spec.Template.Spec.Containers)
		printContainerImages(cronJob.Spec.JobTemplate.Spec.Template.Spec.InitContainers)
	default:
		log.Fatalf("error: unsupported kind %s", gvk.Kind)
	}
}

// printContainerImages prints the container images
func printContainerImages(containers []corev1.Container) {
	for _, container := range containers {
		fmt.Println(container.Image)
	}
}
