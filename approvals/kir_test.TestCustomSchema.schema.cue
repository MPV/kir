// A replacement for the embedded schema, teaching kir about containers that
// carry a field the Kubernetes API does not define.
//
// It extends the official type rather than restating it: corev1.#Container is
// embedded, and the extra field is declared alongside it. Passed with --schema,
// so no rebuild is needed.
package podspec

import corev1 "k8s.io/api/core/v1"

#Container: {
	corev1.#Container
	sidecarPolicy?: string
}

#Containers: [...#Container]

#EphemeralContainers: [...corev1.#EphemeralContainer]
