package k8s

import (
	"fmt"
	"reflect"

	corev1 "k8s.io/api/core/v1"
)

// podSpecType is the type we are hunting for inside decoded objects.
var podSpecType = reflect.TypeOf(corev1.PodSpec{})

// maxDepth bounds the struct walk. The deepest PodSpec in the built-in API is
// CronJob's, at spec.jobTemplate.spec.template.spec — well inside this — so the
// limit only ever guards against a pathological or cyclic type.
const maxDepth = 20

// FindPodSpecs returns every corev1.PodSpec reachable from obj, in declaration
// order.
//
// Rather than enumerating kinds and their PodSpec paths, this walks the decoded
// Go value and matches on type. Any type registered in the client-go scheme is
// therefore understood for free — including ReplicationController and
// PodTemplate, which the previous hand-written type switch omitted — and adding
// a workload kind to k8s.io/api requires no change here.
//
// The trade-off is that this only sees types the scheme can decode: custom
// resources are still invisible. See docs/adr/0009-podspec-discovery.md.
func FindPodSpecs(obj any) []*corev1.PodSpec {
	var found []*corev1.PodSpec
	walk(reflect.ValueOf(obj), &found, 0)
	return found
}

func walk(v reflect.Value, found *[]*corev1.PodSpec, depth int) {
	if depth > maxDepth || !v.IsValid() {
		return
	}

	switch v.Kind() {
	case reflect.Pointer, reflect.Interface:
		if v.IsNil() {
			return
		}
		walk(v.Elem(), found, depth+1)

	case reflect.Struct:
		if v.Type() == podSpecType {
			*found = append(*found, podSpecPointer(v))
			return // a PodSpec never contains another PodSpec
		}
		for i := range v.NumField() {
			if !v.Type().Field(i).IsExported() {
				continue
			}
			walk(v.Field(i), found, depth+1)
		}

	case reflect.Slice, reflect.Array:
		for i := range v.Len() {
			walk(v.Index(i), found, depth+1)
		}
	}
}

// podSpecPointer returns v as a *corev1.PodSpec, copying only when v is not
// addressable (a slice or map element reached by value).
func podSpecPointer(v reflect.Value) *corev1.PodSpec {
	if v.CanAddr() {
		return v.Addr().Interface().(*corev1.PodSpec)
	}
	spec := v.Interface().(corev1.PodSpec)
	return &spec
}

// GetPodSpec returns the first PodSpec in obj, or an error when it has none.
func GetPodSpec(obj any) (*corev1.PodSpec, error) {
	specs := FindPodSpecs(obj)
	if len(specs) == 0 {
		return nil, fmt.Errorf("object does not have a PodSpec")
	}
	return specs[0], nil
}

func GetContainerImages(containers []corev1.Container) []string {
	var images []string
	for _, container := range containers {
		images = append(images, container.Image)
	}
	return images
}

// GetContainersFromObject returns the containers of every PodSpec in obj.
func GetContainersFromObject(obj any) ([]corev1.Container, error) {
	specs := FindPodSpecs(obj)
	if len(specs) == 0 {
		return nil, fmt.Errorf("object does not have a PodSpec")
	}

	var containers []corev1.Container
	for _, podSpec := range specs {
		containers = append(containers, podSpec.Containers...)
		containers = append(containers, podSpec.InitContainers...)
		for _, ec := range podSpec.EphemeralContainers {
			containers = append(containers, corev1.Container(ec.EphemeralContainerCommon))
		}
	}
	return containers, nil
}
