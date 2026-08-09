package yamlparser

import (
	"bufio"
	"errors"
	"fmt"
	"io"

	"github.com/mpv/kir/k8s"
	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/runtime/serializer"
	utilyaml "k8s.io/apimachinery/pkg/util/yaml"
	"k8s.io/client-go/kubernetes/scheme"
)

// ProcessReader reads a (possibly multi-document) YAML stream and returns the
// container images of every supported workload it contains. Documents are
// separated using the Kubernetes YAML reader, which correctly handles leading
// and trailing "---" separators, separators followed by trailing whitespace,
// CRLF line endings, and a final document without a trailing newline.
//
// A document that cannot be processed does not discard the stream. Its failure
// is collected, the documents after it are still read, and whatever images were
// found are returned alongside the joined errors. ADR 0008 has kir print every
// image it finds and surface failures through the exit code; that has to hold
// within a stream as well as between inputs, or one unparseable document in a
// cluster dump costs every image around it.
func ProcessReader(r io.Reader) ([]string, error) {
	var images []string
	var errs []error
	reader := utilyaml.NewYAMLReader(bufio.NewReader(r))
	for {
		doc, err := reader.Read()
		if err == io.EOF {
			break
		}
		if err != nil {
			// The stream can no longer be split into documents, so there is
			// nothing further to read — but what was already found still counts.
			errs = append(errs, fmt.Errorf("error reading YAML document: %v", err))
			break
		}
		imgs, err := ProcessData(doc)
		if err != nil {
			errs = append(errs, err)
			continue
		}
		images = append(images, imgs...)
	}
	return images, errors.Join(errs...)
}

func ProcessData(data []byte) ([]string, error) {
	// Decode the YAML file into a Kubernetes object
	decode := serializer.NewCodecFactory(scheme.Scheme).UniversalDeserializer().Decode
	obj, gvk, err := decode(data, nil, nil)
	if err != nil {
		// Kinds that aren't registered in the scheme (CRDs and other custom
		// resources) are skipped rather than failing the whole stream. Some of
		// them may embed a PodSpec we could inspect; surfacing those ("seen but
		// not detected") is tracked in #75. For now they are skipped silently,
		// like any other non-workload document — see
		// docs/adr/0007-document-classification.md.
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

	// Any other kind (Service, ConfigMap, ...) is a valid object with no images
	// to report, not an error; skip it silently so a single non-workload
	// document does not discard images from the rest of the stream. See
	// docs/adr/0007-document-classification.md.
	return nil, nil
}

// processUnstructured handles one item of a List by feeding it back through
// ProcessData, which already skips non-workload and unregistered kinds. There
// is no kind allow-list to consult: whether an item yields images is decided by
// whether its decoded object contains a PodSpec.
func processUnstructured(item unstructured.Unstructured) ([]string, error) {
	itemData, err := item.MarshalJSON()
	if err != nil {
		return nil, fmt.Errorf("error marshaling item: %v", err)
	}
	return ProcessData(itemData)
}
