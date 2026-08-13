// The shape kir looks for — the official Kubernetes types, not a restatement
// of them.
//
// corev1 is generated from k8s.io/api/core/v1 by `cue get go` (see the Makefile
// target and k8s/cue.mod/gen/). Because those are CUE definitions, they are
// closed: a container carrying a field the Kubernetes API does not define is a
// mismatch, which is what stops a resource with a `containers` field holding
// something else from being read as a workload.
//
// A user schema passed to --schema replaces this file and can import the same
// package, so extending the official type is a two-line exercise rather than a
// transcription.
package podspec

import corev1 "k8s.io/api/core/v1"

#Containers: [...corev1.#Container]

#EphemeralContainers: [...corev1.#EphemeralContainer]
