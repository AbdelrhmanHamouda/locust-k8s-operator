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
	"testing"

	"github.com/stretchr/testify/assert"
)

const testCommandSeed = "locust -f /lotest/src/test.py"
const testMasterHost = "my-test-master"

func TestBuildMasterCommand(t *testing.T) {
	workerReplicas := int32(5)

	cmd := BuildMasterCommand(testCommandSeed, workerReplicas, false)

	// Verify all expected flags are present
	assert.Contains(t, cmd, "locust")
	assert.Contains(t, cmd, "-f")
	assert.Contains(t, cmd, "/lotest/src/test.py")
	assert.Contains(t, cmd, "--master")
	assert.Contains(t, cmd, "--master-port=5557")
	assert.Contains(t, cmd, "--expect-workers=5")
	assert.Contains(t, cmd, "--autostart")
	assert.Contains(t, cmd, "--autoquit")
	assert.Contains(t, cmd, "60")
	assert.Contains(t, cmd, "--enable-rebalancing")
	assert.Contains(t, cmd, "--only-summary")
}

func TestBuildMasterCommand_SplitsCorrectly(t *testing.T) {
	seedWithExtraSpaces := "locust   -f   /lotest/src/test.py"
	workerReplicas := int32(3)

	cmd := BuildMasterCommand(seedWithExtraSpaces, workerReplicas, false)

	// strings.Fields handles multiple spaces correctly
	assert.Equal(t, "locust", cmd[0])
	assert.Equal(t, "-f", cmd[1])
	assert.Equal(t, "/lotest/src/test.py", cmd[2])
}

func TestBuildWorkerCommand(t *testing.T) {
	masterHost := testMasterHost

	cmd := BuildWorkerCommand(testCommandSeed, masterHost, false)

	// Verify all expected flags are present
	assert.Contains(t, cmd, "locust")
	assert.Contains(t, cmd, "-f")
	assert.Contains(t, cmd, "/lotest/src/test.py")
	assert.Contains(t, cmd, "--worker")
	assert.Contains(t, cmd, "--master-port=5557")
	assert.Contains(t, cmd, "--master-host="+testMasterHost)
}

func TestBuildWorkerCommand_MasterHostCorrect(t *testing.T) {
	masterHost := "team-a-load-test-master"

	cmd := BuildWorkerCommand(testCommandSeed, masterHost, false)

	// Find the master-host flag
	found := false
	for _, arg := range cmd {
		if arg == "--master-host=team-a-load-test-master" {
			found = true
			break
		}
	}
	assert.True(t, found, "master-host flag should contain the correct master host")
}

// ===== OTel Flag Tests =====

func TestBuildMasterCommand_OTelDisabled(t *testing.T) {
	workerReplicas := int32(3)

	cmd := BuildMasterCommand(testCommandSeed, workerReplicas, false)

	// --otel flag should NOT be present
	assert.NotContains(t, cmd, "--otel")
}

func TestBuildMasterCommand_OTelEnabled(t *testing.T) {
	workerReplicas := int32(3)

	cmd := BuildMasterCommand(testCommandSeed, workerReplicas, true)

	// --otel flag should be present
	assert.Contains(t, cmd, "--otel")
}

func TestBuildMasterCommand_OTelFlagPosition(t *testing.T) {
	workerReplicas := int32(3)

	cmd := BuildMasterCommand(testCommandSeed, workerReplicas, true)

	// Find positions of --otel and --master
	otelIndex := -1
	masterIndex := -1
	for i, arg := range cmd {
		if arg == "--otel" {
			otelIndex = i
		}
		if arg == "--master" {
			masterIndex = i
		}
	}

	// --otel should appear before --master
	assert.NotEqual(t, -1, otelIndex, "--otel flag should be present")
	assert.NotEqual(t, -1, masterIndex, "--master flag should be present")
	assert.Less(t, otelIndex, masterIndex, "--otel should appear before --master")
}

func TestBuildWorkerCommand_OTelDisabled(t *testing.T) {
	masterHost := testMasterHost

	cmd := BuildWorkerCommand(testCommandSeed, masterHost, false)

	// --otel flag should NOT be present
	assert.NotContains(t, cmd, "--otel")
}

func TestBuildWorkerCommand_OTelEnabled(t *testing.T) {
	masterHost := testMasterHost

	cmd := BuildWorkerCommand(testCommandSeed, masterHost, true)

	// --otel flag should be present
	assert.Contains(t, cmd, "--otel")
}

func TestBuildWorkerCommand_OTelFlagPosition(t *testing.T) {
	masterHost := testMasterHost

	cmd := BuildWorkerCommand(testCommandSeed, masterHost, true)

	// Find positions of --otel and --worker
	otelIndex := -1
	workerIndex := -1
	for i, arg := range cmd {
		if arg == "--otel" {
			otelIndex = i
		}
		if arg == "--worker" {
			workerIndex = i
		}
	}

	// --otel should appear before --worker
	assert.NotEqual(t, -1, otelIndex, "--otel flag should be present")
	assert.NotEqual(t, -1, workerIndex, "--worker flag should be present")
	assert.Less(t, otelIndex, workerIndex, "--otel should appear before --worker")
}
