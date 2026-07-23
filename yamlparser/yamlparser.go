package yamlparser

import (
	"bufio"
	"fmt"
	"io"
	"slices"

	"github.com/mpv/kir/k8s"
	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/runtime/serializer"
	utilyaml "k8s.io/apimachinery/pkg/util/yaml"
	"k8s.io/client-go/kubernetes/scheme"
)

var supportedKinds = []string{"Pod", "Deployment", "DaemonSet", "ReplicaSet", "StatefulSet", "Job", "CronJob"}

// ProcessReader reads a (possibly multi-document) YAML stream and returns the
// container images of every supported workload it contains. Documents are
// separated using the Kubernetes YAML reader, which correctly handles leading
// and trailing "---" separators, separators followed by trailing whitespace,
// CRLF line endings, and a final document without a trailing newline.
func ProcessReader(r io.Reader) ([]string, error) {
	var images []string
	reader := utilyaml.NewYAMLReader(bufio.NewReader(r))
	for {
		doc, err := reader.Read()
		if err == io.EOF {
			break
		}
		if err != nil {
			return nil, fmt.Errorf("error reading YAML document: %v", err)
		}
		imgs, err := ProcessData(doc)
		if err != nil {
			return nil, err
		}
		images = append(images, imgs...)
	}
	return images, nil
}

func ProcessData(data []byte) ([]string, error) {
	// Decode the YAML file into a Kubernetes object
	decode := serializer.NewCodecFactory(scheme.Scheme).UniversalDeserializer().Decode
	obj, gvk, err := decode(data, nil, nil)
	if err != nil {
		// Kinds that aren't registered in the scheme (CRDs and other custom
		// resources) are not workloads we can inspect; skip them rather than
		// failing the whole stream.
		if runtime.IsNotRegisteredError(err) {
			return nil, nil
		}
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

	// Any other kind (Service, ConfigMap, ...) is not a workload; skip it so a
	// single non-workload document does not discard images from the rest of
	// the stream.
	return nil, nil
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
	// Non-workload items inside a List are skipped, mirroring how top-level
	// non-workload documents are handled.
	return nil, nil
}
