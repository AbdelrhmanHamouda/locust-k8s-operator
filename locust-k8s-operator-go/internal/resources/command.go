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

	locustv2 "github.com/AbdelrhmanHamouda/locust-k8s-operator/api/v2"
)

// BuildMasterCommand constructs the command arguments for the master node.
// Uses MasterSpec configuration
func BuildMasterCommand(masterSpec *locustv2.MasterSpec, workerReplicas int32, otelEnabled bool) []string {
	var cmdParts []string
	cmdParts = append(cmdParts, masterSpec.Command)

	// Add --otel flag if enabled (must come before other flags)
	if otelEnabled {
		cmdParts = append(cmdParts, "--otel")
	}

	cmdParts = append(cmdParts,
		"--master",
		fmt.Sprintf("--master-port=%d", MasterPort),
		fmt.Sprintf("--expect-workers=%d", workerReplicas),
	)

	// Add --autostart if enabled (default: true)
	if masterSpec.Autostart == nil || *masterSpec.Autostart {
		cmdParts = append(cmdParts, "--autostart")
	}

	// Add --autoquit if enabled (default: enabled with 60s timeout)
	if masterSpec.Autoquit == nil || masterSpec.Autoquit.Enabled {
		timeout := int32(60) // default
		if masterSpec.Autoquit != nil && masterSpec.Autoquit.Timeout > 0 {
			timeout = masterSpec.Autoquit.Timeout
		}
		cmdParts = append(cmdParts, "--autoquit", fmt.Sprintf("%d", timeout))
	}

	cmdParts = append(cmdParts,
		"--enable-rebalancing",
		"--only-summary",
	)

	cmd := strings.Join(cmdParts, " ")
	return strings.Fields(cmd)
}

// BuildWorkerCommand constructs the command arguments for worker nodes.
// Template: "{seed} [--otel] --worker --master-port=5557 --master-host={master-name}"
func BuildWorkerCommand(commandSeed string, masterHost string, otelEnabled bool) []string {
	var cmdParts []string
	cmdParts = append(cmdParts, commandSeed)

	// Add --otel flag if enabled (must come before other flags)
	if otelEnabled {
		cmdParts = append(cmdParts, "--otel")
	}

	cmdParts = append(cmdParts,
		"--worker",
		fmt.Sprintf("--master-port=%d", MasterPort),
		fmt.Sprintf("--master-host=%s", masterHost),
	)

	cmd := strings.Join(cmdParts, " ")
	return strings.Fields(cmd)
}
