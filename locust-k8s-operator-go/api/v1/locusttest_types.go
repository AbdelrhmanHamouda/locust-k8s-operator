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

package v1

import (
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

// PodLabels defines labels for master and worker pods.
type PodLabels struct {
	// Master defines labels attached to the master pod.
	// +optional
	Master map[string]string `json:"master,omitempty"`

	// Worker defines labels attached to worker pods.
	// +optional
	Worker map[string]string `json:"worker,omitempty"`
}

// PodAnnotations defines annotations for master and worker pods.
type PodAnnotations struct {
	// Master defines annotations attached to the master pod.
	// +optional
	Master map[string]string `json:"master,omitempty"`

	// Worker defines annotations attached to worker pods.
	// +optional
	Worker map[string]string `json:"worker,omitempty"`
}

// LocustTestAffinity defines affinity rules for pod scheduling.
type LocustTestAffinity struct {
	// NodeAffinity defines node affinity rules.
	// +optional
	NodeAffinity *LocustTestNodeAffinity `json:"nodeAffinity,omitempty"`
}

// LocustTestNodeAffinity defines node affinity configuration.
type LocustTestNodeAffinity struct {
	// RequiredDuringSchedulingIgnoredDuringExecution defines required node affinity rules.
	// The map keys are label keys and values are label values that nodes must have.
	// +optional
	RequiredDuringSchedulingIgnoredDuringExecution map[string]string `json:"requiredDuringSchedulingIgnoredDuringExecution,omitempty"`
}

// LocustTestToleration defines a toleration for pod scheduling.
type LocustTestToleration struct {
	// Key is the taint key that the toleration applies to.
	// +kubebuilder:validation:Required
	Key string `json:"key"`

	// Operator represents the relationship between the key and value.
	// +kubebuilder:validation:Required
	// +kubebuilder:validation:Enum=Exists;Equal
	Operator string `json:"operator"`

	// Value is the taint value the toleration matches to.
	// +optional
	Value string `json:"value,omitempty"`

	// Effect indicates the taint effect to match.
	// +kubebuilder:validation:Required
	// +kubebuilder:validation:Enum=NoSchedule;PreferNoSchedule;NoExecute
	Effect string `json:"effect"`
}

// LocustTestSpec defines the desired state of LocustTest.
type LocustTestSpec struct {
	// MasterCommandSeed is the command seed for the master pod.
	// This forms the base of the locust master command.
	// +kubebuilder:validation:Required
	MasterCommandSeed string `json:"masterCommandSeed"`

	// WorkerCommandSeed is the command seed for worker pods.
	// This forms the base of the locust worker command.
	// +kubebuilder:validation:Required
	WorkerCommandSeed string `json:"workerCommandSeed"`

	// WorkerReplicas is the number of worker pods to spawn.
	// +kubebuilder:validation:Required
	// +kubebuilder:validation:Minimum=1
	// +kubebuilder:validation:Maximum=500
	// +kubebuilder:default=1
	WorkerReplicas int32 `json:"workerReplicas"`

	// Image is the Locust container image to use.
	// +kubebuilder:validation:Required
	Image string `json:"image"`

	// ImagePullPolicy defines when to pull the image.
	// +kubebuilder:validation:Enum=Always;IfNotPresent;Never
	// +optional
	ImagePullPolicy string `json:"imagePullPolicy,omitempty"`

	// ImagePullSecrets is a list of secret names for pulling images from private registries.
	// +optional
	ImagePullSecrets []string `json:"imagePullSecrets,omitempty"`

	// ConfigMap is the name of the ConfigMap containing the test file(s).
	// +optional
	ConfigMap string `json:"configMap,omitempty"`

	// LibConfigMap is the name of the ConfigMap containing lib directory files.
	// +optional
	LibConfigMap string `json:"libConfigMap,omitempty"`

	// Labels defines labels to attach to deployed pods.
	// +optional
	Labels *PodLabels `json:"labels,omitempty"`

	// Annotations defines annotations to attach to deployed pods.
	// +optional
	Annotations *PodAnnotations `json:"annotations,omitempty"`

	// Affinity defines affinity rules for pod scheduling.
	// +optional
	Affinity *LocustTestAffinity `json:"affinity,omitempty"`

	// Tolerations defines tolerations for pod scheduling.
	// +optional
	Tolerations []LocustTestToleration `json:"tolerations,omitempty"`
}

// LocustTestStatus defines the observed state of LocustTest.
type LocustTestStatus struct {
	// INSERT ADDITIONAL STATUS FIELD - define observed state of cluster
	// Important: Run "make" to regenerate code after modifying this file
}

// +kubebuilder:object:root=true
// +kubebuilder:subresource:status
// +kubebuilder:resource:shortName=lotest
// +kubebuilder:deprecatedversion:warning="locust.io/v1 LocustTest is deprecated, migrate to locust.io/v2"
// +kubebuilder:printcolumn:name="master_cmd",type=string,JSONPath=`.spec.masterCommandSeed`,description="Master pod command seed"
// +kubebuilder:printcolumn:name="worker_replica_count",type=integer,JSONPath=`.spec.workerReplicas`,description="Number of requested worker pods"
// +kubebuilder:printcolumn:name="Image",type=string,JSONPath=`.spec.image`,description="Locust image"
// +kubebuilder:printcolumn:name="Age",type=date,JSONPath=`.metadata.creationTimestamp`

// LocustTest is the Schema for the locusttests API (v1 - DEPRECATED).
type LocustTest struct {
	metav1.TypeMeta   `json:",inline"`
	metav1.ObjectMeta `json:"metadata,omitempty"`

	Spec   LocustTestSpec   `json:"spec,omitempty"`
	Status LocustTestStatus `json:"status,omitempty"`
}

// +kubebuilder:object:root=true

// LocustTestList contains a list of LocustTest.
type LocustTestList struct {
	metav1.TypeMeta `json:",inline"`
	metav1.ListMeta `json:"metadata,omitempty"`
	Items           []LocustTest `json:"items"`
}

func init() {
	SchemeBuilder.Register(&LocustTest{}, &LocustTestList{})
}
