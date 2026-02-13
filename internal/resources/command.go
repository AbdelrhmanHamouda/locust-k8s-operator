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
	"github.com/go-logr/logr"
)

// operatorManagedFlags is the registry of flags managed by the operator.
// Users should not override these in extraArgs, but if they do, their value takes precedence.
var operatorManagedFlags = map[string]bool{
	"--master":             true,
	"--worker":             true,
	"--master-port":        true,
	"--master-host":        true,
	"--expect-workers":     true,
	"--autostart":          true,
	"--autoquit":           true,
	"--otel":               true,
	"--enable-rebalancing": true,
	"--only-summary":       true,
}

// detectFlagConflicts checks if extraArgs contain operator-managed flags.
// Returns a slice of conflicting arguments.
func detectFlagConflicts(extraArgs []string) []string {
	var conflicts []string
	for _, arg := range extraArgs {
		// Check if arg matches a known operator-managed flag
		// Handle both "--flag=value" and "--flag value" forms
		for flag := range operatorManagedFlags {
			if arg == flag || strings.HasPrefix(arg, flag+"=") {
				conflicts = append(conflicts, arg)
				break
			}
		}
	}
	return conflicts
}

// BuildMasterCommand constructs the command arguments for the master node.
// Uses MasterSpec configuration and appends extraArgs after operator-managed flags.
func BuildMasterCommand(masterSpec *locustv2.MasterSpec, workerReplicas int32, otelEnabled bool, logger logr.Logger) []string {
	var cmdParts []string
	// Split command seed into individual args at append time
	cmdParts = append(cmdParts, strings.Fields(masterSpec.Command)...)

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
		if masterSpec.Autoquit != nil && masterSpec.Autoquit.Timeout >= 0 {
			timeout = masterSpec.Autoquit.Timeout
		}
		cmdParts = append(cmdParts, "--autoquit", fmt.Sprintf("%d", timeout))
	}

	cmdParts = append(cmdParts,
		"--enable-rebalancing",
		"--only-summary",
	)

	// Append extraArgs after operator-managed flags (user flags take precedence via POSIX last-occurrence-wins)
	if len(masterSpec.ExtraArgs) > 0 {
		conflicts := detectFlagConflicts(masterSpec.ExtraArgs)
		if len(conflicts) > 0 {
			logger.Info("User-provided extraArgs override operator-managed flags",
				"conflicts", conflicts,
				"behavior", "user value takes precedence")
		}
		cmdParts = append(cmdParts, masterSpec.ExtraArgs...)
	}

	return cmdParts
}

// BuildWorkerCommand constructs the command arguments for worker nodes.
// Template: "{seed} [--otel] --worker --master-port=5557 --master-host={master-name} [extraArgs...]"
func BuildWorkerCommand(commandSeed string, masterHost string, otelEnabled bool, extraArgs []string, logger logr.Logger) []string {
	var cmdParts []string
	// Split command seed into individual args at append time
	cmdParts = append(cmdParts, strings.Fields(commandSeed)...)

	// Add --otel flag if enabled (must come before other flags)
	if otelEnabled {
		cmdParts = append(cmdParts, "--otel")
	}

	cmdParts = append(cmdParts,
		"--worker",
		fmt.Sprintf("--master-port=%d", MasterPort),
		fmt.Sprintf("--master-host=%s", masterHost),
	)

	// Append extraArgs after operator-managed flags (user flags take precedence via POSIX last-occurrence-wins)
	if len(extraArgs) > 0 {
		conflicts := detectFlagConflicts(extraArgs)
		if len(conflicts) > 0 {
			logger.Info("User-provided extraArgs override operator-managed flags",
				"mode", "worker",
				"conflicts", conflicts,
				"behavior", "user value takes precedence")
		}
		cmdParts = append(cmdParts, extraArgs...)
	}

	return cmdParts
}
