// The shape kir looks for, as data rather than as Go code.
//
// A node in a manifest is treated as a PodSpec when one of its container lists
// unifies with the schema below. #Container is a definition, so CUE closes it:
// a field that is not listed here is a mismatch, which is what stops a resource
// with a `containers` field holding something else from being mistaken for a
// workload.
//
// Field *values* are deliberately loose ({...}, [...]) — kir only needs to
// recognise a container, not validate one. A stricter schema can be generated
// from the Go types with `cue get go k8s.io/api/core/v1`, and either edited in
// place or passed to `kir --schema`.
package podspec

_container: {
	name?:                     string
	image?:                    string
	command?: [...string]
	args?: [...string]
	workingDir?: string
	ports?: [...{...}]
	env?: [...{...}]
	envFrom?: [...{...}]
	resources?: {...}
	resizePolicy?: [...{...}]
	restartPolicy?: string
	volumeMounts?: [...{...}]
	volumeDevices?: [...{...}]
	livenessProbe?: {...}
	readinessProbe?: {...}
	startupProbe?: {...}
	lifecycle?: {...}
	terminationMessagePath?:   string
	terminationMessagePolicy?: string
	imagePullPolicy?:          string
	securityContext?: {...}
	stdin?:     bool
	stdinOnce?: bool
	tty?:       bool
}

#Container: close(_container)

// An ephemeral container is a container plus the field naming its target.
#EphemeralContainer: close(_container & {
	targetContainerName?: string
})

#Containers: [...#Container]

#EphemeralContainers: [...#EphemeralContainer]
