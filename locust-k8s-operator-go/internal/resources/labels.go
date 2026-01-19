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
	"fmt"
	"strings"

	locustv1 "github.com/AbdelrhmanHamouda/locust-k8s-operator/api/v1"
	"github.com/AbdelrhmanHamouda/locust-k8s-operator/internal/config"
)

// NodeName constructs the node name from the CR name and operational mode.
// Format: "{cr-name}-{mode}" with dots replaced by dashes.
// Example: "team-a.load-test" -> "team-a-load-test-master"
func NodeName(crName string, mode OperationalMode) string {
	name := fmt.Sprintf("%s-%s", crName, mode.String())
	return strings.ReplaceAll(name, ".", "-")
}

// BuildLabels constructs the labels for a pod based on the LocustTest CR and mode.
// Includes required labels and merges user-defined labels from the CR spec.
func BuildLabels(lt *locustv1.LocustTest, mode OperationalMode) map[string]string {
	nodeName := NodeName(lt.Name, mode)

	labels := map[string]string{
		LabelApp:       lt.Name,
		LabelPodName:   nodeName,
		LabelManagedBy: ManagedByValue,
		LabelTestName:  lt.Name,
	}

	// Merge user-defined labels
	for k, v := range getUserLabels(lt, mode) {
		labels[k] = v
	}

	return labels
}

// getUserLabels extracts user-defined labels from the CR spec for the given mode.
func getUserLabels(lt *locustv1.LocustTest, mode OperationalMode) map[string]string {
	if lt.Spec.Labels == nil {
		return nil
	}

	switch mode {
	case Master:
		return lt.Spec.Labels.Master
	case Worker:
		return lt.Spec.Labels.Worker
	default:
		return nil
	}
}

// BuildAnnotations constructs the annotations for a pod based on the LocustTest CR and mode.
// Master pods include Prometheus scrape annotations; worker pods do not.
// Merges user-defined annotations from the CR spec.
func BuildAnnotations(lt *locustv1.LocustTest, mode OperationalMode, cfg *config.OperatorConfig) map[string]string {
	annotations := make(map[string]string)

	// Master pods get Prometheus annotations
	if mode == Master {
		annotations[AnnotationPrometheusScrape] = "true"
		annotations[AnnotationPrometheusPath] = MetricsEndpointPath
		annotations[AnnotationPrometheusPort] = fmt.Sprintf("%d", cfg.MetricsExporterPort)
	}

	// Merge user-defined annotations
	for k, v := range getUserAnnotations(lt, mode) {
		annotations[k] = v
	}

	return annotations
}

// getUserAnnotations extracts user-defined annotations from the CR spec for the given mode.
func getUserAnnotations(lt *locustv1.LocustTest, mode OperationalMode) map[string]string {
	if lt.Spec.Annotations == nil {
		return nil
	}

	switch mode {
	case Master:
		return lt.Spec.Annotations.Master
	case Worker:
		return lt.Spec.Annotations.Worker
	default:
		return nil
	}
}
