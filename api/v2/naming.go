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
	"fmt"
	"strings"
)

// Node mode suffixes used when generating resource names from a CR name.
const (
	// NodeModeMaster is the suffix for master-side generated resources.
	NodeModeMaster = "master"
	// NodeModeWorker is the suffix for worker-side generated resources.
	NodeModeWorker = "worker"
)

// SanitizeResourceName converts a CR name into a form usable as a generated
// Kubernetes resource name. CR names may contain dots (they are valid in
// object names), but the operator's generated names — Services, Jobs, pod
// labels, volume names — must not, so dots are replaced with dashes.
func SanitizeResourceName(name string) string {
	return strings.ReplaceAll(name, ".", "-")
}

// GeneratedNodeName returns the resource name the operator generates for a CR
// name and node mode, e.g. ("team-a.load-test", NodeModeMaster) ->
// "team-a-load-test-master".
//
// This is the single source of truth for generated resource names. Anything
// that creates, looks up, or validates against those names must use it —
// computing the name by hand caused a recovery loop for dotted CR names
// (see the fix for issue #50).
func GeneratedNodeName(crName, mode string) string {
	return SanitizeResourceName(fmt.Sprintf("%s-%s", crName, mode))
}
