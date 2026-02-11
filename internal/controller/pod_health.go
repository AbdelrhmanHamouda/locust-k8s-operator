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

package controller

import (
	"context"
	"fmt"
	"regexp"
	"sort"
	"strings"
	"time"

	corev1 "k8s.io/api/core/v1"
	"sigs.k8s.io/controller-runtime/pkg/client"
	logf "sigs.k8s.io/controller-runtime/pkg/log"

	locustv2 "github.com/AbdelrhmanHamouda/locust-k8s-operator/api/v2"
)

const (
	// podStartupGracePeriod is the time to wait before reporting pod failures.
	// This prevents false positives during normal startup (scheduling, image pull, volume mount).
	podStartupGracePeriod = 2 * time.Minute
)

// PodHealthStatus represents the aggregated health status of all pods for a LocustTest.
type PodHealthStatus struct {
	Healthy       bool
	Reason        string
	Message       string
	FailedPods    []PodFailureInfo
	InGracePeriod bool
}

// PodFailureInfo contains details about a failed pod.
type PodFailureInfo struct {
	Name         string
	FailureType  string
	ErrorMessage string
}

// checkPodHealth analyzes all pods owned by the LocustTest and returns their health status.
// Returns PodHealthStatus and optional requeue duration (for grace period handling).
func (r *LocustTestReconciler) checkPodHealth(ctx context.Context, lt *locustv2.LocustTest) (PodHealthStatus, time.Duration) {
	log := logf.FromContext(ctx)

	// List all pods owned by this LocustTest
	podList := &corev1.PodList{}
	listOpts := []client.ListOption{
		client.InNamespace(lt.Namespace),
		client.MatchingLabels{
			"performance-test-name": lt.Name,
		},
	}

	if err := r.List(ctx, podList, listOpts...); err != nil {
		log.Error(err, "Failed to list pods for health check")
		// Return healthy status to avoid blocking on transient errors
		return PodHealthStatus{
			Healthy: true,
			Reason:  locustv2.ReasonPodsHealthy,
			Message: "Pod health check pending",
		}, 0
	}

	// No pods yet - this is normal during initial creation
	if len(podList.Items) == 0 {
		return PodHealthStatus{
			Healthy: true,
			Reason:  locustv2.ReasonPodsStarting,
			Message: "Waiting for pods to be created",
		}, 0
	}

	// Check if we're still in the grace period
	oldestPodCreation := findOldestPodCreationTime(podList.Items)
	gracePeriodRemaining := podStartupGracePeriod - time.Since(oldestPodCreation)

	if gracePeriodRemaining > 0 {
		log.V(1).Info("Pods in startup grace period", "remaining", gracePeriodRemaining)
		return PodHealthStatus{
			Healthy:       true,
			Reason:        locustv2.ReasonPodsStarting,
			Message:       "Pods are starting up",
			InGracePeriod: true,
		}, gracePeriodRemaining
	}

	// Analyze each pod for failures
	var failedPods []PodFailureInfo
	for _, pod := range podList.Items {
		if failure := analyzePodFailure(&pod, lt); failure != nil {
			failedPods = append(failedPods, *failure)
		}
	}

	// If no failures, all pods are healthy
	if len(failedPods) == 0 {
		return PodHealthStatus{
			Healthy: true,
			Reason:  locustv2.ReasonPodsHealthy,
			Message: "All pods are healthy",
		}, 0
	}

	// Categorize and prioritize failures
	failureType, message := buildFailureMessage(failedPods)

	return PodHealthStatus{
		Healthy:    false,
		Reason:     failureType,
		Message:    message,
		FailedPods: failedPods,
	}, 0
}

// analyzePodFailure examines a single pod and returns failure info if the pod is unhealthy.
// Returns nil if the pod is healthy.
func analyzePodFailure(pod *corev1.Pod, lt *locustv2.LocustTest) *PodFailureInfo {
	// Check pod conditions for scheduling failures
	for _, condition := range pod.Status.Conditions {
		if condition.Type == corev1.PodScheduled && condition.Status == corev1.ConditionFalse {
			return &PodFailureInfo{
				Name:         pod.Name,
				FailureType:  locustv2.ReasonPodSchedulingError,
				ErrorMessage: condition.Message,
			}
		}
	}

	// Check init containers
	for _, initStatus := range pod.Status.InitContainerStatuses {
		if failure := analyzeContainerStatus(pod.Name, initStatus, true, lt); failure != nil {
			return failure
		}
	}

	// Check main containers
	for _, containerStatus := range pod.Status.ContainerStatuses {
		if failure := analyzeContainerStatus(pod.Name, containerStatus, false, lt); failure != nil {
			return failure
		}
	}

	// Pod is healthy
	return nil
}

// analyzeContainerStatus checks a container status for failures.
func analyzeContainerStatus(podName string, status corev1.ContainerStatus, isInitContainer bool, lt *locustv2.LocustTest) *PodFailureInfo {
	// Check waiting state
	if status.State.Waiting != nil {
		waiting := status.State.Waiting
		reason := waiting.Reason
		message := waiting.Message

		switch {
		case reason == "CreateContainerConfigError":
			// Extract ConfigMap name if this is a config error
			enhancedMsg := extractConfigMapError(message, lt)
			return &PodFailureInfo{
				Name:         podName,
				FailureType:  locustv2.ReasonPodConfigError,
				ErrorMessage: enhancedMsg,
			}

		case reason == "ImagePullBackOff" || reason == "ErrImagePull" || strings.Contains(reason, "ImagePull"):
			return &PodFailureInfo{
				Name:         podName,
				FailureType:  locustv2.ReasonPodImagePullError,
				ErrorMessage: message,
			}

		case reason == "CrashLoopBackOff":
			return &PodFailureInfo{
				Name:         podName,
				FailureType:  locustv2.ReasonPodCrashLoop,
				ErrorMessage: message,
			}
		}
	}

	// Check terminated state (for init containers or recently failed containers)
	if status.State.Terminated != nil {
		terminated := status.State.Terminated
		if terminated.ExitCode != 0 {
			failureType := locustv2.ReasonPodInitError
			if !isInitContainer {
				failureType = locustv2.ReasonPodCrashLoop
			}
			return &PodFailureInfo{
				Name:         podName,
				FailureType:  failureType,
				ErrorMessage: fmt.Sprintf("Container %s exited with code %d: %s", status.Name, terminated.ExitCode, terminated.Reason),
			}
		}
	}

	return nil
}

// extractConfigMapError enhances ConfigMap error messages with the expected ConfigMap name from spec.
func extractConfigMapError(errorMsg string, lt *locustv2.LocustTest) string {
	// Try to extract ConfigMap name from error message
	// Common patterns:
	// - "configmap \"name\" not found"
	// - "couldn't find key ... in ConfigMap ..."
	configMapRegex := regexp.MustCompile(`[Cc]onfig[Mm]ap\s+"([^"]+)"`)
	matches := configMapRegex.FindStringSubmatch(errorMsg)

	var expectedConfigMap string
	if lt.Spec.TestFiles != nil {
		if lt.Spec.TestFiles.ConfigMapRef != "" {
			expectedConfigMap = lt.Spec.TestFiles.ConfigMapRef
		} else if lt.Spec.TestFiles.LibConfigMapRef != "" {
			expectedConfigMap = lt.Spec.TestFiles.LibConfigMapRef
		}
	}

	if len(matches) > 1 {
		// ConfigMap name found in error
		foundName := matches[1]
		if expectedConfigMap != "" && foundName == expectedConfigMap {
			return fmt.Sprintf("ConfigMap not found (expected: %s). %s", expectedConfigMap, errorMsg)
		}
		return errorMsg
	}

	// Generic config error - add expected ConfigMap if known
	if expectedConfigMap != "" {
		return fmt.Sprintf("ConfigMap not found (expected: %s). %s", expectedConfigMap, errorMsg)
	}

	return errorMsg
}

// buildFailureMessage creates a user-friendly message from pod failures.
// Returns failure type (reason) and formatted message.
func buildFailureMessage(failures []PodFailureInfo) (string, string) {
	if len(failures) == 0 {
		return locustv2.ReasonPodsHealthy, "All pods are healthy"
	}

	// Group failures by type
	failuresByType := make(map[string][]PodFailureInfo)
	for _, f := range failures {
		failuresByType[f.FailureType] = append(failuresByType[f.FailureType], f)
	}

	// Prioritize failure types (most critical first)
	priorityOrder := []string{
		locustv2.ReasonPodConfigError,
		locustv2.ReasonPodImagePullError,
		locustv2.ReasonPodSchedulingError,
		locustv2.ReasonPodCrashLoop,
		locustv2.ReasonPodInitError,
	}

	var primaryType string
	var primaryFailures []PodFailureInfo

	for _, failureType := range priorityOrder {
		if pods, exists := failuresByType[failureType]; exists {
			primaryType = failureType
			primaryFailures = pods
			break
		}
	}

	// Build message for primary failure type
	podNames := make([]string, len(primaryFailures))
	for i, f := range primaryFailures {
		podNames[i] = f.Name
	}

	// Get first error message as example
	exampleError := primaryFailures[0].ErrorMessage

	message := fmt.Sprintf("%s: %d pod(s) affected [%s]: %s",
		primaryType,
		len(primaryFailures),
		strings.Join(podNames, ", "),
		exampleError,
	)

	// Add recovery hint for config errors
	if primaryType == locustv2.ReasonPodConfigError {
		message += ". Create the ConfigMap and the pods will restart automatically."
	}

	return primaryType, message
}

// findOldestPodCreationTime returns the creation time of the oldest pod in the list.
func findOldestPodCreationTime(pods []corev1.Pod) time.Time {
	if len(pods) == 0 {
		return time.Now()
	}

	oldest := pods[0].CreationTimestamp.Time
	for _, pod := range pods[1:] {
		if pod.CreationTimestamp.Time.Before(oldest) {
			oldest = pod.CreationTimestamp.Time
		}
	}

	return oldest
}

// categorizePodFailures groups failures by type and returns them sorted by priority.
func categorizePodFailures(failures []PodFailureInfo) map[string][]PodFailureInfo {
	result := make(map[string][]PodFailureInfo)
	for _, failure := range failures {
		result[failure.FailureType] = append(result[failure.FailureType], failure)
	}

	// Sort each group by pod name for consistent output
	for _, group := range result {
		sort.Slice(group, func(i, j int) bool {
			return group[i].Name < group[j].Name
		})
	}

	return result
}
