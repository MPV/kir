package cueparser

// PodSpec is a simplified schema for Kubernetes PodSpec
#PodSpec: {
	containers?: [...#Container]
	initContainers?: [...#Container]
}

// Container represents a container in a pod
#Container: {
	name: string
	image: string
	// Add other fields as needed
}