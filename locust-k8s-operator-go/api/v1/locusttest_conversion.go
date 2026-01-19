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
	corev1 "k8s.io/api/core/v1"
	"k8s.io/utils/ptr"
	"sigs.k8s.io/controller-runtime/pkg/conversion"

	v2 "github.com/AbdelrhmanHamouda/locust-k8s-operator/api/v2"
)

// ConvertTo converts this v1 LocustTest to the Hub version (v2).
func (src *LocustTest) ConvertTo(dstRaw conversion.Hub) error {
	dst := dstRaw.(*v2.LocustTest)

	// Metadata
	dst.ObjectMeta = src.ObjectMeta

	// Image configuration
	dst.Spec.Image = src.Spec.Image
	if src.Spec.ImagePullPolicy != "" {
		dst.Spec.ImagePullPolicy = corev1.PullPolicy(src.Spec.ImagePullPolicy)
	}
	dst.Spec.ImagePullSecrets = convertImagePullSecretsToV2(src.Spec.ImagePullSecrets)

	// Master configuration (grouped)
	dst.Spec.Master = v2.MasterSpec{
		Command:   src.Spec.MasterCommandSeed,
		Autostart: ptr.To(true),
		Autoquit:  &v2.AutoquitConfig{Enabled: true, Timeout: 60},
	}
	if src.Spec.Labels != nil && src.Spec.Labels.Master != nil {
		dst.Spec.Master.Labels = src.Spec.Labels.Master
	}
	if src.Spec.Annotations != nil && src.Spec.Annotations.Master != nil {
		dst.Spec.Master.Annotations = src.Spec.Annotations.Master
	}

	// Worker configuration (grouped)
	dst.Spec.Worker = v2.WorkerSpec{
		Command:  src.Spec.WorkerCommandSeed,
		Replicas: src.Spec.WorkerReplicas,
	}
	if src.Spec.Labels != nil && src.Spec.Labels.Worker != nil {
		dst.Spec.Worker.Labels = src.Spec.Labels.Worker
	}
	if src.Spec.Annotations != nil && src.Spec.Annotations.Worker != nil {
		dst.Spec.Worker.Annotations = src.Spec.Annotations.Worker
	}

	// Test files configuration
	if src.Spec.ConfigMap != "" || src.Spec.LibConfigMap != "" {
		dst.Spec.TestFiles = &v2.TestFilesConfig{
			ConfigMapRef:    src.Spec.ConfigMap,
			LibConfigMapRef: src.Spec.LibConfigMap,
		}
	}

	// Scheduling configuration
	if src.Spec.Affinity != nil || len(src.Spec.Tolerations) > 0 {
		dst.Spec.Scheduling = &v2.SchedulingConfig{}
		if src.Spec.Affinity != nil {
			dst.Spec.Scheduling.Affinity = convertAffinityToV2(src.Spec.Affinity)
		}
		if len(src.Spec.Tolerations) > 0 {
			dst.Spec.Scheduling.Tolerations = convertTolerationsToV2(src.Spec.Tolerations)
		}
	}

	return nil
}

// ConvertFrom converts the Hub version (v2) to this v1 LocustTest.
// Note: This is a lossy conversion - v2-only fields are not preserved.
func (dst *LocustTest) ConvertFrom(srcRaw conversion.Hub) error {
	src := srcRaw.(*v2.LocustTest)

	// Metadata
	dst.ObjectMeta = src.ObjectMeta

	// Image configuration
	dst.Spec.Image = src.Spec.Image
	dst.Spec.ImagePullPolicy = string(src.Spec.ImagePullPolicy)
	dst.Spec.ImagePullSecrets = convertImagePullSecretsToV1(src.Spec.ImagePullSecrets)

	// Master configuration → flat fields
	dst.Spec.MasterCommandSeed = src.Spec.Master.Command

	// Worker configuration → flat fields
	dst.Spec.WorkerCommandSeed = src.Spec.Worker.Command
	dst.Spec.WorkerReplicas = src.Spec.Worker.Replicas

	// Labels from grouped structure
	if len(src.Spec.Master.Labels) > 0 || len(src.Spec.Worker.Labels) > 0 {
		dst.Spec.Labels = &PodLabels{}
		if len(src.Spec.Master.Labels) > 0 {
			dst.Spec.Labels.Master = src.Spec.Master.Labels
		}
		if len(src.Spec.Worker.Labels) > 0 {
			dst.Spec.Labels.Worker = src.Spec.Worker.Labels
		}
	}

	// Annotations from grouped structure
	if len(src.Spec.Master.Annotations) > 0 || len(src.Spec.Worker.Annotations) > 0 {
		dst.Spec.Annotations = &PodAnnotations{}
		if len(src.Spec.Master.Annotations) > 0 {
			dst.Spec.Annotations.Master = src.Spec.Master.Annotations
		}
		if len(src.Spec.Worker.Annotations) > 0 {
			dst.Spec.Annotations.Worker = src.Spec.Worker.Annotations
		}
	}

	// Test files configuration → flat fields
	if src.Spec.TestFiles != nil {
		dst.Spec.ConfigMap = src.Spec.TestFiles.ConfigMapRef
		dst.Spec.LibConfigMap = src.Spec.TestFiles.LibConfigMapRef
	}

	// Scheduling → flat fields
	if src.Spec.Scheduling != nil {
		if src.Spec.Scheduling.Affinity != nil {
			dst.Spec.Affinity = convertAffinityToV1(src.Spec.Scheduling.Affinity)
		}
		if len(src.Spec.Scheduling.Tolerations) > 0 {
			dst.Spec.Tolerations = convertTolerationsToV1(src.Spec.Scheduling.Tolerations)
		}
		// Note: nodeSelector is lost (v2-only field)
	}

	// The following v2-only fields are NOT preserved in v1:
	// - master.resources, master.extraArgs
	// - worker.resources, worker.extraArgs
	// - testFiles.srcMountPath, testFiles.libMountPath
	// - scheduling.nodeSelector
	// - env (configMapRefs, secretRefs, variables, secretMounts)
	// - volumes, volumeMounts
	// - observability (OpenTelemetry config)
	// - status (v1 has no status subresource fields)

	return nil
}

// =============================================================================
// Helper Functions
// =============================================================================

func convertImagePullSecretsToV2(secrets []string) []corev1.LocalObjectReference {
	if len(secrets) == 0 {
		return nil
	}
	result := make([]corev1.LocalObjectReference, len(secrets))
	for i, s := range secrets {
		result[i] = corev1.LocalObjectReference{Name: s}
	}
	return result
}

func convertImagePullSecretsToV1(secrets []corev1.LocalObjectReference) []string {
	if len(secrets) == 0 {
		return nil
	}
	result := make([]string, len(secrets))
	for i, s := range secrets {
		result[i] = s.Name
	}
	return result
}

func convertAffinityToV2(src *LocustTestAffinity) *corev1.Affinity {
	if src == nil || src.NodeAffinity == nil {
		return nil
	}

	nodeReqs := src.NodeAffinity.RequiredDuringSchedulingIgnoredDuringExecution
	if len(nodeReqs) == 0 {
		return nil
	}

	terms := make([]corev1.NodeSelectorRequirement, 0, len(nodeReqs))
	for key, value := range nodeReqs {
		terms = append(terms, corev1.NodeSelectorRequirement{
			Key:      key,
			Operator: corev1.NodeSelectorOpIn,
			Values:   []string{value},
		})
	}

	return &corev1.Affinity{
		NodeAffinity: &corev1.NodeAffinity{
			RequiredDuringSchedulingIgnoredDuringExecution: &corev1.NodeSelector{
				NodeSelectorTerms: []corev1.NodeSelectorTerm{
					{MatchExpressions: terms},
				},
			},
		},
	}
}

func convertAffinityToV1(src *corev1.Affinity) *LocustTestAffinity {
	if src == nil || src.NodeAffinity == nil {
		return nil
	}

	required := src.NodeAffinity.RequiredDuringSchedulingIgnoredDuringExecution
	if required == nil || len(required.NodeSelectorTerms) == 0 {
		return nil
	}

	// Extract first term's match expressions into v1 format
	// This is a lossy conversion - complex affinity rules may lose data
	nodeReqs := make(map[string]string)
	for _, term := range required.NodeSelectorTerms {
		for _, expr := range term.MatchExpressions {
			if expr.Operator == corev1.NodeSelectorOpIn && len(expr.Values) > 0 {
				nodeReqs[expr.Key] = expr.Values[0]
			}
		}
	}

	if len(nodeReqs) == 0 {
		return nil
	}

	return &LocustTestAffinity{
		NodeAffinity: &LocustTestNodeAffinity{
			RequiredDuringSchedulingIgnoredDuringExecution: nodeReqs,
		},
	}
}

func convertTolerationsToV2(src []LocustTestToleration) []corev1.Toleration {
	if len(src) == 0 {
		return nil
	}
	result := make([]corev1.Toleration, len(src))
	for i, t := range src {
		result[i] = corev1.Toleration{
			Key:      t.Key,
			Operator: corev1.TolerationOperator(t.Operator),
			Value:    t.Value,
			Effect:   corev1.TaintEffect(t.Effect),
		}
	}
	return result
}

func convertTolerationsToV1(src []corev1.Toleration) []LocustTestToleration {
	if len(src) == 0 {
		return nil
	}
	result := make([]LocustTestToleration, len(src))
	for i, t := range src {
		result[i] = LocustTestToleration{
			Key:      t.Key,
			Operator: string(t.Operator),
			Value:    t.Value,
			Effect:   string(t.Effect),
		}
	}
	return result
}
