package yamlparser

import (
	"bytes"
	"fmt"
	"slices"

	"github.com/mpv/kir/cueparser"
	"github.com/mpv/kir/k8s"
	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
	"k8s.io/apimachinery/pkg/runtime/serializer"
	"k8s.io/client-go/kubernetes/scheme"
)

var supportedKinds = []string{"Pod", "Deployment", "DaemonSet", "ReplicaSet", "StatefulSet", "Job", "CronJob"}

// ProcessData processes YAML data and extracts container images
func ProcessData(data []byte) ([]string, error) {
	// First, try to use the CUE-based parser
	cueImages, err := cueparser.ProcessData(data)
	if err == nil && len(cueImages) > 0 {
		return cueImages, nil
	}

	// If CUE parsing fails, fall back to the original implementation
	// Decode the YAML file into a Kubernetes object
	decode := serializer.NewCodecFactory(scheme.Scheme).UniversalDeserializer().Decode
	obj, gvk, err := decode(data, nil, nil)
	if err != nil {
		return nil, err
	}

	var k8sImages []string

	// Check if the object has containers
	if containers, err := k8s.GetContainersFromObject(obj); err == nil {
		k8sImages = append(k8sImages, k8s.GetContainerImages(containers)...)
		return k8sImages, nil
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
			k8sImages = append(k8sImages, imgs...)
		}
		return k8sImages, nil
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

// ProcessKubernetesListYAML processes a Kubernetes List YAML document and extracts container images
func ProcessKubernetesListYAML(data []byte) ([]string, error) {
	// First, try to use the CUE-based parser
	cueImages, err := cueparser.ProcessKubernetesListYAML(data)
	if err == nil && len(cueImages) > 0 {
		return cueImages, nil
	}

	// If CUE parsing fails, fall back to the original implementation
	var k8sImages []string
	docs := bytes.Split(data, []byte("\n---\n"))
	for _, doc := range docs {
		imgs, err := ProcessData(doc)
		if err != nil {
			// Log the error but continue processing other documents
			fmt.Printf("Error processing document: %v\n", err)
			continue
		}
		k8sImages = append(k8sImages, imgs...)
	}
	return k8sImages, nil
}
