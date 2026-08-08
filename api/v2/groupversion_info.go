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

// Package v2 contains API Schema definitions for the locust v2 API group.
// +kubebuilder:object:generate=true
// +groupName=locust.io
package v2

import (
	"k8s.io/apimachinery/pkg/runtime/schema"
	"sigs.k8s.io/controller-runtime/pkg/scheme"
)

var (
	// GroupVersion is group version used to register these objects.
	GroupVersion = schema.GroupVersion{Group: "locust.io", Version: "v2"}

	// SchemeBuilder is used to add go types to the GroupVersionKind scheme.
	//
	// controller-runtime v0.24 deprecated scheme.Builder so that api packages
	// need not import controller-runtime at all. Replacing it means hand-rolling
	// runtime.NewSchemeBuilder here and rewriting the SchemeBuilder.Register
	// call in locusttest_types.go — both kubebuilder-scaffolded — which is more
	// than a dependency bump should carry.
	//nolint:staticcheck // SA1019: deliberate, see above
	SchemeBuilder = &scheme.Builder{GroupVersion: GroupVersion}

	// AddToScheme adds the types in this group-version to the given scheme.
	AddToScheme = SchemeBuilder.AddToScheme
)
