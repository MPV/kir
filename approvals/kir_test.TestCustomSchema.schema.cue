// A replacement for the embedded schema, teaching kir about containers that
// carry a field the Kubernetes API does not define. Passed with --schema, so
// no rebuild is needed.
package podspec

_container: {
	name?:          string
	image?:         string
	sidecarPolicy?: string
}

#Container: close(_container)

#EphemeralContainer: close(_container & {
	targetContainerName?: string
})

#Containers: [...#Container]

#EphemeralContainers: [...#EphemeralContainer]
