package k8s

import (
	"fmt"

	appsv1 "k8s.io/api/apps/v1"
	batchv1 "k8s.io/api/batch/v1"
	corev1 "k8s.io/api/core/v1"
)

// GetPodSpec extracts the PodSpec from a Kubernetes object
func GetPodSpec(obj interface{}) (*corev1.PodSpec, error) {
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

func GetContainerImages(containers []corev1.Container) []string {
	var images []string
	for _, container := range containers {
		images = append(images, container.Image)
	}
	return images
}

func GetContainersFromObject(obj interface{}) ([]corev1.Container, error) {
	podSpec, err := GetPodSpec(obj)
	if err != nil {
		return nil, err
	}

	var containers []corev1.Container
	containers = append(containers, podSpec.Containers...)
	containers = append(containers, podSpec.InitContainers...)
	return containers, nil
}
