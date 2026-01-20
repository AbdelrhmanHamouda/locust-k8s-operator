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
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

func TestPathConflicts_ExactMatch(t *testing.T) {
	assert.True(t, PathConflicts("/lotest/src", "/lotest/src"))
	assert.True(t, PathConflicts("/opt/locust/lib", "/opt/locust/lib"))
}

func TestPathConflicts_Subpath(t *testing.T) {
	// /foo conflicts with /foo/bar because /foo is a prefix
	assert.True(t, PathConflicts("/lotest/src", "/lotest/src/secrets"))
	assert.True(t, PathConflicts("/lotest/src/secrets", "/lotest/src"))

	// Deeper nesting
	assert.True(t, PathConflicts("/opt/locust/lib", "/opt/locust/lib/utils"))
	assert.True(t, PathConflicts("/opt/locust/lib/utils", "/opt/locust/lib"))
}

func TestPathConflicts_NoConflict(t *testing.T) {
	// Completely different paths
	assert.False(t, PathConflicts("/lotest/src", "/etc/certs"))
	assert.False(t, PathConflicts("/opt/locust/lib", "/var/secrets"))

	// Similar prefix but not a subpath
	assert.False(t, PathConflicts("/lotest/src", "/lotest/src2"))
	assert.False(t, PathConflicts("/lotest/src2", "/lotest/src"))
}

func TestPathConflicts_TrailingSlash(t *testing.T) {
	// Trailing slashes should be normalized
	assert.True(t, PathConflicts("/lotest/src/", "/lotest/src"))
	assert.True(t, PathConflicts("/lotest/src", "/lotest/src/"))
	assert.True(t, PathConflicts("/lotest/src/", "/lotest/src/"))
}

func TestValidateSecretMounts_NilEnv(t *testing.T) {
	lt := &LocustTest{
		Spec: LocustTestSpec{
			Env: nil,
		},
	}

	warnings, err := validateSecretMounts(lt)
	assert.NoError(t, err)
	assert.Nil(t, warnings)
}

func TestValidateSecretMounts_EmptyMounts(t *testing.T) {
	lt := &LocustTest{
		Spec: LocustTestSpec{
			Env: &EnvConfig{
				SecretMounts: []SecretMount{},
			},
		},
	}

	warnings, err := validateSecretMounts(lt)
	assert.NoError(t, err)
	assert.Nil(t, warnings)
}

func TestValidateSecretMounts_ValidPath(t *testing.T) {
	lt := &LocustTest{
		Spec: LocustTestSpec{
			Env: &EnvConfig{
				SecretMounts: []SecretMount{
					{Name: "tls-certs", MountPath: "/etc/locust/certs"},
				},
			},
		},
	}

	warnings, err := validateSecretMounts(lt)
	assert.NoError(t, err)
	assert.Nil(t, warnings)
}

func TestValidateSecretMounts_ConflictDefault(t *testing.T) {
	lt := &LocustTest{
		Spec: LocustTestSpec{
			Env: &EnvConfig{
				SecretMounts: []SecretMount{
					{Name: "bad-secret", MountPath: "/lotest/src"},
				},
			},
		},
	}

	warnings, err := validateSecretMounts(lt)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "conflicts with reserved path")
	assert.Contains(t, err.Error(), "/lotest/src")
	assert.Nil(t, warnings)
}

func TestValidateSecretMounts_ConflictLib(t *testing.T) {
	lt := &LocustTest{
		Spec: LocustTestSpec{
			Env: &EnvConfig{
				SecretMounts: []SecretMount{
					{Name: "bad-secret", MountPath: "/opt/locust/lib"},
				},
			},
		},
	}

	_, err := validateSecretMounts(lt)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "conflicts with reserved path")
	assert.Contains(t, err.Error(), "/opt/locust/lib")
}

func TestValidateSecretMounts_ConflictSubpath(t *testing.T) {
	lt := &LocustTest{
		Spec: LocustTestSpec{
			Env: &EnvConfig{
				SecretMounts: []SecretMount{
					{Name: "bad-secret", MountPath: "/lotest/src/secrets"},
				},
			},
		},
	}

	_, err := validateSecretMounts(lt)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "conflicts with reserved path")
}

func TestValidateSecretMounts_CustomTestFilesPath(t *testing.T) {
	lt := &LocustTest{
		Spec: LocustTestSpec{
			TestFiles: &TestFilesConfig{
				ConfigMapRef: "my-scripts",
				SrcMountPath: "/custom/src",
			},
			Env: &EnvConfig{
				SecretMounts: []SecretMount{
					{Name: "secret", MountPath: "/custom/src/secrets"},
				},
			},
		},
	}

	_, err := validateSecretMounts(lt)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "/custom/src")
}

func TestValidateSecretMounts_CustomPathAllowsDefault(t *testing.T) {
	// When using custom paths, the default paths should be allowed
	lt := &LocustTest{
		Spec: LocustTestSpec{
			TestFiles: &TestFilesConfig{
				ConfigMapRef: "my-scripts",
				SrcMountPath: "/custom/src",
			},
			Env: &EnvConfig{
				SecretMounts: []SecretMount{
					// This would conflict with default but we're using custom
					{Name: "secret", MountPath: "/lotest/src"},
				},
			},
		},
	}

	warnings, err := validateSecretMounts(lt)
	// Should pass because we're using custom path, not default
	assert.NoError(t, err)
	assert.Nil(t, warnings)
}

func TestGetReservedPaths_NoTestFiles(t *testing.T) {
	lt := &LocustTest{
		Spec: LocustTestSpec{},
	}

	paths := getReservedPaths(lt)
	assert.Contains(t, paths, DefaultSrcMountPath)
	assert.Contains(t, paths, DefaultLibMountPath)
}

func TestGetReservedPaths_WithConfigMapRef(t *testing.T) {
	lt := &LocustTest{
		Spec: LocustTestSpec{
			TestFiles: &TestFilesConfig{
				ConfigMapRef: "my-scripts",
			},
		},
	}

	paths := getReservedPaths(lt)
	assert.Contains(t, paths, DefaultSrcMountPath)
	assert.Len(t, paths, 1)
}

func TestGetReservedPaths_WithBothConfigMaps(t *testing.T) {
	lt := &LocustTest{
		Spec: LocustTestSpec{
			TestFiles: &TestFilesConfig{
				ConfigMapRef:    "my-scripts",
				LibConfigMapRef: "my-lib",
			},
		},
	}

	paths := getReservedPaths(lt)
	assert.Contains(t, paths, DefaultSrcMountPath)
	assert.Contains(t, paths, DefaultLibMountPath)
	assert.Len(t, paths, 2)
}

func TestGetReservedPaths_CustomPaths(t *testing.T) {
	lt := &LocustTest{
		Spec: LocustTestSpec{
			TestFiles: &TestFilesConfig{
				ConfigMapRef:    "my-scripts",
				LibConfigMapRef: "my-lib",
				SrcMountPath:    "/custom/src",
				LibMountPath:    "/custom/lib",
			},
		},
	}

	paths := getReservedPaths(lt)
	assert.Contains(t, paths, "/custom/src")
	assert.Contains(t, paths, "/custom/lib")
	assert.NotContains(t, paths, DefaultSrcMountPath)
	assert.NotContains(t, paths, DefaultLibMountPath)
}

func TestValidateCreate(t *testing.T) {
	validator := &LocustTestCustomValidator{}
	lt := &LocustTest{
		ObjectMeta: metav1.ObjectMeta{
			Name:      "test",
			Namespace: "default",
		},
		Spec: LocustTestSpec{
			Image: "locustio/locust:2.20.0",
			Master: MasterSpec{
				Command: "locust -f /lotest/src/locustfile.py",
			},
			Worker: WorkerSpec{
				Command:  "locust -f /lotest/src/locustfile.py",
				Replicas: 1,
			},
			Env: &EnvConfig{
				SecretMounts: []SecretMount{
					{Name: "valid-secret", MountPath: "/etc/certs"},
				},
			},
		},
	}

	warnings, err := validator.ValidateCreate(context.Background(), lt)
	require.NoError(t, err)
	assert.Nil(t, warnings)
}

func TestValidateCreate_Invalid(t *testing.T) {
	validator := &LocustTestCustomValidator{}
	lt := &LocustTest{
		ObjectMeta: metav1.ObjectMeta{
			Name:      "test",
			Namespace: "default",
		},
		Spec: LocustTestSpec{
			Image: "locustio/locust:2.20.0",
			Master: MasterSpec{
				Command: "locust -f /lotest/src/locustfile.py",
			},
			Worker: WorkerSpec{
				Command:  "locust -f /lotest/src/locustfile.py",
				Replicas: 1,
			},
			Env: &EnvConfig{
				SecretMounts: []SecretMount{
					{Name: "bad-secret", MountPath: "/lotest/src"},
				},
			},
		},
	}

	warnings, err := validator.ValidateCreate(context.Background(), lt)
	require.Error(t, err)
	assert.Contains(t, err.Error(), "conflicts with reserved path")
	assert.Nil(t, warnings)
}

func TestValidateUpdate(t *testing.T) {
	validator := &LocustTestCustomValidator{}
	oldLt := &LocustTest{
		ObjectMeta: metav1.ObjectMeta{
			Name:      "test",
			Namespace: "default",
		},
		Spec: LocustTestSpec{
			Image: "locustio/locust:2.20.0",
			Master: MasterSpec{
				Command: "locust -f /lotest/src/locustfile.py",
			},
			Worker: WorkerSpec{
				Command:  "locust -f /lotest/src/locustfile.py",
				Replicas: 1,
			},
		},
	}

	newLt := oldLt.DeepCopy()
	newLt.Spec.Env = &EnvConfig{
		SecretMounts: []SecretMount{
			{Name: "valid-secret", MountPath: "/etc/certs"},
		},
	}

	warnings, err := validator.ValidateUpdate(context.Background(), oldLt, newLt)
	require.NoError(t, err)
	assert.Nil(t, warnings)
}

func TestValidateDelete(t *testing.T) {
	validator := &LocustTestCustomValidator{}
	lt := &LocustTest{
		ObjectMeta: metav1.ObjectMeta{
			Name:      "test",
			Namespace: "default",
		},
	}

	warnings, err := validator.ValidateDelete(context.Background(), lt)
	require.NoError(t, err)
	assert.Nil(t, warnings)
}

func TestValidateCreate_WrongType(t *testing.T) {
	validator := &LocustTestCustomValidator{}

	// Pass wrong type
	warnings, err := validator.ValidateCreate(context.Background(), &LocustTestList{})
	require.Error(t, err)
	assert.Contains(t, err.Error(), "expected LocustTest")
	assert.Nil(t, warnings)
}
