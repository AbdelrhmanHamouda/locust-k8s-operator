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
)

// BuildMasterCommand constructs the command arguments for the master node.
// Template: "{seed} [--otel] --master --master-port=5557 --expect-workers={N} --autostart --autoquit 60 --enable-rebalancing --only-summary"
func BuildMasterCommand(commandSeed string, workerReplicas int32, otelEnabled bool) []string {
	var cmdParts []string
	cmdParts = append(cmdParts, commandSeed)

	// Add --otel flag if enabled (must come before other flags)
	if otelEnabled {
		cmdParts = append(cmdParts, "--otel")
	}

	cmdParts = append(cmdParts,
		"--master",
		fmt.Sprintf("--master-port=%d", MasterPort),
		fmt.Sprintf("--expect-workers=%d", workerReplicas),
		"--autostart",
		"--autoquit", "60",
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
