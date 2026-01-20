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
	return validateSecretMounts(lt)
}

// ValidateUpdate implements webhook.CustomValidator.
func (v *LocustTestCustomValidator) ValidateUpdate(ctx context.Context, oldObj, newObj runtime.Object) (admission.Warnings, error) {
	lt, ok := newObj.(*LocustTest)
	if !ok {
		return nil, fmt.Errorf("expected LocustTest but got %T", newObj)
	}
	locusttestlog.Info("validate update", "name", lt.Name)
	return validateSecretMounts(lt)
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
func validateSecretMounts(lt *LocustTest) (admission.Warnings, error) {
	if lt.Spec.Env == nil || len(lt.Spec.Env.SecretMounts) == 0 {
		return nil, nil
	}

	reservedPaths := getReservedPaths(lt)

	for _, sm := range lt.Spec.Env.SecretMounts {
		for _, reserved := range reservedPaths {
			if PathConflicts(sm.MountPath, reserved) {
				return nil, fmt.Errorf(
					"secretMount path %q conflicts with reserved path %q; "+
						"operator uses this path for test files",
					sm.MountPath, reserved)
			}
		}
	}

	return nil, nil
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
