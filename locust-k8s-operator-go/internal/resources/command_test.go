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

	locustv2 "github.com/AbdelrhmanHamouda/locust-k8s-operator/api/v2"
	"github.com/stretchr/testify/assert"
	"k8s.io/utils/ptr"
)

const testCommandSeed = "locust -f /lotest/src/test.py"
const testMasterHost = "my-test-master"

// Helper to create a default MasterSpec for testing
func testMasterSpec() *locustv2.MasterSpec {
	return &locustv2.MasterSpec{
		Command:   testCommandSeed,
		Autostart: ptr.To(true),
		Autoquit:  &locustv2.AutoquitConfig{Enabled: true, Timeout: 60},
	}
}

func TestBuildMasterCommand(t *testing.T) {
	workerReplicas := int32(5)
	masterSpec := testMasterSpec()

	cmd := BuildMasterCommand(masterSpec, workerReplicas, false)

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
	masterSpec := &locustv2.MasterSpec{
		Command:   "locust   -f   /lotest/src/test.py",
		Autostart: ptr.To(true),
	}
	workerReplicas := int32(3)

	cmd := BuildMasterCommand(masterSpec, workerReplicas, false)

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
	masterSpec := testMasterSpec()

	cmd := BuildMasterCommand(masterSpec, workerReplicas, false)

	// --otel flag should NOT be present
	assert.NotContains(t, cmd, "--otel")
}

func TestBuildMasterCommand_OTelEnabled(t *testing.T) {
	workerReplicas := int32(3)
	masterSpec := testMasterSpec()

	cmd := BuildMasterCommand(masterSpec, workerReplicas, true)

	// --otel flag should be present
	assert.Contains(t, cmd, "--otel")
}

func TestBuildMasterCommand_OTelFlagPosition(t *testing.T) {
	workerReplicas := int32(3)
	masterSpec := testMasterSpec()

	cmd := BuildMasterCommand(masterSpec, workerReplicas, true)

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

// ===== Autostart/Autoquit Tests =====

func TestBuildMasterCommand_AutostartDisabled(t *testing.T) {
	masterSpec := &locustv2.MasterSpec{
		Command:   testCommandSeed,
		Autostart: ptr.To(false),
	}

	cmd := BuildMasterCommand(masterSpec, 3, false)

	assert.NotContains(t, cmd, "--autostart")
}

func TestBuildMasterCommand_AutostartDefault(t *testing.T) {
	// When Autostart is nil, default to true
	masterSpec := &locustv2.MasterSpec{
		Command: testCommandSeed,
	}

	cmd := BuildMasterCommand(masterSpec, 3, false)

	assert.Contains(t, cmd, "--autostart")
}

func TestBuildMasterCommand_AutoquitDisabled(t *testing.T) {
	masterSpec := &locustv2.MasterSpec{
		Command:   testCommandSeed,
		Autostart: ptr.To(true),
		Autoquit:  &locustv2.AutoquitConfig{Enabled: false},
	}

	cmd := BuildMasterCommand(masterSpec, 3, false)

	assert.NotContains(t, cmd, "--autoquit")
}

func TestBuildMasterCommand_AutoquitCustomTimeout(t *testing.T) {
	masterSpec := &locustv2.MasterSpec{
		Command:   testCommandSeed,
		Autostart: ptr.To(true),
		Autoquit:  &locustv2.AutoquitConfig{Enabled: true, Timeout: 120},
	}

	cmd := BuildMasterCommand(masterSpec, 3, false)

	assert.Contains(t, cmd, "--autoquit")
	assert.Contains(t, cmd, "120")
	assert.NotContains(t, cmd, "60")
}

func TestBuildMasterCommand_AutoquitDefault(t *testing.T) {
	// When Autoquit is nil, default to enabled with 60s timeout
	masterSpec := &locustv2.MasterSpec{
		Command: testCommandSeed,
	}

	cmd := BuildMasterCommand(masterSpec, 3, false)

	assert.Contains(t, cmd, "--autoquit")
	assert.Contains(t, cmd, "60")
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
