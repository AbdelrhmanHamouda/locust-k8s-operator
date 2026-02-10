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

package v2

import (
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

// ============================================
// MASTER CONFIGURATION
// ============================================

// MasterSpec defines master node configuration.
type MasterSpec struct {
	// Command is the base command for the master node.
	// The operator appends: --master --master-port=5557 --expect-workers=N
	// +kubebuilder:validation:Required
	Command string `json:"command"`

	// Resources defines resource requests and limits for the master pod.
	// +optional
	Resources corev1.ResourceRequirements `json:"resources,omitempty"`

	// Labels for the master pod.
	// +optional
	Labels map[string]string `json:"labels,omitempty"`

	// Annotations for the master pod.
	// +optional
	Annotations map[string]string `json:"annotations,omitempty"`

	// Autostart enables the --autostart flag to start the test automatically.
	// +optional
	// +kubebuilder:default=true
	Autostart *bool `json:"autostart,omitempty"`

	// Autoquit configuration for automatic test termination.
	// +optional
	Autoquit *AutoquitConfig `json:"autoquit,omitempty"`

	// ExtraArgs are additional CLI arguments appended to the command.
	// +optional
	ExtraArgs []string `json:"extraArgs,omitempty"`
}

// AutoquitConfig defines autoquit behavior for the master.
type AutoquitConfig struct {
	// Enabled enables the --autoquit flag.
	// +kubebuilder:default=true
	Enabled bool `json:"enabled"`

	// Timeout in seconds after test completion before quitting.
	// +optional
	// +kubebuilder:validation:Minimum=0
	// +kubebuilder:default=60
	Timeout int32 `json:"timeout,omitempty"`
}

// ============================================
// WORKER CONFIGURATION
// ============================================

// WorkerSpec defines worker node configuration.
type WorkerSpec struct {
	// Command is the base command for worker nodes.
	// The operator appends: --worker --master-host=<service> --master-port=5557
	// +kubebuilder:validation:Required
	Command string `json:"command"`

	// Replicas is the number of worker pods to create.
	// +kubebuilder:validation:Required
	// +kubebuilder:validation:Minimum=1
	// +kubebuilder:validation:Maximum=500
	Replicas int32 `json:"replicas"`

	// Resources defines resource requests and limits for worker pods.
	// +optional
	Resources corev1.ResourceRequirements `json:"resources,omitempty"`

	// Labels for worker pods.
	// +optional
	Labels map[string]string `json:"labels,omitempty"`

	// Annotations for worker pods.
	// +optional
	Annotations map[string]string `json:"annotations,omitempty"`

	// ExtraArgs are additional CLI arguments appended to the command.
	// +optional
	ExtraArgs []string `json:"extraArgs,omitempty"`
}

// ============================================
// TEST FILES CONFIGURATION
// ============================================

// TestFilesConfig defines test file mounting configuration.
type TestFilesConfig struct {
	// ConfigMapRef is the name of the ConfigMap containing locustfile(s).
	// +optional
	ConfigMapRef string `json:"configMapRef,omitempty"`

	// LibConfigMapRef is the name of the ConfigMap containing library files.
	// +optional
	LibConfigMapRef string `json:"libConfigMapRef,omitempty"`

	// SrcMountPath is the mount path for test files.
	// +optional
	// +kubebuilder:default="/lotest/src"
	SrcMountPath string `json:"srcMountPath,omitempty"`

	// LibMountPath is the mount path for library files.
	// +optional
	// +kubebuilder:default="/opt/locust/lib"
	LibMountPath string `json:"libMountPath,omitempty"`
}

// ============================================
// SCHEDULING CONFIGURATION
// ============================================

// SchedulingConfig defines pod scheduling configuration.
type SchedulingConfig struct {
	// Affinity rules for pod scheduling.
	// +optional
	Affinity *corev1.Affinity `json:"affinity,omitempty"`

	// Tolerations for pod scheduling.
	// +optional
	Tolerations []corev1.Toleration `json:"tolerations,omitempty"`

	// NodeSelector for pod scheduling.
	// +optional
	NodeSelector map[string]string `json:"nodeSelector,omitempty"`
}

// ============================================
// ENVIRONMENT INJECTION (Issue #149)
// ============================================

// EnvConfig defines environment variable injection configuration.
type EnvConfig struct {
	// ConfigMapRefs injects all keys from ConfigMaps as environment variables.
	// +optional
	ConfigMapRefs []ConfigMapEnvSource `json:"configMapRefs,omitempty"`

	// SecretRefs injects all keys from Secrets as environment variables.
	// +optional
	SecretRefs []SecretEnvSource `json:"secretRefs,omitempty"`

	// Variables defines specific environment variables.
	// +optional
	Variables []corev1.EnvVar `json:"variables,omitempty"`

	// SecretMounts mounts secrets as files in the container.
	// +optional
	SecretMounts []SecretMount `json:"secretMounts,omitempty"`
}

// ConfigMapEnvSource defines a ConfigMap environment source.
type ConfigMapEnvSource struct {
	// Name of the ConfigMap.
	// +kubebuilder:validation:Required
	Name string `json:"name"`

	// Prefix to add to all keys when injecting as env vars.
	// +optional
	Prefix string `json:"prefix,omitempty"`
}

// SecretEnvSource defines a Secret environment source.
type SecretEnvSource struct {
	// Name of the Secret.
	// +kubebuilder:validation:Required
	Name string `json:"name"`

	// Prefix to add to all keys when injecting as env vars.
	// +optional
	Prefix string `json:"prefix,omitempty"`
}

// SecretMount defines a secret file mount.
type SecretMount struct {
	// Name of the secret to mount.
	// +kubebuilder:validation:Required
	Name string `json:"name"`

	// MountPath is the path where the secret should be mounted.
	// +kubebuilder:validation:Required
	MountPath string `json:"mountPath"`

	// ReadOnly mounts the secret as read-only.
	// +optional
	// +kubebuilder:default=true
	ReadOnly bool `json:"readOnly,omitempty"`
}

// ============================================
// VOLUME MOUNTING (Issue #252)
// ============================================

// TargetedVolumeMount extends VolumeMount with target pod selection.
type TargetedVolumeMount struct {
	corev1.VolumeMount `json:",inline"`

	// Target specifies which pods receive this mount.
	// +optional
	// +kubebuilder:validation:Enum=master;worker;both
	// +kubebuilder:default=both
	Target string `json:"target,omitempty"`
}

// ============================================
// OBSERVABILITY (Issue #72)
// ============================================

// ObservabilityConfig defines observability settings.
type ObservabilityConfig struct {
	// OpenTelemetry configuration for native Locust OTel integration.
	// +optional
	OpenTelemetry *OpenTelemetryConfig `json:"openTelemetry,omitempty"`
}

// OpenTelemetryConfig defines OpenTelemetry integration settings.
type OpenTelemetryConfig struct {
	// Enabled enables OpenTelemetry integration.
	// When true, adds --otel flag to Locust command.
	// +kubebuilder:default=false
	Enabled bool `json:"enabled"`

	// Endpoint is the OTel collector endpoint (e.g., "otel-collector:4317").
	// Required when Enabled is true.
	// +optional
	Endpoint string `json:"endpoint,omitempty"`

	// Protocol for OTel export.
	// +optional
	// +kubebuilder:validation:Enum=grpc;http/protobuf
	// +kubebuilder:default=grpc
	Protocol string `json:"protocol,omitempty"`

	// Insecure skips TLS verification for the collector connection.
	// +optional
	// +kubebuilder:default=false
	Insecure bool `json:"insecure,omitempty"`

	// ExtraEnvVars for additional OTel SDK configuration.
	// +optional
	ExtraEnvVars map[string]string `json:"extraEnvVars,omitempty"`
}

// ============================================
// STATUS
// ============================================

// LocustTestStatus defines the observed state of LocustTest.
type LocustTestStatus struct {
	// Phase is the current lifecycle phase of the test.
	// +kubebuilder:validation:Enum=Pending;Running;Succeeded;Failed
	// +optional
	Phase Phase `json:"phase,omitempty"`

	// ObservedGeneration is the most recent generation observed by the controller.
	// +optional
	ObservedGeneration int64 `json:"observedGeneration,omitempty"`

	// ExpectedWorkers is the number of workers expected to connect.
	// +optional
	ExpectedWorkers int32 `json:"expectedWorkers,omitempty"`

	// ConnectedWorkers is the approximate number of connected workers,
	// derived from the worker Job's Active pod count (Job.Status.Active).
	// This is an approximation as Kubernetes Job.Status.Active may lag behind
	// actual Locust worker connections.
	// +optional
	ConnectedWorkers int32 `json:"connectedWorkers,omitempty"`

	// StartTime is when the test started.
	// +optional
	StartTime *metav1.Time `json:"startTime,omitempty"`

	// CompletionTime is when the test completed.
	// +optional
	CompletionTime *metav1.Time `json:"completionTime,omitempty"`

	// Conditions represent the latest available observations of the test's state.
	// +optional
	// +patchMergeKey=type
	// +patchStrategy=merge
	// +listType=map
	// +listMapKey=type
	Conditions []metav1.Condition `json:"conditions,omitempty" patchStrategy:"merge" patchMergeKey:"type"`
}

// ============================================
// SPEC
// ============================================

// LocustTestSpec defines the desired state of LocustTest.
type LocustTestSpec struct {
	// Image is the container image for Locust pods.
	// +kubebuilder:validation:Required
	Image string `json:"image"`

	// ImagePullPolicy for the Locust container.
	// +optional
	// +kubebuilder:validation:Enum=Always;IfNotPresent;Never
	// +kubebuilder:default=IfNotPresent
	ImagePullPolicy corev1.PullPolicy `json:"imagePullPolicy,omitempty"`

	// ImagePullSecrets for pulling from private registries.
	// +optional
	ImagePullSecrets []corev1.LocalObjectReference `json:"imagePullSecrets,omitempty"`

	// Master configuration for the master node.
	// +kubebuilder:validation:Required
	Master MasterSpec `json:"master"`

	// Worker configuration for worker nodes.
	// +kubebuilder:validation:Required
	Worker WorkerSpec `json:"worker"`

	// TestFiles configuration for locustfile and library mounting.
	// +optional
	TestFiles *TestFilesConfig `json:"testFiles,omitempty"`

	// Scheduling configuration for pod placement.
	// +optional
	Scheduling *SchedulingConfig `json:"scheduling,omitempty"`

	// Env configuration for environment variable injection.
	// +optional
	Env *EnvConfig `json:"env,omitempty"`

	// Volumes to add to pods.
	// +optional
	Volumes []corev1.Volume `json:"volumes,omitempty"`

	// VolumeMounts for the locust container with target selection.
	// +optional
	VolumeMounts []TargetedVolumeMount `json:"volumeMounts,omitempty"`

	// Observability configuration for metrics and tracing.
	// +optional
	Observability *ObservabilityConfig `json:"observability,omitempty"`
}

// ============================================
// ROOT TYPES
// ============================================

// +kubebuilder:object:root=true
// +kubebuilder:subresource:status
// +kubebuilder:resource:shortName=lotest
// +kubebuilder:storageversion
// +kubebuilder:printcolumn:name="Phase",type=string,JSONPath=`.status.phase`,description="Current test phase"
// +kubebuilder:printcolumn:name="Workers",type=integer,JSONPath=`.spec.worker.replicas`,description="Requested worker count"
// +kubebuilder:printcolumn:name="Connected",type=integer,JSONPath=`.status.connectedWorkers`,description="Connected workers"
// +kubebuilder:printcolumn:name="Image",type=string,JSONPath=`.spec.image`,priority=1
// +kubebuilder:printcolumn:name="Age",type=date,JSONPath=`.metadata.creationTimestamp`

// LocustTest is the Schema for the locusttests API.
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
