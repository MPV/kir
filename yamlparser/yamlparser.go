package yamlparser

import (
	"fmt"
	"slices"

	"github.com/mpv/kir/k8s"
	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
	"k8s.io/apimachinery/pkg/runtime/serializer"
	"k8s.io/client-go/kubernetes/scheme"
)

var supportedKinds = []string{"Pod", "Deployment", "DaemonSet", "ReplicaSet", "StatefulSet", "Job", "CronJob"}

func ProcessData(data []byte) ([]string, error) {
	// Decode the YAML file into a Kubernetes object
	decode := serializer.NewCodecFactory(scheme.Scheme).UniversalDeserializer().Decode
	obj, gvk, err := decode(data, nil, nil)
	if err != nil {
		return nil, err
	}

	var images []string

	// Check if the object has containers
	if containers, err := k8s.GetContainersFromObject(obj); err == nil {
		images = append(images, k8s.GetContainerImages(containers)...)
		return images, nil
	}

	// Handle List type separately
	if gvk.Kind == "List" {
		list, ok := obj.(*corev1.List)
		if !ok {
			return nil, fmt.Errorf("not a List")
		}
		for _, item := range list.Items {
			var unstructuredObj unstructured.Unstructured
			if err := unstructuredObj.UnmarshalJSON(item.Raw); err != nil {
				return nil, fmt.Errorf("error unmarshaling item: %v", err)
			}
			imgs, err := processUnstructured(unstructuredObj)
			if err != nil {
				return nil, fmt.Errorf("error processing unstructured item: %v", err)
			}
			images = append(images, imgs...)
		}
		return images, nil
	}

	return nil, fmt.Errorf("unsupported kind %s", gvk.Kind)
}

func processUnstructured(item unstructured.Unstructured) ([]string, error) {
	itemData, err := item.MarshalJSON()
	if err != nil {
		return nil, fmt.Errorf("error marshaling item: %v", err)
	}
	gvk := item.GroupVersionKind()
	if slices.Contains(supportedKinds, gvk.Kind) {
		images, err := ProcessData(itemData)
		if err != nil {
			return nil, fmt.Errorf("error processing data: %v", err)
		}
		return images, nil
	}
	return nil, fmt.Errorf("error: unsupported kind %s in List", gvk.Kind)
}
