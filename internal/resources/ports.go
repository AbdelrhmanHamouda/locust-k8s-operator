/*
Copyright 2026.

Licensed under the Apache License, Version 2.0 (the "License");
you may not use this file except in compliance with the License.
You may obtain a copy of the License at

    http://www.apache.org/licenses/LICENSE-2.0

Unless required by applicable law or agreed to in writing, software
distributed under the License is distributed on an "AS IS" BASIS,
WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
See the License for the specific language governing permissions and
limitations under the License.
*/

package resources

import (
	corev1 "k8s.io/api/core/v1"
)

// MasterPorts returns the container ports for the master node.
// Ports: 5557 (master), 5558 (bind), 8089 (web UI)
func MasterPorts() []corev1.ContainerPort {
	return []corev1.ContainerPort{
		{ContainerPort: MasterPort},
		{ContainerPort: MasterBindPort},
		{ContainerPort: WebUIPort},
	}
}

// WorkerPorts returns the container ports for worker nodes.
// Ports: 8080 (worker)
func WorkerPorts() []corev1.ContainerPort {
	return []corev1.ContainerPort{
		{ContainerPort: WorkerPort},
	}
}

// MasterPortInts returns the master ports as int32 slice (for service creation).
func MasterPortInts() []int32 {
	return []int32{MasterPort, MasterBindPort, WebUIPort}
}

// WorkerPortInts returns the worker ports as int32 slice.
func WorkerPortInts() []int32 {
	return []int32{WorkerPort}
}
