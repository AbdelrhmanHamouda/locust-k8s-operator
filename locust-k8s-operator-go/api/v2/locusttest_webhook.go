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
	"context"
	"fmt"
	"strings"

	"k8s.io/apimachinery/pkg/runtime"
	ctrl "sigs.k8s.io/controller-runtime"
	logf "sigs.k8s.io/controller-runtime/pkg/log"
	"sigs.k8s.io/controller-runtime/pkg/webhook"
	"sigs.k8s.io/controller-runtime/pkg/webhook/admission"
)

// log is for logging in this package.
var locusttestlog = logf.Log.WithName("locusttest-resource")

// Default reserved paths that cannot be used for secret mounts
const (
	DefaultSrcMountPath = "/lotest/src"
	DefaultLibMountPath = "/opt/locust/lib"
)

// Reserved volume name constants
const (
	reservedVolumeNamePrefix = "secret-"
	libVolumeName            = "lib"
)

// LocustTestCustomValidator handles validation for LocustTest resources.
type LocustTestCustomValidator struct{}

// SetupWebhookWithManager sets up the webhook with the Manager.
func (r *LocustTest) SetupWebhookWithManager(mgr ctrl.Manager) error {
	return ctrl.NewWebhookManagedBy(mgr).
		For(r).
		WithValidator(&LocustTestCustomValidator{}).
		Complete()
}

// +kubebuilder:webhook:path=/validate-locust-io-v2-locusttest,mutating=false,failurePolicy=fail,sideEffects=None,groups=locust.io,resources=locusttests,verbs=create;update,versions=v2,name=vlocusttest-v2.kb.io,admissionReviewVersions=v1

var _ webhook.CustomValidator = &LocustTestCustomValidator{}

// ValidateCreate implements webhook.CustomValidator.
func (v *LocustTestCustomValidator) ValidateCreate(ctx context.Context, obj runtime.Object) (admission.Warnings, error) {
	lt, ok := obj.(*LocustTest)
	if !ok {
		return nil, fmt.Errorf("expected LocustTest but got %T", obj)
	}
	locusttestlog.Info("validate create", "name", lt.Name)
	return validateLocustTest(lt)
}

// ValidateUpdate implements webhook.CustomValidator.
func (v *LocustTestCustomValidator) ValidateUpdate(ctx context.Context, oldObj, newObj runtime.Object) (admission.Warnings, error) {
	lt, ok := newObj.(*LocustTest)
	if !ok {
		return nil, fmt.Errorf("expected LocustTest but got %T", newObj)
	}
	locusttestlog.Info("validate update", "name", lt.Name)
	return validateLocustTest(lt)
}

// ValidateDelete implements webhook.CustomValidator.
func (v *LocustTestCustomValidator) ValidateDelete(ctx context.Context, obj runtime.Object) (admission.Warnings, error) {
	lt, ok := obj.(*LocustTest)
	if !ok {
		return nil, fmt.Errorf("expected LocustTest but got %T", obj)
	}
	locusttestlog.Info("validate delete", "name", lt.Name)
	return nil, nil
}

// validateSecretMounts checks that secret mount paths don't conflict with reserved paths.
func validateSecretMounts(lt *LocustTest) error {
	if lt.Spec.Env == nil || len(lt.Spec.Env.SecretMounts) == 0 {
		return nil
	}

	reservedPaths := getReservedPaths(lt)

	for _, sm := range lt.Spec.Env.SecretMounts {
		for _, reserved := range reservedPaths {
			if PathConflicts(sm.MountPath, reserved) {
				return fmt.Errorf(
					"secretMount path %q conflicts with reserved path %q; "+
						"operator uses this path for test files",
					sm.MountPath, reserved)
			}
		}
	}

	return nil
}

// getReservedPaths returns the paths that are reserved by the operator.
// It dynamically calculates based on testFiles configuration.
func getReservedPaths(lt *LocustTest) []string {
	var paths []string

	srcPath := DefaultSrcMountPath
	libPath := DefaultLibMountPath

	if lt.Spec.TestFiles != nil {
		if lt.Spec.TestFiles.SrcMountPath != "" {
			srcPath = lt.Spec.TestFiles.SrcMountPath
		}
		if lt.Spec.TestFiles.LibMountPath != "" {
			libPath = lt.Spec.TestFiles.LibMountPath
		}
	}

	// Only add paths that are actually in use
	if lt.Spec.TestFiles != nil && lt.Spec.TestFiles.ConfigMapRef != "" {
		paths = append(paths, srcPath)
	}
	if lt.Spec.TestFiles != nil && lt.Spec.TestFiles.LibConfigMapRef != "" {
		paths = append(paths, libPath)
	}

	// If no testFiles are configured, still protect default paths
	// since users might add them later
	if len(paths) == 0 {
		paths = []string{srcPath, libPath}
	}

	return paths
}

// PathConflicts checks if two paths would conflict.
// Conflict occurs if one path is a prefix of the other.
func PathConflicts(path1, path2 string) bool {
	// Normalize paths by removing trailing slashes
	p1 := strings.TrimSuffix(path1, "/")
	p2 := strings.TrimSuffix(path2, "/")

	// Check if either is a prefix of the other
	return p1 == p2 ||
		strings.HasPrefix(p1, p2+"/") ||
		strings.HasPrefix(p2, p1+"/")
}

// validateLocustTest runs all validation checks.
func validateLocustTest(lt *LocustTest) (admission.Warnings, error) {
	// Validate secret mounts
	if err := validateSecretMounts(lt); err != nil {
		return nil, err
	}

	// Validate user volumes
	if err := validateVolumes(lt); err != nil {
		return nil, err
	}

	// Validate OTel configuration
	if err := validateOTelConfig(lt); err != nil {
		return nil, err
	}

	return nil, nil
}

// validateOTelConfig validates OpenTelemetry configuration.
func validateOTelConfig(lt *LocustTest) error {
	if lt.Spec.Observability == nil {
		return nil
	}

	otelCfg := lt.Spec.Observability.OpenTelemetry
	if otelCfg == nil {
		return nil
	}

	// If OTel is enabled, endpoint is required
	if otelCfg.Enabled && otelCfg.Endpoint == "" {
		return fmt.Errorf("observability.openTelemetry.endpoint is required when OpenTelemetry is enabled")
	}

	return nil
}

// validateVolumes checks for volume name and mount path conflicts.
func validateVolumes(lt *LocustTest) error {
	// Check volume names
	for _, vol := range lt.Spec.Volumes {
		if err := validateVolumeName(lt, vol.Name); err != nil {
			return err
		}
	}

	// Check mount paths
	reservedPaths := getReservedPaths(lt)
	for _, mount := range lt.Spec.VolumeMounts {
		if err := validateMountPath(mount.MountPath, reservedPaths); err != nil {
			return err
		}
	}

	// Validate that all mounts reference defined volumes
	if err := validateMountReferences(lt); err != nil {
		return err
	}

	return nil
}

// validateVolumeName checks if a volume name conflicts with operator-managed names.
func validateVolumeName(lt *LocustTest, name string) error {
	// Check for reserved prefix
	if strings.HasPrefix(name, reservedVolumeNamePrefix) {
		return fmt.Errorf("volume name %q uses reserved prefix %q", name, reservedVolumeNamePrefix)
	}

	// Check for lib volume name
	if name == libVolumeName {
		return fmt.Errorf("volume name %q is reserved by the operator", name)
	}

	// Check for CR-based names
	masterName := lt.Name + "-master"
	workerName := lt.Name + "-worker"
	if name == masterName || name == workerName {
		return fmt.Errorf("volume name %q conflicts with operator-generated name", name)
	}

	return nil
}

// validateMountPath checks if a mount path conflicts with reserved paths.
func validateMountPath(path string, reservedPaths []string) error {
	for _, reserved := range reservedPaths {
		if PathConflicts(path, reserved) {
			return fmt.Errorf("volumeMount path %q conflicts with reserved path %q", path, reserved)
		}
	}
	return nil
}

// validateMountReferences ensures all mounts reference defined volumes.
func validateMountReferences(lt *LocustTest) error {
	volumeNames := make(map[string]bool)
	for _, vol := range lt.Spec.Volumes {
		volumeNames[vol.Name] = true
	}

	for _, mount := range lt.Spec.VolumeMounts {
		if !volumeNames[mount.Name] {
			return fmt.Errorf("volumeMount %q references undefined volume", mount.Name)
		}
	}

	return nil
}
