package yamlparser

import (
	"bufio"
	"errors"
	"fmt"
	"io"

	"github.com/mpv/kir/k8s"
	utilyaml "k8s.io/apimachinery/pkg/util/yaml"
	"sigs.k8s.io/yaml"
)

// ProcessReader reads a (possibly multi-document) YAML stream and returns the
// container images of every workload it contains, using config to decide where each kind keeps its images. Documents are
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
func ProcessReader(config *k8s.Config, r io.Reader) ([]string, error) {
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
		imgs, err := ProcessData(config, doc)
		if err != nil {
			errs = append(errs, err)
			continue
		}
		images = append(images, imgs...)
	}
	return images, errors.Join(errs...)
}

// ProcessData returns the images in a single manifest document.
//
// The document is decoded into plain Go values rather than into a typed
// Kubernetes object, so a custom resource is reachable on the same terms as a
// built-in: whether it yields images depends only on whether the configuration
// describes it.
func ProcessData(config *k8s.Config, data []byte) ([]string, error) {
	var doc map[string]any
	if err := yaml.Unmarshal(data, &doc); err != nil {
		return nil, err
	}

	// Everything kir accepts is a Kubernetes object, and every Kubernetes
	// object has a kind. Requiring it keeps kir pointed at manifests instead of
	// mining arbitrary YAML (a Helm values.yaml, say) for anything image-like,
	// and it keeps an empty or malformed document an error rather than a silent
	// no-op — the unprocessable tier of
	// docs/adr/0007-document-classification.md, which ADR 0008 surfaces as a
	// non-zero exit.
	if _, ok := doc["kind"]; !ok {
		return nil, fmt.Errorf("Object 'Kind' is missing in %q", data)
	}

	// A document with nothing to report — a Service, a ConfigMap — is not an
	// error. See ADR 0007.
	return config.FindImages(doc), nil
}
